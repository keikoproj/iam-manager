package awsapi_test

import (
	"context"
	"errors"
	"testing"

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
	check.Suite(&IAMAPISuite{t: t})
	check.TestingT(t)
}

func (s *STSAPISuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockI = mock_awsapi.NewMockSTSAPI(s.mockCtrl)
	s.mockSTS = awsapi.STS{
		Client: s.mockI,
	}

	_ = config.LoadProperties("LOCAL")
}

func (s *STSAPISuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *STSAPISuite) TestGetAccountIDSuccess(c *check.C) {
	s.mockI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Times(1).Return(&sts.GetCallerIdentityOutput{}, nil)
	accountID, err := s.mockSTS.GetAccountID(s.ctx)
	c.Assert(err, check.IsNil)
	c.Assert(accountID, check.NotNil)
}

func (s *STSAPISuite) TestGetAccountIDFailed(c *check.C) {
	s.mockI.EXPECT().GetCallerIdentity(&sts.GetCallerIdentityInput{}).Times(1).Return(&sts.GetCallerIdentityOutput{}, errors.New(iam.ErrCodeNoSuchEntityException))
	accountID, err := s.mockSTS.GetAccountID(s.ctx)
	c.Assert(err, check.NotNil)
	c.Assert(accountID, check.Equals, "")
}
