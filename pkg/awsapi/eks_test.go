package awsapi_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"

	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	mock_awsapi "github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
)

type EKSAPISuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
	mockE    *mock_awsapi.MockEKSAPI
	mockEKS  awsapi.EKS
}

func TestEKSAPITestSuite(t *testing.T) {
	check.Suite(&EKSAPISuite{t: t})
	check.TestingT(t)
}

func (s *EKSAPISuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockE = mock_awsapi.NewMockEKSAPI(s.mockCtrl)
	s.mockEKS = awsapi.EKS{
		Client: s.mockE,
	}

	_ = config.LoadProperties("LOCAL")
}

func (s *EKSAPISuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *EKSAPISuite) TestDescribeClusterSuccess(c *check.C) {
	s.mockE.EXPECT().DescribeCluster(&eks.DescribeClusterInput{Name: aws.String("valid_cluster")}).Times(1).Return(&eks.DescribeClusterOutput{
		Cluster: &eks.Cluster{
			Name: aws.String("valid_cluster"),
		},
	}, nil)
	_, err := s.mockEKS.DescribeCluster(s.ctx, "valid_cluster")
	c.Assert(err, check.IsNil)
}

func (s *EKSAPISuite) TestDescribeClusterNotFound(c *check.C) {
	s.mockE.EXPECT().DescribeCluster(&eks.DescribeClusterInput{Name: aws.String("not_found_cluster")}).Times(1).Return(nil, awserr.New(eks.ErrCodeResourceNotFoundException, "", errors.New(eks.ErrCodeResourceNotFoundException)))
	_, err := s.mockEKS.DescribeCluster(s.ctx, "not_found_cluster")
	c.Assert(err, check.NotNil)
}

func (s *EKSAPISuite) TestDescribeClusterClientException(c *check.C) {
	s.mockE.EXPECT().DescribeCluster(&eks.DescribeClusterInput{Name: aws.String("wrong_cluster")}).Times(1).Return(nil, awserr.New(eks.ErrCodeClientException, "", errors.New(eks.ErrCodeClientException)))
	_, err := s.mockEKS.DescribeCluster(s.ctx, "wrong_cluster")
	c.Assert(err, check.NotNil)
}

func (s *EKSAPISuite) TestDescribeClusterServerException(c *check.C) {
	s.mockE.EXPECT().DescribeCluster(&eks.DescribeClusterInput{Name: aws.String("wrong_server_cluster")}).Times(1).Return(nil, awserr.New(eks.ErrCodeServerException, "", errors.New(eks.ErrCodeServerException)))
	_, err := s.mockEKS.DescribeCluster(s.ctx, "wrong_server_cluster")
	c.Assert(err, check.NotNil)
}

func (s *EKSAPISuite) TestDescribeClusterServiceUnavailableException(c *check.C) {
	s.mockE.EXPECT().DescribeCluster(&eks.DescribeClusterInput{Name: aws.String("service_unavailable")}).Times(1).Return(nil, awserr.New(eks.ErrCodeServiceUnavailableException, "", errors.New(eks.ErrCodeServiceUnavailableException)))
	_, err := s.mockEKS.DescribeCluster(s.ctx, "service_unavailable")
	c.Assert(err, check.NotNil)
}
