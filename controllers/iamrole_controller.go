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
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"time"

	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
)

const (
	maxRetryCount = 10
	finalizerName = "iamrole.finalizers.iammanager.keikoproj.io"
)

// IamroleReconciler reconciles a Iamrole object
type IamroleReconciler struct {
	client.Client
	Log       logr.Logger
	IAMClient *awsapi.IAM
}

// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iammanager.keikoproj.io,resources=iamroles/status,verbs=get;update;patch

func (r *IamroleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("iamrole", req.NamespacedName)

	//Get the resource
	var iamRole iammanagerv1alpha1.Iamrole

	if err := r.Get(ctx, req.NamespacedName, &iamRole); err != nil {

		return ctrl.Result{}, ignoreNotFound(err)
	}

	log.Info("I got something", "status", iamRole.Status.State)
	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace)
	// Isit being deleted?
	if iamRole.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		//Lets check if this is very first time use case
		if !containsString(iamRole.ObjectMeta.Finalizers, finalizerName) {
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
		log.Info("iam role name in delete", "role", roleName)
		//r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteInprogress)
		if err := r.IAMClient.DeleteRole(ctx, roleName); err != nil {
			log.Error(err, "msg", "err", err.Error())
			//r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteError)
			//return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		//r.UpdateStatus(ctx, &iamRole, iamRole.Status, iammanagerv1alpha1.DeleteComplete)

		// Ok. Lets delete the finalizer so controller can delete the custom object
		iamRole.ObjectMeta.Finalizers = removeString(iamRole.ObjectMeta.Finalizers, finalizerName)
		r.UpdateMeta(ctx, &iamRole)
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
	log := r.Log.WithValues("id", iamRole.ObjectMeta.UID, "iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))

	role, _ := json.Marshal(iamRole.Spec.PolicyDocument)
	fmt.Println(string(role))
	roleName := fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace)
	input := awsapi.IAMRoleRequest{
		Name:             roleName,
		PolicyName:       fmt.Sprintf("k8s-%s-policy", iamRole.ObjectMeta.Namespace),
		Description:      "bla bla",
		SessionDuration:  3600,
		TrustPolicy:      roleTrust,
		PermissionPolicy: string(role),
	}
	status := iammanagerv1alpha1.IamroleStatus{
		RoleName:   roleName,
		RetryCount: 0,
	}

	//I MIGHT NOT NEED THIS WHOLE SWITCH CASE
	//REVISIT LATER
	switch iamRole.Status.State {

	case "", iammanagerv1alpha1.CreateError:
		// why check with 0?
		//if iamRole.Status.RetryCount != 0 {
		//	status.RetryCount = iamRole.Status.RetryCount + 1
		//}
		//r.UpdateStatus(ctx, iamRole, status, iammanagerv1alpha1.CreateInProgress)
		//This should be zero day use case
		_, err := r.IAMClient.CreateRole(ctx, input)
		if err != nil {
			log.Error(err, "msg", "err", err.Error())
			status.RetryCount = iamRole.Status.RetryCount + 1
			r.UpdateStatus(ctx, iamRole, status, iammanagerv1alpha1.CreateError)
			return ctrl.Result{RequeueAfter: 30 * time.Duration(status.RetryCount) * time.Second}, nil
		}

	case iammanagerv1alpha1.Ready, iammanagerv1alpha1.UpdateError:
		//This means its an update request
		log.Info("lets overwrite iam pol")
		//if iamRole.Status.RetryCount != 0 {
		//	status.RetryCount = iamRole.Status.RetryCount + 1
		//}
		//r.UpdateStatus(ctx, iamRole, status, iammanagerv1alpha1.UpdateInprogress)
		//This should be zero day use case
		_, err := r.IAMClient.CreateRole(ctx, input)
		if err != nil {
			log.Error(err, "msg", "err", err.Error())
			status.RetryCount = iamRole.Status.RetryCount + 1
			r.UpdateStatus(ctx, iamRole, status, iammanagerv1alpha1.UpdateError)
			return ctrl.Result{RequeueAfter: 30 * time.Duration(status.RetryCount) * time.Second}, nil
		}

	}

	// If i reach here, That means it is all good. Lets change the status to Ready
	status.RetryCount = 0
	r.UpdateStatus(ctx, iamRole, status, iammanagerv1alpha1.Ready)
	log.Info("I'm responding back", "status", iamRole.Status.State)
	return ctrl.Result{}, nil
}

func (r *IamroleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iammanagerv1alpha1.Iamrole{}).
		Complete(r)
}

//UpdateStatus function updates the status based on the process step
func (r *IamroleReconciler) UpdateStatus(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole, status iammanagerv1alpha1.IamroleStatus, state iammanagerv1alpha1.State) {
	log := r.Log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	status.State = state
	iamRole.Status = status
	if err := r.Status().Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update status", "err", err.Error(), "status", state)
		panic(err)
	}
}

//UpdateMeta function updates the metadata (mostly finalizers in this case)
func (r *IamroleReconciler) UpdateMeta(ctx context.Context, iamRole *iammanagerv1alpha1.Iamrole) {
	log := r.Log.WithValues("iamrole", fmt.Sprintf("k8s-%s", iamRole.ObjectMeta.Namespace))
	if err := r.Update(ctx, iamRole); err != nil {
		log.Error(err, "Unable to update object metadata (finalizer)", "err", err.Error())
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
