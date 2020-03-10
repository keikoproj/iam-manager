package utils_test

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/utils"
	"gopkg.in/check.v1"
	"testing"
)

type UtilsTestSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestValidateSuite(t *testing.T) {
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
	expect := &utils.TrustPolicy{
		Version: "2012-10-17",
		Statement: []utils.Statement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/trust_role"},
				},
			},
		},
	}

	expected, _ := json.Marshal(expect)
	resp, err := utils.GetTrustPolicy(s.ctx, nil)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRoleSuccess(c *check.C) {
	expect := &utils.TrustPolicy{
		Version: "2012-10-17",
		Statement: []utils.Statement{
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
	input := &v1alpha1.TrustPolicy{
		Principal: v1alpha1.Principal{
			AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRolesSuccess(c *check.C) {
	expect := &utils.TrustPolicy{
		Version: "2012-10-17",
		Statement: []utils.Statement{
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
	input := &v1alpha1.TrustPolicy{
		Principal: v1alpha1.Principal{
			AWS: []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyServiceRoleSuccess(c *check.C) {
	expect := &utils.TrustPolicy{
		Version: "2012-10-17",
		Statement: []utils.Statement{
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
	input := &v1alpha1.TrustPolicy{
		Principal: v1alpha1.Principal{
			Service: "ec2.amazonaws.com",
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyAWSRolesAndServiceRoleSuccess(c *check.C) {
	expect := &utils.TrustPolicy{
		Version: "2012-10-17",
		Statement: []utils.Statement{
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
	input := &v1alpha1.TrustPolicy{
		Principal: v1alpha1.Principal{
			AWS:     []string{"arn:aws:iam::123456789012:role/user_request_role1", "arn:aws:iam::123456789012:role/user_request_role2"},
			Service: "ec2.amazonaws.com",
		},
	}
	resp, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.DeepEquals, string(expected))
}

func (s *UtilsTestSuite) TestGetTrustPolicyServiceRoleInvalidName(c *check.C) {

	input := &v1alpha1.TrustPolicy{
		Principal: v1alpha1.Principal{
			Service: "ec2.amazonws.com",
		},
	}
	_, err := utils.GetTrustPolicy(s.ctx, input)
	c.Assert(err, check.NotNil)
}
