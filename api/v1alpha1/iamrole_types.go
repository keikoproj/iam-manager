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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IamroleSpec defines the desired state of Iamrole
type IamroleSpec struct {
	PolicyDocument PolicyDocument `json:"PolicyDocument"`
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:validation:Required

//PolicyDocument type defines IAM policy struct
type PolicyDocument struct {

	// +kubebuilder:default:="2012-10-17"
	// Version specifies IAM policy version
	// By default, this value is "2012-10-17"
	// +optional
	Version string `json:"Version,omitempty"`

	// Statement allows list of statement object
	Statement []Statement `json:"Statement"`
}

// +kubebuilder:validation:Required
// Statement type defines the AWS IAM policy statement
type Statement struct {
	//Effect on target resource
	Effect Effect `json:"Effect"`

	//Action allowed/denied on specific resources
	Action []string `json:"Action"`

	//Resources defines target resources which IAM policy will be applied
	Resource []string `json:"Resource"`
	// Sid is an optional field which describes the specific statement action
	// +optional
	Sid string `json:"Sid,omitempty"`
}

// Effect describes whether to allow or deny the specific action
// Allowed values are
// - "Allow" : allows the specific action on resources
// - "Deny" : denies the specific action on resources
// +kubebuilder:validation:Enum=Allow;Deny
type Effect string

const (
	//Allow Policy allows policy
	AllowPolicy Effect = "Allow"

	//DenyPolicy denies policy
	DenyPolicy Effect = "Deny"
)

// IamroleStatus defines the observed state of Iamrole
type IamroleStatus struct {
	// Important: Run "make" to regenerate code after modifying this file
	RoleName   string `json:"roleName,omitempty"`
	State      State  `json:"state,omitempty"`
	RetryCount int    `json:"retryCount"`
}

type State string

const (
	CreateInProgress State = "CreateInProgress"
	CreateError      State = "CreateError"

	UpdateInprogress State = "UpdateInProgress"
	UpdateError      State = "UpdateError"

	DeleteInprogress State = "DeleteInprogress"

	Ready State = "Ready"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=iamroles,scope=Namespaced,shortName=iam,singular=iamrole
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="current state of the iam role"
// +kubebuilder:printcolumn:name="RoleName",type="string",JSONPath=".status.roleName",description="Name of the role"
// +kubebuilder:printcolumn:name="RetryCount",type="integer",JSONPath=".status.retryCount",description="Retry count"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="time passed since iamrole creation"
// Iamrole is the Schema for the iamroles API
type Iamrole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IamroleSpec   `json:"spec,omitempty"`
	Status IamroleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IamroleList contains a list of Iamrole
type IamroleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Iamrole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Iamrole{}, &IamroleList{})
}
