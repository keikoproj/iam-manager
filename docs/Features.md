### Supported Features

Following features are supported by IAM Manager

[IAM Roles Management](#iam-roles-management)  
[IAM Role for Service Accounts (IRSA)](#iam-role-for-service-accounts-irsa)  
[AWS Service-Linked Roles](#aws-service-linked-roles)  
[Default Trust Policy for All Roles](#default-trust-policy-for-all-roles)  
[Maximum Number of Roles per Namespace](#maximum-number-of-roles-per-namespace)  
[Attaching Managed IAM Policies for All Roles](#attaching-managed-iam-policies-for-all-roles)  
[Multiple Trust policies](#multiple-trust-policies)

#### Note: Please Note that Permission Boundary will be automatically added for each role and you can configure permission boundary name using config map variable.
```bash
iam.managed.permission.boundary.policy: "permission-boundary-policy-name"
```  

##### IAM Roles Management
Simplest way to create an IAM Role is to provide the PolicyDocument and AssumeRolePolicyDocument. You might have already guessed by now, to keep it simple, iam-manager tried to keep AWS CFN(Cloud Formation) format for IamRole Spec.  
Example:
```bash
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

##### IAM Role for Service Accounts (IRSA)
IAM Manager supports IRSA(IAM Role for service accounts) starting from IAM Manager 0.0.4 version.  
<B>Pre-requisites</B>  
For EKS Clusters, You can provide following params in config map and IAM Manager will automatically sets up OIDC IDP in AWS IAM if it doesn't exist
```bash
iam.irsa.enabled : "true"  
k8s.cluster.name : "EKS Cluster Name"  
```
OR  
You can provide OIDC URL (OIDC IDP must be created in AWS IAM)
```bash 
k8s.cluster.oidc.issuer.url: "<OIDC URL from K8s cluster>"
```
Request:  
Once pre-requisites are completed, Attach following annotation to the pod and IAM Manager will automatically attaches required trust policy to IAM Role.  
```bash
iam.amazonaws.com/irsa-service-account: "service_account_name"
```
IAM Manager will create Service Account if its doesn't exist or update the service account with required annotations.

Note: For kops clusters, you must install AWS [amazon-eks-pod-identity-webhook](https://github.com/aws/amazon-eks-pod-identity-webhook)  
For more info: [AWS Blog](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/)

Example:
```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: iam-manager-iamrole-irsa
  annotations:
    iam.amazonaws.com/irsa-service-account: aws-sa
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
```
This should automatically create service account if it doesn't exist
```yaml
mtvl15367e28a:iam-manager nmogulla$ k get sa aws-sa -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::XXXXXXXXX:role/k8s-iam-manager-iamrole-irsa
  creationTimestamp: "2020-06-01T06:19:41Z"
  name: aws-sa
  namespace: a-test-usw2-test-123
  resourceVersion: "10670180"
  selfLink: /api/v1/namespaces/a-test-usw2-test-123/serviceaccounts/aws-sa
  uid: e1307b2a-a3cf-11ea-8b0e-0a6a4f8e42d4
secrets:
- name: aws-sa-token-bcflm
```


##### AWS Service-Linked Roles
IAM Manager can also be used to create service linked roles for example "eks.amazonaws.com" to allow EKS to perform the required activities to run cluster.

```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: iam-manager-iamrole-eks-role
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
          Service: "eks.amazonaws.com"
```

##### Default Trust Policy for All Roles
There might be a situations where as an administrator you might want to control the trust policy. For example, in KIAM use case every role must be trusted by master server role where kiam server is deployed. That can be configured in IAM Manager using config map variable  
```bash
iam.default.trust.policy.role.arn.list : "Comma delimited AWS IAM Role ARNs"
```
Example:
```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: iam-manager-iamrole-default-trust
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
```
This should automatically add the trust policy with sts:AssumeRole and Prinipal from config map variable value

##### Maximum Number of Roles per Namespace
By default, maximum number of roles per namespace is 1. You can configure the max roles per namespace using config map variable
```bash
iam.role.max.limit.per.namespace : "10"
```

##### Attaching Managed IAM Policies for All Roles
You can attach any managed iam policies to all the roles created by IAM Manager by configuring config map variable  
```bash
iam.managed.policies: "shared.security-policy-20200504-k8s"
```
This might be useful in use cases where security team wants to attach a managed policies for all the roles.  

#### Multiple Trust Policies
You can always use the combination of trust policies. For example: An IAM Role might need to access it from application as well as AwS service.

Example:
```yaml
apiVersion: iammanager.keikoproj.io/v1alpha1
kind: Iamrole
metadata:
  name: iam-manager-iamrole-multiple-trust
  annotations:
    iam.amazonaws.com/irsa-service-account: aws-sa-multiple
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
          Service: "eks.amazonaws.com"
```
This will have both AssumeRoleWithWebIdentity to assume role from a pod and also sts:AssumeRole for "eks.amazonaws.com" to allow access.
