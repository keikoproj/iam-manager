package config

import (
	"context"
	"github.com/keikoproj/iam-manager/internal/k8s"
	"strings"
)

const (
	iamPolicyWhitelist    = "iam.policy.action.prefix.whitelist"
	iamPolicyBlacklist    = "iam.policy.resource.blacklist"
	iamPolicyS3Restricted = "iam.policy.s3.restricted.resource"
)

const (
	separator = ","
)

//Properties struct loads the properties
type Properties struct {
	AllowedPolicyAction       []string
	RestrictedPolicyResources []string
	RestrictedS3Resources     []string
}

//LoadProperties function loads properties from various sources
func LoadProperties(ctx context.Context, kClient *k8s.Client, ns string, cmName string) *Properties {
	res := kClient.GetConfigMap(ctx, ns, cmName)
	allowedActions := strings.Split(res.Data[iamPolicyWhitelist], separator)
	restrictedResources := strings.Split(res.Data[iamPolicyBlacklist], separator)
	restrictedS3Resources := strings.Split(res.Data[iamPolicyS3Restricted], separator)

	return &Properties{
		AllowedPolicyAction:       allowedActions,
		RestrictedPolicyResources: restrictedResources,
		RestrictedS3Resources:     restrictedS3Resources,
	}

}
