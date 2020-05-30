This document explains configmap variables.

| Property                          | Definition                    | Default            | Required/Optional  |
| ----------------------------------|:-----------------------------:| ------------------:|-------------------:|
| iam.policy.action.prefix.whitelist| Allowed IAM policy actions    |                    |Optional            |
| iam.policy.resource.blacklist     | Restricted IAM policy resource|                    |Optional            |
| iam.policy.s3.restricted.resource | Restricted S3 resource        |                    |Optional            |
| aws.accountId                     | AWS account ID where IAM roles are created|        |Optional            |
| iam.managed.policies              | User managed IAM policies     |                    |Optional            |
| iam.managed.permission.boundary.policy| User managed permission boundary policy|k8s-iam-manager-cluster-permission-boundary       |Required            |
| webhook.enabled                   |  Enable webhook?              | false              | Required           |
| iam.role.max.limit.per.namespace      | Maximum number of roles per namespace |   1        | Required |
| aws.region                        | AWS Region                    | us-west-2          | Required |
| iam.default.trust.policy.role.arn.list| Default trust policy role arn list |           | Optional |
| iam.role.derive.from.namespace    | Derive iam role name from namespace? if true it will be k8s-<namespace> | false | Optional|
| controller.desired.frequency      | Controller frequency to check the state of the world (in seconds) | 300  | Optional |
| k8s.cluster.name                  | Name of the cluster           |                    | Optional | 
| k8s.cluster.oidc.issuer.url       | OIDC issuer of the cluster    |                    | Optional |
| iam.irsa.enabled                  | IRSA option enabled