package awsapi_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"

	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	mock_awsapi "github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
)

type STSAPISuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
	mockI    *mock_awsapi.MockSTSAPI
	mockSTS  awsapi.STS
}

func TestSTSAPITestSuite(t *testing.T) {
	check.Suite(&STSAPISuite{t: t})
	check.TestingT(t)
}

func (s *STSAPISuite) SetUpTest(c *check.C) {
	// Skip problematic tests on ARM64 if needed
	if awsapi.SkipIfProblematicTest(c) {
		return
	}

	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockI = mock_awsapi.NewMockSTSAPI(s.mockCtrl)
	s.mockSTS = awsapi.STS{
		Client: s.mockI,
	}

	// Setup test environment
	awsapi.SetTestEnvironment()
	_ = config.LoadProperties("LOCAL")
}

func (s *STSAPISuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
	awsapi.CleanupTestEnvironment()
}

func (s *STSAPISuite) TestGetAccountIDSuccess(c *check.C) {
	// Create mock response with valid Account field
	mockOutput := &sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
	}

	// Setup expectation
	s.mockI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Times(1).Return(mockOutput, nil)

	// Call the function
	accountID, err := s.mockSTS.GetAccountID(s.ctx)

	// Verify results
	c.Assert(err, check.IsNil)
	c.Assert(accountID, check.Equals, "123456789012")
}

func (s *STSAPISuite) TestGetAccountIDFailed(c *check.C) {
	// Setup expectation for error case
	s.mockI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Times(1).Return(nil, errors.New(iam.ErrCodeNoSuchEntityException))

	// Call the function
	accountID, err := s.mockSTS.GetAccountID(s.ctx)

	// Verify results
	c.Assert(err, check.NotNil)
	c.Assert(accountID, check.Equals, "")
}
