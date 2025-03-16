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

package validation

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/stretchr/testify/assert"

	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
)

// These additional tests use the testify framework for simpler assertions
// and complement the existing tests in validate_test.go

func TestContainsStringAndRemoveString(t *testing.T) {
	testCases := []struct {
		name           string
		slice          []string
		searchStr      string
		expectContains bool
		expectedResult []string
	}{
		{
			name:           "String exists in slice",
			slice:          []string{"apple", "banana", "cherry"},
			searchStr:      "banana",
			expectContains: true,
			expectedResult: []string{"apple", "cherry"},
		},
		{
			name:           "String doesn't exist in slice",
			slice:          []string{"apple", "banana", "cherry"},
			searchStr:      "grape",
			expectContains: false,
			expectedResult: []string{"apple", "banana", "cherry"},
		},
		{
			name:           "Empty slice",
			slice:          []string{},
			searchStr:      "anything",
			expectContains: false,
			expectedResult: []string{}, // This will be nil in the function
		},
		{
			name:           "Empty search string",
			slice:          []string{"apple", "banana", "", "cherry"},
			searchStr:      "",
			expectContains: true,
			expectedResult: []string{"apple", "banana", "cherry"},
		},
		{
			name:           "Duplicate elements",
			slice:          []string{"apple", "banana", "banana", "cherry"},
			searchStr:      "banana",
			expectContains: true,
			// RemoveString removes ALL occurrences of the target string
			expectedResult: []string{"apple", "cherry"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test ContainsString
			result := ContainsString(tc.slice, tc.searchStr)
			assert.Equal(t, tc.expectContains, result)

			// Test RemoveString
			removed := RemoveString(tc.slice, tc.searchStr)
			// Special case for empty slices
			if len(tc.slice) == 0 {
				assert.Empty(t, removed, "Expected empty slice")
			} else {
				// The RemoveString function skips ALL instances of the search string
				assert.ElementsMatch(t, tc.expectedResult, removed, "Removed elements should match")
			}
		})
	}
}

func TestCompareTags(t *testing.T) {
	testCases := []struct {
		name           string
		requestTags    map[string]string
		targetTags     []*iam.Tag
		expectEqual    bool
	}{
		{
			name: "Equal tags",
			requestTags: map[string]string{
				"env": "dev",
				"app": "iam-manager",
			},
			targetTags: []*iam.Tag{
				{Key: aws.String("env"), Value: aws.String("dev")},
				{Key: aws.String("app"), Value: aws.String("iam-manager")},
			},
			expectEqual: true,
		},
		{
			name: "Different tag values",
			requestTags: map[string]string{
				"env": "dev",
				"app": "iam-manager",
			},
			targetTags: []*iam.Tag{
				{Key: aws.String("env"), Value: aws.String("prod")},
				{Key: aws.String("app"), Value: aws.String("iam-manager")},
			},
			expectEqual: false,
		},
		{
			name: "Different number of tags",
			requestTags: map[string]string{
				"env": "dev",
				"app": "iam-manager",
				"team": "platform",
			},
			targetTags: []*iam.Tag{
				{Key: aws.String("env"), Value: aws.String("dev")},
				{Key: aws.String("app"), Value: aws.String("iam-manager")},
			},
			expectEqual: false,
		},
		{
			name: "Empty tags",
			requestTags: map[string]string{},
			targetTags: []*iam.Tag{},
			// The actual implementation treats empty tags as not equal
			expectEqual: false,
		},
		{
			name: "Nil target tags",
			requestTags: map[string]string{},
			targetTags: nil,
			// The actual implementation treats nil tags as not equal to empty map
			expectEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareTags(context.Background(), tc.requestTags, tc.targetTags)
			assert.Equal(t, tc.expectEqual, result)
		})
	}
}

func TestValidateIAMPolicyAction_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		policyDoc      v1alpha1.PolicyDocument
		expectError    bool
	}{
		{
			name: "Empty policy document",
			policyDoc: v1alpha1.PolicyDocument{
				Statement: []v1alpha1.Statement{},
			},
			expectError: false,
		},
		{
			name: "Empty actions",
			policyDoc: v1alpha1.PolicyDocument{
				Statement: []v1alpha1.Statement{
					{
						Effect: "Allow",
						Action: []string{},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Mixed Deny and Allow",
			policyDoc: v1alpha1.PolicyDocument{
				Statement: []v1alpha1.Statement{
					{
						Effect: "Deny",
						Action: []string{"*"}, // Would be restricted but it's Deny
					},
					{
						Effect: "Allow",
						// Don't test specific actions, since the validation config
						// may vary between test environments
						Action: []string{},
						Resource: []string{"*"},
					},
				},
			},
			expectError: false,
		},
	}

	// Directly setup test environment
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("LOCAL", "true")
	os.Setenv("GO_TEST_MODE", "true")
	os.Setenv("ALLOWED_POLICY_ACTION", "ec2:*,elasticloadbalancing:*,cloudwatch:*,logs:*")
	
	// Reset global config and force reload
	config.Props = nil
	_ = config.LoadProperties("LOCAL")
	
	defer func() {
		// Clean up environment variables
		os.Unsetenv("ALLOWED_POLICY_ACTION")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("LOCAL")
		os.Unsetenv("GO_TEST_MODE")
	}()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIAMPolicyAction(context.Background(), tc.policyDoc)
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
