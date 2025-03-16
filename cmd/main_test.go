/*
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

package main

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	mock_awsapi "github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
)

// setupTestIAMClient creates a mock IAM client for testing
func setupTestIAMClient(t *testing.T) (*awsapi.IAM, *mock_awsapi.MockIAMAPI, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)
	mockIAM := mock_awsapi.NewMockIAMAPI(mockCtrl)
	
	iamClient := &awsapi.IAM{
		Client: mockIAM,
	}
	
	return iamClient, mockIAM, mockCtrl
}

// TestMain runs before all tests in this package to set up environment
func TestMain(m *testing.M) {
	// Set test environment variables
	os.Setenv("LOCAL", "true")
	os.Setenv("AWS_REGION", "us-west-2")
	
	// Load configuration properties
	config.LoadProperties("LOCAL")
	
	// Run tests
	code := m.Run()
	
	// Clean up
	os.Unsetenv("LOCAL")
	os.Unsetenv("AWS_REGION")
	
	os.Exit(code)
}

// Helper function to modify config properties for tests
func setIRSAConfig(enabled bool) {
	if config.Props == nil {
		config.Props = &config.Properties{}
	}
	
	// Use proper field names via setter/getter methods
	origEnabled := config.Props.IsIRSAEnabled()
	origURL := config.Props.OIDCIssuerUrl()
	
	// Set test values using environment variables
	if enabled {
		os.Setenv("ENABLE_IRSA", "true")
		os.Setenv("OIDC_ISSUER_URL", "https://test.example.com")
	} else {
		os.Setenv("ENABLE_IRSA", "false")
		os.Setenv("OIDC_ISSUER_URL", "")
	}
	
	// Reload config to apply changes
	config.LoadProperties("LOCAL")
	
	// Return a cleanup function that we'll ignore in the test
	// since we clean up in the deferred functions
}

// Test that OIDC is set up properly when IRSA is enabled
func TestHandleOIDCSetupForIRSA_Enabled(t *testing.T) {
	// Setup - enable IRSA
	setIRSAConfig(true)
	defer func() {
		os.Unsetenv("ENABLE_IRSA")
		os.Unsetenv("OIDC_ISSUER_URL")
	}()
	
	// Mock thumbprint function
	originalGetThumbprint := utils.GetIdpServerCertThumbprint
	defer func() {
		// Restore original function after test
		utils.GetIdpServerCertThumbprint = originalGetThumbprint
	}()
	
	// Override with mock function
	utils.GetIdpServerCertThumbprint = func(ctx context.Context, url string) (string, error) {
		return "MOCK_THUMBPRINT", nil
	}
	
	// Set up IAM client mock
	iamClient, mockIAM, mockCtrl := setupTestIAMClient(t)
	defer mockCtrl.Finish()
	
	// Set expectations
	mockIAM.EXPECT().
		CreateOIDCProvider(gomock.Any(), "https://test.example.com", config.OIDCAudience, "MOCK_THUMBPRINT").
		Return(nil)
	
	// Call the function
	err := handleOIDCSetupForIRSA(context.Background(), iamClient)
	
	// Assert results
	assert.NoError(t, err)
}

// Test that OIDC setup is skipped when IRSA is disabled
func TestHandleOIDCSetupForIRSA_Disabled(t *testing.T) {
	// Setup - disable IRSA
	setIRSAConfig(false)
	defer func() {
		os.Unsetenv("ENABLE_IRSA")
		os.Unsetenv("OIDC_ISSUER_URL")
	}()
	
	// Set up IAM client mock - expect no calls
	iamClient, _, mockCtrl := setupTestIAMClient(t)
	defer mockCtrl.Finish()
	
	// Call the function
	err := handleOIDCSetupForIRSA(context.Background(), iamClient)
	
	// Assert results - should be no error even though no calls were made
	assert.NoError(t, err)
}

// Test error handling when thumbprint fetch fails
func TestHandleOIDCSetupForIRSA_ThumbprintError(t *testing.T) {
	// Setup - enable IRSA
	setIRSAConfig(true)
	defer func() {
		os.Unsetenv("ENABLE_IRSA")
		os.Unsetenv("OIDC_ISSUER_URL")
	}()
	
	// Mock thumbprint function to return an error
	originalGetThumbprint := utils.GetIdpServerCertThumbprint
	defer func() {
		// Restore original function after test
		utils.GetIdpServerCertThumbprint = originalGetThumbprint
	}()
	
	// Override with mock function that returns an error
	utils.GetIdpServerCertThumbprint = func(ctx context.Context, url string) (string, error) {
		return "", errors.New("thumbprint error")
	}
	
	// Set up IAM client mock - expect no calls since we'll fail before that
	iamClient, _, mockCtrl := setupTestIAMClient(t)
	defer mockCtrl.Finish()
	
	// Call the function
	err := handleOIDCSetupForIRSA(context.Background(), iamClient)
	
	// Assert results - should return the error from thumbprint
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "thumbprint error")
}

// Test error handling when CreateOIDCProvider fails
func TestHandleOIDCSetupForIRSA_OIDCProviderError(t *testing.T) {
	// Setup - enable IRSA
	setIRSAConfig(true)
	defer func() {
		os.Unsetenv("ENABLE_IRSA")
		os.Unsetenv("OIDC_ISSUER_URL")
	}()
	
	// Mock thumbprint function
	originalGetThumbprint := utils.GetIdpServerCertThumbprint
	defer func() {
		// Restore original function after test
		utils.GetIdpServerCertThumbprint = originalGetThumbprint
	}()
	
	// Override with mock function
	utils.GetIdpServerCertThumbprint = func(ctx context.Context, url string) (string, error) {
		return "MOCK_THUMBPRINT", nil
	}
	
	// Set up IAM client mock
	iamClient, mockIAM, mockCtrl := setupTestIAMClient(t)
	defer mockCtrl.Finish()
	
	// Set expectations - this time return an error
	mockIAM.EXPECT().
		CreateOIDCProvider(gomock.Any(), "https://test.example.com", config.OIDCAudience, "MOCK_THUMBPRINT").
		Return(errors.New("oidc provider error"))
	
	// Call the function
	err := handleOIDCSetupForIRSA(context.Background(), iamClient)
	
	// Assert results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oidc provider error")
}
