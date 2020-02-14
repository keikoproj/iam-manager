This document explains configmap variables.

| Property                          | Definition                    | Required/Optional  |
| ----------------------------------|:-----------------------------:| ------------------:|
| iam.policy.action.prefix.whitelist| Allowed IAM policy actions    | Optional |
| iam.policy.resource.blacklist     | Restricted IAM policy resource| Optional |
| iam.policy.s3.restricted.resource | Restricted S3 resource        | Optional |
| aws.accountId                     | AWS account ID where IAM roles are created| Required |
| iam.managed.policies              | User managed IAM policies     | Optional |
| iam.managed.permission.boundary.policy| User managed permission boundary policy| Required |