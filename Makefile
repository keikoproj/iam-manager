# Image URL to use all building/pushing image targets
IMG         ?= keikoproj/iam-manager:latest

# Tools required to run the full suite of tests properly
OSNAME           ?= $(shell uname -s | tr A-Z a-z)
KUBEBUILDER_VER  ?= 2.2.0
KUBEBUILDER_ARCH ?= amd64

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

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

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
	GOBIN := $(shell go env GOPATH)/bin
else
	GOBIN := $(shell go env GOBIN)
endif

all: manager

.PHONY: kubebuilder
kubebuilder:
	@echo "Downloading and installing Kubebuilder - this requires sudo privileges"
	curl -fsSL -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VER)/kubebuilder_$(KUBEBUILDER_VER)_$(OSNAME)_$(KUBEBUILDER_ARCH).tar.gz"
	rm -rf kubebuilder && mkdir -p kubebuilder
	tar -zxvf kubebuilder_$(KUBEBUILDER_VER)_$(OSNAME)_$(KUBEBUILDER_ARCH).tar.gz --strip-components 1 -C kubebuilder
	sudo cp -rf kubebuilder /usr/local

mock:
	go get -u github.com/golang/mock/mockgen
	@echo "mockgen is in progess"
	@for pkg in $(shell go list ./...) ; do \
		go generate ./... ;\
	done

# Run tests
test: mock generate fmt manifests
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
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet update
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd_no_webhook | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default_no_webhook | kubectl apply -f -

# Install CRDs into a cluster
install_with_webhook: manifests
	kustomize build config/crd | kubectl apply -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy_with_webhook: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# updates the full config yaml file
update: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default_no_webhook > hack/iam-manager.yaml
	kustomize build config/default > hack/iam-manager_with_webhook.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd_no_webhook/bases


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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
