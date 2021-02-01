package config

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	"k8s.io/api/core/v1"
	"strings"
	"testing"
)

type PropertiesSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestPropertiesSuite(t *testing.T) {
	check.Suite(&PropertiesSuite{t: t})
	check.TestingT(t)
}

func (s *PropertiesSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *PropertiesSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

// test local properties for local environment
func (s *PropertiesSuite) TestLoadPropertiesLocalEnvSuccess(c *check.C) {
	Props = nil
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
	c.Assert(Props, check.NotNil)
	c.Assert(Props.AWSAccountID(), check.Equals, "123456789012")
}

// test failure when env is not local and cm is empty
// should not return nil pointer
func (s *PropertiesSuite) TestLoadPropertiesFailedNoCM(c *check.C) {
	Props = nil
	err := LoadProperties("")
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "config map cannot be nil")
}

func (s *PropertiesSuite) TestLoadPropertiesFailedNilCM(c *check.C) {
	Props = nil
	err := LoadProperties("", nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "config map cannot be nil")
}

func (s *PropertiesSuite) TestLoadPropertiesSuccess(c *check.C) {
	Props = nil
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.managed.permission.boundary.policy": "iam-manager-permission-boundary",
			"aws.accountId":                          "123456789012",
			"iam.role.max.limit.per.namespace":       "5",
			"aws.region":                             "us-east-2",
			"webhook.enabled":                        "true",
		},
	}
	err := LoadProperties("", cm)
	c.Assert(err, check.IsNil)
	c.Assert(Props.AWSRegion(), check.Equals, "us-east-2")
	c.Assert(Props.MaxRolesAllowed(), check.Equals, 5)
	c.Assert(Props.IsWebHookEnabled(), check.Equals, true)
	c.Assert(Props.AWSAccountID(), check.Equals, "123456789012")
	c.Assert(strings.HasPrefix(Props.ManagedPermissionBoundaryPolicy(), "arn:aws:iam:"), check.Equals, true)
}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithDefaults(c *check.C) {
	Props = nil
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.managed.permission.boundary.policy": "iam-manager-permission-boundary",
			"aws.accountId":                          "123456789012",
		},
	}
	err := LoadProperties("", cm)
	c.Assert(err, check.IsNil)
	c.Assert(Props.AWSRegion(), check.Equals, "us-west-2")
	c.Assert(Props.MaxRolesAllowed(), check.Equals, 1)
	c.Assert(Props.ControllerDesiredFrequency(), check.Equals, 1800)
	c.Assert(Props.IsWebHookEnabled(), check.Equals, false)
	c.Assert(Props.AWSAccountID(), check.Equals, "123456789012")
	c.Assert(strings.HasPrefix(Props.ManagedPermissionBoundaryPolicy(), "arn:aws:iam:"), check.Equals, true)
	c.Assert(Props.IamRolePattern(), check.Equals, "k8s-{{ .ObjectMeta.Name }}")
	//when an emty string passed split strings gives you array of 1 with ""
	c.Assert(len(Props.ManagedPolicies()), check.Equals, 1)
	c.Assert(Props.ManagedPolicies()[0], check.Equals, "")

}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithDefaultsManagedPoliciesWithNoPrefix(c *check.C) {
	Props = nil
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.managed.permission.boundary.policy": "iam-manager-permission-boundary",
			"aws.accountId":                          "123456789012",
			"iam.managed.policies":                   "DescribeEC2",
		},
	}
	err := LoadProperties("", cm)
	c.Assert(err, check.IsNil)
	c.Assert(Props.AWSRegion(), check.Equals, "us-west-2")
	c.Assert(Props.MaxRolesAllowed(), check.Equals, 1)
	c.Assert(Props.ControllerDesiredFrequency(), check.Equals, 1800)
	c.Assert(Props.IsWebHookEnabled(), check.Equals, false)
	c.Assert(Props.AWSAccountID(), check.Equals, "123456789012")
	c.Assert(strings.HasPrefix(Props.ManagedPermissionBoundaryPolicy(), "arn:aws:iam:"), check.Equals, true)
	//when an emty string passed split strings gives you array of 1 with ""
	c.Assert(len(Props.ManagedPolicies()), check.Equals, 1)
	c.Assert(Props.ManagedPolicies()[0], check.Equals, "arn:aws:iam::123456789012:policy/DescribeEC2")

}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithCustom(c *check.C) {
	Props = nil
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.managed.permission.boundary.policy": "iam-manager-permission-boundary",
			"aws.accountId":                          "123456789012",
			"iam.role.derive.from.namespace":         "true",
			"controller.desired.frequency":           "30",
			"iam.role.max.limit.per.namespace":       "5",
			"iam.role.pattern":                       "pfx-{{ .ObjectMeta.Name }}",
		},
	}
	err := LoadProperties("", cm)
	c.Assert(err, check.IsNil)
	c.Assert(Props.MaxRolesAllowed(), check.Equals, 5)
	c.Assert(Props.ControllerDesiredFrequency(), check.Equals, 30)
	c.Assert(Props.IamRolePattern(), check.Equals, "pfx-{{ .ObjectMeta.Name }}")
}

func (s *PropertiesSuite) TestGetAllowedPolicyAction(c *check.C) {
	value := Props.AllowedPolicyAction()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetRestrictedPolicyResources(c *check.C) {
	value := Props.RestrictedPolicyResources()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetRestrictedS3Resources(c *check.C) {
	value := Props.RestrictedS3Resources()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetManagedPolicies(c *check.C) {
	value := Props.ManagedPolicies()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetAWSAccountID(c *check.C) {
	value := Props.AWSAccountID()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetAWSRegion(c *check.C) {
	value := Props.AWSRegion()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetManagedPermissionBoundaryPolicy(c *check.C) {
	value := Props.ManagedPermissionBoundaryPolicy()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestIsWebhookEnabled(c *check.C) {
	value := Props.IsWebHookEnabled()
	c.Assert(value, check.Equals, false)
}

func (s *PropertiesSuite) TestControllerDesiredFrequency(c *check.C) {
	value := Props.ControllerDesiredFrequency()
	c.Assert(value, check.Equals, 0)
}

func (s *PropertiesSuite) TestIsIRSAEnabled(c *check.C) {
	value := Props.IsIRSAEnabled()
	c.Assert(value, check.Equals, false)
}

func (s *PropertiesSuite) TestControllerClusterName(c *check.C) {
	value := Props.ClusterName()
	c.Assert(value, check.Equals, "k8s_test_keiko")
}

func (s *PropertiesSuite) TestControllerOIDCIssuerUrl(c *check.C) {
	value := Props.OIDCIssuerUrl()
	c.Assert(value, check.Equals, "https://google.com/OIDC")
}

func (s *PropertiesSuite) TestControllerDefaultTrustPolicy(c *check.C) {
	def := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:SERVICE_ACCOUNT_NAME"}}}, {"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`
	value := Props.DefaultTrustPolicy()
	c.Assert(value, check.Equals, def)
}
