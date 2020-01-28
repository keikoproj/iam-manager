package config

import (
	"context"
	"strings"

	"github.com/keikoproj/iam-manager/internal/k8s"
)

const (
	iamPolicyWhitelist    = "iam.policy.action.prefix.whitelist"
	iamPolicyBlacklist    = "iam.policy.resource.blacklist"
	iamPolicyS3Restricted = "iam.policy.s3.restricted.resource"
	awsAccountId          = "aws.accountId"
	awsMasterRole         = "aws.MasterRole"
	managedPolicies       = "iam.managed.policies"
)

const (
	separator = ","
)

//Properties struct loads the properties
type Properties struct {
	AllowedPolicyAction       []string
	RestrictedPolicyResources []string
	RestrictedS3Resources     []string
	AWSAccountId              string
	AWSMasterRole             string
	ManagedPolicies           []string
}

//LoadProperties function loads properties from various sources
func LoadProperties(ctx context.Context, kClient *k8s.Client, ns string, cmName string) *Properties {
	res := kClient.GetConfigMap(ctx, ns, cmName)
	allowedActions := strings.Split(res.Data[iamPolicyWhitelist], separator)
	restrictedResources := strings.Split(res.Data[iamPolicyBlacklist], separator)
	restrictedS3Resources := strings.Split(res.Data[iamPolicyS3Restricted], separator)
	managedPolicies := strings.Split(res.Data[managedPolicies], separator)
	awsAccountId := res.Data[awsAccountId]
	awsMasterRole := res.Data[awsMasterRole]

	return &Properties{
		AllowedPolicyAction:       allowedActions,
		RestrictedPolicyResources: restrictedResources,
		RestrictedS3Resources:     restrictedS3Resources,
		AWSAccountId:              awsAccountId,
		AWSMasterRole:             awsMasterRole,
		ManagedPolicies:           managedPolicies,
	}

}
