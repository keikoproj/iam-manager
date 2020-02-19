
# Image URL to use all building/pushing image targets
IMG ?= keikoproj/iam-manager:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

setup: ; $(info $(M) setting up env variables for test…) @ ## Setup env variables
export LOCAL=true
export ALLOWED_POLICY_ACTION=allowed-action
export RESTRICTED_POLICY_RESOURCES=policy-resource
export RESTRICTED_S3_RESOURCES=s3-resource
export AWS_ACCOUNT_ID=123456789012
export AWS_REGION=us-west-2
export AWS_MASTER_ROLE=
export MANAGED_POLICIES=arn:aws:iam::123456789012:policy/SOMETHING
export MANAGED_PERMISSION_BOUNDARY_POLICY=arn:aws:iam::1123456789012:role/iam-manager-permission-boundary

mock:
	go get -u github.com/golang/mock/mockgen
	@echo "mockgen is in progess"
	@for pkg in $(shell go list ./...) ; do \
		go generate ./... ;\
	done

# Run tests
test: setup mock generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet update
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# updates the full config yaml file
update: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > hack/iam-manager.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.1
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
