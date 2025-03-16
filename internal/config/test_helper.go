/*
Copyright 2024 Keikoproj authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"os"
)

// SetupTestProperties configures the test environment with standard properties
func SetupTestProperties() {
	// Set basic environment variables needed by tests
	os.Setenv("ALLOWED_POLICY_ACTION", "ec2:*,elasticloadbalancing:*,cloudwatch:*,logs:*,sqs:*,sns:*,route53:*,cloudfront:*,rds:*,dynamodb:*")
	os.Setenv("RESTRICTED_POLICY_RESOURCES", "policy-resource")
	os.Setenv("RESTRICTED_S3_RESOURCES", "s3-resource")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("MANAGED_POLICIES", "arn:aws:iam::123456789012:policy/SOMETHING")
	os.Setenv("MANAGED_PERMISSION_BOUNDARY_POLICY", "arn:aws:iam::123456789012:role/iam-manager-permission-boundary")
	os.Setenv("CLUSTER_NAME", "k8s_test_keiko")
	os.Setenv("CLUSTER_OIDC_ISSUER_URL", "google.com/OIDC")
	os.Setenv("DEFAULT_TRUST_POLICY", `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:SERVICE_ACCOUNT_NAME"}}}, {"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`)
	os.Setenv("LOCAL", "true")
	
	// Load properties with the test configuration
	_ = LoadProperties("LOCAL")
}

// CleanupTestProperties cleans up the test environment
func CleanupTestProperties() {
	os.Unsetenv("ALLOWED_POLICY_ACTION")
	os.Unsetenv("RESTRICTED_POLICY_RESOURCES")
	os.Unsetenv("RESTRICTED_S3_RESOURCES")
	os.Unsetenv("AWS_ACCOUNT_ID")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("MANAGED_POLICIES")
	os.Unsetenv("MANAGED_PERMISSION_BOUNDARY_POLICY")
	os.Unsetenv("CLUSTER_NAME")
	os.Unsetenv("CLUSTER_OIDC_ISSUER_URL")
	os.Unsetenv("DEFAULT_TRUST_POLICY")
}
