package utils_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/internal/utils"
	"gopkg.in/check.v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type OIDCTestSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestOIDCTestSuite(t *testing.T) {
	check.Suite(&OIDCTestSuite{t: t})
	check.TestingT(t)
}

func (s *OIDCTestSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *OIDCTestSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *OIDCTestSuite) TestParseIRSAAnnotationEmpty(c *check.C) {
	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name: "iam-role",
		},
	}
	flag, saName := utils.ParseIRSAAnnotation(s.ctx, input)
	c.Assert(flag, check.Equals, false)
	c.Assert(saName, check.HasLen, 0)
}

func (s *OIDCTestSuite) TestParseIRSAAnnotationValid(c *check.C) {
	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name:      "iam-role",
			Namespace: "k8s-namespace-dev",
			Annotations: map[string]string{
				constants.IRSAAnnotation: "default",
			},
		},
	}
	flag, saName := utils.ParseIRSAAnnotation(s.ctx, input)
	c.Assert(flag, check.Equals, true)
	c.Assert(saName, check.Equals, "default")
}

func (s *OIDCTestSuite) TestParseIRSAAnnotationOtherAnnotations(c *check.C) {
	input := &v1alpha1.Iamrole{
		ObjectMeta: v1.ObjectMeta{
			Name: "iam-role",
			Annotations: map[string]string{
				"iam.amazonaws.com/role": "default",
			},
		},
	}
	flag, saName := utils.ParseIRSAAnnotation(s.ctx, input)
	c.Assert(flag, check.Equals, false)
	c.Assert(saName, check.HasLen, 0)
}

func (s *OIDCTestSuite) TestGetIdpServerCertThumbprintSuccess(c *check.C) {
	_, err := utils.GetIdpServerCertThumbprint(s.ctx, "https://www.google.com")
	c.Assert(err, check.IsNil)
}

func (s *OIDCTestSuite) TestGetIdpServerCertThumbprintNotURL(c *check.C) {
	_, err := utils.GetIdpServerCertThumbprint(s.ctx, "#8123not_even_a_url")
	c.Assert(err, check.NotNil)
}

func (s *OIDCTestSuite) TestGetIdpServerCertThumbprintNotHttps(c *check.C) {
	_, err := utils.GetIdpServerCertThumbprint(s.ctx, "http://www.google.com")
	c.Assert(err, check.NotNil)
}
