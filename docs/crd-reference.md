# Iamrole CRD Reference

This document provides a detailed reference for the `Iamrole` custom resource definition (CRD) used by IAM Manager.

## Overview

The `Iamrole` CRD is the primary resource used to define AWS IAM roles in Kubernetes. When you create an `Iamrole` custom resource, IAM Manager creates a corresponding IAM role in AWS with the specified policies and trust relationships.

## Resource Definition

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: example-role
  namespace: default
  # Optional annotation for IRSA integration
  annotations:
    iam.amazonaws.com/irsa-service-account: my-service-account
spec:
  # IAM permissions policy (required)
  PolicyDocument:
    Version: "2012-10-17"  # Optional, defaults to "2012-10-17"
    Statement:
      - Effect: Allow  # Required: "Allow" or "Deny"
        Action:        # Required: List of IAM actions
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:      # Required: List of AWS resources
          - "arn:aws:s3:::mybucket/*"
          - "arn:aws:s3:::mybucket"
        Sid: "AllowS3Access"  # Optional: Statement identifier
  
  # Trust policy (optional)
  # If not specified, a default trust policy will be used
  AssumeRolePolicyDocument:
    Version: "2012-10-17"  # Optional, defaults to "2012-10-17"
    Statement:
      - Effect: Allow
        Action: "sts:AssumeRole"
        Principal:
          AWS:
            - "arn:aws:iam::123456789012:role/KubernetesNode"
        # Optional conditions
        Condition:
          StringEquals:
            "aws:SourceAccount": "123456789012"
          StringLike:
            "aws:username": "admin-*"
  
  # Custom role name (optional)
  # Only available in privileged namespaces
  RoleName: "custom-role-name"
```

## Field Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `PolicyDocument` | Object | Yes | Defines the permissions for the IAM role |
| `AssumeRolePolicyDocument` | Object | No | Defines which entities can assume the role (trust policy) |
| `RoleName` | String | No | Custom name for the IAM role (only for privileged namespaces) |

### PolicyDocument Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Version` | String | No | Policy language version (defaults to "2012-10-17") |
| `Statement` | Array | Yes | List of policy statements |

### Statement Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Effect` | String | Yes | Either "Allow" or "Deny" |
| `Action` | Array of Strings | Yes | List of AWS API actions to allow or deny |
| `Resource` | Array of Strings | Yes | List of AWS resources the actions apply to |
| `Sid` | String | No | Statement identifier for logging and debugging |

### AssumeRolePolicyDocument Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Version` | String | No | Policy language version (defaults to "2012-10-17") |
| `Statement` | Array | Yes | List of trust policy statements |

### Trust Policy Statement Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Effect` | String | No | Either "Allow" or "Deny" (defaults to "Allow") |
| `Action` | String | No | The action to allow/deny (typically "sts:AssumeRole") |
| `Principal` | Object | Yes | The entity that can assume the role |
| `Condition` | Object | No | Additional conditions on the trust relationship |

### Principal Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `AWS` | String/Array | No* | AWS account/role/user ARN(s) that can assume the role |
| `Service` | String | No* | AWS service that can assume the role (e.g., "ec2.amazonaws.com") |
| `Federated` | String | No* | Federated identity provider (e.g., OIDC provider) |

*At least one of AWS, Service, or Federated must be specified.

### Condition Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `StringEquals` | Map | No | Exact string matching conditions |
| `StringLike` | Map | No | String pattern matching conditions using wildcards |

## Status Fields

The `Iamrole` resource includes the following status fields that are populated by the controller:

| Field | Description |
|-------|-------------|
| `roleName` | The name of the IAM role in AWS |
| `roleARN` | The Amazon Resource Name (ARN) of the IAM role |
| `roleID` | The unique identifier of the IAM role |
| `state` | Current state of the IAM role (Ready, Error, etc.) |
| `retryCount` | Number of reconciliation attempts |
| `errorDescription` | Description of any errors that occurred |
| `lastUpdatedTimestamp` | When the role was last updated |

## Annotations

| Annotation | Description |
|------------|-------------|
| `iam.amazonaws.com/irsa-service-account` | Specifies the service account that can use this IAM role (for IRSA) |

## Examples

### Basic Role with S3 Access

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: s3-reader
  namespace: default
spec:
  PolicyDocument:
    Statement:
      - Effect: "Allow"
        Action:
          - "s3:GetObject"
          - "s3:ListBucket"
        Resource:
          - "arn:aws:s3:::example-bucket/*"
          - "arn:aws:s3:::example-bucket"
        Sid: "AllowS3Access"
  AssumeRolePolicyDocument:
    Statement:
      - Effect: "Allow"
        Action: "sts:AssumeRole"
        Principal:
          AWS:
            - "arn:aws:iam::123456789012:role/KubernetesNode"
```

### IRSA Role for Pod Service Account

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
        Action:
          - "dynamodb:GetItem"
          - "dynamodb:PutItem"
        Resource:
          - "arn:aws:dynamodb:*:*:table/my-table"
```

## Common States

| State | Description |
|-------|-------------|
| `Ready` | The IAM role has been successfully created/updated in AWS |
| `Error` | An error occurred during creation/update |
| `PolicyNotAllowed` | The policy contains actions that are not allowed |
| `PermissionDenied` | IAM Manager does not have sufficient permissions |
| `InvalidSpecification` | The spec contains invalid configuration |
| `InProgress` | The role is being created or updated |
