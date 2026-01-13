package awsapi_test

import (
	"errors"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"gopkg.in/check.v1"
)

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyHasAccess(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "dynamodb:*",
                    "Resource": "*"
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessNewPolicyHasAccess(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "s3:*",
                    "Resource": "*"
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*existing policy doesn't have DynamoDB access to the same AWS account, and new permission policy has DynamoDB access to the same AWS account*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessGetRoleFailure(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(nil, errors.New("get role failed"))

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*get role failed: VALID_ROLE*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessGetRolePolicyFailure(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(nil, errors.New("get role policy failed"))

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyAndNewRequestHaveAccess(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "dynamodb:*",
                    "Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.IsNil)
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyOtherAccountAndNewRequestSameAccount(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "dynamodb:*",
                    "Resource": "arn:aws:dynamodb:us-west-2:987654321098:table/OtherTable"
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*existing policy doesn't have DynamoDB access to the same AWS account, and new permission policy has DynamoDB access to the same AWS account*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyNoAccessAndNewRequestHasAccess(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "s3:*",
                    "Resource": "*"
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*existing policy doesn't have DynamoDB access to the same AWS account, and new permission policy has DynamoDB access to the same AWS account*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyNoAccessAndNewRequestHasAccessWithArrayFields(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": ["s3:*", "ec2:*"],
                    "Resource": ["arn:aws:s3:::example_bucket", "arn:aws:ec2:us-west-2:123456789012:instance/*"]
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*existing policy doesn't have DynamoDB access to the same AWS account, and new permission policy has DynamoDB access to the same AWS account*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessInvalidRoleArn(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("INVALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("invalid-arn-format"),
		},
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "INVALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*invalid ARN format*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessExistingPolicyOtherAccountAndNewRequestSameAccountWithArrayResources(c *check.C) {
	s.mockIAM.DisallowSameAccountDynamoDBAccess = true
	s.mockI.EXPECT().GetRole(&iam.GetRoleInput{RoleName: aws.String("VALID_ROLE")}).Times(1).Return(&iam.GetRoleOutput{
		Role: &iam.Role{
			Arn: aws.String("arn:aws:iam::123456789012:role/VALID_ROLE"),
		},
	}, nil)
	s.mockI.EXPECT().GetRolePolicy(&iam.GetRolePolicyInput{RoleName: aws.String("VALID_ROLE"), PolicyName: aws.String("VALID_POLICY")}).Times(1).Return(&iam.GetRolePolicyOutput{
		PolicyDocument: aws.String(url.QueryEscape(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": "dynamodb:*",
                    "Resource": ["arn:aws:dynamodb:us-west-2:987654321098:table/OtherTable", "arn:aws:dynamodb:us-west-2:987654321098:table/OtherTable2"]
                }
            ]
        }`)),
	}, nil)

	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": ["arn:aws:dynamodb:us-west-2:123456789012:table/MyTable"]}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.NotNil)
	c.Assert(err.Error(), check.Matches, "*existing policy doesn't have DynamoDB access to the same AWS account, and new permission policy has DynamoDB access to the same AWS account*")
}

func (s *IAMAPISuite) TestValidateAllowedDynamoDBAccessFeatureFlagDisabled(c *check.C) {
	// When DisallowSameAccountDynamoDBAccess is false, validation should be skipped entirely
	s.mockIAM.DisallowSameAccountDynamoDBAccess = false

	// No mock expectations needed since the function returns early without making any AWS API calls
	req := awsapi.IAMRoleRequest{Name: "VALID_ROLE", PolicyName: "VALID_POLICY", PermissionPolicy: `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "dynamodb:*", "Resource": "*"}]}`}
	err := s.mockIAM.ValidateAllowSameAccountDynamoDBAccess(s.ctx, req)
	c.Assert(err, check.IsNil)
}
