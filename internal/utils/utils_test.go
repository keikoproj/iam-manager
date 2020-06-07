package utils_test

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"gopkg.in/check.v1"
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

func (s *UtilsTestSuite) TestGetTrustPolicyDefaultRole(c *check.C) {
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

	input := &v1alpha1.Iamrole{
		Spec: v1alpha1.IamroleSpec{},
	}

	expected, _ := json.Marshal(expect)
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
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
