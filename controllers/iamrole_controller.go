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

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/log"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"net/url"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
)

const (
	maxRetryCount = 100
	finalizerName = "iamrole.finalizers.iammanager.keikoproj.io"
	requestId     = "request_id"
)

// IamroleReconciler reconciles a Iamrole object
type IamroleReconciler struct {
	client.Client
	IAMClient *awsapi.IAM
}

// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles/status,verbs=get;update;patch

func (r *IamroleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "iamrole_controller", "Reconcile")
	log.WithValues("iamrole", req.NamespacedName)
	log.Info("Start of the request")
	//Get the resource
	var iamRole iammanagerv1alpha1.Iamrole

	if err := r.Get(ctx, req.NamespacedName, &iamRole); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace)
	// Isit being deleted?
	if iamRole.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		//Lets check if this is very first time use case
		if !containsString(iamRole.ObjectMeta.Finalizers, finalizerName) {
			log.Info("New iamrole resource. Adding the finalizer", "finalizer", finalizerName)
			iamRole.ObjectMeta.Finalizers = append(iamRole.ObjectMeta.Finalizers, finalizerName)
			r.UpdateMeta(ctx, &iamRole)
		}
		return r.HandleReconcile(ctx, &iamRole)

	} else {
		//oh oh.. This is delete use case
		//Lets make sure to clean up the iam role
		if iamRole.Status.RetryCount != 0 {
			iamRole.Status.RetryCount = iamRole.Status.RetryCount + 1
		}
		log.Info("Iamrole delete request")
		r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteInprogress)
		if err := r.IAMClient.DeleteRole(ctx, roleName); err != nil {
			log.Error(err, "Unable to delete the role")
			//i got to fix this
			r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteError)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		//r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteComplete)

		// Ok. Lets delete the finalizer so controller can delete the custom object
		log.Info("Removing finalizer from Iamrole")
		iamRole.ObjectMeta.Finalizers = removeString(iamRole.ObjectMeta.Finalizers, finalizerName)
		r.UpdateMeta(ctx, &iamRole)
		log.Info("Successfully deleted iam role")
	}

	return ctrl.Result{}, nil
}

var roleTrust = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::000065563193:role/masters.ops-prim-ppd.cluster.k8s.local"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

//HandleReconcile function handles all the reconcile
func (r *IamroleReconciler) HandleReconcile(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "HandleReconcile")
	log.WithValues("iamrole", iamRole.Name)
	log.Info("state of the custom resource ", "state", iamRole.Status.State)
	role, _ := json.Marshal(iamRole.Spec.PolicyDocument)
	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace)
	input := awsapi.IAMRoleRequest{
		Name:             roleName,
		PolicyName:       config.InlinePolicyName,
		Description:      "#DO NOT DELETE#. Managed by iam-manager",
		SessionDuration:  3600,
		TrustPolicy:      roleTrust,
		PermissionPolicy: string(role),
	}

	switch iamRole.Status.State {
	case "":
		// This is first time use case so lets update the status with create in progress
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{}, iammanagerv1alpha1.CreateInProgress)
		_, err := r.IAMClient.CreateRole(ctx, input)
		if err != nil {
			log.Error(err, "error in creating a role")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1}, iammanagerv1alpha1.CreateError)
			return ctrl.Result{RequeueAfter: 30 * time.Duration(iamRole.Status.RetryCount+1) * time.Second}, nil
		}
		//Ok. Successful!!
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName}, iammanagerv1alpha1.Ready)

	case iammanagerv1alpha1.CreateInProgress, iammanagerv1alpha1.UpdateInprogress, iammanagerv1alpha1.DeleteInprogress:
		//Nothing to do.. previous operation is in progress
		log.Info("reconcile is in progress for the given role")
		return ctrl.Result{}, nil

	case iammanagerv1alpha1.CreateError, iammanagerv1alpha1.UpdateError:
		// Check the retry count before proceed further
		if iamRole.Status.RetryCount > maxRetryCount {
			// if it exceeds number of retries, lets just keep it in error state and do not retry anymore
			return ctrl.Result{}, nil
		}

		//update the status and retry again
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{}, iammanagerv1alpha1.UpdateInprogress)
		_, err := r.IAMClient.CreateRole(ctx, input)
		if err != nil {
			// If we are still getting error in retrial, increase the retryCount and requeue it.
			log.Error(err, "error in creating/updating a role")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1}, iammanagerv1alpha1.CreateError)
			return ctrl.Result{RequeueAfter: 30 * time.Duration(iamRole.Status.RetryCount+1) * time.Second}, nil
		}

		//Ok. Successful!!
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName}, iammanagerv1alpha1.Ready)

	case iammanagerv1alpha1.Ready:
		// This can be update request or a duplicate Requeue for the previous status change to Ready

		// Check with state of the world to figure out if this event is because of status update

		targetPolicy, err := r.IAMClient.GetRolePolicy(ctx, input)
		if err != nil {
			// THIS SHOULD NEVER HAPPEN
			// Just requeue in case if it happens
			log.Error(err, "error in verifying the status of the iam role with state of the world")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1}, iammanagerv1alpha1.CreateError)
			return ctrl.Result{RequeueAfter: 30 * time.Duration(iamRole.Status.RetryCount+1) * time.Second}, nil
		}
		if !comparePolicy(ctx, input.PermissionPolicy, *targetPolicy) {
			// if reached here, it means its an update request
			log.Info("Update request event. Proceeding with update operation")
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{}, iammanagerv1alpha1.UpdateInprogress)
			_, err = r.IAMClient.CreateRole(ctx, input)
			if err != nil {
				log.Error(err, "error in creating a role")
				r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 1}, iammanagerv1alpha1.UpdateError)
				return ctrl.Result{RequeueAfter: 30 * time.Duration(1) * time.Second}, nil
			}
			//OK. Successful!!
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName}, iammanagerv1alpha1.Ready)
		} else {
			log.Info("No change in the incoming policy compare to state of the world(external AWS IAM) policy")
		}

	default:
		log.Error(errors.New("Unknown State"), "This should not happen")
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1}, iammanagerv1alpha1.UpdateError)
		return ctrl.Result{RequeueAfter: 30 * time.Duration(iamRole.Status.RetryCount+1) * time.Second}, nil
	}

	log.Info("Successfully reconciled")
	return ctrl.Result{}, nil
}

//SetupWithManager sets up manager with controller
func (r *IamroleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iammanagerv1alpha1.Iamrole{}).
		Complete(r)
}

//UpdateStatus function updates the status based on the process step
func (r *IamroleReconciler) UpdateStatus(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole, status iammanagerv1alpha1.IamroleStatus, state iammanagerv1alpha1.State) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "UpdateStatus")
	log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	status.State = state
	iamRole.Status = status
	if err := r.Status().Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update status", "status", state)
		panic(err)
	}
}

//UpdateMeta function updates the metadata (mostly finalizers in this case)
func (r *IamroleReconciler) UpdateMeta(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "UpdateMeta")
	log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	if err := r.Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update object metadata (finalizer)")
		panic(err)
	}
}

//containsString  Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.
*/
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func comparePolicy(ctx context.Context, request string, target string) bool {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "comparePolicy")

	d, _ := url.QueryUnescape(target)
	dest := iammanagerv1alpha1.PolicyDocument{}
	err := json.Unmarshal([]byte(d), &dest)
	if err != nil {
		log.Error(err, "failed to unmarshal policy document", target)
	}

	req := iammanagerv1alpha1.PolicyDocument{}
	err = json.Unmarshal([]byte(request), &req)
	if err != nil {
		log.Error(err, "failed to marshal policy document", request)
	}
	//compare
	if reflect.DeepEqual(req, dest) {
		log.Info("input policy and target policy are equal")
		return true
	}
	return false
}
