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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/k8s"
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
	"strings"
	"time"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles/status,verbs=get;update;patch

func (r *IamroleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic occured %v", err)
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

	// Is it being deleted?
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
			//Get the roleName from status
			roleName := iamRole.Status.RoleName
			if err := r.IAMClient.DeleteRole(ctx, roleName); err != nil {
				log.Error(err, "Unable to delete the role")
				//i got to fix this
				r.UpdateStatus(ctx, &iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, LastUpdatedTimestamp: metav1.Now(), ErrorDescription: err.Error(), State: iammanagerv1alpha1.Error})
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

	return successRequeueIt()
}

//HandleReconcile function handles all the reconcile
func (r *IamroleReconciler) HandleReconcile(ctx context.Context, req ctrl.Request, iamRole *iammanagerv1alpha1.Iamrole) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "HandleReconcile")
	log = log.WithValues("iam_role_cr", iamRole.Name)
	log.Info("state of the custom resource ", "state", iamRole.Status.State)
	ns := v1.Namespace{}
	if iamRole.Status.RoleName == "" && iamRole.Spec.RoleName != "" {
		//Get Namespace metadata
		ns2, err := k8s.NewK8sClientDoOrDie().GetNamespace(ctx, iamRole.Namespace)
		if err != nil {
			log.Error(err, "unable to get namespace metadata. please update Cluster Role to allow namespace get operations")
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "unable to get namespace metadata. please update Cluster Role to allow namespace get operations "+err.Error())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		ns = *ns2
	}

	roleName, err := utils.GenerateRoleName(ctx, iamRole, *config.Props, &ns)
	log.V(1).Info("roleName constructed successfully", "roleName", roleName)

	if err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to construct iam role name to error "+err.Error())
		// It is not clear to me that we want to requeue here - as this is a fairly permanent
		// error. Is there a better pattern here?
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	input, status, err := r.ConstructCreateIAMRoleInput(ctx, iamRole, roleName)
	if err != nil {
		if status == nil {
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to construct iam role due to error "+err.Error())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: status.RoleName, ErrorDescription: status.ErrorDescription, State: status.State, LastUpdatedTimestamp: metav1.Now()})
	}

	var requeueTime float64
	switch iamRole.Status.State {
	case iammanagerv1alpha1.Ready:

		// This can be update request or a duplicate Requeue for the previous status change to Ready
		// Check with state of the world to figure out if this event is because of status update
		targetRole, err := r.IAMClient.GetRole(ctx, *input)
		if err != nil {
			// THIS SHOULD NEVER HAPPEN
			// Just requeue in case if it happens
			log.Error(err, "error in verifying the status of the iam role with state of the world")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.Error, LastUpdatedTimestamp: metav1.Now()}, 3000)

		}

		targetPolicy, err := r.IAMClient.GetRolePolicy(ctx, *input)
		if err != nil {
			// THIS SHOULD NEVER HAPPEN
			// Just requeue in case if it happens
			log.Error(err, "error in verifying the status of the iam role with state of the world")
			log.Info("retry count error", "count", iamRole.Status.RetryCount)
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.Error}, 3000)

		}

		if validation.CompareRole(ctx, *input, targetRole, *targetPolicy) {
			log.Info("No change in the incoming policy compare to state of the world(external AWS IAM) policy")
			r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName, ErrorDescription: "", RoleID: aws.StringValue(targetRole.Role.RoleId), RoleARN: aws.StringValue(targetRole.Role.Arn), LastUpdatedTimestamp: iamRole.Status.LastUpdatedTimestamp, State: iammanagerv1alpha1.Ready})

			return successRequeueIt()
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
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: errMsg, State: iammanagerv1alpha1.RolesMaxLimitReached, LastUpdatedTimestamp: metav1.Now()})
		}
		fallthrough
	default:

		// Default behavior on new Iamrole resource state is to go off and create it
		resp, err := r.IAMClient.EnsureRole(ctx, *input)
		if err != nil {
			log.Error(err, "error in creating a role")
			state := iammanagerv1alpha1.Error

			// This check verifies whether or not the IAM Role somehow already exists, but is allocated to another namespace based on the tag applied to it.
			if strings.Contains(err.Error(), awsapi.RoleExistsAlreadyForOtherNamespace) {
				state = iammanagerv1alpha1.RoleNameNotAvailable

				//Role itself is not created
				roleName = ""
			}
			r.Recorder.Event(iamRole, v1.EventTypeWarning, string(state), "Unable to create/update iam role due to error "+err.Error())
			return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error(), State: state, LastUpdatedTimestamp: metav1.Now()}, requeueTime)
		}

		//OK. Successful!!
		// Is this IRSA role? If yes, Create/update Service Account with required annotation
		flag, saName := utils.ParseIRSAAnnotation(ctx, iamRole)
		if flag {
			if err := k8s.NewK8sManagerClient(r.Client).CreateOrUpdateServiceAccount(ctx, saName, iamRole.Namespace, resp.RoleARN); err != nil {
				log.Error(err, "error in updating service account for IRSA role")
				r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update service account for IRSA role due to error "+err.Error())
				return r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: iamRole.Status.RetryCount + 1, RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.Error, LastUpdatedTimestamp: metav1.Now()}, requeueTime)
			}
		}

		r.Recorder.Event(iamRole, v1.EventTypeNormal, string(iammanagerv1alpha1.Ready), "Successfully created/updated iam role")
		r.UpdateStatus(ctx, iamRole, iammanagerv1alpha1.IamroleStatus{RetryCount: 0, RoleName: roleName, ErrorDescription: "", RoleID: resp.RoleID, RoleARN: resp.RoleARN, LastUpdatedTimestamp: metav1.Now(), State: iammanagerv1alpha1.Ready})
	}
	log.Info("Successfully reconciled")

	return successRequeueIt()
}

//ConstructInput function constructs input for
func (r *IamroleReconciler) ConstructCreateIAMRoleInput(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole, roleName string) (*awsapi.IAMRoleRequest, *iammanagerv1alpha1.IamroleStatus, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "ConstructInput")
	log.WithValues("iamrole", iamRole.Name)
	role, _ := json.Marshal(iamRole.Spec.PolicyDocument)

	//Validate IAM Policy and Resource
	if err := validation.ValidateIAMPolicyAction(ctx, iamRole.Spec.PolicyDocument); err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.PolicyNotAllowed), "Unable to create/update iam role due to error "+err.Error())
		return nil, &iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.PolicyNotAllowed}, err
	}

	if err := validation.ValidateIAMPolicyResource(ctx, iamRole.Spec.PolicyDocument); err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.PolicyNotAllowed), "Unable to create/update iam role due to error "+err.Error())
		return nil, &iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.PolicyNotAllowed}, err
	}

	trustPolicy, err := utils.GetTrustPolicy(ctx, iamRole)
	if err != nil {
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update iam role due to error "+err.Error())
		return nil, &iammanagerv1alpha1.IamroleStatus{RoleName: roleName, ErrorDescription: err.Error(), State: iammanagerv1alpha1.Error}, err
	}

	//Attach some tags
	tags := map[string]string{
		"managedBy": "iam-manager",
		"Namespace": iamRole.Namespace,
	}

	if config.Props.ClusterName() != "" {
		tags["Cluster"] = config.Props.ClusterName()
	}

	input := &awsapi.IAMRoleRequest{
		Name:                            roleName,
		PolicyName:                      config.InlinePolicyName,
		Description:                     "#DO NOT DELETE#. Managed by iam-manager",
		SessionDuration:                 43200,
		TrustPolicy:                     trustPolicy,
		PermissionPolicy:                string(role),
		ManagedPermissionBoundaryPolicy: config.Props.ManagedPermissionBoundaryPolicy(),
		ManagedPolicies:                 config.Props.ManagedPolicies(),
		Tags:                            tags,
	}

	return input, nil, nil
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
func (r *IamroleReconciler) UpdateStatus(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole, status iammanagerv1alpha1.IamroleStatus, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "UpdateStatus")
	log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	if status.RoleARN == "" {
		status.RoleARN = iamRole.Status.RoleARN
	}
	if status.RoleID == "" {
		status.RoleID = iamRole.Status.RoleID
	}

	if iamRole.Status.LastUpdatedTimestamp.IsZero() {
		status.LastUpdatedTimestamp = metav1.Now()
	}

	iamRole.Status = status
	if err := r.Status().Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update status", "status", status.State)
		r.Recorder.Event(iamRole, v1.EventTypeWarning, string(iammanagerv1alpha1.Error), "Unable to create/update status due to error "+err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if status.State != iammanagerv1alpha1.Error {
		return successRequeueIt()
	}

	//if wait time is specified, requeue it after provided time
	if len(requeueTime) == 0 {
		requeueTime[0] = 0
	}

	log.Info("Requeue time", "time", requeueTime[0])
	return ctrl.Result{RequeueAfter: time.Duration(requeueTime[0]) * time.Millisecond}, nil
}

//UpdateMeta function updates the metadata (mostly finalizers in this case)
func (r *IamroleReconciler) UpdateMeta(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole) {
	log := log.Logger(ctx, "controllers", "iamrole_controller", "UpdateMeta")
	log = log.WithValues("iam_role_cr", iamRole.ObjectMeta.Name)

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

//successRequeueIt function requeues it after defined time
func successRequeueIt() (ctrl.Result, error) {

	return ctrl.Result{RequeueAfter: time.Duration(config.Props.ControllerDesiredFrequency()) * time.Second}, nil
}
