# Developer Guide for IAM Manager

This guide provides instructions for developers who want to contribute to the iam-manager project.

## Project Overview

IAM Manager is built using [Kubebuilder](https://book.kubebuilder.io/), a framework for building Kubernetes APIs using custom resource definitions (CRDs). Kubebuilder provides scaffolding tools to quickly create new APIs, controllers, and webhook components. Understanding Kubebuilder will greatly help in comprehending the iam-manager codebase structure and development workflow.

## Development Environment Setup

### Prerequisites

- Go 1.19+ (check the current version in go.mod)
- Kubernetes cluster for testing (minikube, kind, or a remote cluster)
- Docker for building images
- kubectl CLI
- kustomize
- controller-gen
- AWS account with IAM permissions for testing

### Clone the Repository

```bash
git clone https://github.com/keikoproj/iam-manager.git
cd iam-manager
```

### Install Required Tools

The Makefile can help install required development tools:

```bash
# Install controller-gen
make controller-gen

# Install kustomize
make kustomize

# Install mockgen (for tests)
make mockgen
```

**Note for ARM64 users**: There are known issues with some versions of controller-gen on ARM64 architecture. If you encounter issues, try specifying a compatible version in the Makefile:

```makefile
# Try version v0.13.0 or v0.17.0 for ARM64
controller-gen:
    $(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0)
```

### Build the Project

```bash
# Build the manager binary
make

# Run the manager locally (outside the cluster)
make run
```

## Setting Up AWS Resources

For local development, you'll need to set up the necessary AWS resources:

```bash
# Set environment variables
export AWS_ACCOUNT_ID=123456789012
export AWS_REGION=us-west-2
export CLUSTER_NAME=my-cluster

# Create the AWS resources using CloudFormation
aws cloudformation create-stack \
  --stack-name iam-manager-dev-resources \
  --template-body file://hack/iam-manager-cfn.yaml \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameters ParameterKey=ClusterName,ParameterValue=$CLUSTER_NAME
```

## Running Tests

### Unit Tests

```bash
# Run unit tests
make test
```

### Integration Tests

For integration tests, you need a Kubernetes cluster and AWS access:

```bash
# Set up environment variables for integration tests
export KUBECONFIG=~/.kube/config
export AWS_REGION=us-west-2
export AWS_PROFILE=your-aws-profile

# Run integration tests
make integration-test
```

## Creating and Deploying Custom Builds

### Building Docker Images

To build a custom Docker image:

```bash
# Build the controller image
make docker-build IMG=your-registry/iam-manager:your-tag

# Push the image to your registry
make docker-push IMG=your-registry/iam-manager:your-tag
```

### Deploying Custom Builds

Deploy your custom build to a cluster:

```bash
# Deploy with your custom image
make deploy IMG=your-registry/iam-manager:your-tag
```

## Code Structure

Here's an overview of the project structure:

```
.
├── api/                    # API definitions (CRDs)
│   └── v1alpha1/           # API version
├── cmd/                    # Entry points
├── config/                 # Kubernetes YAML manifests
├── controllers/            # Reconciliation logic
│   └── iamrole_controller.go # Main controller logic
├── pkg/                    # Shared packages
│   ├── awsapi/             # AWS API client wrapper
│   └── k8s/                # Kubernetes helpers
└── hack/                   # Development scripts
```

### Key Components

- **api/v1alpha1**: Contains the CRD definitions, including the Iamrole type.
- **controllers**: Contains the controller that reconciles the Iamrole custom resources.
- **pkg/awsapi**: Implements the AWS API client for IAM operations.

## Making Changes

### Adding a New Feature

1. Create a new branch: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Add tests for your changes
4. Run tests: `make test`
5. Build and verify: `make`
6. Commit changes with DCO signature: `git commit -s -m "Your commit message"`
7. Push changes: `git push origin feature/your-feature-name`
8. Create a pull request

### Adding New API Fields

To add new fields to the Iamrole CRD:

1. Modify the `api/v1alpha1/iamrole_types.go` file
2. Run code generation: `make generate`
3. Update CRDs: `make manifests`
4. Update the controller reconciliation logic to handle the new fields

## Debugging

### Running the Controller Locally

For easier debugging, you can run the controller outside the cluster:

```bash
# Run the controller locally
make run
```

### Remote Debugging

You can use Delve for remote debugging:

```bash
# Install Delve if you don't have it
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with Delve
dlv debug ./cmd/manager/main.go -- --kubeconfig=$HOME/.kube/config
```

### Verbose Logging

To enable debug logs:

```bash
# When running locally
make run ARGS="--zap-log-level=debug"

# In a deployed controller
kubectl edit deployment iam-manager-controller-manager -n iam-manager-system
# Add environment variable LOG_LEVEL=debug
```

## Code Generation

iam-manager uses kubebuilder and controller-gen for code generation.

### Kubebuilder and controller-gen

IAM Manager follows the [Kubebuilder](https://book.kubebuilder.io/) project structure and conventions. The project was initially scaffolded using Kubebuilder, which set up:

- API types in `api/v1alpha1/`
- Controller logic in `controllers/`
- Configuration files in `config/`
- Main entry point in `cmd/manager/main.go`

When you make changes to the API types, you need to regenerate various files:

```bash
# Generate CRDs
make manifests

# Generate code (deepcopy methods, etc.)
make generate
```

### Adding New API Types

To add a new Custom Resource Definition:

```bash
# Use kubebuilder to scaffold a new API
kubebuilder create api --group iammanager --version v1alpha1 --kind YourNewResource

# This will create:
# - api/v1alpha1/yournewresource_types.go
# - controllers/yournewresource_controller.go
# - And update main.go to include the new controller
```

After scaffolding, you'll need to:
1. Define your API schema in the `_types.go` file
2. Implement the reconciliation logic in the controller
3. Regenerate the manifests and code as described above

## Working with Webhooks

iam-manager uses validating webhooks to enforce security policies. To modify webhook logic:

1. Edit the validation logic in `api/v1alpha1/iamrole_webhook.go`
2. Regenerate manifests: `make manifests`
3. Deploy the changes: `make deploy`

## Continuous Integration

The project uses GitHub Actions for CI. When you submit a PR, the CI will:

1. Run unit tests
2. Build the controller image
3. Verify code generation is up-to-date
4. Check code style

Make sure all CI checks pass before requesting a review.

## Releasing

To create a new release:

1. Update version tags in all relevant files
2. Run tests and ensure they pass
3. Create a git tag: `git tag -a v0.x.y -m "Release v0.x.y"`
4. Push the tag: `git push origin v0.x.y`
5. Create a release on GitHub with release notes
