package config

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	v1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

type PropertiesSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestPropertiesSuite(t *testing.T) {
	check.Suite(&PropertiesSuite{t: t})
	check.TestingT(t)
}

func (s *PropertiesSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *PropertiesSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

// test local properties for local environment
func (s *PropertiesSuite) TestLoadPropertiesLocalEnvSuccess(c *check.C) {
	//c.Assert(Props, check.IsNil)
	err := LoadProperties("Local")
	c.Assert(err, check.IsNil)
	c.Assert(Props, check.NotNil)
	c.Assert(Props.AWSAccountID(), check.Equals, "123456789012")
}

// test failure when env is not local and cm is empty
// should not return nil pointer
func (s *PropertiesSuite) TestLoadPropertiesFailedNoCM(c *check.C) {
	err := LoadProperties("")
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "config map cannot be nil")
}

func (s *PropertiesSuite) TestLoadPropertiesFailedNilCM(c *check.C) {
	err := LoadProperties("", nil)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Equals, "config map cannot be nil")
}

func (s *PropertiesSuite) TestLoadPropertiesSuccess(c *check.C) {
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.managed.permission.boundary.policy": "iam-manager-permission-boundary",
		},
	}
	err := LoadProperties("", cm)
	c.Assert(err, check.IsNil)
	c.Assert(strings.HasPrefix(Props.ManagedPermissionBoundaryPolicy(), "arn:aws:iam:"), check.Equals, true)
}
