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
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Documentation](#documentation)
- [Version Compatibility](#version-compatibility)
- [Contributing](#contributing)

## Overview

iam-manager simplifies AWS IAM role management within Kubernetes clusters by providing a declarative approach through custom resources. It enables namespace-scoped IAM role creation, enforces security best practices, and integrates with AWS IAM Role for Service Accounts (IRSA).

Originally developed at Intuit to manage IAM roles across 200+ clusters and 8000+ namespaces, iam-manager allows application teams to create and update IAM roles as part of their GitOps deployment pipelines, eliminating manual IAM policy management. This enables a "single manifest" approach where teams can manage both Kubernetes resources and IAM permissions together. For more details on the design principles and origin story, see the [Managing IAM Roles as K8s Resources](https://medium.com/keikoproj/managing-iam-roles-as-k8s-resources-aa00c5c4447f) article.

## Requirements

- Kubernetes cluster 1.16+
- AWS IAM permissions to create/update/delete roles
- AWS account with permission boundary policy configured
- Cert-manager (for webhook validation, optional)

## Features

iam-manager provides a comprehensive set of features for IAM role management:

- [IAM Roles Management](docs/features.md#iam-roles-management) - Create, update, and delete IAM roles through Kubernetes resources
- [IAM Role for Service Accounts (IRSA)](docs/features.md#iam-role-for-service-accounts-irsa) - Integration with AWS IAM Roles for Service Accounts
- [AWS Service-Linked Roles](docs/features.md#aws-service-linked-roles) - Support for service-linked roles
- [Default Trust Policy for All Roles](docs/features.md#default-trust-policy-for-all-roles) - Enforce consistent trust policies
- [Maximum Number of Roles per Namespace](docs/features.md#maximum-number-of-roles-per-namespace) - Governance controls
- [Attaching Managed IAM Policies for All Roles](docs/features.md#attaching-managed-iam-policies-for-all-roles) - Simplified policy management
- [Multiple Trust policies](docs/features.md#multiple-trust-policies) - Flexible trust relationship configuration

## Quick Start

The fastest way to install iam-manager is to use the provided installation script:

```bash
git clone https://github.com/keikoproj/iam-manager.git
cd iam-manager
./hack/install.sh [cluster_name] [aws_region] [aws_profile]
```

For detailed installation instructions, configuration options, and prerequisites, see the [Installation Guide](docs/install.md).

## Usage

Here's a minimal example of an IAM role for accessing S3:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: s3-reader-role
  namespace: default
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:
          - "arn:aws:s3:::your-bucket-name/*"
        Sid: "AllowS3Access"
```

For IRSA (IAM Roles for Service Accounts) integration:

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: app-role
  namespace: default
  annotations:
    iam.amazonaws.com/irsa-service-account: app-service-account
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action: ["s3:GetObject"]
        Resource: ["arn:aws:s3:::your-bucket-name/*"]
```

For detailed examples and usage patterns, see the [examples directory](examples/) and the [CRD Reference](docs/crd-reference.md).

## Documentation

Comprehensive documentation is available:

- [Architecture Documentation](docs/architecture.md)
- [Quick Start Guide](docs/quickstart.md)
- [Design Documentation](docs/design.md)
- [Configuration Options](docs/configmap-properties.md)
- [Developer Guide](docs/developer-guide.md)
- [AWS Integration](docs/aws-integration.md)
- [AWS Security](docs/aws-security.md)
- [Features](docs/features.md)
- [Installation Guide](docs/install.md)
- [CRD Reference](docs/crd-reference.md)
- [Troubleshooting Guide](docs/troubleshooting.md)

## Version Compatibility

| iam-manager Version | Kubernetes Version | Go Version | Key Features |
|---------------------|-------------------|------------|--------------|
| v0.22.0 | 1.16 - 1.25 | 1.19+ | IRSA regional endpoint configuration |
| v0.21.0 | 1.16 - 1.24 | 1.18+ | Enhanced security features |
| v0.20.0 | 1.16 - 1.23 | 1.17+ | Improved reconciliation controller |
| v0.19.0 | 1.16 - 1.22 | 1.16+ | IRSA support improvements |
| v0.18.0 | 1.16 - 1.21 | 1.15+ | Custom role naming |

For detailed information about each release, see the [GitHub Releases page](https://github.com/keikoproj/iam-manager/releases).

## Contributing

Please check [CONTRIBUTING.md](CONTRIBUTING.md) before contributing.

<!-- Markdown link -->
[GithubMaintainedUrl]: https://github.com/keikoproj/iam-manager/graphs/commit-activity
[GithubPrsUrl]: https://github.com/keikoproj/iam-manager/pulls
[SlackUrl]: https://keikoproj.slack.com/app_redirect?channel=iam-manager

[ReleaseImg]: https://img.shields.io/github/release/keikoproj/iam-manager.svg
[ReleaseUrl]: https://github.com/keikoproj/iam-manager/releases

[BuildStatusImg]: https://github.com/keikoproj/iam-manager/actions/workflows/unit_test.yaml/badge.svg
[BuildMasterUrl]: https://github.com/keikoproj/iam-manager/actions/workflows/unit_test.yaml

[CodecovImg]: https://codecov.io/gh/keikoproj/iam-manager/branch/master/graph/badge.svg
[CodecovUrl]: https://codecov.io/gh/keikoproj/iam-manager

[GoReportImg]: https://goreportcard.com/badge/github.com/keikoproj/iam-manager
[GoReportUrl]: https://goreportcard.com/report/github.com/keikoproj/iam-manager