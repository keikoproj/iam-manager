package awsapi_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"

	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
)

type IAMAPISuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
	mockI    *mock_awsapi.MockIAMAPI
	mockIAM  awsapi.IAM
}

func TestIAMAPITestSuite(t *testing.T) {
	check.Suite(&IAMAPISuite{t: t})
	check.TestingT(t)
}

func (s *IAMAPISuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockI = mock_awsapi.NewMockIAMAPI(s.mockCtrl)
	s.mockIAM = awsapi.IAM{
		Client: s.mockI,
	}

	_ = config.LoadProperties("LOCAL")
}

func (s *IAMAPISuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

//############

func (s *IAMAPISuite) TestEnsureRoleSuccess(c *check.C) {
	// Use a properly formatted ARN with sufficient length for AWS SDK validation
	validPolicyArn := "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary"
	
	// First, it will try to get the role and fail with NoSuchEntity
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	
	// Then it will try to create the role
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{
		RoleName:               aws.String("VALID_ROLE"),
		PermissionsBoundary:    aws.String(validPolicyArn),
		MaxSessionDuration:     aws.Int64(3600),
		AssumeRolePolicyDocument: aws.String("SOMETHING"),
		Description:            aws.String(""),
	}).Times(1).Return(&iam.CreateRoleOutput{
		Role: &iam.Role{
			RoleId: aws.String("ABCDE1234"),
			Arn:    aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	
	// Then list role tags
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)
	
	// Then tag the role
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{
		RoleName: aws.String("VALID_ROLE"),
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}).Times(1).Return(&iam.TagRoleOutput{}, nil)
	
	// Add permission boundary
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String("VALID_ROLE"),
		PermissionsBoundary: aws.String(validPolicyArn),
	}).Times(1).Return(&iam.PutRolePermissionsBoundaryOutput{}, nil)
	
	// Put role policy
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyDocument: aws.String("SOMETHING"),
		RoleName:       aws.String("VALID_ROLE"),
		PolicyName:     aws.String("VALID_POLICY"),
	}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)
	
	// Attach role policy
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"),
		RoleName:  aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.AttachRolePolicyOutput{}, nil)
	
	// Update role
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{
		RoleName:           aws.String("VALID_ROLE"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String(""),
	}).Times(1).Return(&iam.UpdateRoleOutput{}, nil)
	
	// Update assume role policy
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String("VALID_ROLE"),
		PolicyDocument: aws.String("SOMETHING"),
	}).Times(1).Return(&iam.UpdateAssumeRolePolicyOutput{}, nil)
	
	// Create request
	req := awsapi.IAMRoleRequest{
		Name:                          "VALID_ROLE",
		PolicyName:                    "VALID_POLICY",
		PermissionPolicy:              "SOMETHING",
		SessionDuration:               3600,
		TrustPolicy:                   "SOMETHING",
		ManagedPermissionBoundaryPolicy: validPolicyArn,
		ManagedPolicies:               []string{"arn:aws:iam::123456789012:policy/SOMETHING"},
		Tags: map[string]string{
			"managedBy": "iam-manager",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	
	// Verify the results
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
}

func (s *IAMAPISuite) TestEnsureRoleSuccessWithNoManagedPolicies(c *check.C) {
	// Use a properly formatted ARN with sufficient length for AWS SDK validation
	validPolicyArn := "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary"
	
	// First, it will try to get the role and fail with NoSuchEntity
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	
	// Then it will try to create the role
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{
		RoleName:               aws.String("VALID_ROLE"),
		PermissionsBoundary:    aws.String(validPolicyArn),
		MaxSessionDuration:     aws.Int64(3600),
		AssumeRolePolicyDocument: aws.String("SOMETHING"),
		Description:            aws.String(""),
	}).Times(1).Return(&iam.CreateRoleOutput{
		Role: &iam.Role{
			RoleId: aws.String("ABCDE1234"),
			Arn:    aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	
	// Then list role tags
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)
	
	// Then tag the role
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{
		RoleName: aws.String("VALID_ROLE"),
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}).Times(1).Return(&iam.TagRoleOutput{}, nil)
	
	// Add permission boundary
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String("VALID_ROLE"),
		PermissionsBoundary: aws.String(validPolicyArn),
	}).Times(1).Return(&iam.PutRolePermissionsBoundaryOutput{}, nil)
	
	// Put role policy
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{
		PolicyDocument: aws.String("SOMETHING"),
		RoleName:       aws.String("VALID_ROLE"),
		PolicyName:     aws.String("VALID_POLICY"),
	}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)
	
	// Note: No AttachRolePolicy call because there are no managed policies
	
	// Update role
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{
		RoleName:           aws.String("VALID_ROLE"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String(""),
	}).Times(1).Return(&iam.UpdateRoleOutput{}, nil)
	
	// Update assume role policy
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String("VALID_ROLE"),
		PolicyDocument: aws.String("SOMETHING"),
	}).Times(1).Return(&iam.UpdateAssumeRolePolicyOutput{}, nil)
	
	// Create request with empty managed policies array
	req := awsapi.IAMRoleRequest{
		Name:                          "VALID_ROLE",
		PolicyName:                    "VALID_POLICY",
		PermissionPolicy:              "SOMETHING",
		SessionDuration:               3600,
		TrustPolicy:                   "SOMETHING",
		ManagedPermissionBoundaryPolicy: validPolicyArn,
		ManagedPolicies:               []string{""},
		Tags: map[string]string{
			"managedBy": "iam-manager",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	
	// Verify the results
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
}

func (s *IAMAPISuite) TestEnsureRoleFailsIfGetRoleAndCreateRoleConflict(c *check.C) {
	// Use a properly formatted ARN with sufficient length for AWS SDK validation
	validPolicyArn := "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary"
	
	// First, try to get the role and return NoSuchEntity
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	
	// Then try to create role, but return EntityAlreadyExists
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String("VALID_ROLE"),
		PermissionsBoundary:      aws.String(validPolicyArn),
		MaxSessionDuration:       aws.Int64(3600),
		AssumeRolePolicyDocument: aws.String("SOMETHING"),
		Description:              aws.String(""),
	}).Times(1).Return(nil, awserr.New(iam.ErrCodeEntityAlreadyExistsException, "", errors.New(iam.ErrCodeEntityAlreadyExistsException)))
	
	// Create request with all required fields
	req := awsapi.IAMRoleRequest{
		Name:                            "VALID_ROLE",
		PolicyName:                      "VALID_POLICY",
		PermissionPolicy:                "SOMETHING",
		SessionDuration:                 3600,
		TrustPolicy:                     "SOMETHING",
		ManagedPermissionBoundaryPolicy: validPolicyArn,
		ManagedPolicies:                 []string{"arn:aws:iam::123456789012:policy/SOMETHING"},
		Tags: map[string]string{
			"managedBy": "iam-manager",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	
	// Verify the results - expect an error
	c.Assert(err, check.NotNil)
	c.Assert(resp, check.DeepEquals, &awsapi.IAMRoleResponse{})
}

func (s *IAMAPISuite) TestEnsureRoleWithRoleOwnedByOtherNamespace(c *check.C) {
	// Use a properly formatted ARN with sufficient length for AWS SDK validation
	validPolicyArn := "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary"
	
	// Mock GetRole with a role owned by another namespace
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			RoleId:  aws.String("ABCDE1234"),
			Arn:     aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
			RoleName: aws.String("VALID_ROLE"),
		},
	}, nil)
	
	// Mock ListRoleTags with tags for a different namespace
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("Namespace"),
				Value: aws.String("other-namespace"),
			},
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)
	
	// Create request for this namespace
	req := awsapi.IAMRoleRequest{
		Name:                            "VALID_ROLE",
		PolicyName:                      "VALID_POLICY", 
		PermissionPolicy:                "SOMETHING",
		SessionDuration:                 3600,
		TrustPolicy:                     "SOMETHING",
		ManagedPermissionBoundaryPolicy: validPolicyArn,
		ManagedPolicies:                 []string{"arn:aws:iam::123456789012:policy/SOMETHING"},
		Tags: map[string]string{
			"managedBy": "iam-manager",
			"Namespace": "this-namespace",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	
	// Verify the results - expect an error that includes information about the conflict
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, ".*"+awsapi.RoleExistsAlreadyForOtherNamespace+".*")
	c.Assert(resp, check.DeepEquals, &awsapi.IAMRoleResponse{})
}

func (s *IAMAPISuite) TestGetOrCreateRoleSuccessNewRole(c *check.C) {
	// Use a properly formatted ARN with sufficient length for AWS SDK validation
	validPolicyArn := "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary"
	
	// First, try to get the role and return NoSuchEntity
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	
	// Then create the role
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{
		RoleName:                 aws.String("VALID_ROLE"),
		PermissionsBoundary:      aws.String(validPolicyArn),
		MaxSessionDuration:       aws.Int64(3600),
		AssumeRolePolicyDocument: aws.String("SOMETHING"),
		Description:              aws.String(""),
	}).Times(1).Return(&iam.CreateRoleOutput{
		Role: &iam.Role{
			RoleId: aws.String("ABCDE1234"),
			Arn:    aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	
	// Create request
	req := awsapi.IAMRoleRequest{
		Name:                            "VALID_ROLE",
		PolicyName:                      "VALID_POLICY",
		PermissionPolicy:                "SOMETHING",
		SessionDuration:                 3600,
		TrustPolicy:                     "SOMETHING",
		ManagedPermissionBoundaryPolicy: validPolicyArn,
		ManagedPolicies:                 []string{"arn:aws:iam::123456789012:policy/SOMETHING"},
		Tags: map[string]string{
			"managedBy": "iam-manager",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.GetOrCreateRole(s.ctx, req)
	
	// Verify the results
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
}

func (s *IAMAPISuite) TestGetOrCreateRoleSuccessExistsRole(c *check.C) {
	// Get the role - it exists
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{
		RoleName: aws.String("VALID_ROLE"),
	}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			RoleId: aws.String("ABCDE1234"),
			Arn:    aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	
	// Create request
	req := awsapi.IAMRoleRequest{
		Name:                            "VALID_ROLE",
		PolicyName:                      "VALID_POLICY",
		PermissionPolicy:                "SOMETHING",
		SessionDuration:                 3600,
		TrustPolicy:                     "SOMETHING",
		ManagedPermissionBoundaryPolicy: "arn:aws:iam::123456789012:policy/iam-manager-permission-boundary",
		ManagedPolicies:                 []string{"arn:aws:iam::123456789012:policy/SOMETHING"},
		Tags: map[string]string{
			"managedBy": "iam-manager",
		},
	}
	
	// Call the function being tested
	resp, err := s.mockIAM.GetOrCreateRole(s.ctx, req)
	
	// Verify the results
	c.Assert(err, check.IsNil)
	c.Assert(resp, check.NotNil)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
}

// TestCreateRoleInvalidRequest tests that an invalid role request returns an error
func (s *IAMAPISuite) TestCreateRoleInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}
