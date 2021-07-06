package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ServiceAccountRequest struct
type ServiceAccountRequest struct {
	Namespace          string
	IamRoleARN         string
	ServiceAccountName string
}

// EnsureServiceAccount decomposes a ServiceAccountRequest into the right behavior and
// desired target state of a ServiceAccount<->Iamrole mapping, and implements that.
//
// Current functionality is limited - but I am to increase the complexity and control
// of how this mapping works, which and I want to keep as little complexity in the
// CreateOrUpdateServiceAccount function as possible.
func (c *Client) EnsureServiceAccount(ctx context.Context, req ServiceAccountRequest) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "EnsureServiceAccount")
	log = log.WithValues("request", req)

	// TODO: Implement a check in the future to figure out whether or not we _were supposed_ to manage the
	// ServiceAccount at one point, and should now _remove_ the annotation from it or not.

	// First thing, check if the SA exists already. If it does, we will avoid a failed Create call.
	sa, err := c.GetOrCreateServiceAccount(ctx, req.ServiceAccountName, req.Namespace)
	if err != nil {
		log.Error(err, "Unable to get or create ServiceAccount!")
		return nil, err
	}
	log.V(1).Info("Got existing SA", "serviceaccount", sa)

	// If the SA exists, let's check its annotations now. If they match, we don't need to do any
	// more work and we can return!
	currentVal, _ := sa.Annotations[constants.ServiceAccountRoleAnnotation]

	// Does it match? If so, we're done here. Return.
	if currentVal == req.IamRoleARN {
		log.V(1).Info("Existing ServiceAccount looks good")
		return sa, nil
	}

	// Log this out because its a useful datapoint
	log.Info(fmt.Sprintf(
		`Found existing ServiceAccount - but %s annotation
		 does not match expected value.`, constants.ServiceAccountRoleAnnotation))

	// At this point, patch the ServiceAccount to set the annotation that we expect
	sa, err = c.PatchServiceAccountAnnotation(ctx, sa.Name, sa.Namespace, constants.ServiceAccountRoleAnnotation, req.IamRoleARN)
	if err != nil {
		msg := fmt.Sprintf("Failed to update service account %s due to %v", sa.Name, err)
		log.Error(err, msg)
		return nil, errors.New(msg)
	}

	// Everything worked out!
	return sa, nil
}

// GetOrCreateServiceAccount will get an SA and return it, or create it (and return it) if necessary
func (c *Client) GetOrCreateServiceAccount(ctx context.Context, saName string, ns string) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "GetOrCreateServiceAccount")
	log = log.WithValues("name", saName, "namespace", ns)

	// First try to get an existing SA...
	sa, err := c.GetServiceAccount(ctx, saName, ns)
	if sa != nil {
		return sa, nil
	}
	if err != nil {
		log.Info("Did not find existing ServiceAccount, will create one instead")
	}

	// If we got here, then we need to create the SA instead...
	sa, err = c.CreateServiceAccount(ctx, saName, ns)
	if err != nil {
		log.Error(err, "Unable to create ServiceAccount")
		return nil, err
	}
	return sa, nil
}

// GetServiceAccount returns back a ServiceAccount resource if it exists in K8S
func (c *Client) GetServiceAccount(ctx context.Context, saName string, ns string) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "GetServiceAccount")
	log = log.WithValues("name", saName, "namespace", ns)

	sa, err := c.Cl.CoreV1().ServiceAccounts(ns).Get(saName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// CreateServiceAccount returns back a ServiceAccount resource if it exists in K8S
func (c *Client) CreateServiceAccount(ctx context.Context, saName string, ns string) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "CreateServiceAccount")
	log = log.WithValues("name", saName, "namespace", ns)

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: ns,
		},
	}

	sa, err := c.Cl.CoreV1().ServiceAccounts(ns).Create(sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// PatchServiceAccountAnnotation will issue a patch call to the K8S API to update an annotation on a
// particular ServiceAccount. We use a patch here so that we do not need to worry about mutating any
// other state on a ServiceAccount object, nor do we have to worry about any out-of-order update issues.
func (c *Client) PatchServiceAccountAnnotation(ctx context.Context, saName string, ns string, annotation string, value string) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "PatchServiceAccountAnnotation")
	log = log.WithValues("annotation", annotation, "value", value)
	patch := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s": "%s"}}}`, annotation, value))

	log.V(1).Info("Applying patch for ServiceAccount", "patch", string(patch))
	sa, err := c.Cl.CoreV1().ServiceAccounts(ns).Patch(saName, types.StrategicMergePatchType, patch)
	if err != nil {
		return nil, err
	}
	return sa, nil
}
