# IAM Manager Design Documentation

This document outlines the design principles, architecture decisions, and trade-offs considered during the development of IAM Manager.

## Design Goals

IAM Manager was designed with the following goals in mind:

1. **Declarative Management**: Enable management of AWS IAM roles using Kubernetes-native declarative principles
2. **Security First**: Implement strong security boundaries to prevent privilege escalation
3. **GitOps Compatibility**: Allow IAM roles to be version-controlled alongside application code
4. **Namespace Isolation**: Provide isolation between different teams/namespaces
5. **Kubernetes Integration**: Leverage Kubernetes patterns and extend its API

## Core Design Principles

### 1. Separation of Concerns

IAM Manager separates the following concerns:

- **Role Definition**: Handled by the Iamrole CR
- **Permission Boundaries**: Managed separately and enforced by the controller
- **Trust Relationships**: Configured based on use case (standard or IRSA)
- **Validation Logic**: Implemented in admission webhooks

This separation allows for more flexible and maintainable code, as each component handles a specific aspect of IAM role management.

### 2. Kubernetes Extension Pattern

IAM Manager follows the Kubernetes extension pattern:

- Custom Resource Definitions (CRDs) to extend the Kubernetes API
- Custom controllers to reconcile desired vs. actual state
- Admission webhooks for validation
- Operator pattern for reconciliation

This approach provides a seamless integration with Kubernetes, allowing users to manage AWS IAM roles using familiar kubectl commands and GitOps workflows.

### 3. Secure by Default

Security is a primary concern in IAM Manager's design:

- **Permission Boundaries**: All roles are created with a permission boundary that limits their maximum permissions
- **Validation Webhooks**: Prevent creation of roles with excessive permissions
- **Namespace Restrictions**: Limit the number of roles per namespace
- **Role Naming Conventions**: Enforce consistent naming with prefixes to identify managed roles
- **Resource Tagging**: Tag all AWS resources for auditing and management

### 4. Reconciliation Loop

IAM Manager uses a controller-based reconciliation loop:

1. Watch for changes to Iamrole CRs
2. Compare desired state (CR) with actual state (AWS)
3. Make changes to bring actual state in line with desired state
4. Update status to reflect current state

This pattern ensures that even if changes are made directly in AWS, the controller will detect and revert them to match the desired state defined in Kubernetes.

## Architecture Decisions

### Decision 1: Kubernetes Custom Resources

**Decision**: Use Kubernetes CRDs to represent IAM roles.

**Alternatives Considered**:
- External database to store role definitions
- Annotations on namespaces or service accounts
- External API server

**Rationale**:
- CRDs provide a native Kubernetes experience
- Built-in validation and versioning
- Can leverage existing Kubernetes RBAC
- Works with existing GitOps tools

### Decision 2: Permission Boundaries

**Decision**: Use AWS IAM Permission Boundaries to limit the maximum permissions of created roles.

**Alternatives Considered**:
- Policy validation only
- Custom approval workflow
- Limited IAM permissions for the controller

**Rationale**:
- Permission boundaries provide a hard limit at the AWS level
- Even if the controller is compromised, it cannot create overly permissive roles
- Clear separation between what permissions are allowed vs. what permissions are granted

### Decision 3: Namespace-Scoped Resources

**Decision**: Make Iamrole resources namespace-scoped rather than cluster-scoped.

**Alternatives Considered**:
- Cluster-scoped resources with namespace field
- Custom namespace-based isolation

**Rationale**:
- Aligns with Kubernetes' namespace isolation model
- Allows for RBAC to be applied at namespace level
- Teams can manage their own IAM roles without affecting others
- Prevents naming conflicts between different teams

### Decision 4: AWS API Integration

**Decision**: Use the AWS SDK for Go to directly interact with the AWS API.

**Alternatives Considered**:
- AWS CloudFormation
- AWS CDK
- Terraform

**Rationale**:
- Direct API access provides more control and better error handling
- Faster reconciliation as there's no need to wait for external provisioning tools
- Lower dependency footprint

### Decision 5: IRSA Integration

**Decision**: Support IAM Roles for Service Accounts through annotations.

**Alternatives Considered**:
- Separate CRD for IRSA roles
- External service account mapping

**Rationale**:
- Annotations provide a simple, declarative way to associate roles with service accounts
- Consistent with how IRSA works in EKS
- Minimizes the learning curve for users familiar with IRSA

## Trade-offs

### Trade-off 1: Controller Permissions

**Context**: The controller needs permissions to create and manage IAM roles.

**Trade-off**: Giving the controller broad IAM permissions could be a security risk, but too limited permissions would restrict functionality.

**Decision**: Use a combination of:
- Permission boundaries to limit what the controller can create
- IAM conditions to restrict actions to specific role name patterns
- Resource tagging to identify controller-managed resources

### Trade-off 2: Validation vs. Flexibility

**Context**: Strict validation prevents misuse but can limit legitimate use cases.

**Trade-off**: Stricter validation improves security but reduces flexibility.

**Decision**: Implement configurable validation rules that can be adjusted by cluster administrators based on their security requirements.

### Trade-off 3: State Management

**Context**: How to handle reconciliation between Kubernetes state and AWS state.

**Trade-off**: Continuously checking and updating state provides better consistency but increases API calls and potential rate limiting.

**Decision**: Implement periodic full reconciliation combined with event-driven updates to balance consistency with performance.

### Trade-off 4: Role Naming

**Context**: How to name IAM roles created by the controller.

**Trade-off**: Including namespace/name in role names provides clarity but can exceed AWS name length limits.

**Decision**: Use a deterministic naming scheme with prefixes and truncation strategies to handle long names while maintaining uniqueness.

## Future Design Considerations

1. **Multi-Account Support**: Extend IAM Manager to create roles across multiple AWS accounts
2. **Enhanced Metrics and Auditing**: Add more detailed metrics and audit logs
3. **Advanced Policy Templates**: Support for policy templates and inheritance
4. **Cross-Namespace References**: Allow referencing roles across namespaces with proper authorization
5. **Integration with External Secrets Management**: Better integration with secrets management systems for sensitive credentials

## Lessons Learned

1. **AWS API Limitations**: Working within AWS API rate limits and consistency model
2. **CRD Versioning**: Importance of forward-compatible CRD designs for upgrades
3. **Error Handling**: Comprehensive error handling for eventual consistency
4. **Performance Optimization**: Balancing reconciliation frequency with performance
5. **Security Considerations**: Defense in depth approach to IAM security
