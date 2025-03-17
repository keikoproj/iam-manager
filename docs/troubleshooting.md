# IAM Manager Troubleshooting Guide

This guide provides solutions for common issues you might encounter when using iam-manager.

## Table of Contents
- [Installation Issues](#installation-issues)
- [Controller Issues](#controller-issues)
- [IAM Role Creation Issues](#iam-role-creation-issues)
- [IRSA Issues](#irsa-issues)
- [AWS Permission Issues](#aws-permission-issues)
- [Webhook Issues](#webhook-issues)
- [Collecting Information for Bug Reports](#collecting-information-for-bug-reports)
- [Viewing Controller Logs](#viewing-controller-logs)
- [Common Error States and Resolutions](#common-error-states-and-resolutions)
- [Checking IAM Role Status in AWS](#checking-iam-role-status-in-aws)

## Installation Issues

### Controller Pod Not Starting

**Symptoms**: The iam-manager-controller-manager pod is not starting or stays in a pending/crash loop state.

**Possible Causes and Solutions**:

1. **Resource constraints**:
   ```bash
   kubectl describe pod -n iam-manager-system iam-manager-controller-manager
   ```
   Look for resource-related errors. If necessary, adjust resource requests/limits in the Deployment.

2. **Image pull issues**:
   ```bash
   kubectl get events -n iam-manager-system
   ```
   Look for image pull errors. Check if the image exists and is accessible from your cluster.

3. **RBAC issues**:
   ```bash
   kubectl describe pod -n iam-manager-system iam-manager-controller-manager
   ```
   Look for RBAC permission errors. Ensure the service account has necessary permissions:
   ```bash
   kubectl describe clusterrole iam-manager-manager-role
   kubectl describe clusterrolebinding iam-manager-manager-rolebinding
   ```

### CRDs Not Installing

**Symptoms**: Custom Resource Definitions are not created after installation.

**Solution**:
```bash
# Install CRDs manually
kubectl apply -f config/crd/bases/
```

### Controller-Gen Issues on ARM64

**Symptoms**: Build failures on ARM64 architecture with controller-gen errors.

**Solution**:
1. Use a compatible version of controller-gen:
   ```bash
   make controller-gen
   ```

2. If issues persist, try specifying a different version in the Makefile:
   ```bash
   # In the Makefile, update the controller-gen version
   # For example:
   # controller-gen: ## Download controller-gen locally if necessary.
   #     $(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0)
   ```

## Controller Issues

### Controller Not Reconciling Resources

**Symptoms**: Custom resources are created but the controller doesn't appear to be processing them.

**Possible Causes and Solutions**:

1. **Controller crashes or errors**:
   ```bash
   kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager
   ```
   Look for error messages that might indicate why reconciliation is failing.

2. **AWS credentials issues**:
   ```bash
   kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager | grep -i aws
   ```
   Check if there are AWS authentication or permission errors.

3. **Configuration issues**:
   ```bash
   kubectl get configmap -n iam-manager-system iammanager-config -o yaml
   ```
   Verify the ConfigMap has correct AWS account ID, region, and other settings.

### High CPU/Memory Usage

**Symptoms**: The controller pod uses excessive CPU or memory.

**Solutions**:
- Check if there are too many IAM roles being processed simultaneously
- Look for reconciliation loops where the same resource is constantly being updated
- Adjust the controller's resource limits if necessary

## IAM Role Creation Issues

### IAM Roles Not Being Created in AWS

**Symptoms**: The Iamrole CR shows "Pending" status or error messages about AWS role creation.

**Possible Causes and Solutions**:

1. **AWS API credentials issue**:
   - Verify that the controller has proper AWS credentials
   - Check if the controller pod's service account has the correct IAM role (for IRSA)
   - Check if the AWS IAM user/role has sufficient permissions

2. **Permission boundary issues**:
   - Make sure the permission boundary policy specified in the ConfigMap exists in AWS
   - Verify that the controller has permission to attach the permission boundary

3. **Policy validation errors**:
   - Check if the PolicyDocument in the Iamrole CR is valid
   - Look for syntax errors in the policy JSON
   - Ensure actions and resources follow AWS IAM syntax

### Role Deletion Hangs

**Symptoms**: Attempts to delete an Iamrole CR hang indefinitely.

**Solution**:
```bash
# Remove finalizers from the resource
kubectl patch iamrole <name> -n <namespace> --type json -p '[{"op":"remove","path":"/metadata/finalizers"}]'
```

## IRSA Issues

### IRSA Not Working with Service Accounts

**Symptoms**: Pods using the service account cannot assume the IAM role.

**Possible Causes and Solutions**:

1. **Missing or incorrect annotation**:
   - Verify the Iamrole CR has the correct annotation: `iam.amazonaws.com/irsa-service-account`
   - Make sure the service account exists in the same namespace

2. **OIDC provider issues**:
   - Check if the EKS cluster has an OIDC provider configured
   - Verify the OIDC provider URL in the ConfigMap matches the cluster's OIDC provider

3. **Trust relationship issues**:
   - Check the trust relationship on the IAM role in AWS
   - Verify it includes the correct service account and OIDC provider

### Trust Policy Not Being Applied Correctly

**Symptoms**: The IAM role is created, but the trust policy doesn't include the service account.

**Solution**:
1. Check the controller logs for errors related to trust policy updates
2. Verify the IRSA annotation format is correct
3. Make sure the OIDC provider URL is correctly set in the ConfigMap

## AWS Permission Issues

### Insufficient Permissions for Controller

**Symptoms**: The controller fails to create or modify IAM roles.

**Solution**:
1. Check AWS CloudTrail logs for denied API calls
2. Ensure the IAM policy for the controller includes all necessary permissions:
   - iam:CreateRole
   - iam:DeleteRole
   - iam:GetRole
   - iam:PutRolePolicy
   - iam:DeleteRolePolicy
   - iam:AttachRolePolicy
   - iam:DetachRolePolicy
   - iam:TagRole
   - iam:ListRolePolicies
   - iam:ListAttachedRolePolicies

### Permission Boundary Conflicts

**Symptoms**: The controller cannot create roles due to permission boundary issues.

**Solution**:
1. Verify the permission boundary policy exists in your AWS account
2. Check that the controller has permission to use the boundary
3. Make sure the boundary policy is properly formatted

## Webhook Issues

### Validation Webhook Rejecting Resources

**Symptoms**: When creating or updating Iamrole CRs, requests are rejected by the webhook.

**Possible Causes and Solutions**:

1. **Policy contains disallowed actions**:
   - Check if your policy contains actions not allowed by the webhook
   - Review the controller logs for specific validation errors

2. **Webhook certificate issues**:
   - Check if the webhook's certificates are valid
   - Verify cert-manager is properly set up if used for certificate management

3. **Webhook configuration errors**:
   - Check the webhook configuration:
     ```bash
     kubectl get validatingwebhookconfiguration -l app=iam-manager
     ```

## Collecting Information for Bug Reports

When filing a bug report, please include:

1. **Controller logs**:
   ```bash
   kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager --tail=200
   ```

2. **Custom resource definitions**:
   ```bash
   kubectl get crd | grep iammanager.keikoproj.io
   ```

3. **Custom resource samples**:
   ```bash
   kubectl get iamrole -A -o yaml
   ```

4. **Kubernetes and controller versions**:
   ```bash
   kubectl version
   kubectl get deployment -n iam-manager-system iam-manager-controller-manager -o jsonpath='{.spec.template.spec.containers[0].image}'
   ```

5. **AWS environment information**:
   - AWS region
   - EKS cluster version (if applicable)
   - Whether IRSA is being used

6. **Reproduction steps**:
   - Clear steps to reproduce the issue
   - Sample resources that demonstrate the problem

## Viewing Controller Logs

To view the iam-manager controller logs for diagnostic purposes:

```bash
kubectl logs -n iam-manager-system deployment/iam-manager-controller-manager
```

For more detailed logging, you can modify the logging level in the ConfigMap or deployment:

```bash
kubectl edit deployment iam-manager-controller-manager -n iam-manager-system
```

Add or modify environment variables:
```yaml
env:
- name: LOG_LEVEL
  value: "debug"
```

## Common Error States and Resolutions

When an Iamrole resource shows an error state, here's what each state means and how to resolve it:

| State | Description | Resolution |
|-------|-------------|------------|
| `Error` | General error occurred during reconciliation | Check logs and error description in the status |
| `PolicyNotAllowed` | The policy contains actions that are not allowed by the permission boundary | Modify the policy to use only allowed actions |
| `PermissionDenied` | IAM Manager does not have sufficient permissions | Check AWS permissions for the controller role |
| `InvalidSpecification` | The spec contains invalid configuration | Validate your Iamrole YAML against the CRD spec |
| `InProgress` | The role is being created or updated | No action needed; the controller is processing the request |

## Checking IAM Role Status in AWS

Sometimes it's useful to verify the IAM role state directly in AWS:

```bash
# Get the IAM role name from the Iamrole status
ROLE_NAME=$(kubectl get iamrole -n <namespace> <name> -o jsonpath='{.status.roleName}')

# Check if the role exists in AWS
aws iam get-role --role-name $ROLE_NAME --profile <your-profile>

# Check permissions
aws iam list-role-policies --role-name $ROLE_NAME --profile <your-profile>
aws iam list-attached-role-policies --role-name $ROLE_NAME --profile <your-profile>
