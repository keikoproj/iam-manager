package config

// Global constants
const (
	// InlinePolicyName defines user managed inline policy
	InlinePolicyName = "custom"

	// IamManagerNamespaceName is the namespace name where iam-manager controllers are running
	IamManagerNamespaceName = "iam-manager-system"

	// IamManagerConfigMapName is the config map name for iam-manager namespace
	IamManagerConfigMapName = "iam-manager-iamroles-v1alpha1-configmap"
)

const (
	// iam policy action prefix
	propertyIamPolicyWhitelist = "iam.policy.action.prefix.whitelist"

	// iam policy to blacklist resource
	propertyIamPolicyBlacklist = "iam.policy.resource.blacklist"

	// iam policy for restricting s3 resources
	propertyIamPolicyS3Restricted = "iam.policy.s3.restricted.resource"

	// aws region
	propertyAwsRegion = "aws.region"

	//enable webhook property
	propertyAWSAccountID = "aws.accountId"

	// user managed policies
	propertyManagedPolicies = "iam.managed.policies"

	// user managed permission boundary policy
	propertyPermissionBoundary = "iam.managed.permission.boundary.policy"

	//enable webhook property
	propertyWebhookEnabled = "webhook.enabled"

	//golang-templated pattern to use for iam role name generation
	propertyIamRolePattern = "iam.role.pattern"

	//max allowed aws iam roles per namespace
	propertyMaxIamRoles = "iam.role.max.limit.per.namespace"

	//propertyDesiredStateFrequency is a configurable param to make sure to check the external state (in seconds). default to 30 mins (1800 seconds)
	propertyDesiredStateFrequency = "controller.desired.frequency"

	//propertyClusterName can be used to set cluster name
	propertyClusterName = "k8s.cluster.name"

	//propertyIRSAEnabled can be used to enable/disable IRSA support by IAM-Manager
	propertyIRSAEnabled = "iam.irsa.enabled"

	//propertyK8sClusterOIDCIssuerUrl can be used to provide OIDC issuer url
	propertyK8sClusterOIDCIssuerUrl = "k8s.cluster.oidc.issuer.url"

	//propertyDefaultTrustPolicy can be used to provide default trust policy
	propertyDefaultTrustPolicy = "iam.default.trust.policy"
)

const (
	separator = ","

	OIDCAudience = "sts.amazonaws.com"

	IRSAAnnotation = "iam.amazonaws.com/irsa-service-account"
)
