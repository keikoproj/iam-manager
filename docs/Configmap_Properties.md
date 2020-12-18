This document explains configmap variables.

| Property                          | Definition                    | Default            | Required/Optional  |
| ----------------------------------|:-----------------------------:| ------------------:|-------------------:|
| iam.policy.action.prefix.whitelist| Allowed IAM policy actions    |                    |Optional            |
| iam.policy.resource.blacklist     | Restricted IAM policy resource|                    |Optional            |
| iam.policy.s3.restricted.resource | Restricted S3 resource        |                    |Optional            |
| aws.accountId                     | AWS account ID where IAM roles are created|        |Optional            |
| iam.managed.policies              | User managed IAM policies     |                    |Optional            |
| iam.managed.permission.boundary.policy| User managed permission boundary policy|k8s-iam-manager-cluster-permission-boundary       |Required            |
| webhook.enabled                   |  Enable webhook?              | `false             | Required           |
| iam.role.max.limit.per.namespace  | Maximum number of roles per namespace |   1        | Required |
| aws.region                        | AWS Region                    | `us-west-2`        | Required |
| iam.default.trust.policy          | Default trust policy role. This must follow v1alpha1.AssumeRolePolicyDocument syntax|           | Optional |
| [iam.role.prefix](#iamroleprefix) | See docs below...           | `k8s`              | Optional           |
| [iam.role.separator](#iamroleseparator) | See docs below...     | `-`                | Optional           |
| iam.role.derive.from.namespace    | Derive iam role name from namespace? if true it will be k8s-<namespace> | false | Optional|
| controller.desired.frequency      | Controller frequency to check the state of the world (in seconds) | 300  | Optional |
| k8s.cluster.name                  | Name of the cluster           |                    | Optional | 
| k8s.cluster.oidc.issuer.url       | OIDC issuer of the cluster    |                    | Optional |
| iam.irsa.enabled                  | Enable IRSA option?           | `false`            | Optional |


## `iam.role.prefix`

_Default_: `k8s`

All IAM roles created and managed by the controller will use this prefix. This
helps organize IAM roles within your AWS Account, and can be used to ensure
uniqueness between different EKS Clusters within the same AWS account.

**Critical Note: Read [this](#note-about-changing-iamroleprefix-and-iamroleseparator)
before changing this setting**

**Note: Changes to your IAM Policy may be required if you customize this**

## `iam.role.separator`

_Default_: `-`

This is the separator between the [IAM Role Prefix](#iamroleprefix) and the
desired IAM Role Name.

**Critical Note: Read [this](#note-about-changing-iamroleprefix-and-iamroleseparator)
before changing this setting**

**Note: Changes to your IAM Policy may be required if you customize this**

## Note about changing `iam.role.prefix` and `iam.role.separator`

If you have existing `IAMRole` resources in your cluster, and you make a change
to the `iam.role.prefix` or `iam.role.separator` setting - the controller will
reconcile the situation by creating NEW IAM roles. It will _not_ however clean
up the old roles - thus you will have left over unused IAM roles in your account.

Get these settings right from the beginning, or be prepared to clean up the left
over roles.
