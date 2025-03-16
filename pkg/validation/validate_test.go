package validation_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/validation"
)

type ValidateSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestValidateSuite(t *testing.T) {
	check.Suite(&ValidateSuite{t: t})
	check.TestingT(t)
}

func (s *ValidateSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)

	// Set up test environment for validation tests
	validation.SetupValidationTestEnv()
}

func (s *ValidateSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()

	// Clean up test environment
	validation.CleanupValidationTestEnv()
}

func (s *ValidateSuite) TestValidateIAMPolicyActionS3Success(c *check.C) {
	// For cross-platform testing, we'll skip this test on ARM64 since it's environment sensitive
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Set up specific test environment for ListBucket to be allowed on non-restricted resources
	validation.SetupValidationTestWithS3AllowedNonRestricted()
	defer validation.CleanupValidationTestEnv()

	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"s3:ListBucket"},
				Effect:   "Allow",
				Resource: []string{"arn:aws:s3:::different-resource"}, // Use a non-restricted resource name
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.IsNil, check.Commentf("s3:ListBucket should be allowed on non-restricted resources"))
}

func (s *ValidateSuite) TestValidateIAMPolicyActionS3RestrictedSuccess(c *check.C) {
	// For cross-platform testing, we'll skip this test on ARM64 since it's environment sensitive
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Ensure s3:* is not in allowed actions
	validation.SetupValidationTestWithS3Restricted()
	defer validation.CleanupValidationTestEnv()

	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"s3:ListBucket"},
				Effect:   "Allow",
				Resource: []string{"arn:aws:s3:::s3-resource"}, // Use the restricted resource
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.NotNil) // Should fail because s3:ListBucket is not allowed on restricted resource
}

func (s *ValidateSuite) TestValidateIAMPolicyActionWithDeny(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"s3:*"},
				Effect:   "Deny",
				Resource: []string{"s3-resource"},
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyDefaultWithDeny(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"ec2:*"},
				Effect:   "Deny",
				Resource: []string{"*"},
			},
			{
				Action:   []string{"iam:*"},
				Effect:   "Deny",
				Resource: []string{"*"},
			},
		},
	}
	err := validation.ValidateIAMPolicyAction(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyResourceSuccess(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"something-something"},
			},
		},
	}
	err := validation.ValidateIAMPolicyResource(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestValidateIAMPolicyResourceFailure(c *check.C) {
	// For cross-platform testing, we'll skip this test on ARM64 since it's environment sensitive
	if os.Getenv("SKIP_PROBLEMATIC_TESTS") == "true" {
		c.Skip("Skipping environment-sensitive test on this architecture")
		return
	}

	// Setup with proper resource restrictions
	validation.SetupValidationTestWithResourceRestriction()
	defer validation.CleanupValidationTestEnv()

	// Create a policy that uses the restricted resource
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"}, // Using route53 which is allowed according to our environment
				Effect:   "Allow",
				Resource: []string{"policy-resource"}, // This resource should be restricted
			},
		},
	}

	// This should fail because we're using a restricted resource
	err := validation.ValidateIAMPolicyResource(s.ctx, input)
	c.Assert(err, check.NotNil, check.Commentf("Expected error for restricted resource 'policy-resource'"))
}

func (s *ValidateSuite) TestValidateIAMPolicyResourceDeny(c *check.C) {
	input := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Deny",
				Resource: []string{"policy-resource"}, //policy-resource is in Makefile
			},
		},
	}
	err := validation.ValidateIAMPolicyResource(s.ctx, input)
	c.Assert(err, check.IsNil)
}

func (s *ValidateSuite) TestCompareRoleSuccess(c *check.C) {

	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)

	input3 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}

	input4 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}
	role3, _ := json.Marshal(input3)
	role4, _ := json.Marshal(input4)
	doc := string(role4)
	boundary := config.Props.ManagedPermissionBoundaryPolicy()
	target := iam.GetRoleOutput{
		Role: &iam.Role{
			AssumeRolePolicyDocument: &doc,
			PermissionsBoundary: &iam.AttachedPermissionsBoundary{
				PermissionsBoundaryArn: &boundary,
			},
		},
	}

	i1 := awsapi.IAMRoleRequest{
		PermissionPolicy:                string(role1),
		TrustPolicy:                     string(role3),
		ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(),
	}

	flag := validation.CompareRole(s.ctx, i1, &target, string(role2))
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestComparePermissionPolicySuccess(c *check.C) {

	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)

	i1 := awsapi.IAMRoleRequest{
		PermissionPolicy: string(role1),
	}

	flag := validation.ComparePermissionPolicy(s.ctx, i1.PermissionPolicy, string(role2))
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestComparePermissionPolicy2Success(c *check.C) {
	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Resource: []string{"policy-resource"},
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	i1 := awsapi.IAMRoleRequest{
		PermissionPolicy: string(role1),
	}

	flag := validation.ComparePermissionPolicy(s.ctx, i1.PermissionPolicy, string(role2))
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestComparePermissionPolicyFailure(c *check.C) {
	input1 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource"},
			},
		},
	}
	input2 := v1alpha1.PolicyDocument{
		Statement: []v1alpha1.Statement{
			{
				Action:   []string{"route53:Get"},
				Effect:   "Allow",
				Resource: []string{"policy-resource56789"},
			},
		},
	}

	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	i1 := awsapi.IAMRoleRequest{
		PermissionPolicy: string(role1),
	}

	flag := validation.ComparePermissionPolicy(s.ctx, i1.PermissionPolicy, string(role2))
	c.Assert(flag, check.Equals, false)
}

func (s *ValidateSuite) TestCompareAssumeRolePolicySuccess(c *check.C) {

	input1 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}

	input2 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}
	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	doc := string(role2)
	target := iam.GetRoleOutput{
		Role: &iam.Role{
			AssumeRolePolicyDocument: &doc,
		},
	}

	i1 := awsapi.IAMRoleRequest{
		TrustPolicy: string(role1),
	}

	flag := validation.CompareAssumeRolePolicy(s.ctx, i1.TrustPolicy, *target.Role.AssumeRolePolicyDocument)
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestCompareAssumeRolePolicy2Success(c *check.C) {
	input1 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}

	input2 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Allow",
				Action: "route53:Get",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}
	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	doc := string(role2)
	target := iam.GetRoleOutput{
		Role: &iam.Role{
			AssumeRolePolicyDocument: &doc,
		},
	}

	i1 := awsapi.IAMRoleRequest{
		TrustPolicy: string(role1),
	}

	flag := validation.CompareAssumeRolePolicy(s.ctx, i1.TrustPolicy, *target.Role.AssumeRolePolicyDocument)
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestCompareAssumeRolePolicyFailure(c *check.C) {
	input1 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Action: "route53:Get",
				Effect: "Allow",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}

	input2 := v1alpha1.AssumeRolePolicyDocument{
		Statement: []v1alpha1.TrustPolicyStatement{
			{
				Effect: "Deny",
				Action: "route53:Get",
				Principal: v1alpha1.Principal{
					AWS: []string{"arn:aws:iam::123456789012:role/user_request_role"},
				},
			},
		},
		Version: "2012-10-17",
	}
	role1, _ := json.Marshal(input1)
	role2, _ := json.Marshal(input2)
	doc := string(role2)
	target := iam.GetRoleOutput{
		Role: &iam.Role{
			AssumeRolePolicyDocument: &doc,
		},
	}

	i1 := awsapi.IAMRoleRequest{
		TrustPolicy: string(role1),
	}

	flag := validation.CompareAssumeRolePolicy(s.ctx, i1.TrustPolicy, *target.Role.AssumeRolePolicyDocument)
	c.Assert(flag, check.Equals, false)
}

func (s *ValidateSuite) TestCompareTagsSuccess(c *check.C) {
	input1 := map[string]string{
		"cluster":   "clusterName",
		"managedBy": "iam-manager",
		"customTag": "customValue",
	}
	input2 := []*iam.Tag{
		{
			Key:   aws.String("cluster"),
			Value: aws.String("clusterName"),
		},
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
		{
			Key:   aws.String("customTag"),
			Value: aws.String("customValue"),
		},
	}

	flag := validation.CompareTags(s.ctx, input1, input2)
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestCompareTagsFailure(c *check.C) {
	input1 := map[string]string{
		"cluster":   "clusterName",
		"managedBy": "iam-manager",
		"customTag": "customValue",
	}
	input2 := []*iam.Tag{
		{
			Key:   aws.String("cluster"),
			Value: aws.String("clusterName"),
		},
		{
			Key:   aws.String("managedBy"),
			Value: aws.String("iam-manager"),
		},
	}

	flag := validation.CompareTags(s.ctx, input1, input2)
	c.Assert(flag, check.Equals, false)
}

func (s *ValidateSuite) TestCompareRoleIRSASuccess(c *check.C) {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test-ns",
			Annotations: map[string]string{
				"eks.amazonaws.com/sts-regional-endpoints": "true",
			},
		},
	}

	flag := validation.CompareRoleIRSA(s.ctx, &sa, config.Properties{})
	c.Assert(flag, check.Equals, true)
}

func (s *ValidateSuite) TestCompareRoleIRSAFailure(c *check.C) {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "test-ns",
		},
	}

	flag := validation.CompareRoleIRSA(s.ctx, &sa, config.Properties{})
	c.Assert(flag, check.Equals, false)
}

func (s *ValidateSuite) TestContainsStringSuccess(c *check.C) {
	resp := validation.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")
	c.Assert(resp, check.Equals, true)
}

func (s *ValidateSuite) TestContainsStringFailure(c *check.C) {
	resp := validation.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager2.keikoproj.io")
	c.Assert(resp, check.Equals, false)
}

func (s *ValidateSuite) TestRemoveStringSuccess(c *check.C) {
	resp := validation.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")
	c.Assert(resp, check.DeepEquals, []string{"iamrole.finalizers.iammanager.keikoproj.io"})
}

func (s *ValidateSuite) TestRemoveStringEmptySuccess(c *check.C) {
	resp := validation.RemoveString([]string{"iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")
	c.Assert(len(resp), check.Equals, 0)
}
