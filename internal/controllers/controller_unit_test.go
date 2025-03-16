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
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	mock_awsapi "github.com/keikoproj/iam-manager/pkg/awsapi/mocks"
)

func init() {
	// Set up test environment variables
	os.Setenv("LOCAL", "true")
	os.Setenv("AWS_REGION", "us-west-2")
	
	// Load configuration
	config.LoadProperties("LOCAL")
}

// setupTest configures a test environment for the controller tests
func setupTest(t *testing.T) (*gomock.Controller, *mock_awsapi.MockIAMAPI, *IamroleReconciler, client.Client) {
	// Setup controller
	mockCtrl := gomock.NewController(t)
	mockIAM := mock_awsapi.NewMockIAMAPI(mockCtrl)

	// Create a scheme with IAM Manager types
	scheme := runtime.NewScheme()
	_ = iammanagerv1alpha1.AddToScheme(scheme)
	
	// Create a fake client with the scheme
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create controller
	reconciler := &IamroleReconciler{
		Client:    k8sClient,
		IAMClient: &awsapi.IAM{Client: mockIAM},
		Recorder:  &record.FakeRecorder{},
	}

	return mockCtrl, mockIAM, reconciler, k8sClient
}

// TestSuccessRequeueIt tests the successRequeueIt function which requeues reconciliation
func TestSuccessRequeueIt(t *testing.T) {
	// Call the function
	result, err := successRequeueIt()

	// Verify results
	assert.NoError(t, err, "Expected no error")
	assert.True(t, result.Requeue, "Expected Requeue to be true")
	assert.Equal(t, float64(config.Props.ControllerDesiredFrequency()), result.RequeueAfter.Seconds(), "Expected correct requeue time")
}

// TestIgnoreNotFound tests the ignoreNotFound function which filters errors
func TestIgnoreNotFound(t *testing.T) {
	// Test with nil error
	err := ignoreNotFound(nil)
	assert.NoError(t, err, "Expected nil error to remain nil")

	// Test with non-NotFound error
	testErr := assert.AnError
	err = ignoreNotFound(testErr)
	assert.Error(t, err, "Expected non-NotFound error to be returned")
	assert.Equal(t, testErr, err, "Expected original error to be returned")

	// Note: Testing with a real NotFound error requires more complex setup with k8s types
}

// TestUpdateStatus tests the UpdateStatus function
func TestUpdateStatus(t *testing.T) {
	// Setup
	_, _, reconciler, k8sClient := setupTest(t)

	// Create test IAM role
	iamRole := &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-role",
			Namespace: "default",
		},
		Status: iammanagerv1alpha1.IamroleStatus{
			State: iammanagerv1alpha1.Ready,
		},
	}

	// Create the role in the fake client
	err := k8sClient.Create(context.Background(), iamRole)
	assert.NoError(t, err, "Expected role creation to succeed")

	// Test status update
	newStatus := iammanagerv1alpha1.IamroleStatus{
		State:      iammanagerv1alpha1.Error,
		RetryCount: 1,
		RoleName:   "test-role-updated",
	}

	// Update status
	requeueTime := 30.0
	result, err := reconciler.UpdateStatus(context.Background(), iamRole, newStatus, requeueTime)

	// Verify results
	assert.NoError(t, err, "Expected status update to succeed")
	assert.True(t, result.Requeue, "Expected Requeue to be true")
	assert.Equal(t, requeueTime, result.RequeueAfter.Seconds(), "Expected correct requeue time")

	// Verify the status was updated
	updatedRole := &iammanagerv1alpha1.Iamrole{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test-role"}, updatedRole)
	assert.NoError(t, err, "Expected to fetch the updated role")
	assert.Equal(t, iammanagerv1alpha1.Error, updatedRole.Status.State, "Expected status state to be updated")
	assert.Equal(t, "test-role-updated", updatedRole.Status.RoleName, "Expected status role name to be updated")
	assert.Equal(t, 1, updatedRole.Status.RetryCount, "Expected retry count to be updated")
}

// TestUpdateMeta tests the UpdateMeta function which manages finalizers
func TestUpdateMeta(t *testing.T) {
	// Setup
	_, _, reconciler, k8sClient := setupTest(t)

	// Create test IAM role without finalizer
	iamRole := &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-role-meta",
			Namespace: "default",
		},
	}

	// Create the role in the fake client
	err := k8sClient.Create(context.Background(), iamRole)
	assert.NoError(t, err, "Expected role creation to succeed")

	// Update metadata (adds finalizer)
	reconciler.UpdateMeta(context.Background(), iamRole)

	// Verify the finalizer was added
	updatedRole := &iammanagerv1alpha1.Iamrole{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test-role-meta"}, updatedRole)
	assert.NoError(t, err, "Expected to fetch the updated role")
	assert.Contains(t, updatedRole.Finalizers, finalizerName, "Expected finalizer to be added")

	// Test with deletion timestamp set
	deleteTime := metav1.Now()
	updatedRole.DeletionTimestamp = &deleteTime
	err = k8sClient.Update(context.Background(), updatedRole)
	assert.NoError(t, err, "Expected update with deletion timestamp to succeed")

	// Get the updated role with deletion timestamp
	deletingRole := &iammanagerv1alpha1.Iamrole{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test-role-meta"}, deletingRole)
	assert.NoError(t, err, "Expected to fetch the role being deleted")

	// This is a special case - finalizer should be kept during deletion
	// because the controller is responsible for resource cleanup
	reconciler.UpdateMeta(context.Background(), deletingRole)

	// Get the final state after second UpdateMeta call
	finalRole := &iammanagerv1alpha1.Iamrole{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "test-role-meta"}, finalRole)
	assert.NoError(t, err, "Expected to fetch the role after second UpdateMeta")
	assert.Contains(t, finalRole.Finalizers, finalizerName, "Expected finalizer to still be present")

	// After resource cleanup in a real scenario, the controller would remove the finalizer
}

// TestConstructCreateIAMRoleInput tests the ConstructCreateIAMRoleInput function
func TestConstructCreateIAMRoleInput(t *testing.T) {
	// Setup
	_, _, reconciler, _ := setupTest(t)

	ctx := context.Background()

	// Create test IAM role with minimal spec
	iamRole := &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-role-input",
			Namespace: "default",
			UID:       "test-uid",
		},
		Spec: iammanagerv1alpha1.IamroleSpec{
			PolicyDocument: iammanagerv1alpha1.PolicyDocument{
				Statement: []iammanagerv1alpha1.Statement{
					{
						Effect:   iammanagerv1alpha1.AllowPolicy,
						Action:   []string{"s3:GetObject"},
						Resource: []string{"*"},
					},
				},
			},
		},
	}

	// Test input construction
	roleName := "test-constructed-role"
	input, status, err := reconciler.ConstructCreateIAMRoleInput(ctx, iamRole, roleName)

	// Verify results
	assert.NoError(t, err, "Expected input construction to succeed")
	assert.NotNil(t, input, "Expected non-nil input")
	assert.NotNil(t, status, "Expected non-nil status")

	// Verify input values
	assert.Equal(t, roleName, input.Name, "Expected correct role name")
	assert.Equal(t, roleName, status.RoleName, "Expected correct role name in status")
	assert.NotEmpty(t, input.PermissionPolicy, "Expected policy to be constructed")
	assert.NotEmpty(t, input.TrustPolicy, "Expected trust policy to be constructed")
	assert.Contains(t, input.Tags, "managedBy", "Expected managedBy tag")
	assert.Equal(t, "iam-manager", input.Tags["managedBy"], "Expected correct managedBy tag value")
}
