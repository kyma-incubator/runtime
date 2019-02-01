/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package function

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	runtimev1alpha1 "github.com/kyma-incubator/runtime/pkg/apis/runtime/v1alpha1"
	"github.com/pborman/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimeUtil "github.com/kyma-incubator/runtime/pkg/utils"
)

var log = logf.Log.WithName("controller")

// Add creates a new Function Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFunction{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("function-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Function
	err = c.Watch(&source.Kind{Type: &runtimev1alpha1.Function{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Function - change this for objects you create
	// err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &runtimev1alpha1.Function{},
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

var _ reconcile.Reconciler = &ReconcileFunction{}

// ReconcileFunction reconciles a Function object
type ReconcileFunction struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=runtime.kyma-project.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=runtime.kyma-project.io,resources=functions/status,verbs=get;update;patch
func (r *ReconcileFunction) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	fnConfigName := "fn-config"
	fnConfigNamespace := "default"
	fnConfigNameEnv := os.Getenv("CONTROLLER_CONFIGMAP")
	fnConfigNamespaceEnv := os.Getenv("CONTROLLER_CONFIGMAP_NS")
	if len(fnConfigNameEnv) > 0 {
		fnConfigName = fnConfigNameEnv
	}
	if len(fnConfigNamespaceEnv) > 0 {
		fnConfigNamespace = fnConfigNamespaceEnv
	}
	fnConfig := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: fnConfigName, Namespace: fnConfigNamespace}, fnConfig)

	if err != nil {
		log.Info("Unable to read Function controller config: %v from Namespace: %v", fnConfigName, fnConfigNamespace)
		return reconcile.Result{}, err
	}

	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		fmt.Printf("Error in reading ConfigMap: %v", err)
		return reconcile.Result{}, err
	}
	// Fetch the Function instance
	fn := &runtimev1alpha1.Function{}
	err = r.Get(context.TODO(), request.NamespacedName, fn)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	data := make(map[string]string)
	data["handler"] = "handler.main"
	data["handler.js"] = fn.Spec.Function
	if len(strings.Trim(fn.Spec.Deps, " ")) == 0 {
		data["package.json"] = "{}"
	} else {
		data["package.json"] = fn.Spec.Deps
	}

	// Managing a ConfigMap
	deployCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    fn.Labels,
			Namespace: fn.Namespace,
			Name:      fn.Name,
		},
		Data: data,
	}
	if err := controllerutil.SetControllerReference(fn, deployCm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundCm := &corev1.ConfigMap{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deployCm.Name, Namespace: deployCm.Namespace}, foundCm)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		err = r.Create(context.TODO(), deployCm)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(deployCm.Data, foundCm.Data) {
		foundCm.Data = deployCm.Data
		log.Info("Updating ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		err = r.Update(context.TODO(), foundCm)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	// Managing a resource of type Service.serving.knative.dev

	dockerRegistry := rnInfo.RegistryInfo
	randomStr := uuid.NewRandom().String()[:8]
	imageName := fmt.Sprintf("%s/%s-%s:%s", dockerRegistry, fn.Namespace, fn.Name, randomStr)
	deployService := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    fn.Labels,
			Namespace: fn.Namespace,
			Name:      fn.Name,
		},
		Spec: runtimeUtil.GetServiceSpec(imageName, *fn, rnInfo),
	}

	if err := controllerutil.SetControllerReference(fn, deployService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundService := &servingv1alpha1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deployService.Name, Namespace: deployService.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Service", "namespace", deployService.Namespace, "name", deployService.Name)
		err = r.Create(context.TODO(), deployService)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		fmt.Printf("Error while creating: %v", err)
		return reconcile.Result{}, err
	}
	if !reflect.DeepEqual(deployService.Spec, deployService.Spec) {
		foundService.Spec = deployService.Spec
		log.Info("Updating Service", "namespace", deployService.Namespace, "name", deployService.Name)
		err = r.Update(context.TODO(), foundService)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil

}
