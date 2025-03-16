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

package utils

import (
	"os"
	
	"github.com/keikoproj/iam-manager/internal/config"
)

// SetupIRSATestEnvironment sets up the test environment for IRSA tests
func SetupIRSATestEnvironment() {
	// Set environment variables needed for IRSA tests
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("CLUSTER_OIDC_ISSUER_URL", "google.com/OIDC")
	os.Setenv("DEFAULT_TRUST_POLICY", `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow","Principal": {"Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/OIDC_PROVIDER"},"Action": "sts:AssumeRoleWithWebIdentity","Condition": {"StringEquals": {"OIDC_PROVIDER:sub": "system:serviceaccount:{{.NamespaceName}}:SERVICE_ACCOUNT_NAME"}}}, {"Effect": "Allow","Principal": {"AWS": ["arn:aws:iam::{{.AccountID}}:role/trust_role"]},"Action": "sts:AssumeRole"}]}`)
	os.Setenv("LOCAL", "true")
	os.Setenv("IRSA_ENABLED", "true")
	
	// Load properties with the test configuration
	_ = config.LoadProperties("LOCAL")
}

// CleanupIRSATestEnvironment cleans up the IRSA test environment
func CleanupIRSATestEnvironment() {
	os.Unsetenv("AWS_ACCOUNT_ID")
	os.Unsetenv("CLUSTER_OIDC_ISSUER_URL")
	os.Unsetenv("DEFAULT_TRUST_POLICY")
	os.Unsetenv("IRSA_ENABLED")
}
