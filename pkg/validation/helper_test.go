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
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/stretchr/testify/assert"

	"github.com/keikoproj/iam-manager/api/v1alpha1"
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
			expectedResult: []string{},
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
			// Note: RemoveString only removes the first occurrence
			expectedResult: []string{"apple", "banana", "cherry"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test ContainsString
			result := ContainsString(tc.slice, tc.searchStr)
			assert.Equal(t, tc.expectContains, result)

			// Test RemoveString
			removed := RemoveString(tc.slice, tc.searchStr)
			assert.Equal(t, tc.expectedResult, removed)
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
			expectEqual: true,
		},
		{
			name: "Nil target tags",
			requestTags: map[string]string{},
			targetTags: nil,
			expectEqual: true,
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
						Action: []string{"s3:GetObject"}, // Allowed
						Resource: []string{"arn:aws:s3:::example-bucket/*"},
					},
				},
			},
			expectError: false,
		},
	}

	// Set up test environment using existing helper functions
	SetupTestEnvironment()
	SetupValidationTestWithS3AllowedNonRestricted()
	
	defer func() {
		CleanupValidationTestEnv()
		CleanupTestEnvironment()
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
