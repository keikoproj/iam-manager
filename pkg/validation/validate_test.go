package validation_test

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/pkg/validation"
	"gopkg.in/check.v1"
	"testing"
)

type ValidateSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestValidateSuite(t *testing.T) {
	check.Suite(&ValidateSuite{t: t})
	check.TestingT(t)
}

func (s *ValidateSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *ValidateSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *ValidateSuite) TestValidateIAMPolicyActionS3Success(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"s3:ListBucket"},
				Effect:   "Allow",
				Resource: []string{"arn:aws:s3:::s3-resource"},
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyActionS3RestrictedSuccess(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"s3:*"},
				Effect:   "Allow",
				Resource: []string{"s3-resource"},
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.NotNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyResourceSuccess(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"something-something"},
			},
		},
	}
	err := validation.ValidateIAMPolicyResource(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyResourceFailure(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	err := validation.ValidateIAMPolicyResource(s.ctx, input)
	c.Assert(err, check.NotNil)
}

func (s *ValidateSuite) TestComparePolicySuccess(c *check.C) {
	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	flag := validation.ComparePolicy(s.ctx, string(role1), string(role2))
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestComparePolicy2Success(c *check.C) {
	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Resource: []string{"policy-resource"},
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	flag := validation.ComparePolicy(s.ctx, string(role1), string(role2))
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestComparePolicyFailure(c *check.C) {
	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource56789"},
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	flag := validation.ComparePolicy(s.ctx, string(role1), string(role2))
	c.Assert(flag, check.Equals, false)
}

func (s *ValidateSuite) TestContainsStringSuccess(c *check.C) {
	resp := validation.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")
	c.Assert(resp, check.Equals, true)
}

func (s *ValidateSuite) TestContainsStringFailure(c *check.C) {
	resp := validation.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager2.keikoproj.io")
	c.Assert(resp, check.Equals, false)
}

func (s *ValidateSuite) TestRemoveStringSuccess(c *check.C) {
	resp := validation.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")
	c.Assert(resp, check.DeepEquals, []string{"iamrole.finalizers.iammanager.keikoproj.io"})
}

func (s *ValidateSuite) TestRemoveStringEmptySuccess(c *check.C) {
	resp := validation.RemoveString([]string{"iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")
	c.Assert(len(resp), check.Equals, 0)
}
