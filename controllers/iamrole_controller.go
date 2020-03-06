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
	"errors"
	"fmt"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/log"
	"github.com/keikoproj/iam-manager/pkg/validation"
	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"math"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
)

const (
	finalizerName = "iamrole.finalizers.iammanager.keikoproj.io"
	requestId     = "request_id"
	//2 minutes
	maxWaitTime = 120000
)

// IamroleReconciler reconciles a Iamrole object
type IamroleReconciler struct {
	client.Client
	IAMClient *awsapi.IAM
	Recorder  record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create
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

	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Name)
	// Isit being deleted?
	if iamRole.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		//Lets check if this is very first time use case
		if !validation.ContainsString(iamRole.ObjectMeta.Finalizers, finalizerName) {
			log.Info("New iamrole resource. Adding the finalizer", "finalizer", finalizerName)
			iamRole.ObjectMeta.Finalizers = append(iamRole.ObjectMeta.Finalizers, finalizerName)
			r.UpdateMeta(ctx, &iamRole)
		}
		return r.HandleReconcile(ctx, req, &iamRole)

	} else {
		//oh oh.. This is delete use case
		//Lets make sure to clean up the iam role
		if iamRole.Status.RetryCount != 0 {
			iamRole.Status.RetryCount = iamRole.Status.RetryCount + 1
		}
		log.Info("Iamrole delete request")
		if iamRole.Status.State != iammanagerv1alpha1.PolicyNotAllowed {
			if err := r.IAMClient.DeleteRole(ctx, roleName); err != nil {
				log.Error(err, "Unable to delete the role")
				//i got to fix this
				r.UpdateStatus(ctx, &iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, ErrorDescription: err.Error()}, iammanagerv1alpha1.Error)
				r.Recorder.Event(&iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "unable to delete the role due to "+err.Error())

				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
		}

		// Ok. Lets delete the finalizer so controller can delete the custom object
		log.Info("Removing finalizer from Iamrole")
		iamRole.ObjectMeta.Finalizers = validation.RemoveString(iamRole.ObjectMeta.Finalizers, finalizerName)
		r.UpdateMeta(ctx, &iamRole)
		log.Info("Successfully deleted iam role")
		r.Recorder.Event(&iamRole, v1.EventTypeNormal, "Deleted", "Successfully deleted iam role")
	}

	return ctrl.Result{}, nil
}

//HandleReconcile function handles all the reconcile
func (r *IamroleReconciler) HandleReconcile(ctx context.Context, req ctrl.Request, iamRole *iammanagerv1alpha1.Iamrole) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "HandleReconcile")
	log.WithValues("iamrole", iamRole.Name)
	log.Info("state of the custom resource ", "state", iamRole.Status.State)
	role, _ := json.Marshal(iamRole.Spec.PolicyDocument)
	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Name)

	trustPolicy, err := utils.GetTrustPolicy(ctx, &iamRole.Spec.TrustPolicy)
	if err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
		return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.Error)
	}
	input := awsapi.IAMRoleRequest{
		Name:                            roleName,
		PolicyName:                      config.InlinePolicyName,
		Description:                     "#DO NOT DELETE#. Managed by iam-manager",
		SessionDuration:                 3600,
		TrustPolicy:                     trustPolicy,
		PermissionPolicy:                string(role),
		ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(),
		ManagedPolicies:                 config.Props.ManagedPolicies(),
	}
	//Validate IAM Policy and Resource
	if err := validation.ValidateIAMPolicyAction(ctx, iamRole.Spec.PolicyDocument); err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.PolicyNotAllowed), "Unable to create/update iam role due to error "+err.Error())
		return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.PolicyNotAllowed)
	}

	if err := validation.ValidateIAMPolicyResource(ctx, iamRole.Spec.PolicyDocument); err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.PolicyNotAllowed), "Unable to create/update iam role due to error "+err.Error())
		return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.PolicyNotAllowed)
	}

	var requeueTime float64
	switch iamRole.Status.State {
	case iammanagerv1alpha1.Ready:

		// This can be update request or a duplicate Requeue for the previous status change to Ready
		// Check with state of the world to figure out if this event is because of status update
		targetRole, err := r.IAMClient.GetRole(ctx, input)
		if err != nil {
			// THIS SHOULD NEVER HAPPEN
			// Just requeue in case if it happens
			log.Error(err, "error in verifying the status of the iam role with state of the world")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.Error, 3000)

		}

		targetPolicy, err := r.IAMClient.GetRolePolicy(ctx, input)
		if err != nil {
			// THIS SHOULD NEVER HAPPEN
			// Just requeue in case if it happens
			log.Error(err, "error in verifying the status of the iam role with state of the world")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.Error, 3000)

		}

		if validation.CompareRole(ctx, input, targetRole, *targetPolicy) {
			log.Info("No change in the incoming policy compare to state of the world(external AWS IAM) policy")
			return ctrl.Result{}, nil
		}
		fallthrough

	case iammanagerv1alpha1.Error:
		// Needs to check if it is just error retrial or user changed anything
		// if user modified the input we shouldn't wait and directly fallthrough

		if iamRole.Status.RetryCount != 0 {
			//Lets do exponential back off
			// 2^x
			waitTime := math.Pow(2, float64(iamRole.Status.RetryCount+1)) * 100
			requeueTime = waitTime
			if waitTime > maxWaitTime {
				requeueTime = maxWaitTime
			}
			log.V(1).Info("Going to requeue it as part of exponential back off after this try", "count", iamRole.Status.RetryCount+1, "time in ms", requeueTime)
		}
		fallthrough
	case "", iammanagerv1alpha1.PolicyNotAllowed, iammanagerv1alpha1.RolesMaxLimitReached:
		//Validate Number of successful IAM roles

		var iamRoles iammanagerv1alpha1.IamroleList
		if err := r.List(ctx, &iamRoles, client.InNamespace(req.Namespace)); err != nil {
			return ctrl.Result{}, ignoreNotFound(err)
		}

		log.Info("Total Number of roles", "count", len(iamRoles.Items), "allowed", config.Props.MaxRolesAllowed())

		if config.Props.MaxRolesAllowed() < len(iamRoles.Items) {
			errMsg := "maximum number of allowed roles reached. You must delete any existing role before proceeding further"
			log.Error(errors.New(errMsg), errMsg)
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.RolesMaxLimitReached), errMsg)
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: errMsg}, iammanagerv1alpha1.RolesMaxLimitReached)
		}
		fallthrough
	default:

		_, err := r.IAMClient.CreateRole(ctx, input)
		if err != nil {
			log.Error(err, "error in creating a role")
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error()}, iammanagerv1alpha1.Error, requeueTime)
		}
		//OK. Successful!!
		r.Recorder.Event(iamRole, v1.EventTypeNormal, string(iammanagerv1alpha1.Ready), "Successfully created/updated iam role")
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName, ErrorDescription: ""}, iammanagerv1alpha1.Ready)
	}
	log.Info("Successfully reconciled")

	return ctrl.Result{}, nil
}

type StatusUpdatePredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating generation change
func (StatusUpdatePredicate) Update(e event.UpdateEvent) bool {
	log := log.Logger(context.Background(), "controllers", "iamrole_controller", "HandleReconcile")
	if e.MetaOld == nil {
		log.Error(nil, "Update event has no old metadata", "event", e)
		return false
	}
	if e.ObjectOld == nil {
		log.Error(nil, "Update event has no old runtime object to update", "event", e)
		return false
	}
	if e.ObjectNew == nil {
		log.Error(nil, "Update event has no new runtime object for update", "event", e)
		return false
	}
	if e.MetaNew == nil {
		log.Error(nil, "Update event has no new metadata", "event", e)
		return false
	}
	oldObj := e.ObjectOld.(*iammanagerv1alpha1.Iamrole)
	newObj := e.ObjectNew.(*iammanagerv1alpha1.Iamrole)

	if oldObj.Status != newObj.Status {
		return false
	}
	return true
}

//SetupWithManager sets up manager with controller
//GenerationChangedPredicate will take care of not allowing to trigger reconcile for every time status update happens
func (r *IamroleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	//Lets try to predicate based on Status retry count
	return ctrl.NewControllerManagedBy(mgr).
		For(&iammanagerv1alpha1.Iamrole{}).
		WithEventFilter(StatusUpdatePredicate{}).
		Complete(r)
}

//UpdateStatus function updates the status based on the process step
func (r *IamroleReconciler) UpdateStatus(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole, status iammanagerv1alpha1.IamroleStatus, state iammanagerv1alpha1.State, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "UpdateStatus")
	log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	status.State = state
	iamRole.Status = status
	if err := r.Status().Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update status", "status", state)
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update status due to error "+err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	//if wait time is specified, requeue it after provided time
	if len(requeueTime) == 0 {
		requeueTime[0] = 0
	}

	if state != iammanagerv1alpha1.Error {
		return ctrl.Result{}, nil
	}
	log.Info("Requeue time", "time", requeueTime[0])
	return ctrl.Result{RequeueAfter: time.Duration(requeueTime[0]) * time.Millisecond}, nil
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
