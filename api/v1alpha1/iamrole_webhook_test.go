package v1alpha1

import (
	"context"
	"github.com/keikoproj/iam-manager/pkg/k8s/mocks"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/internal/config"
	"gopkg.in/check.v1"
	"testing"
)

type WebhookSuite struct {
	t             *testing.T
	ctx           context.Context
	mockCtrl      *gomock.Controller
	mockk8sClient *mock_client.MockIface
	iamRole       *Iamrole
}

func TestWebhookSuite(t *testing.T) {
	check.Suite(&WebhookSuite{t: t})
	check.TestingT(t)
}

func (s *WebhookSuite) SetUpTest(c *check.C) {
	// Could be done in
	err := config.LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)

	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockk8sClient = mock_client.NewMockIface(s.mockCtrl)

	s.iamRole = &Iamrole{}
	s.iamRole.Default()

	// A very basic policy
	s.iamRole.Spec.PolicyDocument.Statement = []Statement{
		{
			Effect:   "Allowed",
			Action:   []string{"s3:PutObject", "s3:DeleteObject"},
			Resource: []string{"arn:aws:s3:::bucket_name"},
		},
	}
}

func (s *WebhookSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *WebhookSuite) TestValidateCreate(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(
		config.Props.MaxRolesAllowed()-1, nil)

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.IsNil)
}

func (s *WebhookSuite) TestTooManyRolesOnCreate(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(config.Props.MaxRolesAllowed(), nil)

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}

func (s *WebhookSuite) TestTooManyRolesOnUpdate(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(config.Props.MaxRolesAllowed()+1, nil)

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}

func (s *WebhookSuite) TestNameTooLong(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(0, nil)

	// Limit is 52 characters
	s.iamRole.ObjectMeta.Name = strings.Repeat("a", 56)

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}

func (s *WebhookSuite) TestRestrictedPolicyResources(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(0, nil)

	s.iamRole.Spec.PolicyDocument.Statement = []Statement{
		{
			Effect: "Allowed",
			Action: []string{"policy-resource:Create"},
			// Not allowed by config
			Resource: []string{"arn:aws:policy-resource:::res/name"},
		},
	}

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}

func (s *WebhookSuite) TestRestrictedAction(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(0, nil)

	s.iamRole.Spec.PolicyDocument.Statement = []Statement{
		{
			Effect:   "Allowed",
			Action:   []string{"ec2:RunInstances"},
			Resource: []string{"*"},
		},
	}

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}

func (s *WebhookSuite) TestRestrictedS3Resource(c *check.C) {
	s.mockk8sClient.EXPECT().IamrolesCount(s.ctx, s.iamRole.ObjectMeta.Namespace).Return(0, nil)

	s.iamRole.Spec.PolicyDocument.Statement = []Statement{
		{
			Effect:   "Allowed",
			Action:   []string{"s3:*"},
			Resource: []string{"s3-resource", "arn:aws:s3:::bucket_name"},
		},
	}

	err := s.iamRole.validateIAMPolicy(s.ctx, false, s.mockk8sClient)
	c.Assert(err, check.NotNil)
}
