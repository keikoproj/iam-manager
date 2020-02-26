This document explains configmap variables.

| Property                          | Definition                    | Default            | Required/Optional  |
| ----------------------------------|:-----------------------------:| ------------------:|-------------------:|
| iam.policy.action.prefix.whitelist| Allowed IAM policy actions    |                    |Optional            |
| iam.policy.resource.blacklist     | Restricted IAM policy resource|                    |Optional            |
| iam.policy.s3.restricted.resource | Restricted S3 resource        |                    |Optional            |
| aws.accountId                     | AWS account ID where IAM roles are created|        |Optional            |
| iam.managed.policies              | User managed IAM policies     |                    |Optional            |
| iam.managed.permission.boundary.policy| User managed permission boundary policy|k8s-iam-manager-cluster-permission-boundary       |Required            |
| webhook.enabled                   |  Enable webhhok?              | false              | Required           |
| iam.role.max.limit.per.namespace      | Maximum number of roles per namespace |   1        | Required |
| aws.region                        | AWS Region                    | us-west-2          | Required |