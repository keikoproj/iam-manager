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

package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseURL(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expectedOut string
		expectError bool
	}{
		{
			name:        "Valid HTTPS URL with port",
			url:         "https://example.com:443",
			expectedOut: "example.com:443",
			expectError: false,
		},
		{
			name:        "Valid HTTPS URL without port",
			url:         "https://example.com",
			expectedOut: "example.com:443",
			expectError: false,
		},
		{
			name:        "Invalid URL",
			url:         "not-a-url",
			expectedOut: "",
			expectError: true,
		},
		{
			name:        "Non-HTTPS URL",
			url:         "http://example.com",
			expectedOut: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseURL(context.Background(), tc.url)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOut, result)
			}
		})
	}
}

func TestParseIRSAAnnotation(t *testing.T) {
	testCases := []struct {
		name           string
		annotations    map[string]string
		expectedFound  bool
		expectedArns   []string
	}{
		{
			name: "Valid single ARN",
			annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/test-role",
			},
			expectedFound: true,
			expectedArns: []string{"arn:aws:iam::123456789012:role/test-role"},
		},
		{
			name: "Valid multiple ARNs",
			annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "[\"arn:aws:iam::123456789012:role/test-role-1\", \"arn:aws:iam::123456789012:role/test-role-2\"]",
			},
			expectedFound: true,
			expectedArns: []string{
				"arn:aws:iam::123456789012:role/test-role-1", 
				"arn:aws:iam::123456789012:role/test-role-2",
			},
		},
		{
			name: "Empty annotation value",
			annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "",
			},
			expectedFound: false,
			expectedArns: nil,
		},
		{
			name: "No IRSA annotation",
			annotations: map[string]string{
				"some-other-annotation": "some-value",
			},
			expectedFound: false,
			expectedArns: nil,
		},
		{
			name: "Invalid JSON in array annotation",
			annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "[\"invalid-json",
			},
			expectedFound: true,
			expectedArns: []string{"[\"invalid-json"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			iamRole := &iammanagerv1alpha1.Iamrole{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tc.annotations,
				},
			}
			
			found, arns := ParseIRSAAnnotation(context.Background(), iamRole)
			
			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedArns, arns)
		})
	}
}

// Note: The GetIdpServerCertThumbprint function requires network access
// and makes real TLS connections, making it challenging to test in a unit test.
// For proper testing, this would need to be mocked or tested in integration tests.
//
// A more testable design would be to inject a client or dialer interface
// that can be mocked in tests.
