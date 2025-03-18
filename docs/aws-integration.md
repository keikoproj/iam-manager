# AWS Integration for IAM Manager

This guide explains how to integrate iam-manager with AWS services, particularly focusing on IAM Roles for Service Accounts (IRSA) on Amazon EKS clusters.

## Overview

IAM Manager works by creating and managing AWS IAM roles based on Kubernetes custom resources. This integration requires proper AWS setup, including:

1. Permission boundaries to control the maximum permissions allowed
2. IAM policies for the iam-manager controller
3. OIDC provider configuration for EKS clusters (for IRSA)

## IAM Manager Controller Permissions

The iam-manager controller requires specific AWS IAM permissions to manage IAM roles. These permissions should be restricted using the principle of least privilege.

### Required Permissions

The controller needs permissions to:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:GetRole",
        "iam:PassRole",
        "iam:TagRole",
        "iam:UntagRole",
        "iam:ListRoleTags",
        "iam:PutRolePolicy",
        "iam:GetRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:ListRolePolicies",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:ListAttachedRolePolicies"
      ],
      "Resource": [
        "arn:aws:iam::*:role/k8s-*"
      ],
      "Condition": {
        "StringEquals": {
          "iam:PermissionsBoundary": "arn:aws:iam::*:policy/iam-manager-permission-boundary"
        }
      }
    }
  ]
}
```

### Setting Up Controller IAM

You can set up the required IAM resources using the provided CloudFormation template:

```bash
aws cloudformation create-stack \
  --stack-name iam-manager-resources \
  --template-body file://hack/iam-manager-cfn.yaml \
  --capabilities CAPABILITY_NAMED_IAM \
  --parameters ParameterKey=ClusterName,ParameterValue=your-cluster-name
```

## IAM Roles for Service Accounts (IRSA)

### Setting Up IRSA Controller Access

When running iam-manager on EKS, you should use IRSA to grant the controller access to AWS:

1. **Verify OIDC Provider Configuration**

   ```bash
   aws eks describe-cluster --name your-cluster-name --query "cluster.identity.oidc.issuer" --output text
   ```

   If this command returns a URL, your cluster has an OIDC provider configured. If not, set it up:

   ```bash
   eksctl utils associate-iam-oidc-provider --cluster your-cluster-name --approve
   ```

2. **Create an IAM Role for the Controller**

   Use eksctl to create the role and associate it with the iam-manager service account:

   ```bash
   eksctl create iamserviceaccount \
     --name iam-manager-controller \
     --namespace iam-manager-system \
     --cluster your-cluster-name \
     --attach-policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/IamManagerControllerPolicy \
     --approve
   ```

3. **Configure Controller Deployment**

   Ensure the deployment uses the correct service account:

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: iam-manager-controller-manager
     namespace: iam-manager-system
   spec:
     template:
       spec:
         serviceAccountName: iam-manager-controller
   ```

### Using IAM Manager for IRSA

IAM Manager can create roles specifically for IRSA:

1. **Configure ConfigMap with OIDC Provider URL**

   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: iammanager-config
     namespace: iam-manager-system
   data:
     cluster.oidc-provider-url: "https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE"
   ```

2. **Create an IAM Role with IRSA Annotation**

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
             - "s3:GetObject"
           Resource:
             - "arn:aws:s3:::my-bucket/*"
   ```

3. **Use the Role in Your Application**

   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: app
     namespace: default
   spec:
     template:
       spec:
         serviceAccountName: app-service-account
   ```

## Working with Permission Boundaries

IAM Manager uses permission boundaries to limit the maximum permissions that can be granted to roles it creates.

### Understanding Permission Boundaries

A permission boundary is an IAM policy that sets the maximum permissions that a role can have. Even if a role has a very permissive policy attached, the effective permissions are limited by the boundary.

For example, if a role policy allows `s3:*` but the permission boundary only allows `s3:GetObject`, the role can only perform the GetObject action.

### Configuring Permission Boundaries

IAM Manager applies a permission boundary to all roles it creates. The boundary is configured in the CloudFormation template and referenced in the ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iammanager-config
  namespace: iam-manager-system
data:
  defaults.permission-boundary-policy: "iam-manager-permission-boundary"
```

### Customizing Permission Boundaries

To customize the permission boundary, update the CloudFormation template with your desired permissions before creating the stack.

## Troubleshooting AWS Integration

### Common Issues

1. **Controller Cannot Create Roles**

   - Check if the controller has the correct IAM permissions
   - Verify the permission boundary policy exists
   - Check controller logs for specific error messages

2. **IRSA Not Working**

   - Verify the OIDC provider URL in the ConfigMap
   - Check that the service account specified in the annotation exists
   - Ensure the pod is using the correct service account

3. **Incorrect Permissions in Created Roles**

   - Check the permission boundary policy
   - Review the policy document in the Iamrole CR
   - Verify the role was created successfully in AWS

### Debugging AWS API Calls

To debug AWS API calls, enable verbose logging in the controller:

```bash
kubectl edit deployment iam-manager-controller-manager -n iam-manager-system
```

Add environment variables:

```yaml
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: AWS_SDK_GO_LOG_LEVEL
          value: "Debug"
        - name: LOG_LEVEL
          value: "debug"
```

## Security Best Practices

1. **Use Least Privilege Policies**
   - Restrict the controller's IAM permissions to only what it needs
   - Use specific resources in role policies instead of wildcards

2. **Implement Strong Boundary Policies**
   - Design permission boundaries that restrict access to sensitive resources
   - Regularly review and update boundaries as security requirements change

3. **Audit Role Usage**
   - Regularly audit the roles created by iam-manager
   - Monitor AWS CloudTrail logs for suspicious activity

4. **Manage Controller Access**
   - Restrict access to modify the iam-manager controller and its resources
   - Use Kubernetes RBAC to control who can create and modify IAM roles

## References

- [AWS IAM Permission Boundaries](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html)
- [EKS IRSA Documentation](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [AWS SDK for Go Documentation](https://docs.aws.amazon.com/sdk-for-go/api/)
