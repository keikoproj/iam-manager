package k8s

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/keikoproj/iam-manager/pkg/log"
)

//CreateServiceAccount adds the service account
func (c *Client) CreateOrUpdateServiceAccount(ctx context.Context, saName string, ns string, roleARN string, regionalEndpointDisabled bool) error {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "CreateOrUpdateServiceAccount")

	sa := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      saName,
			Namespace: ns,
			Annotations: map[string]string{
				"eks.amazonaws.com/role-arn": roleARN,
			},
		},
	}
	if !regionalEndpointDisabled {
		sa.ObjectMeta.Annotations["eks.amazonaws.com/sts-regional-endpoints"] = "true"
	}

	//_, err := c.cl.CoreV1().ServiceAccounts(ns).Create(sa)
	log.V(1).Info("Service Account creation is in progress")
	err := c.rCl.Create(ctx, sa)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("Failed to create service account %s in namespace %s due to %v", sa.Name, ns, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Service account already exists. Trying to update", "serviceAccount", sa.Name, "namespace", ns)
		err = c.rCl.Update(ctx, sa)
		if err != nil {
			msg := fmt.Sprintf("Failed to update service account %s due to %v", sa.Name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
	}
	log.Info("Service account got created successfully", "serviceAccount", sa.Name, "namespace", ns)
	return nil
}
