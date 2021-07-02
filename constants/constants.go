package constants

// Global constants
const (
	// InlinePolicyName defines user managed inline policy
	InlinePolicyName = "custom"

	// IamManagerNamespaceName is the namespace name where iam-manager controllers are running
	IamManagerNamespaceName = "iam-manager-system"

	// IamManagerConfigMapName is the config map name for iam-manager namespace
	IamManagerConfigMapName = "iam-manager-iamroles-v1alpha1-configmap"

	OIDCAudience = "sts.amazonaws.com"

	IRSAAnnotation = "iam.amazonaws.com/irsa-service-account"

	ServiceAccountRoleAnnotation = "eks.amazonaws.com/role-arn"

	IamManagerPrivilegedNamespaceAnnotation = "iammanager.keikoproj.io/privileged"
)
