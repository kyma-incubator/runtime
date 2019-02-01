package utils_test

import (
	"testing"

	"github.com/kyma-incubator/runtime/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

func TestNewRuntimeInfo(t *testing.T) {
	cm := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
		},
	}
	ri, err := utils.New(cm)
	if err != nil {
		t.Fatalf("Error creating a new runtime object: %v", err)
	}
	if ri.ServiceAccount != "test" {
		t.Fatalf("Expected: %s Got: %s", "test", ri.ServiceAccount)
	}

	if ri.RegistryInfo != "foo" {
		t.Fatalf("Expected: %s Got: %s", "foo", ri.RegistryInfo)
	}
}

func TestDockerFileConfigMapName(t *testing.T) {
	runtime := "nodejs8"
	cm := &corev1.ConfigMap{
		Data: map[string]string{
			"serviceAccountName": "test",
			"dockerRegistry":     "foo",
			"runtimes": `[
				{
					"ID": "nodejs8",
					"DockerFileName": "dockerfile-nodejs8",
				}
			]`,
		},
	}
	ri, err := utils.New(cm)
	if err != nil {
		t.Fatalf("Error creating runtime info obj: %v", err)
	}
	dockerFileCMName := ri.DockerFileConfigMapName(runtime)
	if dockerFileCMName != "dockerfile-nodejs8" {
		t.Fatalf("Expected: %s Got: %s", "dockerfile-nodejs8", dockerFileCMName)
	}
}
