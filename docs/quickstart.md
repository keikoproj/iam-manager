# IAM Manager Quickstart Guide

This guide will help you quickly get started with iam-manager by walking through the installation process and creating your first IAM role.

## Prerequisites

Before you begin, ensure you have:

- A Kubernetes cluster (v1.16+)
- `kubectl` configured with admin access to your cluster
- AWS CLI configured with appropriate permissions to create/modify IAM roles
- An IAM role or user with permissions to create and manage IAM roles

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/keikoproj/iam-manager.git
cd iam-manager
```

### 2. Create Required AWS Resources

You need to create the necessary AWS resources, including permission boundaries, before deploying iam-manager.

```bash
# Set your AWS account ID and region
export AWS_ACCOUNT_ID=123456789012
export AWS_REGION=us-west-2

# Create the AWS resources using CloudFormation
aws cloudformation create-stack \
  --stack-name iam-manager-resources \
  --template-body file://hack/iam-manager-cfn.yaml \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameters ParameterKey=ClusterName,ParameterValue=your-cluster-name
```

This creates:
- Permission boundaries for IAM roles
- IAM policy for the iam-manager controller
- Trust relationships for your cluster

### 3. Update the ConfigMap

Edit the ConfigMap to match your environment:

```bash
# Open the ConfigMap YAML file
vim config/default/iammanager.keikoproj.io_iamroles-configmap.yaml

# Update the following values:
# - AWS account ID
# - AWS region
# - Cluster name
# - OIDC provider URL (for EKS with IRSA)
```

### 4. Deploy the Controller

```bash
# Apply CRDs
kubectl apply -f config/crd/bases/

# Deploy the controller
make deploy
```

### 5. Verify the Installation

```bash
kubectl get pods -n iam-manager-system
```

You should see the iam-manager-controller-manager pod running.

## Creating Your First IAM Role

### Basic IAM Role

Create a file named `my-first-role.yaml`:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: my-first-role
  namespace: default
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:
          - "arn:aws:s3:::my-bucket/*"
          - "arn:aws:s3:::my-bucket"
        Sid: "AllowS3Access"
  AssumeRolePolicyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Action: "sts:AssumeRole"
        Principal:
          AWS:
            - "arn:aws:iam::123456789012:role/your-trusted-role"
```

Apply the role to your cluster:

```bash
kubectl apply -f my-first-role.yaml
```

### Check Role Status

```bash
kubectl get iamrole my-first-role -n default -o yaml
```

You should see the status field populated with information about your role, including its ARN and whether it's ready.

## Creating an IAM Role for Service Accounts (IRSA)

If you're using EKS and want to leverage IRSA, create a role with an annotation:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: app-service-account-role
  namespace: default
  annotations:
    iam.amazonaws.com/irsa-service-account: my-service-account
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:
          - "arn:aws:s3:::my-bucket/*"
          - "arn:aws:s3:::my-bucket"
        Sid: "AllowS3Access"
```

Apply it to your cluster:

```bash
kubectl apply -f irsa-role.yaml
```

## Using IAM Roles in Your Applications

### For Standard IAM Roles

Add the ARN to your application's AWS SDK configuration or use the AWS SDK's profile feature to assume the role.

### For IRSA Roles

1. Ensure your pod uses the service account specified in the IRSA annotation:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      serviceAccountName: my-service-account  # This must match the IRSA annotation
      containers:
      - name: app
        image: my-app:latest
```

2. The AWS SDK will automatically use the IAM role's credentials when making API calls from this pod.

## Next Steps

- Explore more [IAM Manager Examples](../examples/)
- Learn about [Configuration Options](configmap-properties.md)
- Read the [Architecture Documentation](architecture.md)
- Set up [AWS Integration](aws-integration.md) for advanced scenarios
- Review [AWS Security](AWS_Security.md) features

## Troubleshooting

If you encounter issues, check the [Troubleshooting Guide](troubleshooting.md) or view the controller logs:

```bash
kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager
```
