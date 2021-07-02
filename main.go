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
	"flag"
	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/controllers"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"github.com/keikoproj/iam-manager/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
	//setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = iammanagerv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var debug bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&debug, "debug", false, "Enable Debug?")
	flag.Parse()

	log.New()
	log := log.Logger(context.Background(), "main", "setup")

	go config.RunConfigMapInformer(context.Background())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
		LeaderElectionID:   "controller-leader-election-helper",
	})

	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.V(1).Info("Setting up reconciler with manager")
	log.Info("region ", "region", config.Props.AWSRegion())

	iamClient := awsapi.NewIAM(config.Props.AWSRegion())
	if err := handleOIDCSetupForIRSA(context.Background(), iamClient); err != nil {
		log.Error(err, "unable to complete/verify oidc setup for IRSA")
	}

	if err = (&controllers.IamroleReconciler{
		Client:    mgr.GetClient(),
		IAMClient: iamClient,
		Recorder:  k8s.NewK8sClientDoOrDie().SetUpEventHandler(context.Background()),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Iamrole")
		os.Exit(1)
	}

	//Get the client
	iammanagerv1alpha1.NewWClient()
	if config.Props.IsWebHookEnabled() {
		log.Info("Registering webhook")
		if err = (&iammanagerv1alpha1.Iamrole{}).SetupWebhookWithManager(mgr); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "Iamrole")
			os.Exit(1)
		}
	}

	// +kubebuilder:scaffold:builder

	log.Info("Registering controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

//handleOIDCSetupForIRSA will be used to setup the OIDC in AWS IAM
func handleOIDCSetupForIRSA(ctx context.Context, iamClient *awsapi.IAM) error {
	log := log.Logger(ctx, "main", "handleOIDCSetupForIRSA")

	//Creating OIDC provider if config map has an entry

	if config.Props.IsIRSAEnabled() {
		//Fetch cert thumb print
		thumbprint, err := utils.GetIdpServerCertThumbprint(context.Background(), config.Props.OIDCIssuerUrl())
		if err != nil {
			log.Error(err, "unable to get the OIDC IDP server thumbprint")
			return err
		}

		err = iamClient.CreateOIDCProvider(ctx, config.Props.OIDCIssuerUrl(), constants.OIDCAudience, thumbprint)
		if err != nil {
			log.Error(err, "unable to setup OIDC with the url", "url", config.Props.OIDCIssuerUrl())
			return err
		}
		log.Info("OIDC provider setup is successfully completed")

	}

	return nil
}
