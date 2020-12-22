package utils_test

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"gopkg.in/check.v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type UtilsTestSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestUtilsTestSuite(t *testing.T) {
	check.Suite(&UtilsTestSuite{t: t})
	check.TestingT(t)
}

func (s *UtilsTestSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *UtilsTestSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *UtilsTestSuite) TestDefaultTrustPolicyNoGoTemplate(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::123456789012:role/trust_role"]},"Action": "sts:AssumeRole"}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
		},
	}
	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(*resp, check.DeepEquals, expect)

}

func (s *UtilsTestSuite) TestDefaultTrustPolicyEmptyString(c *check.C) {
	tD := ""
	_, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.NotNil)

}

func (s *UtilsTestSuite) TestDefaultTrustPolicyInvalidJsonString(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": ["Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`
	_, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.NotNil)
}

func (s *UtilsTestSuite) TestDefaultTrustPolicyUnknownGoTemplateValue(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": ["Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountI}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`
	_, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.NotNil)
}

func (s *UtilsTestSuite) TestDefaultTrustPolicyWithGoTemplate(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
		},
	}
	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(*resp, check.DeepEquals, expect)

}
func (s *UtilsTestSuite) TestDefaultTrustPolicyMultipleNoGoTemplate(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::123456789012:role/trust_role"]},"Action": "sts:AssumeRole"}, {"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:SERVICE_ACCOUNT_NAMESPACE:SERVICE_ACCOUNT_NAME"}}}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER",
				},
				Condition: &v1alpha1.Condition{
					StringEquals: map[string]string{
						"OIDC_PROVIDER:sub": "system:serviceaccount:SERVICE_ACCOUNT_NAMESPACE:SERVICE_ACCOUNT_NAME",
					},
				},
			},
		},
	}
	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(resp.Statement[0], check.DeepEquals, expect.Statement[0])
	c.Assert(len(resp.Statement), check.Equals, len(expect.Statement))
	c.Assert(*resp.Statement[1].Condition, check.DeepEquals, *expect.Statement[1].Condition)
}

func (s *UtilsTestSuite) TestDefaultTrustPolicyMultipleWithGoTemplate(c *check.C) {
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::123456789012:role/trust_role"]},"Action": "sts:AssumeRole"}, {"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:SERVICE_ACCOUNT_NAME"}}}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER",
				},
				Condition: &v1alpha1.Condition{
					StringEquals: map[string]string{
						"OIDC_PROVIDER:sub": "system:serviceaccount:valid_namespace:SERVICE_ACCOUNT_NAME",
					},
				},
			},
		},
	}
	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(*resp, check.DeepEquals, expect)

}

func (s *UtilsTestSuite) TestGetTrustPolicyDefaultRoleWithMultiple(c *check.C) {
	//Add Env variable
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER",
				},
				Condition: &v1alpha1.Condition{
					StringEquals: map[string]string{
						"OIDC_PROVIDER:sub": "system:serviceaccount:valid_namespace:SERVICE_ACCOUNT_NAME",
					},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
		},
	}

	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "valid_namespace",
		},
		Spec: v1alpha1.IamroleSpec{},
	}

	expected, _ := json.Marshal(expect)
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyDefaultRoleWithMultipleAndStringLikeWithNoGoTemplate(c *check.C) {
	//Add Env variable
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::123456789012:role/trust_role"]},"Action": "sts:AssumeRole"}, {"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringLike": {"OIDC_PROVIDER:sub": "system:serviceaccount:SERVICE_ACCOUNT_NAMESPACE:*"}}}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER",
				},
				Condition: &v1alpha1.Condition{
					StringLike: map[string]string{
						"OIDC_PROVIDER:sub": "system:serviceaccount:SERVICE_ACCOUNT_NAMESPACE:*",
					},
				},
			},
		},
	}

	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(*resp, check.DeepEquals, expect)
}

func (s *UtilsTestSuite) TestGetTrustPolicyDefaultRoleWithMultipleAndStringLikeWithGoTemplate(c *check.C) {
	//Add Env variable
	tD := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}, {"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringLike": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:*"}}}]}`
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER",
				},
				Condition: &v1alpha1.Condition{
					StringLike: map[string]string{
						"OIDC_PROVIDER:sub": "system:serviceaccount:valid_namespace:*",
					},
				},
			},
		},
	}

	resp, err := utils.DefaultTrustPolicy(s.ctx, tD, "valid_namespace")
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(*resp, check.DeepEquals, expect)
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRoleSuccess(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)
	tPolicy := v1alpha1.TrustPolicyStatement{
		Effect: "Allow",
		Action: "sts:AssumeRole",
		Principal: v1alpha1.Principal{
			AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
		},
	}
	input := &v1alpha1.Iamrole{
		Spec: v1alpha1.IamroleSpec{
			AssumeRolePolicyDocument: &v1alpha1.AssumeRolePolicyDocument{
				Version: "2012-10-17",
				Statement: []v1alpha1.TrustPolicyStatement{
					tPolicy,
				},
			},
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRolesSuccess(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)
	tPolicy := v1alpha1.TrustPolicyStatement{
		Effect: "Allow",
		Action: "sts:AssumeRole",
		Principal: v1alpha1.Principal{
			AWS: []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
		},
	}

	input := &v1alpha1.Iamrole{
		Spec: v1alpha1.IamroleSpec{
			AssumeRolePolicyDocument: &v1alpha1.AssumeRolePolicyDocument{
				Version: "2012-10-17",
				Statement: []v1alpha1.TrustPolicyStatement{
					tPolicy,
				},
			},
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyServiceRoleSuccess(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					Service: "ec2.amazonaws.com",
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)
	tPolicy := v1alpha1.TrustPolicyStatement{
		Effect: "Allow",
		Action: "sts:AssumeRole",
		Principal: v1alpha1.Principal{
			Service: "ec2.amazonaws.com",
		},
	}

	input := &v1alpha1.Iamrole{
		Spec: v1alpha1.IamroleSpec{
			AssumeRolePolicyDocument: &v1alpha1.AssumeRolePolicyDocument{
				Version: "2012-10-17",
				Statement: []v1alpha1.TrustPolicyStatement{
					tPolicy,
				},
			},
		},
	}

	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRolesAndServiceRoleSuccess(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS:     []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
					Service: "ec2.amazonaws.com",
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)
	tPolicy := v1alpha1.TrustPolicyStatement{
		Effect: "Allow",
		Action: "sts:AssumeRole",
		Principal: v1alpha1.Principal{
			AWS:     []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
			Service: "ec2.amazonaws.com",
		},
	}

	input := &v1alpha1.Iamrole{
		Spec: v1alpha1.IamroleSpec{
			AssumeRolePolicyDocument: &v1alpha1.AssumeRolePolicyDocument{
				Version: "2012-10-17",
				Statement: []v1alpha1.TrustPolicyStatement{
					tPolicy,
				},
			},
		},
	}

	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyWithIRSAAnnotation(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/google.com/OIDC",
				},
				Condition: &v1alpha1.Condition{
					StringEquals: map[string]string{
						"google.com/OIDC:sub": "system:serviceaccount:k8s-namespace-dev:default",
					},
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)

	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name:      "iam-role",
			Namespace: "k8s-namespace-dev",
			Annotations: map[string]string{
				config.IRSAAnnotation: "default",
			},
		},
	}

	roleString, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(roleString, check.Equals, string(expected))

}

func (s *UtilsTestSuite) TestGetTrustPolicyWithIRSAAnnotationAndServiceRoleInRequest(c *check.C) {
	expect := v1alpha1.AssumeRolePolicyDocument{
		Version: "2012-10-17",
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRoleWithWebIdentity",
				Principal: v1alpha1.Principal{
					Federated: "arn:aws:iam::123456789012:oidc-provider/google.com/OIDC",
				},
				Condition: &v1alpha1.Condition{
					StringEquals: map[string]string{
						"google.com/OIDC:sub": "system:serviceaccount:k8s-namespace-dev:default",
					},
				},
			},
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					Service: "ec2.amazonaws.com",
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)

	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name:      "iam-role",
			Namespace: "k8s-namespace-dev",
			Annotations: map[string]string{
				config.IRSAAnnotation: "default",
			},
		},
		Spec: v1alpha1.IamroleSpec{
			AssumeRolePolicyDocument: &v1alpha1.AssumeRolePolicyDocument{
				Version: "2012-10-17",
				Statement: []v1alpha1.TrustPolicyStatement{
					{
						Effect: "Allow",
						Action: "sts:AssumeRole",
						Principal: v1alpha1.Principal{
							Service: "ec2.amazonaws.com",
						},
					},
				},
			},
		},
	}

	roleString, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(roleString, check.Equals, string(expected))

}

func (s *UtilsTestSuite) TestGenerateNameFunction(c *check.C) {
	config.Props = nil
	cm := &v12.ConfigMap{
		Data: map[string]string{
			"iam.role.derive.from.namespace": "false",
			"iam.role.prefix":                "pfx",
			"iam.role.separator":             "+",
		},
	}
	err := config.LoadProperties("", cm)
	c.Assert(err, check.Equals, nil)

	resource := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-ns",
		},
	}
	c.Assert(utils.GenerateRoleName(*resource, *config.Props), check.Equals, "pfx+foo")
}

func (s *UtilsTestSuite) TestGenerateNameFunctionWithDeriveFromNamespaceEnabled(c *check.C) {
	config.Props = nil
	cm := &v12.ConfigMap{
		Data: map[string]string{
			"aws.accountId":                  "123456789012", // Required mock for testing
			"iam.role.derive.from.namespace": "true",
			"iam.role.prefix":                "pfx",
			"iam.role.separator":             "+",
		},
	}
	err := config.LoadProperties("", cm)
	c.Assert(err, check.Equals, nil)

	resource := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-ns",
		},
	}
	c.Assert(utils.GenerateRoleName(*resource, *config.Props), check.Equals, "pfx+test-ns")
}
