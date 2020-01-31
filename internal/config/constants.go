package config

// Global constants
var (
	IamManagedPermissionBoundaryPolicy = "arn:aws:iam::%s:policy/iam-manager-permission-boundary"
	ManagedPolicies                    []string
	AwsAccountId                       string
	InlinePolicyName                   = "custom"
)
