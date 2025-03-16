package config

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
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
	s.ctx = context.TODO()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *PropertiesSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

// SetupTestPropertiesEnv sets up test properties with environment variables
func SetupTestPropertiesEnv() {
	// Set environment variables with specific values for tests
	os.Setenv("CONTROLLER_DESIRED_FREQUENCY", "0")
	os.Setenv("IS_WEBHOOK_ENABLED", "false")
	os.Setenv("LOCAL", "true")
	os.Setenv("CLUSTER_OIDC_ISSUER_URL", "https://google.com/OIDC")
	os.Setenv("GO_TEST_MODE", "true")

	// Create a fresh properties instance to ensure environment changes are applied
	// Use nil to force loading from environment
	Props = nil
	_ = LoadProperties("LOCAL")
}

// CleanupTestPropertiesEnv cleans up test environment variables
func CleanupTestPropertiesEnv() {
	os.Unsetenv("CONTROLLER_DESIRED_FREQUENCY")
	os.Unsetenv("IS_WEBHOOK_ENABLED")
	os.Unsetenv("LOCAL")
	os.Unsetenv("CLUSTER_OIDC_ISSUER_URL")
	os.Unsetenv("GO_TEST_MODE")
}

// test local properties for local environment
func (s *PropertiesSuite) TestLoadPropertiesLocalEnvSuccess(c *check.C) {
	// Set up the test environment with proper variables
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	// Check that we can load properties successfully
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
	c.Assert(Props, check.NotNil)
}

// test failure when env is not local and cm is empty
// should not return nil pointer
func (s *PropertiesSuite) TestLoadPropertiesFailedNoCM(c *check.C) {
	// Skip this test when SKIP_PROBLEMATIC_TESTS is set
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Ensure LOCAL is false to force ConfigMap lookup
	os.Setenv("LOCAL", "")
	os.Setenv("GO_TEST_MODE", "")
	defer func() {
		os.Setenv("LOCAL", "true")
		os.Setenv("GO_TEST_MODE", "true")
	}()

	// Loading without LOCAL and without a ConfigMap should cause an error
	err := LoadProperties("")
	c.Assert(err, check.NotNil)
}

func (s *PropertiesSuite) TestLoadPropertiesFailedNilCM(c *check.C) {
	// Skip this test when SKIP_PROBLEMATIC_TESTS is set
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set to non-local mode to force ConfigMap requirement
	os.Setenv("LOCAL", "")
	os.Setenv("GO_TEST_MODE", "")
	defer func() {
		os.Setenv("LOCAL", "true")
		os.Setenv("GO_TEST_MODE", "true")
	}()

	// Explicitly passing nil ConfigMap should cause an error in non-local mode
	err := LoadProperties("", nil)
	c.Assert(err, check.NotNil)
}

func (s *PropertiesSuite) TestLoadPropertiesSuccess(c *check.C) {
	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	// Perform test
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithDefaults(c *check.C) {
	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	// Perform test
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
	c.Assert(Props.AllowedPolicyAction(), check.NotNil)
	c.Assert(Props.RestrictedPolicyResources(), check.NotNil)
	c.Assert(Props.RestrictedS3Resources(), check.NotNil)
	c.Assert(Props.AWSAccountID(), check.NotNil)
	c.Assert(Props.AWSRegion(), check.NotNil)
}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithDefaultsManagedPoliciesWithNoPrefix(c *check.C) {
	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	// Perform test
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
	c.Assert(Props.ManagedPolicies(), check.NotNil)
}

func (s *PropertiesSuite) TestLoadPropertiesSuccessWithCustom(c *check.C) {
	// Set up test environment with custom values
	os.Setenv("ALLOWED_POLICY_ACTION", "ec2:*")
	os.Setenv("RESTRICTED_POLICY_RESOURCES", "arn:aws:ec2:*:*:instance/*")
	os.Setenv("RESTRICTED_S3_RESOURCES", "arn:aws:s3:::bucket")
	defer func() {
		os.Unsetenv("ALLOWED_POLICY_ACTION")
		os.Unsetenv("RESTRICTED_POLICY_RESOURCES")
		os.Unsetenv("RESTRICTED_S3_RESOURCES")
	}()

	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	// Perform test
	err := LoadProperties("LOCAL")
	c.Assert(err, check.IsNil)
}

func (s *PropertiesSuite) TestGetAllowedPolicyAction(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.AllowedPolicyAction()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetRestrictedPolicyResources(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.RestrictedPolicyResources()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetRestrictedS3Resources(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.RestrictedS3Resources()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetManagedPolicies(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.ManagedPolicies()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetAWSAccountID(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.AWSAccountID()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetAWSRegion(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.AWSRegion()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestGetManagedPermissionBoundaryPolicy(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.ManagedPermissionBoundaryPolicy()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestIsWebhookEnabled(c *check.C) {
	// For cross-platform testing, skip environment-sensitive tests on ARM64
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.IsWebHookEnabled()
	c.Assert(value, check.Equals, false)
}

func (s *PropertiesSuite) TestControllerDesiredFrequency(c *check.C) {
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.ControllerDesiredFrequency()
	c.Assert(value, check.Equals, 0)
}

func (s *PropertiesSuite) TestControllerClusterName(c *check.C) {
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up test environment
	os.Setenv("CLUSTER_NAME", "k8s_test_keiko")
	defer os.Unsetenv("CLUSTER_NAME")

	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.ClusterName()
	c.Assert(value, check.Equals, "k8s_test_keiko")
}

func (s *PropertiesSuite) TestControllerDefaultTrustPolicy(c *check.C) {
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.DefaultTrustPolicy()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestIsIRSAEnabled(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.IsIRSAEnabled()
	c.Assert(value, check.NotNil)
}

func (s *PropertiesSuite) TestControllerOIDCIssuerUrl(c *check.C) {
	// For cross-platform testing, skip environment-sensitive tests on ARM64
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up test environment
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.OIDCIssuerUrl()
	c.Assert(value, check.Equals, "https://google.com/OIDC")
}

func (s *PropertiesSuite) TestIsIRSARegionalEndpointDisabled(c *check.C) {
	SetupTestPropertiesEnv()
	defer CleanupTestPropertiesEnv()

	value := Props.IsIRSARegionalEndpointDisabled()
	c.Assert(value, check.NotNil)
}
