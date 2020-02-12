package config

import (
	"context"
	"fmt"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"github.com/keikoproj/iam-manager/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
)

var (
	Props                              *Properties
	IamManagedPermissionBoundaryPolicy = "arn:aws:iam::%s:policy/%s"
)

type Properties struct {
	allowedPolicyAction             []string
	restrictedPolicyResources       []string
	restrictedS3Resources           []string
	awsAccountID                    string
	awsMasterRole                   string
	managedPolicies                 []string
	managedPermissionBoundaryPolicy string
}

func init() {
	log := log.Logger(context.Background(), "internal.config.properties", "init")
	k8sClient, err := k8s.NewK8sClient()
	if err != nil {
		log.Error(err, "unable to create new k8s client")
		panic(err)
	}
	res := k8sClient.GetConfigMap(context.Background(), IamManagerNamespaceName, IamManagerConfigMapName)

	// load properties into a global variable
	LoadProperties(res)
	log.Info("Loaded properties in init func")
}

func LoadProperties(cm *v1.ConfigMap) {
	allowedPolicyAction := strings.Split(cm.Data[propertyIamPolicyWhitelist], separator)
	restrictedPolicyResources := strings.Split(cm.Data[propertyIamPolicyBlacklist], separator)
	restrictedS3Resources := strings.Split(cm.Data[propertyIamPolicyS3Restricted], separator)
	managedPolicies := strings.Split(cm.Data[propertyManagedPolicies], separator)
	awsAccountID := cm.Data[propertyAwsAccountID]
	awsMasterRole := cm.Data[propertyAwsMasterRole]

	managedPermissionBoundaryPolicy := cm.Data[propertyPermissionBoundary]
	if !strings.HasPrefix(managedPermissionBoundaryPolicy, "arn:aws:iam::") {
		managedPermissionBoundaryPolicy = fmt.Sprintf(IamManagedPermissionBoundaryPolicy, awsAccountID, managedPermissionBoundaryPolicy)
	}

	Props = &Properties{
		allowedPolicyAction:             allowedPolicyAction,
		restrictedPolicyResources:       restrictedPolicyResources,
		restrictedS3Resources:           restrictedS3Resources,
		awsAccountID:                    awsAccountID,
		awsMasterRole:                   awsMasterRole,
		managedPolicies:                 managedPolicies,
		managedPermissionBoundaryPolicy: managedPermissionBoundaryPolicy,
	}
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

func (p *Properties) AWSMasterRole() string {
	return p.awsMasterRole
}

func (p *Properties) ManagedPermissionBoundaryPolicy() string {
	return p.managedPermissionBoundaryPolicy
}

func RunConfigMapInformer(ctx context.Context) {
	log := log.Logger(context.Background(), "internal.config.properties", "RunConfigMapInformer")
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
	log := log.Logger(context.Background(), "internal.config.properties", "onUpdate")
	oldCM := old.(*v1.ConfigMap)
	newCM := new.(*v1.ConfigMap)
	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}
	log.Info("Updating config map", "new revision ", newCM.ResourceVersion)
	LoadProperties(newCM)
}
