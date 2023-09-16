package v1alpha1

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/keikoproj/iam-manager/pkg/logging"
)

/**
 * This function is used to retrieve all IAM-Roles from the cluster across all namespaces.
 * It will return a list of IAM-Roles in structured format.
 */
func ListIamRoles(ctx context.Context, c client.Client) ([]*Iamrole, error) {
	log := logging.Logger(ctx, "k8s", "client", "ListIamRoles")

	var uRoleList *unstructured.UnstructuredList = &unstructured.UnstructuredList{}
	var iamRoles []*Iamrole = []*Iamrole{}
	var err error
	var b []byte
	var IamroleGroupVersionKind = schema.GroupVersionKind{
		Group:   "iammanager.keikoproj.io",
		Version: "v1alpha1",
		Kind:    "Iamrole",
	}
	uRoleList.SetGroupVersionKind(IamroleGroupVersionKind)

	if err = c.List(ctx, uRoleList, &client.ListOptions{}); err != nil {
		log.Error(err, "unable to list iamroles resources")
		return iamRoles, err
	}

	if b, err = json.Marshal(uRoleList.Items); err != nil {
		log.Error(err, "unable to marshal iamroles resources")
		return iamRoles, err
	}

	if err = json.Unmarshal(b, &iamRoles); err != nil {
		log.Error(err, "unable to unmarshal iamroles resources")
		return iamRoles, err
	}

	return iamRoles, nil
}

func GetIamRole(ctx context.Context, c client.Client, name, namespace string) (*Iamrole, error) {
	log := logging.Logger(ctx, "k8s", "client", "GetIamRole")
	log.V(1).Info("get api call for iamrole")

	var uRole *unstructured.Unstructured = &unstructured.Unstructured{}
	var iamRole *Iamrole = &Iamrole{}
	var err error
	var b []byte
	var IamroleGroupVersionKind = schema.GroupVersionKind{
		Group:   "iammanager.keikoproj.io",
		Version: "v1alpha1",
		Kind:    "Iamrole",
	}
	uRole.SetGroupVersionKind(IamroleGroupVersionKind)

	if err = c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, uRole); err != nil {
		log.Error(err, "unable to get iamrole resource")
		return iamRole, err
	}

	if b, err = json.Marshal(uRole); err != nil {
		log.Error(err, "unable to marshal iamrole resource")
		return iamRole, err
	}

	if err = json.Unmarshal(b, iamRole); err != nil {
		log.Error(err, "unable to unmarshal iamrole resource")
		return iamRole, err
	}

	return iamRole, nil
}
