# iam-manager

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)][GithubMaintainedUrl]
[![PR](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)][GithubPrsUrl]
[![slack](https://img.shields.io/badge/slack-join%20the%20conversation-ff69b4.svg)][SlackUrl]

[![Release][ReleaseImg]][ReleaseUrl]
[![Build Status][BuildStatusImg]][BuildMasterUrl]
[![codecov][CodecovImg]][CodecovUrl]
[![Go Report Card][GoReportImg]][GoReportUrl]

A Kubernetes operator that manages AWS IAM roles for namespaces and service accounts using custom resources.

## Table of Contents
- [Overview](#overview)
- [Requirements](#requirements)
- [Features](#features)
- [Installation](#installation)
  - [Quick Installation](#quick-installation)
  - [Advanced Installation](#advanced-installation)
- [Usage](#usage)
  - [Basic IAM Role](#basic-iam-role)
  - [IAM Role for Service Accounts (IRSA)](#iam-role-for-service-accounts-irsa)
- [Configuration](#configuration)
- [Security](#security)
- [Troubleshooting](#troubleshooting)
- [Version Compatibility](#version-compatibility)
- [Contributing](#contributing)

## Overview

iam-manager simplifies AWS IAM role management within Kubernetes clusters by providing a declarative approach through custom resources. It enables namespace-scoped IAM role creation, enforces security best practices, and integrates with AWS IAM Role for Service Accounts (IRSA).

Originally developed at Intuit to manage IAM roles across 200+ clusters and 8000+ namespaces, iam-manager allows application teams to create and update IAM roles as part of their GitOps deployment pipelines, eliminating manual IAM policy management. This enables a "single manifest" approach where teams can manage both Kubernetes resources and IAM permissions together. For more details on the design principles and origin story, see the [Managing IAM Roles as K8s Resources](https://medium.com/keikoproj/managing-iam-roles-as-k8s-resources-aa00c5c4447f) article.

For a more detailed view of the architecture, see the [Architecture Documentation](docs/architecture.md).

## Requirements

- Kubernetes cluster 1.16+
- Access to AWS IAM
- Proper AWS credentials with permissions to create/update/delete IAM roles
- cert-manager (for webhook functionality)

## Features

iam-manager provides a comprehensive set of features for IAM role management:

- [IAM Roles Management](docs/Features.md#iam-roles-management) - Create, update, and delete IAM roles through Kubernetes resources
- [IAM Role for Service Accounts (IRSA)](docs/Features.md#iam-role-for-service-accounts-irsa) - Integration with AWS IAM Roles for Service Accounts
- [AWS Service-Linked Roles](docs/Features.md#aws-service-linked-roles) - Support for service-linked roles
- [Default Trust Policy for All Roles](docs/Features.md#default-trust-policy-for-all-roles) - Enforce consistent trust policies
- [Maximum Number of Roles per Namespace](docs/Features.md#maximum-number-of-roles-per-namespace) - Governance controls
- [Attaching Managed IAM Policies for All Roles](docs/Features.md#attaching-managed-iam-policies-for-all-roles) - Simplified policy management
- [Multiple Trust policies](docs/Features.md#multiple-trust-policies) - Flexible trust relationship configuration

## Installation

### Prerequisites

- Kubernetes cluster admin access (`kubectl` configured with admin permissions)
- AWS account with Administrator access
- Export necessary environment variables:

```bash
export KUBECONFIG=/path/to/your/kubeconfig
export AWS_PROFILE=your_aws_profile
```

### Quick Installation

The simplest way to install iam-manager is to use the provided installation script:

1. Customize the allowed policies in [allowed_policies.txt](hack/allowed_policies.txt)
2. Modify the [config_map](hack/iammanager.keikoproj.io_iamroles-configmap.yaml) for your environment
3. Run the installation script:

```bash
./hack/install.sh [cluster_name] [aws_region] [aws_profile]
```

Example:
```bash
./hack/install.sh eks-prod-cluster us-west-2 admin_profile
```

### Advanced Installation

For more detailed installation options, including:
- Enabling webhooks for validation
- Setting up with KIAM or IRSA
- Custom configurations

Please refer to the [Installation Guide](docs/Install.md).

## Usage

### Basic IAM Role

Create an IAM role by applying a YAML configuration:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: s3-reader-role
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:
          - "arn:aws:s3:::your-bucket-name/*"
          - "arn:aws:s3:::your-bucket-name"
        Sid: "AllowS3Access"
  AssumeRolePolicyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: "Allow"
        Action: "sts:AssumeRole"
        Principal:
          AWS:
            - "arn:aws:iam::<ACCOUNT_ID>:role/your-trusted-role"
```

Apply the configuration to your namespace:
```bash
kubectl apply -f iam_role.yaml -n your-namespace
```

### IAM Role for Service Accounts (IRSA)

For EKS clusters with IRSA enabled, define an IAM role for a specific service account:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: app-service-account-role
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "dynamodb:GetItem"
          - "dynamodb:PutItem"
        Resource:
          - "arn:aws:dynamodb:*:*:table/your-table"
        Sid: "AllowDynamoDBAccess"
  # For IRSA, the trust policy is automatically configured
  # based on your EKS OIDC provider
```

This allows you to bind an IAM role to a specific Kubernetes service account in your namespace.

## Configuration

iam-manager is configured through a ConfigMap. See [ConfigMap Properties](docs/Configmap_Properties.md) for details on available options, including:

- Permission boundaries
- Trust policies
- Role naming
- Resource limits
- AWS region settings

## Security

Security is a primary concern when managing IAM roles. iam-manager implements several security features:

- **Permission boundaries**: Limits the scope of created roles using AWS IAM Permission Boundaries (the actual permissions are the intersection of the permission boundary and the IAM role policy)
- **Namespace-level role restrictions**: Limits the number of roles per namespace
- **Policy action and resource restrictions**: Whitelisted policies can be configured through the config map
- **Audit logging**: Tracks all role creation and modification events
- **Controller security**: The iam-manager controller itself operates with limited permissions and can only:
  - Create roles with pre-defined permission boundaries
  - Create roles with pre-defined name patterns (e.g., k8s-*)
  - Delete roles that have specific tags applied by the controller

For detailed security information, see [AWS Security](docs/AWS_Security.md).

## Troubleshooting

### Common Issues

1. **Role creation fails**
   - Check AWS credentials
   - Verify permission boundary exists
   - Look for validation errors in the iam-manager logs

2. **WebHook issues**
   - Ensure cert-manager is installed and working
   - Check certificate validity
   - Verify webhook service is running

3. **IRSA not working**
   - Confirm EKS cluster has OIDC provider configured
   - Check service account annotations
   - Verify trust relationship is correct

### Viewing Logs

```bash
kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager
```

## Version Compatibility

| iam-manager Version | Kubernetes Version | AWS SDK Version | Notes |
|---------------------|--------------------| ---------------|-------|
| v0.22.0 | 1.22+ | v1.50+ | Current development version |
| v0.20.0 | 1.20+ | v1.50.11 | Latest release; Bug fixes and dependency updates |
| v0.19.0 | 1.19+ | v1.46+ | Upgraded Go version, controller-runtime, controller-gen |
| v0.18.0 | 1.18+ | v1.45+ | Fixed periodical reconcile |
| v0.17.0 | 1.18+ | v1.46.2 | Added support for multiple trusts with condition |
| v0.16.0 | 1.18+ | v1.45+ | Added support for KIAM and IRSA working together |

## Contributing

Please check [CONTRIBUTING.md](CONTRIBUTING.md) before contributing.

<!-- Markdown link -->
[install]: docs/README.md
[ext_link]: https://upload.wikimedia.org/wikipedia/commons/d/d9/VisualEditor_-_Icon_-_External-link.svg

[ReleaseImg]: https://img.shields.io/github/release/keikoproj/iam-manager.svg
[ReleaseUrl]: https://github.com/keikoproj/iam-manager/releases/latest

[GithubMaintainedUrl]: https://github.com/keikoproj/iam-manager/graphs/commit-activity
[GithubPrsUrl]: https://github.com/keikoproj/iam-manager/pulls
[SlackUrl]: https://keikoproj.slack.com/messages/iam-manager

[BuildStatusImg]: https://github.com/keikoproj/iam-manager/actions/workflows/unit_test.yaml/badge.svg
[BuildMasterUrl]: https://github.com/keikoproj/iam-manager/actions/workflows/unit_test.yaml

[CodecovImg]: https://codecov.io/gh/keikoproj/iam-manager/branch/master/graph/badge.svg
[CodecovUrl]: https://codecov.io/gh/keikoproj/iam-manager

[GoReportImg]: https://goreportcard.com/badge/github.com/keikoproj/iam-manager
[GoReportUrl]: https://goreportcard.com/report/github.com/keikoproj/iam-manager