/*
Copyright 2025 Keikoproj authors.

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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/controllers" // Changed import path
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/logging"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(iammanagerv1alpha1.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var enableWebhook bool
	var webhookPort int
	var props string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableWebhook, "enable-webhook", false, "Enable webhook server")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "Webhook server port")
	flag.StringVar(&props, "properties", "LOCAL", "Property file location. Default is empty which means use local env.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	// Initialize logging
	logging.New()
	log := logging.Logger(context.Background(), "main", "setup")

	// Set up controller-runtime style logging
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Load configurations
	err := config.LoadProperties(props)
	if err != nil {
		setupLog.Error(err, "Unable to load properties")
		os.Exit(1)
	}

	// Create controller-runtime manager
	webhookServer := webhook.NewServer(webhook.Options{
		Port: webhookPort,
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "955d5887.keikoproj.io",
		WebhookServer:          webhookServer,
	})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	// Create AWS API client
	log.Info("region ", "region", config.Props.AWSRegion())
	iamClient := awsapi.NewIAM(config.Props.AWSRegion())

	// Setup OIDC for IRSA if needed
	if os.Getenv("LOCAL") != "true" {
		if err = handleOIDCSetupForIRSA(context.Background(), iamClient); err != nil {
			setupLog.Error(err, "Error while setting up OIDC for IRSA")
			os.Exit(1)
		}
	}

	// Setup controller with Kubebuilder v4 patterns
	reconciler := &controllers.IamroleReconciler{ // Changed import path
		Client:    mgr.GetClient(),
		IAMClient: iamClient,
		Recorder:  mgr.GetEventRecorderFor("iamrole-controller"),
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Unable to create controller", "controller", "Iamrole")
		os.Exit(1)
	}

	// Setup webhooks if enabled
	iammanagerv1alpha1.NewWClient()
	if config.Props.IsWebHookEnabled() || enableWebhook {
		setupLog.Info("Registering webhook")
		if err = (&iammanagerv1alpha1.Iamrole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Unable to create webhook", "webhook", "Iamrole")
			os.Exit(1)
		}
	}
	// +kubebuilder:scaffold:builder

	// Setup health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}

// handleOIDCSetupForIRSA will be used to setup the OIDC in AWS IAM
func handleOIDCSetupForIRSA(ctx context.Context, iamClient *awsapi.IAM) error {
	log := logging.Logger(ctx, "main", "handleOIDCSetupForIRSA")

	// Creating OIDC provider if config map has an entry
	if config.Props.IsIRSAEnabled() {
		// Fetch cert thumb print
		thumbprint, err := utils.GetIdpServerCertThumbprint(context.Background(), config.Props.OIDCIssuerUrl())
		if err != nil {
			log.Error(err, "unable to get the OIDC IDP server thumbprint")
			return err
		}

		err = iamClient.CreateOIDCProvider(ctx, config.Props.OIDCIssuerUrl(), config.OIDCAudience, thumbprint)
		if err != nil {
			log.Error(err, "unable to setup OIDC with the url", "url", config.Props.OIDCIssuerUrl())
			return err
		}
		log.Info("OIDC provider setup is successfully completed")
	}

	return nil
}
