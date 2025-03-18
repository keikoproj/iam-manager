# IAM Manager ConfigMap Properties

This document describes all the configuration options available in the iam-manager ConfigMap.

## ConfigMap Structure

The iam-manager ConfigMap consists of several sections that control different aspects of the controller's behavior:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iammanager-config
  namespace: iam-manager-system
data:
  # AWS configuration
  aws.account: "123456789012"
  aws.region: "us-west-2"
  
  # Cluster configuration
  cluster.name: "my-cluster"
  cluster.oidc-provider-url: "https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
  
  # Controller configuration
  controller.reconcile.interval: "5m"
  controller.max-concurrent-reconciles: "5"
  
  # IAM role defaults
  defaults.permission-boundary-policy: "iam-manager-permission-boundary"
  defaults.role-name-prefix: "k8s-"
  
  # Policy validation
  policy.validation.enabled: "true"
  policy.allowed-actions: "s3:*,dynamodb:*,sqs:*,sns:*"
  
  # Role limits
  limits.max-roles-per-namespace: "5"
  
  # IRSA configuration
  iam.irsa.enabled: "true"
  iam.irsa.regional.endpoint.disabled: "false"
  
  # Custom role naming pattern
  iam.role.pattern: "k8s-{{ .ObjectMeta.Name }}"
```

## Available Configuration Options

### AWS Settings

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `aws.account` | N/A | The AWS account ID where IAM roles will be created | Required |
| `aws.region` | `us-west-2` | The AWS region for API operations | Required |
| `aws.role-prefix` | `k8s-` | Prefix for all IAM roles created by the controller | Optional |
| `aws.tags` | `managed-by=iam-manager` | Tags to apply to all created IAM roles | Optional |
| `aws.accountId` | Empty | AWS account ID where IAM roles are created (legacy syntax) | Optional |

### Cluster Settings

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `cluster.name` | N/A | Name of the Kubernetes cluster | Required |
| `cluster.oidc-provider-url` | Empty | OIDC provider URL for EKS clusters (required for IRSA) | Optional |
| `cluster.domain` | `cluster.local` | Kubernetes cluster domain | Optional |
| `k8s.cluster.name` | Empty | Alternative name of the cluster (legacy syntax) | Optional |
| `k8s.cluster.oidc.issuer.url` | Empty | Alternative OIDC issuer URL (legacy syntax) | Optional |

### Controller Settings

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `controller.reconcile.interval` | `5m` | How often to run full reconciliation | Optional |
| `controller.max-concurrent-reconciles` | `5` | Maximum number of concurrent reconciles | Optional |
| `controller.status-update-interval` | `1m` | How often to update status for resources | Optional |
| `controller.manager-workers` | `10` | Number of worker threads in the controller manager | Optional |
| `controller.desired.frequency` | `300` | Controller frequency to check state in seconds (legacy syntax) | Optional |

### IAM Role Defaults

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `defaults.permission-boundary-policy` | `iam-manager-permission-boundary` | Default permission boundary to apply to roles | Required |
| `iam.managed.permission.boundary.policy` | `k8s-iam-manager-cluster-permission-boundary` | Alternative permission boundary name (legacy syntax) | Required |
| `defaults.trust-policy` | AWS account trust | Default trust policy if not specified | Optional |
| `iam.default.trust.policy` | Empty | Default trust policy role (legacy syntax) | Optional |
| `defaults.role-name-prefix` | `k8s-` | Prefix for IAM role names | Optional |
| `defaults.path` | `/` | Path for IAM roles | Optional |
| `iam.managed.policies` | Empty | User managed IAM policies to attach to all roles | Optional |

### Policy Validation

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `policy.validation.enabled` | `true` | Whether to validate policies against allowed actions | Optional |
| `webhook.enabled` | `false` | Enable validation webhook (legacy syntax) | Required |
| `policy.allowed-actions` | Empty | Comma-separated list of allowed IAM actions | Optional |
| `iam.policy.action.prefix.whitelist` | Empty | Allowed IAM policy actions (legacy syntax) | Optional |
| `policy.denied-actions` | Empty | Comma-separated list of denied IAM actions | Optional |
| `policy.allowed-resources` | `*` | Comma-separated list of allowed resource patterns | Optional |
| `iam.policy.resource.blacklist` | Empty | Restricted IAM policy resources (legacy syntax) | Optional |
| `iam.policy.s3.restricted.resource` | Empty | Restricted S3 resources (legacy syntax) | Optional |

### Role Limits

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `limits.max-roles-per-namespace` | `5` | Maximum number of IAM roles allowed per namespace | Optional |
| `iam.role.max.limit.per.namespace` | `1` | Maximum number of roles per namespace (legacy syntax) | Required |
| `limits.max-policy-size` | `10KB` | Maximum size of policy documents | Optional |

### IRSA Configuration

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `iam.irsa.enabled` | `false` | Enable IAM Roles for Service Accounts integration | Optional |
| `iam.irsa.regional.endpoint.disabled` | `false` | Disable regional STS endpoints for IRSA | Optional |

### Role Naming Pattern

| Property | Default | Description | Required |
|----------|---------|-------------|----------|
| `iam.role.pattern` | `k8s-{{ .ObjectMeta.Name }}` | Go template for IAM role names | Optional |

## Custom Role Naming Pattern

### `iam.role.pattern`

_Default_: `k8s-{{ .ObjectMeta.Name }}`

All IAM roles created by the controller will use this [GoLang template](https://golang.org/pkg/text/template/)
to generate the final IAM Role Name. The default setting works fine if you have
a single cluster - but if you want to operate multiple clusters in the same AWS
account you will need to make sure the controllers do not conflict.

The [`Iamrole`](/api/v1alpha1/iamrole_types.go) object is passed into the Go templating engine, enabling
you to use any object field found in that role. For example 
`mycluster-{{ .ObjectMeta.Namespace }}-{{ .ObjectMeta.Name }}`.

All IAM roles created and managed by the controller will use this pattern. This
helps organize IAM roles within your AWS Account, and can be used to ensure
uniqueness between different EKS Clusters within the same AWS account.

**Critical Note: If you have existing `IAMRole` resources in your cluster, and you make a change to
the `iam.role.pattern` setting - the controller will reconcile the situation by
creating NEW IAM roles. It will _not_ however clean up the old roles - thus you
will have left over unused IAM roles in your account.**

Get these settings right from the beginning, or be prepared to clean up the left
over roles.

## IRSA Regional Endpoints

### `iam.irsa.regional.endpoint.disabled`

_Default_: `false`

Information about Service Account regional endpoints can be found 
[here](https://github.com/aws/amazon-eks-pod-identity-webhook#aws_sts_regional_endpoints-injection).
By default, iam-manager will inject `eks.amazonaws.com/sts-regional-endpoints: "true"` as an annotation on service
accounts specified in IamRoles. Setting this property to `true` will disable this injection and remove the annotation so endpoint will default
back to global endpoint in us-east-1.

## Environment Variables

The following environment variables can be used to override ConfigMap settings:

| Environment Variable | Corresponding ConfigMap Key | Description |
|----------------------|----------------------------|-------------|
| `AWS_REGION` | `aws.region` | AWS region |
| `AWS_ACCOUNT_ID` | `aws.account` | AWS account ID |
| `CLUSTER_NAME` | `cluster.name` | Kubernetes cluster name |
| `RECONCILE_INTERVAL` | `controller.reconcile.interval` | Reconciliation interval |
| `LOG_LEVEL` | `logging.level` | Logging level |

These can be set in the controller Deployment specification.