package utils

import (
	"fmt"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	runtimev1alpha1 "github.com/kyma-incubator/runtime/pkg/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GetServiceSpec gets ServiceSpec for a function
func GetServiceSpec(imageName string, fn runtimev1alpha1.Function, rnInfo *RuntimeInfo) servingv1alpha1.ServiceSpec {
	defaultMode := int32(420)
	buildContainer := getBuildContainer(imageName, fn, rnInfo)
	volumes := []corev1.Volume{
		{
			Name: "dockerfile-vol",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rnInfo.DockerFileConfigMapName(fn.Spec.Runtime),
					},
				},
			},
		},
		{
			Name: "func-vol",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fn.Name,
					},
				},
			},
		},
	}

	// TODO: Make it constant for nodejs8/nodejs6
	envVarsForRevision := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}

	return servingv1alpha1.ServiceSpec{
		RunLatest: &servingv1alpha1.RunLatestType{
			Configuration: servingv1alpha1.ConfigurationSpec{
				Build: &servingv1alpha1.RawExtension{
					BuildSpec: &buildv1alpha1.BuildSpec{
						ServiceAccountName: rnInfo.ServiceAccount,
						Steps: []corev1.Container{
							*buildContainer,
						},
						Volumes: volumes,
					},
				},
				RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
					Spec: servingv1alpha1.RevisionSpec{
						Container: corev1.Container{
							Image: imageName,
							Env:   envVarsForRevision,
						},
					},
				},
			},
		},
	}
}

func getBuildContainer(imageName string, fn runtimev1alpha1.Function, riUtil *RuntimeInfo) *corev1.Container {
	destination := fmt.Sprintf("--destination=%s", imageName)
	buildContainer := corev1.Container{
		Name:  "build-and-push",
		Image: "gcr.io/kaniko-project/executor",
		Args:  []string{"--dockerfile=/workspace/Dockerfile", destination},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "dockerfile-vol", //TODO: make it configurable
				MountPath: "/workspace",
			},
			{
				Name:      "func-vol",
				MountPath: "/src",
			},
		},
	}

	return &buildContainer
}
