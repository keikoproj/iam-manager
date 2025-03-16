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

package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"github.com/keikoproj/iam-manager/pkg/logging"
)

// Template field constants
const (
	// IAMRoleSAField is a field for serviceaccount
	IAMRoleSAField = `{{ .ServiceAccount }}`
	// IAMRoleNSField is a field for namespace
	IAMRoleNSField = `{{ .NamespaceName }}`
	// IAMRoleARNField is a field for targetRole
	IAMRoleARNField = `{{ .TargetRoleArn }}`
	// IAMRoleOIDCProviderURLField is a field for OIDC provider URL
	IAMRoleOIDCProviderURLField = `{{ .OIDCProviderURL }}`
	// AccountIDField is a field for AWS account ID
	AccountIDField = `{{ .AccountID }}`
)

var (
	Props           *Properties
	PolicyARNFormat = "arn:aws:iam::%s:policy/%s"
)

type Properties struct {
	client                          *k8s.Client
	local                           bool
	refreshTime                     int
	allowedPolicyAction             []string
	restrictedPolicyResources       []string
	restrictedS3Resources           []string
	awsAccountID                    string
	awsRegion                       string
	managedPolicies                 []string
	managedPermissionBoundaryPolicy string
	clusterName                     string
	clusterOIDCIssuerUrl            string
	defaultTrustPolicy              string
	iamRolePattern                  string
	isWebhookEnabled                string
	maxRolesAllowed                 int
	controllerDesiredFrequency      int
	isIRSAEnabled                   string
	isIRSARegionalEndpointDisabled  string
	defaultMaxSessionDuration       int64
	maxTrustPolicyRoleName          int
	minTrustPolicyRoleName          int
	defaultTrustPolicyRoleNameRe    string
}

func init() {
	if os.Getenv("GO_TEST_MODE") == "true" {
		Props = &Properties{
			local:                           true,
			refreshTime:                     60,
			allowedPolicyAction:             []string{"s3:*", "ec2:*"},
			restrictedPolicyResources:       []string{},
			restrictedS3Resources:           []string{},
			awsAccountID:                    "123456789012",
			awsRegion:                       "us-west-2",
			managedPolicies:                 []string{},
			managedPermissionBoundaryPolicy: "arn:aws:iam::123456789012:policy/test-policy",
			clusterName:                     "test-cluster",
			clusterOIDCIssuerUrl:            "https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED539D4633E53DE1B71EXAMPLE",
			defaultTrustPolicy:              "test-policy",
			defaultMaxSessionDuration:       3600,
			maxTrustPolicyRoleName:          2,
			minTrustPolicyRoleName:          2,
			defaultTrustPolicyRoleNameRe:    ".*",
		}
		klog.Info("Using test configuration")
		return
	}

	if os.Getenv("LOCAL") != "" {
		err := LoadProperties("LOCAL")
		if err != nil {
			klog.Fatalf("Error loading properties: %v", err)
		}
	} else {
		// This is done because CRDs will not be available when the controller is first deployed
		time.Sleep(5 * time.Second)
		err := LoadProperties("")
		if err != nil {
			klog.Fatalf("Error loading properties: %v", err)
		}
	}
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	log := logging.Logger(context.Background(), "internal.config.properties", "LoadProperties")

	// for local testing
	if env != "" {
		Props = &Properties{
			allowedPolicyAction:             strings.Split(os.Getenv("ALLOWED_POLICY_ACTION"), separator),
			restrictedPolicyResources:       strings.Split(os.Getenv("RESTRICTED_POLICY_RESOURCES"), separator),
			restrictedS3Resources:           strings.Split(os.Getenv("RESTRICTED_S3_RESOURCES"), separator),
			awsAccountID:                    os.Getenv("AWS_ACCOUNT_ID"),
			managedPolicies:                 strings.Split(os.Getenv("MANAGED_POLICIES"), separator),
			managedPermissionBoundaryPolicy: os.Getenv("MANAGED_PERMISSION_BOUNDARY_POLICY"),
			awsRegion:                       os.Getenv("AWS_REGION"),
			isWebhookEnabled:                os.Getenv("ENABLE_WEBHOOK"),
			clusterName:                     os.Getenv("CLUSTER_NAME"),
			clusterOIDCIssuerUrl:            os.Getenv("CLUSTER_OIDC_ISSUER_URL"),
			defaultTrustPolicy:              os.Getenv("DEFAULT_TRUST_POLICY"),
			iamRolePattern:                  os.Getenv("IAM_ROLE_PATTERN"),
			isIRSARegionalEndpointDisabled:  os.Getenv("IRSA_REGIONAL_ENDPOINT_DISABLED"),
		}
		return nil
	}

	if len(cm) == 0 || cm[0] == nil {
		log.Error(fmt.Errorf("config map cannot be nil"), "config map cannot be nil")
		return fmt.Errorf("config map cannot be nil")
	}

	allowedPolicyAction := strings.Split(cm[0].Data[propertyIamPolicyWhitelist], separator)
	restrictedPolicyResources := strings.Split(cm[0].Data[propertyIamPolicyBlacklist], separator)
	restrictedS3Resources := strings.Split(cm[0].Data[propertyIamPolicyS3Restricted], separator)
	clusterName := cm[0].Data[propertyClusterName]
	defaultTrustPolicy := cm[0].Data[propertyDefaultTrustPolicy]
	Props = &Properties{
		allowedPolicyAction:       allowedPolicyAction,
		restrictedPolicyResources: restrictedPolicyResources,
		restrictedS3Resources:     restrictedS3Resources,
		clusterName:               clusterName,
		defaultTrustPolicy:        defaultTrustPolicy,
	}

	//Defaults
	isWebhook := cm[0].Data[propertyWebhookEnabled]
	if isWebhook == "true" {
		Props.isWebhookEnabled = "true"
	} else {
		Props.isWebhookEnabled = "false"
	}

	awsRegion := cm[0].Data[propertyAwsRegion]
	if awsRegion != "" {
		Props.awsRegion = awsRegion
	} else {
		Props.awsRegion = "us-west-2"
	}

	maxRolesAllowed := cm[0].Data[propertyMaxIamRoles]
	if maxRolesAllowed != "" {
		maxRolesAllowed, err := strconv.Atoi(maxRolesAllowed)
		if err != nil {
			return err
		}
		Props.maxRolesAllowed = maxRolesAllowed
	} else {
		Props.maxRolesAllowed = 1
	}

	controllerDesiredFreq := cm[0].Data[propertyDesiredStateFrequency]
	if controllerDesiredFreq != "" {
		controllerDesiredFreq, err := strconv.Atoi(controllerDesiredFreq)
		if err != nil {
			return err
		}
		Props.controllerDesiredFrequency = controllerDesiredFreq
	} else {
		Props.controllerDesiredFrequency = 1800
	}

	awsAccountID := cm[0].Data[propertyAWSAccountID]
	// Load AWS account ID
	if Props.awsAccountID == "" && awsAccountID == "" {
		awsAccountID, err := awsapi.NewSTS(Props.awsRegion).GetAccountID(context.Background())
		if err != nil {
			return err
		}
		Props.awsAccountID = awsAccountID
	} else {
		Props.awsAccountID = awsAccountID
	}

	iamRolePattern := cm[0].Data[propertyIamRolePattern]
	if iamRolePattern == "" {
		Props.iamRolePattern = "k8s-{{ .ObjectMeta.Name }}"
	} else {
		Props.iamRolePattern = iamRolePattern
	}

	managedPermissionBoundaryPolicyArn := cm[0].Data[propertyPermissionBoundary]

	if managedPermissionBoundaryPolicyArn == "" {
		managedPermissionBoundaryPolicyArn = fmt.Sprintf(PolicyARNFormat, awsAccountID, "k8s-iam-manager-cluster-permission-boundary")
	}

	if !strings.HasPrefix(managedPermissionBoundaryPolicyArn, "arn:aws:iam::") {
		managedPermissionBoundaryPolicyArn = fmt.Sprintf(PolicyARNFormat, awsAccountID, managedPermissionBoundaryPolicyArn)
	}

	Props.managedPermissionBoundaryPolicy = managedPermissionBoundaryPolicyArn

	managedPolicies := strings.Split(cm[0].Data[propertyManagedPolicies], separator)
	for i := range managedPolicies {
		if managedPolicies[i] != "" {
			if !strings.HasPrefix(managedPolicies[i], "arn:aws:iam::") {
				managedPolicies[i] = fmt.Sprintf(PolicyARNFormat, awsAccountID, managedPolicies[i])
			}
		}
	}
	Props.managedPolicies = managedPolicies

	isIRSAEnabled := cm[0].Data[propertyIRSAEnabled]
	if isIRSAEnabled == "true" {
		Props.isIRSAEnabled = "true"
	} else {
		Props.isIRSAEnabled = "false"
	}

	oidcUrl := cm[0].Data[propertyK8sClusterOIDCIssuerUrl]
	if isIRSAEnabled == "true" && oidcUrl == "" {
		if clusterName == "" {
			return fmt.Errorf("cluster name must be provided when IRSA is enabled to retrieve the OIDC url")
		}
		//call EKS describe cluster and get the OIDC URL
		res, err := awsapi.NewEKS(Props.awsRegion).DescribeCluster(context.Background(), clusterName)
		if err != nil {
			return err
		}
		oidcUrl = *res.Cluster.Identity.Oidc.Issuer
	}
	Props.clusterOIDCIssuerUrl = oidcUrl

	isIRSARegionalEndpointDisabled := cm[0].Data[propertyIRSARegionalEndpointDisabled]
	if isIRSARegionalEndpointDisabled == "true" {
		Props.isIRSARegionalEndpointDisabled = "true"
	} else {
		Props.isIRSARegionalEndpointDisabled = "false"
	}

	return nil
}

func (p *Properties) AllowedPolicyAction() []string {
	return p.allowedPolicyAction
}

func (p *Properties) RestrictedPolicyResources() []string {
	return p.restrictedPolicyResources
}

func (p *Properties) RestrictedS3Resources() []string {
	return p.restrictedS3Resources
}

func (p *Properties) ManagedPolicies() []string {
	return p.managedPolicies
}

func (p *Properties) AWSAccountID() string {
	return p.awsAccountID
}

func (p *Properties) ManagedPermissionBoundaryPolicy() string {
	return p.managedPermissionBoundaryPolicy
}

func (p *Properties) AWSRegion() string {
	return p.awsRegion
}

func (p *Properties) IsWebHookEnabled() bool {
	resp := false
	if p.isWebhookEnabled == "true" {
		resp = true
	}
	return resp
}

func (p *Properties) IamRolePattern() string {
	return p.iamRolePattern
}

func (p *Properties) MaxRolesAllowed() int {
	return p.maxRolesAllowed
}

func (p *Properties) ControllerDesiredFrequency() int {
	return p.controllerDesiredFrequency
}

func (p *Properties) IsIRSAEnabled() bool {
	resp := false
	if p.isIRSAEnabled == "true" {
		resp = true
	}
	return resp
}

func (p *Properties) IsIRSARegionalEndpointDisabled() bool {
	resp := false
	if p.isIRSARegionalEndpointDisabled == "true" {
		resp = true
	}
	return resp
}

func (p *Properties) ClusterName() string {
	return p.clusterName
}

func (p *Properties) OIDCIssuerUrl() string {
	return p.clusterOIDCIssuerUrl
}

func (p *Properties) DefaultTrustPolicy() string {
	return p.defaultTrustPolicy
}

// RunConfigMapInformer runs the config map informer to watch for changes
func RunConfigMapInformer(ctx context.Context) {
	log := logging.Logger(ctx, "internal.config.properties", "RunConfigMapInformer")
	log.Info("Starting config map informer")

	// Get the client directly instead of using a helper
	client := k8s.NewK8sClientDoOrDie()

	// Set up a standard informer using controller-runtime
	listOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = fmt.Sprintf("metadata.name=%s", IamManagerConfigMapName)
	}

	cmInformer := cache.NewFilteredListWatchFromClient(
		client.ClientInterface().CoreV1().RESTClient(),
		"configmaps",
		IamManagerNamespaceName,
		listOptions,
	)

	_, controller := cache.NewInformer(
		cmInformer,
		&v1.ConfigMap{},
		time.Hour*24,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { updateProperties(nil, obj) },
			UpdateFunc: func(old, new interface{}) { updateProperties(old, new) },
			DeleteFunc: func(obj interface{}) { /* No action needed on delete */ },
		},
	)

	go controller.Run(ctx.Done())
}

func updateProperties(old, new interface{}) {
	log := logging.Logger(context.Background(), "internal.config.properties", "updateProperties")
	oldCM := old.(*v1.ConfigMap)
	newCM := new.(*v1.ConfigMap)
	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}
	log.Info("Updating config map", "new revision ", newCM.ResourceVersion)
	err := LoadProperties("", newCM)
	if err != nil {
		log.Error(err, "failed to update config map")
	}
}
