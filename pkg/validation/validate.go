package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/internal/utils"
	"github.com/keikoproj/iam-manager/pkg/awsapi"
	"github.com/keikoproj/iam-manager/pkg/logging"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

//ValidateIAMPolicyAction validates policy action
func ValidateIAMPolicyAction(ctx context.Context, pDoc v1alpha1.PolicyDocument) *field.Error {
	log := logging.Logger(ctx, "pkg.validation", "ValidateIAMPolicyAction")

	//Check the incoming policy actions
	for _, statement := range pDoc.Statement {
		if statement.Effect == "Deny" {
			//This should ignore the validation for all Deny action
			continue
		}
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
	log := logging.Logger(ctx, "pkg.validation", "ValidateIAMPolicyResource")

	//Check the incoming policy resource
	for _, statement := range pDoc.Statement {
		if statement.Effect == "Deny" {
			//This should ignore the validation for all Deny action
			continue
		}
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

//CompareRole function compares input role to target role
func CompareRole(ctx context.Context, request awsapi.IAMRoleRequest, targetRole *iam.GetRoleOutput, targetRolePolicy string) bool {
	log := logging.Logger(ctx, "pkg.validation", "ComparePolicy")

	// Step 1: Compare the permission policy
	if !ComparePermissionPolicy(ctx, request.PermissionPolicy, targetRolePolicy) {
		return false
	}

	//Step 2: Compare Assume Role Policy Document
	if !CompareAssumeRolePolicy(ctx, request.TrustPolicy, *targetRole.Role.AssumeRolePolicyDocument) {
		return false
	}
	//Step 3: Compare Permission Boundary
	if !reflect.DeepEqual(request.ManagedPermissionBoundaryPolicy, *targetRole.Role.PermissionsBoundary.PermissionsBoundaryArn) {
		log.Info("input permission boundary and target permission boundary are NOT equal")
		return false
	}

	//Step 4: Compare Tags
	if !CompareTags(ctx, request.Tags, targetRole.Role.Tags) {
		return false
	}

	return true
}

//CompareRole function compares input role to target role
func CompareRoleIRSA(ctx context.Context, sa *v1.ServiceAccount, props config.Properties) bool {
	// Check if sts-regional-endpoint annotation is set to the expected value
	exists, val := utils.ParseIRSARegionalEndpointAnnotation(ctx, sa)
	// If the regional endpoint disabled flag is not set, make sure the annotation exists and is set to true
	if !props.IsIRSARegionalEndpointDisabled() && (!exists || val != "true") {
		return false
		// If the regional endpoint disabled flag is set to true, make sure the annotation either doesn't exist or is set to false
	} else if props.IsIRSARegionalEndpointDisabled() && (exists && val != "false") {
		return false
	}

	return true
}

//ComparePermissionPolicy compares role policy from request and response
func ComparePermissionPolicy(ctx context.Context, request string, target string) bool {
	log := logging.Logger(ctx, "pkg.validation", "ComparePermissionPolicy")

	d, _ := url.QueryUnescape(target)
	dest := v1alpha1.PolicyDocument{}
	err := json.Unmarshal([]byte(d), &dest)
	if err != nil {
		log.Error(err, "failed to unmarshal policy document")
	}

	req := v1alpha1.PolicyDocument{}
	err = json.Unmarshal([]byte(request), &req)
	if err != nil {
		log.Error(err, "failed to marshal policy document")
	}
	//compare
	if !reflect.DeepEqual(req, dest) {
		log.Info("input policy and target policy are NOT equal", "req", req, "dest", dest)
		return false
	}

	return true
}

//CompareAssumeRolePolicy compares assume role policy from request and response
func CompareAssumeRolePolicy(ctx context.Context, request string, target string) bool {
	log := logging.Logger(ctx, "pkg.validation", "CompareAssumeRolePolicy")

	a, _ := url.QueryUnescape(target)
	destAssume := v1alpha1.AssumeRolePolicyDocument{}
	err := json.Unmarshal([]byte(a), &destAssume)
	if err != nil {
		log.Error(err, "failed to unmarshal assume role policy document")
	}

	reqAssume := v1alpha1.AssumeRolePolicyDocument{}
	err = json.Unmarshal([]byte(request), &reqAssume)
	if err != nil {
		log.Error(err, "failed to marshal assume role policy document")
	}
	//compare
	if !reflect.DeepEqual(reqAssume, destAssume) {
		log.Info("input assume role policy and target assume role policy are NOT equal", "req", reqAssume, "dest", destAssume)
		return false
	}

	return true
}

//CompareTags compares tags from request and response
func CompareTags(ctx context.Context, request map[string]string, target []*iam.Tag) bool {
	log := logging.Logger(ctx, "pkg.validation", "CompareTags")
	log.Info("start CompareTags")

	var targetTags map[string]string
	if len(target) > 0 {
		targetTags = make(map[string]string)
		if target != nil {
			for _, tag := range target {
				if tag != nil && tag.Key != nil && tag.Value != nil && aws.StringValue(tag.Key) != "" {
					key := aws.StringValue(tag.Key)
					value := aws.StringValue(tag.Value)
					targetTags[key] = value
				}
			}
		}
	}

	//compare
	if !reflect.DeepEqual(request, targetTags) {
		log.Info("input tags and target tags are NOT equal", "req", request, "dest", targetTags)
		return false
	}

	return true
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
