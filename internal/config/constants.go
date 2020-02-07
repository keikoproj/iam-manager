package config

// Global constants
var (
	IamManagedPermissionBoundaryPolicy = "arn:aws:iam::%s:policy/iam-manager-permission-boundary"
)

const (
	InlinePolicyName        = "custom"
	IamManagerNamespaceName = "iam-manager-system"
	IamManagerConfigMapName = "iam-manager-iamroles-v1alpha1-configmap"
)

const (
	propertyIamPolicyWhitelist    = "iam.policy.action.prefix.whitelist"
	propertyIamPolicyBlacklist    = "iam.policy.resource.blacklist"
	propertyIamPolicyS3Restricted = "iam.policy.s3.restricted.resource"
	propertyAwsAccountID          = "aws.accountId"
	propertyAwsMasterRole         = "aws.MasterRole"

	// user managed policies
	propertyManagedPolicies = "iam.managed.policies"
)

const (
	separator = ","
)
