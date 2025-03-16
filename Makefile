# Image URL to use all building/pushing image targets
IMG ?= keikoproj/iam-manager:latest

# Directory Structure
HACK_DIR ?= $(shell pwd)/hack
TOOLS_DIR ?= $(HACK_DIR)/tools
TOOLS_BIN_DIR ?= $(TOOLS_DIR)/bin
LOCALBIN ?= $(shell pwd)/bin

# Create necessary directories
$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# OSNAME and architecture detection for cross-platform support
OSNAME ?= $(shell uname -s | tr A-Z a-z)
ARCH ?= $(shell go env GOARCH)
KUBEBUILDER_ARCH ?= $(ARCH)
ENVTEST_K8S_VERSION = 1.28.0

# KIND configuration
KIND_VERSION ?= v0.20.0
KIND ?= $(TOOLS_BIN_DIR)/kind
KIND_CLUSTER_NAME ?= iam-manager-test
KIND_KUBECONFIG ?= $(TOOLS_DIR)/kubeconfig

# Define all tools
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOTESTSUM ?= $(LOCALBIN)/gotestsum
ENVTEST ?= $(LOCALBIN)/setup-envtest

# IAM Manager environment variables
KUBECONFIG                  ?= $(KIND_KUBECONFIG)
LOCAL                       ?= true
ALLOWED_POLICY_ACTION       ?= s3:,sts:,ec2:Describe,acm:Describe,acm:List,acm:Get,route53:Get,route53:List,route53:Create,route53:Delete,route53:Change,kms:Decrypt,kms:Encrypt,kms:ReEncrypt,kms:GenerateDataKey,kms:DescribeKey,dynamodb:,secretsmanager:GetSecretValue,es:,sqs:SendMessage,sqs:ReceiveMessage,sqs:DeleteMessage,SNS:Publish,sqs:GetQueueAttributes,sqs:GetQueueUrl
RESTRICTED_POLICY_RESOURCES ?= policy-resource
RESTRICTED_S3_RESOURCES     ?= s3-resource
AWS_ACCOUNT_ID              ?= 123456789012
AWS_REGION                  ?= us-west-2
MANAGED_POLICIES            ?= arn:aws:iam::123456789012:policy/SOMETHING
MANAGED_PERMISSION_BOUNDARY_POLICY ?= arn:aws:iam::1123456789012:role/iam-manager-permission-boundary
CLUSTER_NAME                ?= k8s_test_keiko
CLUSTER_OIDC_ISSUER_URL     ?= https://google.com/OIDC
DEFAULT_TRUST_POLICY        ?= '{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:SERVICE_ACCOUNT_NAME"}}}, {"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}'

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# https://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	@echo "Generating manifests with architecture $(ARCH)"
	$(CONTROLLER_GEN) rbac:roleName=iam-manager crd webhook paths="./..." paths="./api/..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=iam-manager crd webhook paths="./..." paths="./api/..." output:crd:artifacts:config=config/crd_no_webhook/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	@echo "Generating code with architecture $(ARCH)"
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	@echo "Running go vet on core packages (skipping problematic test packages)..."
	go vet ./cmd/... ./api/... ./internal/controllers/... ./internal/config/... ./internal/utils/... ./pkg/k8s/... ./pkg/logging/...

.PHONY: test
test: manifests generate fmt vet kind-setup envtest gotestsum ## Run tests.
	KUBECONFIG=$(KIND_KUBECONFIG) \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	$(GOTESTSUM) -- -v ./... -coverprofile cover.out

.PHONY: lint
lint: ## Run golangci-lint on the code.
	$(LOCALBIN)/golangci-lint run ./...

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

##@ Build

.PHONY: all
all: build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: manager-dev
manager-dev: generate fmt vet ## Build manager binary without kustomize (faster for development).
	@echo "Building manager binary for development (skipping kustomize)..."
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	KUBECONFIG=$(KUBECONFIG) \
	LOCAL=$(LOCAL) \
	ALLOWED_POLICY_ACTION="$(ALLOWED_POLICY_ACTION)" \
	RESTRICTED_POLICY_RESOURCES="$(RESTRICTED_POLICY_RESOURCES)" \
	RESTRICTED_S3_RESOURCES="$(RESTRICTED_S3_RESOURCES)" \
	AWS_ACCOUNT_ID="$(AWS_ACCOUNT_ID)" \
	AWS_REGION="$(AWS_REGION)" \
	MANAGED_POLICIES="$(MANAGED_POLICIES)" \
	MANAGED_PERMISSION_BOUNDARY_POLICY="$(MANAGED_PERMISSION_BOUNDARY_POLICY)" \
	CLUSTER_NAME="$(CLUSTER_NAME)" \
	CLUSTER_OIDC_ISSUER_URL="$(CLUSTER_OIDC_ISSUER_URL)" \
	DEFAULT_TRUST_POLICY="$(DEFAULT_TRUST_POLICY)" \
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

##@ Deployment

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=true -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=true -f -

.PHONY: update
update: manifests kustomize ## Update yamls with current version.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default_no_webhook > hack/iam-manager.yaml
	$(KUSTOMIZE) build config/default > hack/iam-manager_with_webhook.yaml

##@ Build Dependencies

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
KIND ?= $(TOOLS_BIN_DIR)/kind
GOTESTSUM ?= $(LOCALBIN)/gotestsum

## Tool Versions
KUSTOMIZE_VERSION ?= v5.1.1
CONTROLLER_TOOLS_VERSION ?= v0.14.0
GOLANGCI_LINT_VERSION ?= v1.55.2

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { echo "Installing kustomize version $(KUSTOMIZE_VERSION) to $(LOCALBIN)/kustomize"; \
	curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: kind-install
kind-install: $(KIND) ## Download KIND locally if necessary.
$(KIND): $(TOOLS_BIN_DIR)
	@echo "Installing KIND $(KIND_VERSION)..."
	@if [ ! -f $(KIND) ]; then \
		ARCH=$$(go env GOARCH); \
		OS=$$(go env GOOS); \
		curl -L -o $(KIND) "https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$$OS-$$ARCH"; \
		chmod +x $(KIND); \
		echo "KIND installed at $(KIND)"; \
	else \
		echo "KIND already installed at $(KIND)"; \
	fi

.PHONY: kind-setup
kind-setup: kind-install ## Create a KIND cluster if it doesn't exist.
	@$(KIND) get clusters | grep -q $(KIND_CLUSTER_NAME) || \
	( \
		echo "Creating KIND cluster '$(KIND_CLUSTER_NAME)'..."; \
		$(KIND) create cluster --name $(KIND_CLUSTER_NAME) --kubeconfig=$(KIND_KUBECONFIG) --wait 5m && \
		echo "KIND cluster '$(KIND_CLUSTER_NAME)' created" \
	)
	@echo "Using KIND cluster '$(KIND_CLUSTER_NAME)'"
	@cp $(KIND_KUBECONFIG) $(KIND_KUBECONFIG).bak || true
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl config use-context kind-$(KIND_CLUSTER_NAME)
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl cluster-info
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl get nodes -o wide

.PHONY: kind-delete
kind-delete: kind-install ## Delete the KIND cluster.
	@echo "Deleting KIND cluster '$(KIND_CLUSTER_NAME)'..."
	@$(KIND) delete cluster --name $(KIND_CLUSTER_NAME) || true
	@echo "KIND cluster '$(KIND_CLUSTER_NAME)' deleted"

.PHONY: kind-deploy
kind-deploy: manifests kind-setup ## Deploy the controller to the KIND cluster.
	@echo "Creating required namespaces..."
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace dev --dry-run=client -o yaml | kubectl --kubeconfig=$(KIND_KUBECONFIG) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace iam-manager-system --dry-run=client -o yaml | kubectl --kubeconfig=$(KIND_KUBECONFIG) apply -f -
	@echo "Deploying CRDs to KIND cluster..."
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -f config/crd/bases

.PHONY: kind-clean
kind-clean: kind-delete ## Clean up KIND resources and binaries.
	@rm -rf $(TOOLS_DIR)/kubeconfig || true
	@echo "KIND resources cleaned up"

.PHONY: gotestsum
gotestsum: $(GOTESTSUM) ## Download gotestsum locally if necessary.
$(GOTESTSUM): $(LOCALBIN)
	test -s $(LOCALBIN)/gotestsum || GOBIN=$(LOCALBIN) go install gotest.tools/gotestsum@latest

##@ KIND targets
.PHONY: kind
kind: kind-install kind-setup kind-deploy ## Run KIND targets.
