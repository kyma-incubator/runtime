
# Image URL to use all building/pushing image targets
IMG ?= runtime-controller:latest

all: test manager

# Run tests
test: generate fmt vet
	go test -v ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/kyma-incubator/function-controller/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install:
	kubectl apply -f config/crds/runtime_v1alpha1_function.yaml

# CreateResource creates a resource in the cluster
create-resource:
	kubectl apply -f config/samples

# DeleteResource creates a resource in the cluster
delete-resource:
	kubectl delete -f config/samples

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
# deploy: manifests
# 	kubectl apply -f config/crds
# 	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
# manifests:
	# go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
# docker-build: test
docker-build:
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}
