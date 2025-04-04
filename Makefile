# Image URL to use all building/pushing image targets
IMG         ?= keikoproj/iam-manager:latest

# Tools required to run the full suite of tests properly
OSNAME           ?= $(shell uname -s | tr A-Z a-z)
KUBEBUILDER_ARCH ?= amd64
ENVTEST_K8S_VERSION = 1.28.0

LOCALBIN ?= $(shell pwd)/bin
# Export local bin to path for all recipes
export PATH := $(LOCALBIN):$(PATH)

## Tool Binaries
MOCKGEN ?= $(LOCALBIN)/mockgen
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

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

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
	GOBIN := $(shell go env GOPATH)/bin
else
	GOBIN := $(shell go env GOBIN)
endif

all: manager

# Build manager binary
manager: $(LOCALBIN)/manager
$(LOCALBIN)/manager: generate fmt mock vet update
	go build -o $(LOCALBIN)/manager cmd/main.go

mock: $(MOCKGEN)
	@echo "mockgen is in progess"
	@for pkg in $(shell go list ./...) ; do \
		go generate ./... ;\
	done

# Run tests
test: mock generate fmt manifests envtest
	KUBECONFIG=$(KUBECONFIG) \
	LOCAL=$(LOCAL) \
	ALLOWED_POLICY_ACTION=$(ALLOWED_POLICY_ACTION) \
	RESTRICTED_POLICY_RESOURCES=$(RESTRICTED_POLICY_RESOURCES) \
	RESTRICTED_S3_RESOURCES=$(RESTRICTED_S3_RESOURCES) \
	AWS_ACCOUNT_ID=$(AWS_ACCOUNT_ID) \
	AWS_REGION=$(AWS_REGION) \
	MANAGED_POLICIES=$(MANAGED_POLICIES) \
	MANAGED_PERMISSION_BOUNDARY_POLICY=$(MANAGED_PERMISSION_BOUNDARY_POLICY) \
	CLUSTER_NAME=$(CLUSTER_NAME) \
	CLUSTER_OIDC_ISSUER_URL="$(CLUSTER_OIDC_ISSUER_URL)" \
	DEFAULT_TRUST_POLICY=$(DEFAULT_TRUST_POLICY) \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

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

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd_no_webhook/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet: mock
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


## Tool Versions
MOCKGEN_VERSION ?= v1.6.0
KUSTOMIZE_VERSION ?= v4.2.0
CONTROLLER_TOOLS_VERSION ?= v0.17.0

$(MOCKGEN): $(LOCALBIN) ## Download mockgen if necessary.
	GOBIN=$(LOCALBIN) go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	rm -f $(KUSTOMIZE) || true
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(LOCALBIN):
	mkdir -p $(LOCALBIN)
