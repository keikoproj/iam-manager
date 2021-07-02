package awsapi_test

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
	"gopkg.in/check.v1"
	"testing"
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
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(&iam.CreateRoleOutput{Role: &iam.Role{RoleId: aws.String("ABCDE1234"), Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE")}}, nil)
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("VALID_ROLE"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(&iam.TagRoleOutput{}, nil)
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.AttachRolePolicyOutput{}, nil)
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(&iam.UpdateRoleOutput{}, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyDocument: aws.String("SOMETHING")}).Times(1).Return(&iam.UpdateAssumeRolePolicyOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: config.Props.ManagedPolicies(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
	c.Assert(resp, check.NotNil)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestEnsureRoleSuccessWithNoManagedPolicies(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(&iam.CreateRoleOutput{Role: &iam.Role{RoleId: aws.String("ABCDE1234"), Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE")}}, nil)
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("VALID_ROLE"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(&iam.TagRoleOutput{}, nil)
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)
	//s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.AttachRolePolicyOutput{}, nil)
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(&iam.UpdateRoleOutput{}, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyDocument: aws.String("SOMETHING")}).Times(1).Return(&iam.UpdateAssumeRolePolicyOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: []string{""}, Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	resp, err := s.mockIAM.EnsureRole(s.ctx, req)
	c.Assert(resp.RoleARN, check.Equals, "arn:aws:iam::123456789012:role/VALID_ROLE")
	c.Assert(resp.RoleID, check.Equals, "ABCDE1234")
	c.Assert(resp, check.NotNil)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestEnsureRoleFailsIfGetRoleAndCreateRoleConflict(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeEntityAlreadyExistsException, "", errors.New(iam.ErrCodeEntityAlreadyExistsException)))
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: config.Props.ManagedPolicies(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.EnsureRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestEnsureRoleWithRoleOwnedByOtherNamespace(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(&iam.CreateRoleOutput{Role: &iam.Role{RoleId: aws.String("ABCDE1234"), Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE")}}, nil)
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
			{
				Key:   aws.String("Namespace"),
				Value: aws.String("namespace_name"),
			},
		},
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: config.Props.ManagedPolicies(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.EnsureRole(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*"+awsapi.RoleExistsAlreadyForOtherNamespace)
}

//###########
func (s *IAMAPISuite) TestVerifyTagsSuccess(c *check.C) {
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
		},
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.VerifyTags(s.ctx, req)
	c.Assert(err, check.IsNil)

}

func (s *IAMAPISuite) TestVerifyTagsDifferentClusters(c *check.C) {
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
			{
				Key:   aws.String("Cluster"),
				Value: aws.String("different-cluster"),
			},
		},
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
		"Cluster":   "cluster_name",
	}}
	_, err := s.mockIAM.VerifyTags(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*"+awsapi.RoleExistsAlreadyForOtherNamespace)

}

func (s *IAMAPISuite) TestVerifyTagsDifferentNamespace(c *check.C) {
	s.mockI.EXPECT().ListRoleTags(&iam.ListRoleTagsInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRoleTagsOutput{
		Tags: []*iam.Tag{
			{
				Key:   aws.String("managedBy"),
				Value: aws.String("iam-manager"),
			},
			{
				Key:   aws.String("Namespace"),
				Value: aws.String("different-namespace"),
			},
		},
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
		"Cluster":   "namespace_name",
	}}
	_, err := s.mockIAM.VerifyTags(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*"+awsapi.RoleExistsAlreadyForOtherNamespace)

}

func (s *IAMAPISuite) TestTagRoleSuccess(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("VALID_ROLE"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(&iam.TagRoleOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestTagRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("NO_SUCH_ENTITY"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestTagRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("SERVICE_FAILURE"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestTagRoleFailureInvalidInput(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("INVALID_INPUT"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	req := awsapi.IAMRoleRequest{Name: "INVALID_INPUT", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestTagRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("LIMIT_EXCEEDED"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "LIMIT_EXCEEDED", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestTagRoleFailureUnattachable(c *check.C) {
	s.mockI.EXPECT().TagRole(&iam.TagRoleInput{RoleName: aws.String("UN_ATTACHABLE"), Tags: []*iam.Tag{
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	req := awsapi.IAMRoleRequest{Name: "UN_ATTACHABLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.TagRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//###########

func (s *IAMAPISuite) TestAddPermissionBoundarySuccess(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("MALFORMED_POLICY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("TOO_MANY_REQUEST"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("NO_SUCH_ENTITY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestAddPermissionBoundaryFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("SERVICE_FAILURE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("UNMODIFIABLE_POLICY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailurePolicyNotAttachable(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("POLICY_NOT_ATTACHABLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	req := awsapi.IAMRoleRequest{Name: "POLICY_NOT_ATTACHABLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAddPermissionBoundaryFailureInvalidInput(c *check.C) {
	s.mockI.EXPECT().PutRolePermissionsBoundary(&iam.PutRolePermissionsBoundaryInput{RoleName: aws.String("INVALID_INPUT"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy())}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	req := awsapi.IAMRoleRequest{Name: "INVALID_INPUT", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	err := s.mockIAM.AddPermissionBoundary(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//###########

func (s *IAMAPISuite) TestUpdateRoleSuccess(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.UpdateAssumeRolePolicyOutput{}, nil)
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestUpdateRoleInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("MALFORMED_POLICY"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("TOO_MANY_REQUEST"),
		MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("NO_SUCH_ENTITY"),
		MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestUpdateRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("SERVICE_FAILURE"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("UNMODIFIABLE_POLICY"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailurePolicyNotAttachable(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("POLICY_NOT_ATTACHABLE"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	req := awsapi.IAMRoleRequest{Name: "POLICY_NOT_ATTACHABLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleFailureInvalidInput(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("INVALID_INPUT"),
		MaxSessionDuration: aws.Int64(3600),
		Description:        aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	req := awsapi.IAMRoleRequest{Name: "INVALID_INPUT", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleAssumeRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleAssumeRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleAssumeRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestUpdateRoleAssumeRoleFailureInvalidInput(c *check.C) {
	s.mockI.EXPECT().UpdateRole(&iam.UpdateRoleInput{RoleName: aws.String("VALID_ROLE"), MaxSessionDuration: aws.Int64(3600), Description: aws.String("")}).Times(1).Return(nil, nil)
	s.mockI.EXPECT().UpdateAssumeRolePolicy(&iam.UpdateAssumeRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING"}
	_, err := s.mockIAM.UpdateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//####################

func (s *IAMAPISuite) TestAttachInlineRolePolicySuccess(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.PutRolePolicyOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestAttachInlineRolePolicyInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachInlineRolePolicyFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("MALFORMED_POLICY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachInlineRolePolicyFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("TOO_MANY_REQUEST"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachInlineRolePolicyFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("NO_SUCH_ENTITY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestAttachInlineRolePolicyFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("SERVICE_FAILURE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachInlineRolePolicyFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().PutRolePolicy(&iam.PutRolePolicyInput{PolicyDocument: aws.String("SOMETHING"), RoleName: aws.String("UNMODIFIABLE_POLICY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.AttachInlineRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//###########

func (s *IAMAPISuite) TestGetRolePolicySuccess(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestGetRolePolicyInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRolePolicyFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("MALFORMED_POLICY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRolePolicyFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("TOO_MANY_REQUEST"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRolePolicyFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("NO_SUCH_ENTITY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestGetRolePolicyFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("SERVICE_FAILURE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRolePolicyFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("UNMODIFIABLE_POLICY"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRolePolicy(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//###########

func (s *IAMAPISuite) TestGetOrCreateRoleSuccessNewRole(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("VALID_ROLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(&iam.CreateRoleOutput{Role: &iam.Role{RoleId: aws.String("ABCDE1234"), Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE")}}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: config.Props.ManagedPolicies(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.GetOrCreateRole(s.ctx, req)
	//_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestGetOrCreateRoleSuccessExistsRole(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(
		&iam.GetRoleOutput{Role: &iam.Role{RoleId: aws.String("ABCDE1234"), Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE")}}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(), ManagedPolicies: config.Props.ManagedPolicies(), Tags: map[string]string{
		"managedBy": "iam-manager",
	}}
	_, err := s.mockIAM.GetOrCreateRole(s.ctx, req)
	c.Assert(err, check.IsNil)
}

//###########

func (s *IAMAPISuite) TestCreateRoleInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("MALFORMED_POLICY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("TOO_MANY_REQUEST"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("NO_SUCH_ENTITY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestCreateRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("SERVICE_FAILURE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("UNMODIFIABLE_POLICY"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailurePolicyNotAttachable(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("POLICY_NOT_ATTACHABLE"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	req := awsapi.IAMRoleRequest{Name: "POLICY_NOT_ATTACHABLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateRoleFailureInvalidInput(c *check.C) {
	s.mockI.EXPECT().CreateRole(&iam.CreateRoleInput{RoleName: aws.String("INVALID_INPUT"), PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()), MaxSessionDuration: aws.Int64(3600), AssumeRolePolicyDocument: aws.String("SOMETHING"), Description: aws.String("")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	req := awsapi.IAMRoleRequest{Name: "INVALID_INPUT", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING", SessionDuration: 3600, TrustPolicy: "SOMETHING", ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy()}
	_, err := s.mockIAM.CreateRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//###########

func (s *IAMAPISuite) TestGetRoleSuccess(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{}, nil)
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestGetRoleInvalidRequest(c *check.C) {
	req := awsapi.IAMRoleRequest{PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRoleFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	req := awsapi.IAMRoleRequest{Name: "MALFORMED_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	req := awsapi.IAMRoleRequest{Name: "TOO_MANY_REQUEST", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	req := awsapi.IAMRoleRequest{Name: "NO_SUCH_ENTITY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestGetRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	req := awsapi.IAMRoleRequest{Name: "SERVICE_FAILURE", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestGetRoleFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	req := awsapi.IAMRoleRequest{Name: "UNMODIFIABLE_POLICY", PolicyName: "VALID_POLICY", PermissionPolicy: "SOMETHING"}
	_, err := s.mockIAM.GetRole(s.ctx, req)
	c.Assert(err, check.NotNil)
}

//############################

func (s *IAMAPISuite) TestAttachManagedRolePolicySuccess(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.AttachRolePolicyOutput{}, nil)
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "VALID_ROLE")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "MALFORMED_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "TOO_MANY_REQUEST")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "NO_SUCH_ENTITY")
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "SERVICE_FAILURE")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "UNMODIFIABLE_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureInvalidPolicyDocument(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "INVALID_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestAttachManagedRolePolicyFailureUnattachablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().AttachRolePolicy(&iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	err := s.mockIAM.AttachManagedRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "UNATTACHABLE_POLICY")
	c.Assert(err, check.NotNil)
}

//############################

func (s *IAMAPISuite) TestDeleteRoleSuccess(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.DeleteRoleOutput{}, nil)
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "VALID_ROLE")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)
	err := s.mockIAM.DeleteRole(s.ctx, "MALFORMED_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "TOO_MANY_REQUEST")
	c.Assert(err, check.NotNil)
}
func (s *IAMAPISuite) TestDeleteRoleFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "NO_SUCH_ENTITY")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureNoSuchEntityAssumeRole(c *check.C) {
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))

	err := s.mockIAM.DeleteRole(s.ctx, "NO_SUCH_ENTITY")
	c.Assert(err, check.IsNil)
}
func (s *IAMAPISuite) TestDeleteRoleFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "SERVICE_FAILURE")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "UNMODIFIABLE_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureInvalidPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "INVALID_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteRoleFailureUnattachablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRole(&iam.DeleteRoleInput{RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	s.mockI.EXPECT().ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(&iam.ListAttachedRolePoliciesOutput{}, nil)
	s.mockI.EXPECT().ListRolePolicies(&iam.ListRolePoliciesInput{RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(&iam.ListRolePoliciesOutput{}, nil)

	err := s.mockIAM.DeleteRole(s.ctx, "UNATTACHABLE_POLICY")
	c.Assert(err, check.NotNil)
}

//############################

func (s *IAMAPISuite) TestDeleteInlinePolicySuccess(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.DeleteRolePolicyOutput{}, nil)
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "VALID_ROLE")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "MALFORMED_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "TOO_MANY_REQUEST")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "NO_SUCH_ENTITY")
	c.Assert(err, check.IsNil)
}
func (s *IAMAPISuite) TestDeleteInlinePolicyFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "SERVICE_FAILURE")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "UNMODIFIABLE_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureInvalidPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "INVALID_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDeleteInlinePolicyFailureUnattachablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DeleteRolePolicy(&iam.DeleteRolePolicyInput{PolicyName: aws.String("SOMETHING"), RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	err := s.mockIAM.DeleteInlinePolicy(s.ctx, "SOMETHING", "UNATTACHABLE_POLICY")
	c.Assert(err, check.NotNil)
}

//#############

func (s *IAMAPISuite) TestDetachRolePolicySuccess(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.DetachRolePolicyOutput{}, nil)
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "VALID_ROLE")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureMalformedPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("MALFORMED_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeMalformedPolicyDocumentException, "", errors.New(iam.ErrCodeMalformedPolicyDocumentException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "MALFORMED_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureLimitExceeded(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("TOO_MANY_REQUEST")}).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "TOO_MANY_REQUEST")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureNoSuchEntity(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("NO_SUCH_ENTITY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeNoSuchEntityException, "", errors.New(iam.ErrCodeNoSuchEntityException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "NO_SUCH_ENTITY")
	c.Assert(err, check.IsNil)
}
func (s *IAMAPISuite) TestDetachRolePolicyFailureServiceFailure(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("SERVICE_FAILURE")}).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "SERVICE_FAILURE")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureUnmodififiablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("UNMODIFIABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeUnmodifiableEntityException, "", errors.New(iam.ErrCodeUnmodifiableEntityException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "UNMODIFIABLE_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureInvalidPolicyDocument(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("INVALID_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "INVALID_POLICY")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestDetachRolePolicyFailureUnattachablePolicyDocument(c *check.C) {
	s.mockI.EXPECT().DetachRolePolicy(&iam.DetachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::123456789012:policy/SOMETHING"), RoleName: aws.String("UNATTACHABLE_POLICY")}).Times(1).Return(nil, awserr.New(iam.ErrCodePolicyNotAttachableException, "", errors.New(iam.ErrCodePolicyNotAttachableException)))
	err := s.mockIAM.DetachRolePolicy(s.ctx, "arn:aws:iam::123456789012:policy/SOMETHING", "UNATTACHABLE_POLICY")
	c.Assert(err, check.NotNil)
}

// ################

func (s *IAMAPISuite) TestCreateOIDCProviderSuccess(c *check.C) {
	input := &iam.CreateOpenIDConnectProviderInput{
		ThumbprintList: []*string{aws.String("valid_thumbprint")},
		Url:            aws.String("https://server.example.com"),
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
	}
	s.mockI.EXPECT().CreateOpenIDConnectProvider(input).Times(1).Return(&iam.CreateOpenIDConnectProviderOutput{
		OpenIDConnectProviderArn: aws.String("valid_arn"),
	}, nil)

	err := s.mockIAM.CreateOIDCProvider(s.ctx, "https://server.example.com", config.OIDCAudience, "valid_thumbprint")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestCreateOIDCProviderInvalidInput(c *check.C) {
	input := &iam.CreateOpenIDConnectProviderInput{
		ThumbprintList: []*string{aws.String("invalid_thumbprint")},
		Url:            aws.String("https://server.example.com"),
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
	}
	s.mockI.EXPECT().CreateOpenIDConnectProvider(input).Times(1).Return(nil, awserr.New(iam.ErrCodeInvalidInputException, "", errors.New(iam.ErrCodeInvalidInputException)))

	err := s.mockIAM.CreateOIDCProvider(s.ctx, "https://server.example.com", config.OIDCAudience, "invalid_thumbprint")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateOIDCProviderAlreadyExists(c *check.C) {
	input := &iam.CreateOpenIDConnectProviderInput{
		ThumbprintList: []*string{aws.String("already_exists_thumbprint")},
		Url:            aws.String("https://server.example.com"),
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
	}
	s.mockI.EXPECT().CreateOpenIDConnectProvider(input).Times(1).Return(nil, awserr.New(iam.ErrCodeEntityAlreadyExistsException, "", errors.New(iam.ErrCodeEntityAlreadyExistsException)))

	err := s.mockIAM.CreateOIDCProvider(s.ctx, "https://server.example.com", config.OIDCAudience, "already_exists_thumbprint")
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestCreateOIDCProviderLimitExceeded(c *check.C) {
	input := &iam.CreateOpenIDConnectProviderInput{
		ThumbprintList: []*string{aws.String("limit_exceeded_thumbprint")},
		Url:            aws.String("https://server.example.com"),
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
	}
	s.mockI.EXPECT().CreateOpenIDConnectProvider(input).Times(1).Return(nil, awserr.New(iam.ErrCodeLimitExceededException, "", errors.New(iam.ErrCodeLimitExceededException)))

	err := s.mockIAM.CreateOIDCProvider(s.ctx, "https://server.example.com", config.OIDCAudience, "limit_exceeded_thumbprint")
	c.Assert(err, check.NotNil)
}

func (s *IAMAPISuite) TestCreateOIDCProviderServiceFailure(c *check.C) {
	input := &iam.CreateOpenIDConnectProviderInput{
		ThumbprintList: []*string{aws.String("failure_thumbprint")},
		Url:            aws.String("https://server.example.com"),
		ClientIDList:   []*string{aws.String("sts.amazonaws.com")},
	}
	s.mockI.EXPECT().CreateOpenIDConnectProvider(input).Times(1).Return(nil, awserr.New(iam.ErrCodeServiceFailureException, "", errors.New(iam.ErrCodeServiceFailureException)))

	err := s.mockIAM.CreateOIDCProvider(s.ctx, "https://server.example.com", config.OIDCAudience, "failure_thumbprint")
	c.Assert(err, check.NotNil)
}
