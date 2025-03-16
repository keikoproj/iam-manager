# Image URL to use all building/pushing image targets
IMG         ?= keikoproj/iam-manager:latest

# Tools required to run the full suite of tests properly
OSNAME           ?= $(shell uname -s | tr A-Z a-z)
ARCH             ?= $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/' | sed 's/arm64/arm64/')
KUBEBUILDER_ARCH ?= $(ARCH)
ENVTEST_K8S_VERSION = 1.28.0

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Tool versions
CONTROLLER_GEN_VERSION ?= v0.14.0
ENVTEST_VERSION ?= latest
KUSTOMIZE_VERSION ?= v5.1.1

# Produce CRDs that work across Kubernetes versions
CRD_OPTIONS ?= "crd:crdVersions=v1"

KUBECONFIG                  ?= $(HOME)/.kube/config
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

ENVTEST ?= $(LOCALBIN)/setup-envtest
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
KUSTOMIZE ?= $(LOCALBIN)/kustomize

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
	GOBIN := $(shell go env GOPATH)/bin
else
	GOBIN := $(shell go env GOBIN)
endif

all: manager

# Generate mocks with mockgen
.PHONY: mock
mock:
	@echo "Generating mocks with architecture $(ARCH)"
	@echo "Installing mockgen..."
	@go install github.com/golang/mock/mockgen@v1.6.0
	@echo "Generating awsapi mocks..."
	
	@if [ ! -d pkg/awsapi/mocks ]; then mkdir -p pkg/awsapi/mocks; fi
	
	# Use source mode for more reliable mock generation (especially on ARM64)
	@echo "Generating IAM API mock..."
	@mockgen -source=pkg/awsapi/iam.go -destination=pkg/awsapi/mocks/mock_iamapi.go -package=mock_awsapi
	
	@echo "Generating STS API mock..."
	@mockgen -source=pkg/awsapi/sts.go -destination=pkg/awsapi/mocks/mock_stsapi.go -package=mock_awsapi
	
	@echo "Generating EKS API mock..."
	@mockgen -source=pkg/awsapi/eks.go -destination=pkg/awsapi/mocks/mock_eksapi.go -package=mock_awsapi
	
	@echo "Mock generation complete"

# Run tests
.PHONY: test
test: manifests generate fmt vet envtest ## Run tests
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path --arch=$(ARCH))" \
	GO_TEST_MODE=true \
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run CI tests
.PHONY: ci-test
ci-test: manifests generate fmt vet ## Run CI tests
	GO_TEST_MODE=true \
	go test -coverprofile=coverage.out -covermode=atomic ./...

# Build manager binary
manager: generate fmt vet update
	go build -o bin/manager cmd/main.go

# Build manager binary without CRD generation or kustomize (for faster builds)
.PHONY: manager-dev
manager-dev: generate fmt vet
	@echo "Building manager binary for development (skipping kustomize)..."
	go build -o bin/manager cmd/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd_no_webhook | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default_no_webhook | kubectl apply -f -

# Install CRDs into a cluster
install_with_webhook: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy_with_webhook: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# updates the full config yaml file
update: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default_no_webhook > hack/iam-manager.yaml
	$(KUSTOMIZE) build config/default > hack/iam-manager_with_webhook.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	@echo "Running go vet on core packages (skipping problematic test packages)..."
	go vet ./cmd/... ./api/... ./internal/controllers/... ./internal/config/... ./internal/utils/... ./pkg/k8s/... ./pkg/logging/...

# Generate code
.PHONY: generate
generate: controller-gen ## Generate code
	@echo "Generating code with architecture $(ARCH)"
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build:
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

docker-buildx:
	@echo "Building for multi-platforms"
	@docker buildx build --platform linux/amd64,linux/arm64 -t ${IMG} .

docker-buildx-push:
	@echo "Building for multi-platforms (and pushing)"
	@docker buildx build --platform linux/amd64,linux/arm64 --push -t ${IMG} .

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	@echo "Installing kustomize version $(KUSTOMIZE_VERSION) to $(LOCALBIN)/kustomize"
	@if [ ! -f $(KUSTOMIZE) ] || ! $(KUSTOMIZE) version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "Installing new kustomize version $(KUSTOMIZE_VERSION)"; \
		rm -f $(KUSTOMIZE); \
		GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION); \
	else \
		echo "Kustomize already exists with correct version"; \
	fi

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen kustomize ## Generate manifests e.g. CRD, RBAC etc.
	@echo "Generating manifests with architecture $(ARCH)"
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=iam-manager webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=iam-manager webhook paths="./..." output:crd:artifacts:config=config/crd_no_webhook/bases
