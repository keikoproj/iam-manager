package config

import (
	"context"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"github.com/keikoproj/iam-manager/pkg/log"
)

func init() {
	log := log.Logger(context.Background(), "init", "LoadProperties")
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
