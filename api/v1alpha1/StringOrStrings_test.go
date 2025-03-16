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

package v1alpha1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOrStringsMarshalJSON(t *testing.T) {
	testCases := []struct {
		name           string
		input          StringOrStrings
		expectedOutput string
	}{
		{
			name:           "Empty",
			input:          StringOrStrings{},
			expectedOutput: "null",
		},
		{
			name:           "Single string",
			input:          StringOrStrings{"arn:aws:s3:::example-bucket/*"},
			expectedOutput: "[\"arn:aws:s3:::example-bucket/*\"]",
		},
		{
			name:           "Multiple strings",
			input:          StringOrStrings{"arn:aws:s3:::example-bucket/*", "arn:aws:s3:::another-bucket/*"},
			expectedOutput: "[\"arn:aws:s3:::example-bucket/*\",\"arn:aws:s3:::another-bucket/*\"]",
		},
		{
			name:           "Single empty string",
			input:          StringOrStrings{""},
			expectedOutput: "[\"\"]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the input
			result, err := json.Marshal(tc.input)
			require.NoError(t, err)
			
			// Compare the result with expected output
			assert.Equal(t, tc.expectedOutput, string(result))
		})
	}
}

func TestStringOrStringsUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedOutput StringOrStrings
		expectError    bool
	}{
		{
			name:           "Empty array",
			input:          "[]",
			expectedOutput: StringOrStrings{},
			expectError:    false,
		},
		{
			name:           "Single string in array",
			input:          "[\"arn:aws:s3:::example-bucket/*\"]",
			expectedOutput: StringOrStrings{"arn:aws:s3:::example-bucket/*"},
			expectError:    false,
		},
		{
			name:           "Multiple strings in array",
			input:          "[\"arn:aws:s3:::example-bucket/*\", \"arn:aws:s3:::another-bucket/*\"]",
			expectedOutput: StringOrStrings{"arn:aws:s3:::example-bucket/*", "arn:aws:s3:::another-bucket/*"},
			expectError:    false,
		},
		{
			name:           "Single string (not in array)",
			input:          "\"arn:aws:s3:::example-bucket/*\"",
			expectedOutput: StringOrStrings{"arn:aws:s3:::example-bucket/*"},
			expectError:    false,
		},
		{
			name:           "Empty string (not in array)",
			input:          "\"\"",
			expectedOutput: StringOrStrings{""},
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			input:          "{invalid-json",
			expectedOutput: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result StringOrStrings
			
			// Unmarshal the input
			err := json.Unmarshal([]byte(tc.input), &result)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, result)
			}
		})
	}
}

// TestStringOrStringsRoundTrip tests marshaling and then unmarshaling to ensure data integrity
func TestStringOrStringsRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		input StringOrStrings
	}{
		{
			name:  "Empty",
			input: StringOrStrings{},
		},
		{
			name:  "Single string",
			input: StringOrStrings{"arn:aws:s3:::example-bucket/*"},
		},
		{
			name:  "Multiple strings",
			input: StringOrStrings{"arn:aws:s3:::example-bucket/*", "arn:aws:s3:::another-bucket/*"},
		},
		{
			name:  "Single empty string",
			input: StringOrStrings{""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal the input
			marshaled, err := json.Marshal(tc.input)
			require.NoError(t, err)
			
			// Unmarshal back to StringOrStrings
			var unmarshaled StringOrStrings
			err = json.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)
			
			// Special case for empty slices - they become nil slices when unmarshaled from null
			if len(tc.input) == 0 {
				assert.Len(t, unmarshaled, 0, "Expected an empty result")
			} else {
				// Verify the round-trip preserves data
				assert.Equal(t, tc.input, unmarshaled)
			}
		})
	}
}
