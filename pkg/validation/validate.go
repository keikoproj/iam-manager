package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/log"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/url"
	"reflect"
	"strings"
)

//ValidateIAMPolicyAction validates policy action
func ValidateIAMPolicyAction(ctx context.Context, pDoc v1alpha1.PolicyDocument) *field.Error {
	log := log.Logger(ctx, "pkg.validation", "ValidateIAMPolicyAction")

	//Check the incoming policy actions
	for _, statement := range pDoc.Statement {
		for _, action := range statement.Action {
			isAllowed := false
			for _, prefix := range config.Props.AllowedPolicyAction() {

				if strings.HasPrefix(action, prefix) {
					isAllowed = true
					break
				}

			}
			//This line shouldn't be executed unless if there is restricted action or end of the loop
			if !isAllowed {
				err := fmt.Sprintf("restricted action %s included in the request", action)
				log.Error(errors.New(err), err)
				return field.Forbidden(field.NewPath("spec").Child("PolicyDocument").Child("Action"), fmt.Sprintf("restricted action %s included in the request", action))
			}
			//This is special case-- May be only for Intuit
			if strings.HasPrefix(action, "s3:") {
				for _, resource := range statement.Resource {
					for _, res := range config.Props.RestrictedS3Resources() {
						isAllowed := false
						if resource != res {
							isAllowed = true
							break
						}

						//This line shouldn't be executed unless if there is restricted action or end of the loop
						if !isAllowed {
							err := fmt.Sprintf("restricted resource %s included in the request", resource)
							log.Error(errors.New(err), err)
							return field.Forbidden(field.NewPath("spec").Child("PolicyDocument").Child("Resource"), fmt.Sprintf("restricted resource %s included in the request", resource))
						}
					}
				}
			}
		}
	}
	return nil
}

//ValidateIAMPolicyResource validates policy resource
func ValidateIAMPolicyResource(ctx context.Context, pDoc v1alpha1.PolicyDocument) *field.Error {
	log := log.Logger(ctx, "pkg.validation", "ValidateIAMPolicyResource")

	//Check the incoming policy resource
	for _, statement := range pDoc.Statement {
		for _, resource := range statement.Resource {
			isAllowed := true
			for _, res := range config.Props.RestrictedPolicyResources() {

				if strings.Contains(resource, res) {
					isAllowed = false
					break
				}
			}
			//This line shouldn't be executed unless if there is restricted action or end of the loop
			if !isAllowed {
				err := fmt.Sprintf("restricted resource %s included in the request", resource)
				log.Error(errors.New(err), err)
				return field.Forbidden(field.NewPath("spec").Child("PolicyDocument").Child("Resource"), fmt.Sprintf("restricted resource %s included in the request", resource))
			}
		}
	}
	return nil
}

//ComparePolicy function compares input policy doc to target policy doc
func ComparePolicy(ctx context.Context, request string, target string) bool {
	log := log.Logger(ctx, "pkg.validation", "ComparePolicy")

	d, _ := url.QueryUnescape(target)
	dest := v1alpha1.PolicyDocument{}
	err := json.Unmarshal([]byte(d), &dest)
	if err != nil {
		log.Error(err, "failed to unmarshal policy document", target)
	}

	req := v1alpha1.PolicyDocument{}
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

//ContainsString  Helper functions to check from a slice of strings.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

//RemoveString Helper function to check remove string
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
