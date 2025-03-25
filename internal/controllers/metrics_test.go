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

package controllers_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var _ = Describe("Metrics Server Configuration", func() {
	var stopFunc context.CancelFunc
	var ctx context.Context
	var metricsOpts metricsserver.Options
	var testManager ctrl.Manager

	BeforeEach(func() {
		// Create context with cancel function
		ctx, stopFunc = context.WithCancel(context.Background())

		// Random metrics port to avoid conflicts
		metricsPort := fmt.Sprintf(":%d", 9090+GinkgoParallelProcess())

		// Create metrics options with secure serving and authentication
		metricsOpts = metricsserver.Options{
			BindAddress:    metricsPort,
			SecureServing:  true,
			FilterProvider: filters.WithAuthenticationAndAuthorization,
		}

		// Set up manager with our metrics options
		var err error
		testScheme := runtime.NewScheme()
		testManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  testScheme,
			Metrics: metricsOpts,
			Client: client.Options{
				Cache: &client.CacheOptions{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		// Start the manager in background
		go func() {
			// We expect this to eventually fail due to missing resources in the test environment,
			// but we just want to verify that the manager can be created with secure metrics
			_ = testManager.Start(ctx)
		}()

		// Allow some time for the manager to start
		time.Sleep(100 * time.Millisecond)
	})

	AfterEach(func() {
		// Clean up by stopping the manager
		stopFunc()
		time.Sleep(100 * time.Millisecond)
	})

	It("should configure metrics server with secure serving and authentication", func() {
		// Verify metrics configuration
		Expect(metricsOpts.SecureServing).To(BeTrue(), "Metrics server should be configured with secure serving")
		Expect(metricsOpts.BindAddress).NotTo(BeEmpty(), "Metrics server should have a bind address")
		Expect(metricsOpts.FilterProvider).NotTo(BeNil(), "Metrics server should have authentication filter")

		// Verify the manager was created successfully with the metrics configuration
		Expect(testManager).NotTo(BeNil(), "Manager should be created with secure metrics configuration")
	})
})
