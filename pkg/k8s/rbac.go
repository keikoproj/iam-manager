package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//ServiceAcountRequest struct
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
func (c *Client) EnsureServiceAccount(ctx context.Context, req ServiceAccountRequest) error {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "EnsureServiceAccount")
	log = log.WithValues("request", req)

	// TODO: Implement a check in the future to figure out whether or not we _were supposed_ to manage the
	// ServiceAccount at one point, and should now _remove_ the annotation from it or not.

	// First thing, check if the SA exists already. If it does, we will avoid a failed Create call.
	sa, _ := c.GetServiceAccount(ctx, req.ServiceAccountName, req.Namespace)
	log.V(1).Info("Got existing SA", "serviceaccount", sa)

	// If the SA exists, let's check its annotations now. If they match, we don't need to do any
	// more work and we can return!
	if sa != nil {
		// Get the annotation value. Check it against the desired Role ARN.
		currentVal, _ := sa.Annotations[constants.ServiceAccountRoleAnnotation]

		// Does it match? If so, we're done here. Return.
		if currentVal == req.IamRoleARN {
			log.V(1).Info("Existing ServiceAccount looks good")
			return nil
		}

		// Log this out because its a useful datapoint
		log.Info(fmt.Sprintf(
			`Found existing ServiceAccount - but %s annotation
			 does not match expected value.`, constants.ServiceAccountRoleAnnotation))
	}

	// At this point, we know the user wants us to ensure the ServiceAccount has the right Annotation applied to it.
	err := c.CreateOrUpdateServiceAccount(ctx, req.ServiceAccountName, req.Namespace, req.IamRoleARN)
	if err != nil {
		return err
	}

	// Everything worked out!
	return nil
}

//CreateOrUpdateServiceAccount adds the service account or updates the account if it already exists.
func (c *Client) CreateOrUpdateServiceAccount(ctx context.Context, saName string, ns string, roleARN string) error {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "CreateOrUpdateServiceAccount")

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: ns,
			Annotations: map[string]string{
				constants.ServiceAccountRoleAnnotation: roleARN,
			},
		},
	}

	log.V(1).Info("Service Account creation is in progress")
	err := c.rCl.Create(ctx, sa)
	if err != nil {
		// If the error was _not_ an already-exists issue, then fail quickly.
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("Failed to create service account %s in namespace %s due to %v", sa.Name, ns, err)
			log.Error(err, msg)
			return errors.New(msg)
		}

		log.Info("Service account already exists. Trying to update", "serviceAccount", sa.Name, "namespace", ns)
		//err = c.rCl.Update(ctx, sa)
		err = c.PatchServiceAccountAnnotation(ctx, saName, ns, constants.ServiceAccountRoleAnnotation, roleARN)
		if err != nil {
			msg := fmt.Sprintf("Failed to update service account %s due to %v", sa.Name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
	}
	log.Info("Service account got created successfully", "serviceAccount", sa.Name, "namespace", ns)
	return nil
}

// GetServiceAccount returns back a ServiceAccount resource if it exists in K8S
func (c *Client) GetServiceAccount(ctx context.Context, saName string, ns string) (*corev1.ServiceAccount, error) {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "GetServiceAccount")
	log = log.WithValues("name", saName, "namespace", ns)

	sa := &corev1.ServiceAccount{}
	err := c.rCl.Get(ctx, client.ObjectKey{Namespace: ns, Name: saName}, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// PatchServiceAccountAnnotation will issue a patch call to the K8S API to update an annotation on a
// particular ServiceAccount. We use a patch here so that we do not need to worry about mutating any
// other state on a ServiceAccount object, nor do we have to worry about any out-of-order update issues.
func (c *Client) PatchServiceAccountAnnotation(ctx context.Context, saName string, ns string, annotation string, value string) error {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "PatchServiceAccountAnnotation")
	log = log.WithValues("annotation", annotation, "value", value)
	patch := []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s": "%s"}}}`, annotation, value))

	log.V(1).Info("Applying patch for ServiceAccount", "patch", string(patch))
	err := c.rCl.Patch(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      saName,
		},
	}, client.RawPatch(types.StrategicMergePatchType, patch))
	if err != nil {
		return err
	}
	return nil
}
