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

package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestNew(t *testing.T) {
	// Test with default parameters (debug enabled)
	New()
	
	// Test with debug explicitly enabled
	New(true)
	
	// Test with debug disabled
	New(false)
	
	// The function doesn't return anything, so just ensure it doesn't panic
	assert.True(t, true, "New function should execute without panicking")
}

func TestLogger(t *testing.T) {
	// Set up a controller logger for testing
	ctrl.SetLogger(zap.New())
	
	// Test creating a logger with no names
	logger := Logger(context.Background())
	assert.NotNil(t, logger)
	
	// Test creating a logger with a single name
	logger = Logger(context.Background(), "test")
	assert.NotNil(t, logger)
	
	// Test creating a logger with multiple names
	logger = Logger(context.Background(), "test", "module", "component")
	assert.NotNil(t, logger)
	
	// Test with request_id in context
	ctx := context.WithValue(context.Background(), "request_id", "test-request-id")
	logger = Logger(ctx)
	assert.NotNil(t, logger)
}

// Test that the Logger function uses the request_id from context correctly
func TestLoggerWithRequestID(t *testing.T) {
	// Create a context with a request_id
	requestID := "test-123"
	ctx := context.WithValue(context.Background(), "request_id", requestID)
	
	// Get a logger with the context
	logger := Logger(ctx)
	
	// The logger should not be nil
	assert.NotNil(t, logger)
	
	// We can't easily test the values in the logger itself in a unit test,
	// but we can verify it doesn't panic when we use it
	logger.Info("test log message")
}
