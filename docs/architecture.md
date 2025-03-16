# IAM Manager Architecture

This document provides an overview of the IAM Manager architecture and its interaction with Kubernetes and AWS components.

## Architecture Diagram

```mermaid
graph TD
    %% Define styles
    classDef k8s fill:#326ce5,color:white,stroke:white,stroke-width:2px
    classDef aws fill:#FF9900,color:white,stroke:white,stroke-width:2px
    classDef security fill:#3f8428,color:white,stroke:white,stroke-width:2px
    classDef app fill:#764ABC,color:white,stroke:white,stroke-width:2px
    
    %% Kubernetes Components
    User([DevOps/User])
    User -->|Create/Update/Delete| IAMRoleCR[IAMRole CR]
    
    subgraph Kubernetes Cluster
        IAMRoleCR:::k8s
        APIServer[Kubernetes API Server]:::k8s
        Webhook[Validation Webhook]:::security
        IAMControllerPod[IAM Manager Controller]:::k8s
        ServiceAccount[Kubernetes Service Accounts]:::k8s
        AppPods[Application Pods]:::app
        
        IAMRoleCR -->|Submit| APIServer
        APIServer -->|Validate| Webhook
        Webhook -->|Policy Validation| APIServer
        APIServer -->|Watch| IAMControllerPod
        IAMControllerPod -->|Update Status| APIServer
        ServiceAccount -->|Mount Token| AppPods
    end
    
    %% AWS Components
    subgraph AWS Account
        IAMRoles[IAM Roles]:::aws
        PermBoundary[Permission Boundaries]:::security
        TrustPolicy[Trust Relationships]:::security
        AwsServices[AWS Services]:::aws
        
        IAMControllerPod -->|Create/Update/Delete| IAMRoles
        IAMRoles -->|Limited by| PermBoundary
        IAMRoles -->|Define| TrustPolicy
        AppPods -->|AssumeRole| IAMRoles
        IAMRoles -->|Access| AwsServices
    end
    
    %% IRSA Specific
    subgraph IRSA Components
        EKSOIDCProvider[EKS OIDC Provider]:::aws
        
        IAMControllerPod -->|Configure| TrustPolicy
        ServiceAccount -->|Token Used By| EKSOIDCProvider
        EKSOIDCProvider -->|Authenticate| TrustPolicy
    end
```

## Component Descriptions

### Kubernetes Components

- **IAMRole CR**: Custom Resource that defines the desired IAM role configuration
- **Validation Webhook**: Ensures IAM policies comply with allowed policies and resource limits
- **IAM Manager Controller**: Reconciles IAMRole CRs with actual AWS IAM roles
- **Service Accounts**: Kubernetes identities that can be associated with IAM roles (IRSA)
- **Application Pods**: Workloads that use IAM roles to access AWS services

### AWS Components

- **IAM Roles**: AWS Identity & Access Management roles created and managed by the controller
- **Permission Boundaries**: Limit the maximum permissions that can be granted to roles
- **Trust Relationships**: Define which entities can assume the roles
- **AWS Services**: Cloud services accessed using the IAM roles

### IRSA Components

- **EKS OIDC Provider**: Allows Kubernetes service accounts to authenticate to AWS and assume IAM roles

## Workflows

1. **Creation Flow**:
   - User creates an IAMRole CR in a Kubernetes namespace
   - Webhook validates the policy against allowed actions and resources
   - Controller creates an AWS IAM role with permission boundary
   - Status is updated with the role ARN and creation state

2. **IRSA Flow**:
   - IAMRole CR includes a service account annotation
   - Controller configures the trust policy to allow the service account to assume the role
   - Pods using the service account can access AWS resources via the role

3. **Security Controls**:
   - Permission boundaries limit the maximum permissions
   - Namespace-level role restrictions control proliferation of roles
   - Validation webhooks prevent creation of overly permissive policies
