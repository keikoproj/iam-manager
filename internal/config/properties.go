package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"github.com/keikoproj/iam-manager/pkg/logging"
)

var (
	Props           *Properties
	PolicyARNFormat = "arn:aws:iam::%s:policy/%s"
)

type Properties struct {
	allowedPolicyAction             []string
	restrictedPolicyResources       []string
	restrictedS3Resources           []string
	awsAccountID                    string
	managedPolicies                 []string
	managedPermissionBoundaryPolicy string
	awsRegion                       string
	isWebhookEnabled                string
	maxRolesAllowed                 int
	controllerDesiredFrequency      int
	clusterName                     string
	isIRSAEnabled                   string
	clusterOIDCIssuerUrl            string
	defaultTrustPolicy              string
	iamRolePattern                  string
	isIRSARegionalEndpointDisabled  string
}

func init() {
	var err error
	log := logging.Logger(context.Background(), "internal.config.properties", "init")

	// Testing environment check - use environment variable to detect test mode
	if os.Getenv("GO_TEST_MODE") == "true" || os.Getenv("LOCAL") == "true" {
		err := LoadProperties("LOCAL")
		if err != nil {
			log.Error(err, "failed to load properties for local environment")
			return
		}
		log.Info("Loaded properties in init func for tests")
		return
	}

	// Production mode - try to connect to Kubernetes
	k8sClient, err := k8s.NewK8sClient()
	if err != nil {
		log.Error(err, "unable to create new k8s client")
		// Don't panic in tests, just log the error
		if os.Getenv("FAIL_ON_K8S_ERROR") == "true" {
			panic(err)
		}
		log.Info("Couldn't connect to Kubernetes, using default values")
		return
	}
	res := k8sClient.GetConfigMap(context.Background(), IamManagerNamespaceName, IamManagerConfigMapName)

	// load properties into a global variable
	err = LoadProperties("", res)
	if err != nil {
		log.Error(err, "failed to load properties")
		// Don't panic in tests, just log the error
		if os.Getenv("FAIL_ON_K8S_ERROR") == "true" {
			panic(err)
		}
		log.Info("Failed to load properties, using defaults")
		return
	}
	log.Info("Loaded properties in init func")
}

func LoadProperties(env string, cm ...*v1.ConfigMap) error {
	log := logging.Logger(context.Background(), "internal.config.properties", "LoadProperties")

	if env == "LOCAL" {
		log.Info("Loading local properties")
		Props = &Properties{
			allowedPolicyAction:             []string{"ec2:*", "elasticloadbalancing:*", "cloudwatch:*", "logs:*", "sqs:*", "sns:*", "s3:*", "cloudfront:*", "rds:*", "dynamodb:*", "route53:*"},
			restrictedPolicyResources:       []string{"RESRICTED-RESOURCE-VALUES-SHOULD-COME-FROM-CONFIGMAP"},
			restrictedS3Resources:           []string{"RESRICTED-S3-RESOURCE-VALUES-SHOULD-COME-FROM-CONFIGMAP"},
			awsAccountID:                    getEnvOrDefault("AWS_ACCOUNT_ID", "ACCOUNT-ID-SHOULD-COME-FROM-CONFIGMAP"),
			managedPolicies:                 []string{},
			managedPermissionBoundaryPolicy: getEnvOrDefault("PERMISSION_BOUNDARY_POLICY", "MANAGED-PERMISSION-BOUNDARY-POLICY-SHOULD-COME-FROM-CONFIGMAP"),
			awsRegion:                       getEnvOrDefault("AWS_REGION", "us-west-2"),
			isWebhookEnabled:                getEnvOrDefault("WEBHOOK_ENABLED", "true"),
			maxRolesAllowed:                 10,
			controllerDesiredFrequency:      3600,
			clusterName:                     getEnvOrDefault("CLUSTER_NAME", "local-cluster"),
			isIRSAEnabled:                   getEnvOrDefault("IRSA_ENABLED", "false"),
			clusterOIDCIssuerUrl:            getEnvOrDefault("OIDC_ISSUER_URL", ""),
			defaultTrustPolicy:              getEnvOrDefault("DEFAULT_TRUST_POLICY", ""),
			iamRolePattern:                  getEnvOrDefault("IAM_ROLE_PATTERN", ""),
			isIRSARegionalEndpointDisabled:  getEnvOrDefault("IRSA_REGIONAL_ENDPOINT_DISABLED", "false"),
		}
		return nil
	}

	// If no cm provided, meaning the caller explicitly sent nil cm, use the default properties
	if len(cm) == 0 || cm[0] == nil {
		log.Info("No configmap provided, using default properties")
		Props = &Properties{}
		return nil
	}

	props := Properties{}
	allowedPolicyAction := strings.Split(cm[0].Data[propertyIamPolicyWhitelist], separator)
	restrictedPolicyResources := strings.Split(cm[0].Data[propertyIamPolicyBlacklist], separator)
	restrictedS3Resources := strings.Split(cm[0].Data[propertyIamPolicyS3Restricted], separator)
	clusterName := cm[0].Data[propertyClusterName]
	defaultTrustPolicy := cm[0].Data[propertyDefaultTrustPolicy]
	props = Properties{
		allowedPolicyAction:       allowedPolicyAction,
		restrictedPolicyResources: restrictedPolicyResources,
		restrictedS3Resources:     restrictedS3Resources,
		clusterName:               clusterName,
		defaultTrustPolicy:        defaultTrustPolicy,
	}

	//Defaults
	isWebhook := cm[0].Data[propertyWebhookEnabled]
	if isWebhook == "true" {
		props.isWebhookEnabled = "true"
	} else {
		props.isWebhookEnabled = "false"
	}

	awsRegion := cm[0].Data[propertyAwsRegion]
	if awsRegion != "" {
		props.awsRegion = awsRegion
	} else {
		props.awsRegion = "us-west-2"
	}

	maxRolesAllowed := cm[0].Data[propertyMaxIamRoles]
	if maxRolesAllowed != "" {
		maxRolesAllowed, err := strconv.Atoi(maxRolesAllowed)
		if err != nil {
			return err
		}
		props.maxRolesAllowed = maxRolesAllowed
	} else {
		props.maxRolesAllowed = 1
	}

	controllerDesiredFreq := cm[0].Data[propertyDesiredStateFrequency]
	if controllerDesiredFreq != "" {
		controllerDesiredFreq, err := strconv.Atoi(controllerDesiredFreq)
		if err != nil {
			return err
		}
		props.controllerDesiredFrequency = controllerDesiredFreq
	} else {
		props.controllerDesiredFrequency = 1800
	}

	awsAccountID := cm[0].Data[propertyAWSAccountID]
	// Load AWS account ID
	if props.awsAccountID == "" && awsAccountID == "" {
		awsAccountID, err := awsapi.NewSTS(props.awsRegion).GetAccountID(context.Background())
		if err != nil {
			return err
		}
		props.awsAccountID = awsAccountID
	} else {
		props.awsAccountID = awsAccountID
	}

	iamRolePattern := cm[0].Data[propertyIamRolePattern]
	if iamRolePattern == "" {
		props.iamRolePattern = "k8s-{{ .ObjectMeta.Name }}"
	} else {
		props.iamRolePattern = iamRolePattern
	}

	managedPermissionBoundaryPolicyArn := cm[0].Data[propertyPermissionBoundary]

	if managedPermissionBoundaryPolicyArn == "" {
		managedPermissionBoundaryPolicyArn = fmt.Sprintf(PolicyARNFormat, awsAccountID, "k8s-iam-manager-cluster-permission-boundary")
	}

	if !strings.HasPrefix(managedPermissionBoundaryPolicyArn, "arn:aws:iam::") {
		managedPermissionBoundaryPolicyArn = fmt.Sprintf(PolicyARNFormat, awsAccountID, managedPermissionBoundaryPolicyArn)
	}

	props.managedPermissionBoundaryPolicy = managedPermissionBoundaryPolicyArn

	managedPolicies := strings.Split(cm[0].Data[propertyManagedPolicies], separator)
	for i := range managedPolicies {
		if managedPolicies[i] != "" {
			if !strings.HasPrefix(managedPolicies[i], "arn:aws:iam::") {
				managedPolicies[i] = fmt.Sprintf(PolicyARNFormat, awsAccountID, managedPolicies[i])
			}
		}
	}
	props.managedPolicies = managedPolicies

	isIRSAEnabled := cm[0].Data[propertyIRSAEnabled]
	if isIRSAEnabled == "true" {
		props.isIRSAEnabled = "true"
	} else {
		props.isIRSAEnabled = "false"
	}

	oidcUrl := cm[0].Data[propertyK8sClusterOIDCIssuerUrl]
	if isIRSAEnabled == "true" && oidcUrl == "" {
		if clusterName == "" {
			return fmt.Errorf("cluster name must be provided when IRSA is enabled to retrieve the OIDC url")
		}
		//call EKS describe cluster and get the OIDC URL
		res, err := awsapi.NewEKS(props.awsRegion).DescribeCluster(context.Background(), clusterName)
		if err != nil {
			return err
		}
		oidcUrl = *res.Cluster.Identity.Oidc.Issuer
	}
	props.clusterOIDCIssuerUrl = oidcUrl

	isIRSARegionalEndpointDisabled := cm[0].Data[propertyIRSARegionalEndpointDisabled]
	if isIRSARegionalEndpointDisabled == "true" {
		props.isIRSARegionalEndpointDisabled = "true"
	} else {
		props.isIRSARegionalEndpointDisabled = "false"
	}

	Props = &props
	log.Info("Loaded properties from configmap")
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

func RunConfigMapInformer(ctx context.Context) {
	log := logging.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
	cmInformer := k8s.GetConfigMapInformer(ctx, IamManagerNamespaceName, IamManagerConfigMapName)
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: updateProperties,
	},
	)
	log.Info("Starting config map informer")
	cmInformer.Run(ctx.Done())
	log.Info("Cancelling config map informer")
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

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
