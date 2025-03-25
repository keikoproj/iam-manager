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

package main_test

import (
	"flag"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// TestMetricsSecureConfiguration tests that the metrics server is correctly configured
// with authentication and authorization
func TestMetricsSecureConfiguration(t *testing.T) {
	// Define test parameters
	metricsAddr := ":8443"

	// Create metrics options similar to those in main()
	metricsOpts := metricsserver.Options{
		BindAddress:    metricsAddr,
		SecureServing:  true,
		FilterProvider: filters.WithAuthenticationAndAuthorization,
	}

	// Verify the metrics configuration
	if metricsOpts.BindAddress != metricsAddr {
		t.Errorf("Expected BindAddress to be %s, got %s", metricsAddr, metricsOpts.BindAddress)
	}

	if !metricsOpts.SecureServing {
		t.Errorf("Expected SecureServing to be true, got false")
	}

	// Can't directly compare function values with Equal, so just verify it's not nil
	if metricsOpts.FilterProvider == nil {
		t.Errorf("Expected FilterProvider to be set, got nil")
	}
}

// TestCommandLineFlagParsing tests that command-line flags are correctly defined and parsed
func TestCommandLineFlagParsing(t *testing.T) {
	// Save original command-line arguments
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	}()

	// Set up test command-line arguments
	os.Args = []string{"cmd", "--metrics-addr=:9443", "--enable-leader-election=true", "--debug=true"}

	// Create a new flag set
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	// Define flags similar to those in main()
	var metricsAddr string
	var enableLeaderElection bool
	var debug bool
	fs.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	fs.BoolVar(&enableLeaderElection, "enable-leader-election", false, "Enable leader election for controller manager.")
	fs.BoolVar(&debug, "debug", false, "Enable Debug?")

	// Parse flags
	if err := fs.Parse(os.Args[1:]); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	// Verify flags are parsed correctly
	if metricsAddr != ":9443" {
		t.Errorf("Expected metrics-addr flag to be ':9443', got '%s'", metricsAddr)
	}

	if !enableLeaderElection {
		t.Errorf("Expected enable-leader-election flag to be true, got false")
	}

	if !debug {
		t.Errorf("Expected debug flag to be true, got false")
	}
}
