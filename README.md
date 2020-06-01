# iam-manager

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)][GithubMaintainedUrl]
[![PR](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)][GithubPrsUrl]
[![slack](https://img.shields.io/badge/slack-join%20the%20conversation-ff69b4.svg)][SlackUrl]

![version](https://img.shields.io/badge/version-0.0.1-blue.svg?cacheSeconds=2592000)
[![Build Status][BuildStatusImg]][BuildMasterUrl]
[![codecov][CodecovImg]][CodecovUrl]
[![Go Report Card][GoReportImg]][GoReportUrl]


AWS IAM role management for K8s namespaces inside cluster using k8s CRD Operator. 

#### Security:

Security will be a main concern when we design a solution to create/update/delete IAM roles inside a cluster independently. iam-manager uses AWS IAM Permission Boundary concept along with other solutions to secure the implementation. Please check [AWS Security](docs/AWS_Security.md) for more details.

#### Supported Features

Following features are supported by IAM Manager

[IAM Roles Management](docs/Features.md#iam-roles-management)  
[IAM Role for Service Accounts (IRSA)](docs/Features.md#iam-role-for-service-accounts-irsa)  
[AWS Service-Linked Roles](docs/Features.md#aws-service-linked-roles)  
[Default Trust Policy for All Roles](docs/Features.md#default-trust-policy-for-all-roles)  
[Maximum Number of Roles per Namespace](docs/Features.md#maximum-number-of-roles-per-namespace)  
[Attaching Managed IAM Policies for All Roles](docs/Features.md#attaching-managed-iam-policies-for-all-roles)  
[Multiple Trust policies](docs/Features.md#multiple-trust-policies)

##### iam-manager config-map
This [document](docs/Configmap_Properties.md) provide explanation on configmap variables.

#### Additional Info  
iam-manager is built using kubebuilder project and like any other kubebuilder project iam-manager also uses cert-manager to manage the SSL certs for webhooks.


#### Usage:  
Following is the sample Iamrole spec. 

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: iam-manager-iamrole
spec:
  # Add fields here
  PolicyDocument:
    Statement:
      -
        Effect: "Allow"
        Action:
          - "s3:Get*"
        Resource:
          - "arn:aws:s3:::intu-oim*"
        Sid: "AllowS3Access"
  AssumeRolePolicyDocument:
    Version: "2012-10-17"
    Statement:
      -
        Effect: "Allow"
        Action: "sts:AssumeRole"
        Principal:
          AWS:
            - "arn:aws:iam::XXXXXXXXXXX:role/20190504-k8s-kiam-role"
```

To submit, kubectl apply -f iam_role.yaml --ns namespace1

#### Installation:
 
Simplest way to install iam-manager along with the role required for it to do the job is to run [install.sh](hack/install.sh) command.  

Update the allowed policies in [allowed_policies.txt](hack/allowed_policies.txt) and config map properties [config_map](hack/iammanager.keikoproj.io_iamroles-configmap.yaml) as per your environment before you run install.sh.

Note: You must be cluster admin and have exported KUBECONFIG and also has Administrator access to underlying AWS account and have the credentials exported.

example:
```bash
export KUBECONFIG=/Users/myhome/.kube/admin@eks-dev2-k8s  
export AWS_PROFILE=admin_123456789012_account
./install.sh [cluster_name] [aws_region] [aws_profile]
./install.sh eks-dev2-k8s us-west-2 aws_profile

```

To enable web hook or/and also update your installation of iam-manager to work with kiam please check [Installation](docs/Install.md) for detailed instructions.


## ❤ Contributing ❤

Please see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## Developer Guide

Please see [DEVELOPER.md](.github/DEVELOPER.md).

<!-- Markdown link -->
[install]: docs/README.md
[ext_link]: https://upload.wikimedia.org/wikipedia/commons/d/d9/VisualEditor_-_Icon_-_External-link.svg


[GithubMaintainedUrl]: https://github.com/keikoproj/iam-manager/graphs/commit-activity
[GithubPrsUrl]: https://github.com/keikoproj/iam-manager/pulls
[SlackUrl]: https://keikoproj.slack.com/messages/iam-manager

[BuildStatusImg]: https://travis-ci.org/keikoproj/iam-manager.svg?branch=master
[BuildMasterUrl]: https://travis-ci.org/keikoproj/iam-manager

[CodecovImg]: https://codecov.io/gh/keikoproj/iam-manager/branch/master/graph/badge.svg
[CodecovUrl]: https://codecov.io/gh/keikoproj/iam-manager

[GoReportImg]: https://goreportcard.com/badge/github.com/keikoproj/iam-manager
[GoReportUrl]: https://goreportcard.com/report/github.com/keikoproj/iam-manager