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
	"fmt"
	"hash/adler32"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IamroleSpec defines the desired state of Iamrole
type IamroleSpec struct {
	PolicyDocument PolicyDocument `json:"PolicyDocument"`
	// +optional
	AssumeRolePolicyDocument *AssumeRolePolicyDocument `json:"AssumeRolePolicyDocument,omitempty"`
	// RoleName can be passed only for privileged namespaces. This will be respected only during new iamrole creation and will be ignored during iamrole update
	// Please check the documentation for more on how to configure privileged namespace using annotation for iam-manager
	// +optional
	RoleName string `json:"RoleName,omitempty"`
}

// +kubebuilder:validation:Required

// PolicyDocument type defines IAM policy struct
type PolicyDocument struct {

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
	//Effect allowed/denied
	Effect Effect `json:"Effect"`

	//Action allowed on specific resources
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

// +optional
type AssumeRolePolicyDocument struct {
	// Version specifies IAM policy version
	// By default, this value is "2012-10-17"
	// +optional
	Version string `json:"Version,omitempty"`

	// Statement allows list of TrustPolicyStatement objects
	// +optional
	Statement []TrustPolicyStatement `json:"Statement,omitempty"`
}

// TrustPolicy struct holds Trust policy
// +optional
type TrustPolicyStatement struct {
	Sid string `json:"Sid,omitempty"`
	//Effect allowed/denied
	Effect Effect `json:"Effect,omitempty"`
	//Action can be performed
	Action string `json:"Action,omitempty"`
	// +optional
	Principal Principal `json:"Principal,omitempty"`
	// +optional
	Condition *Condition `json:"Condition,omitempty"`
}

func (tps *TrustPolicyStatement) Checksum() string {
	return fmt.Sprintf("%x", adler32.Checksum([]byte(fmt.Sprintf("%+v", tps))))
}

// Principal struct holds AWS principal
// +optional
type Principal struct {
	// +optional
	AWS StringOrStrings `json:"AWS,omitempty"`
	// +optional
	Service string `json:"Service,omitempty"`
	// +optional
	Federated string `json:"Federated,omitempty"`
}

// Condition struct holds Condition
// +optional
type Condition struct {
	//StringEquals can be used to define Equal condition
	// +optional
	StringEquals map[string]string `json:"StringEquals,omitempty"`
	//StringLike can be used for regex as supported by AWS
	// +optional
	StringLike map[string]string `json:"StringLike,omitempty"`
}

const (
	//Allow Policy allows policy
	AllowPolicy Effect = "Allow"

	//DenyPolicy denies policy
	DenyPolicy Effect = "Deny"
)

// IamroleStatus defines the observed state of Iamrole
type IamroleStatus struct {
	//RoleName represents the name of the iam role created in AWS
	RoleName string `json:"roleName,omitempty"`
	//RoleARN represents the ARN of an IAM role
	RoleARN string `json:"roleARN,omitempty"`
	//RoleID represents the unique ID of the role which can be used in S3 policy etc
	RoleID string `json:"roleID,omitempty"`
	//State of the resource
	State State `json:"state,omitempty"`
	//RetryCount in case of error
	RetryCount int `json:"retryCount"`
	//ErrorDescription in case of error
	// +optional
	ErrorDescription string `json:"errorDescription,omitempty"`
	//LastUpdatedTimestamp represents the last time the iam role has been modified
	// +optional
	LastUpdatedTimestamp metav1.Time `json:"lastUpdatedTimestamp,omitempty"`
}

type State string

const (
	Ready                State = "Ready"
	Error                State = "Error"
	PolicyNotAllowed     State = "PolicyNotAllowed"
	RolesMaxLimitReached State = "RolesMaxLimitReached"
	RoleNameNotAvailable State = "RoleNameNotAvailable"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=iamroles,scope=Namespaced,shortName=iam,singular=iamrole
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="current state of the iam role"
// +kubebuilder:printcolumn:name="RoleName",type="string",JSONPath=".status.roleName",description="Name of the role"
// +kubebuilder:printcolumn:name="RetryCount",type="integer",JSONPath=".status.retryCount",description="Retry count"
// +kubebuilder:printcolumn:name="LastUpdatedTimestamp",type="string",format="date-time",JSONPath=".status.lastUpdatedTimestamp",description="last updated iam role timestamp"
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
