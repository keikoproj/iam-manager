package awsapi

//go:generate mockgen -destination=mocks/mock_iamiface.go -package=mock_awsapi github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
////go:generate mockgen -destination=mocks/mock_iam.go -package=mock_awsapi github.com/keikoproj/iam-manager/pkg/awsapi IAMIface

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/keikoproj/iam-manager/pkg/log"
)

//IAMIface defines interface methods
type IAMIface interface {
	CreateRole(ctx context.Context, req IAMRoleRequest)
	UpdateRole(ctx context.Context, req IAMRoleRequest)
	DeleteRole(ctx context.Context, roleName string)
	AttachInlineRolePolicy(ctx context.Context, req IAMRoleRequest)
	AddPermissionBoundary(ctx context.Context, req IAMRoleRequest) error
	GetRolePolicy(ctx context.Context, req IAMRoleRequest) bool
}

const (
	RoleAlreadyExistsError = "Please choose a different name"
)

//IAMRoleRequest struct
type IAMRoleRequest struct {
	Name                            string
	PolicyName                      string
	Description                     string
	SessionDuration                 int64
	TrustPolicy                     string
	PermissionPolicy                string
	ManagedPermissionBoundaryPolicy string
	ManagedPolicies                 []string
	Tags                            map[string]string
}

type IAMRoleResponse struct {
	RoleARN string
	RoleID  string
}

type IAM struct {
	Client iamiface.IAMAPI
}

func NewIAM(region string) *IAM {

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		panic(err)
	}
	return &IAM{
		Client: iam.New(sess),
	}
}

// CreateRole creates/updates the role
func (i *IAM) CreateRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "CreateRole")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(req.TrustPolicy),
		RoleName:                 aws.String(req.Name),
		Description:              aws.String(req.Description),
		MaxSessionDuration:       aws.Int64(req.SessionDuration),
		PermissionsBoundary:      aws.String(req.ManagedPermissionBoundaryPolicy),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return nil, err
	}

	roleAlreadyExists := false
	iResp, err := i.Client.CreateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// Update the role to the latest spec if it is existed already
			case iam.ErrCodeEntityAlreadyExistsException:
				roleAlreadyExists = true
				log.Info(iam.ErrCodeEntityAlreadyExistsException)
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}
		if !roleAlreadyExists {
			return nil, err
		}
	}

	resp := &IAMRoleResponse{}

	if !roleAlreadyExists {
		resp.RoleARN = aws.StringValue(iResp.Role.Arn)
		resp.RoleID = aws.StringValue(iResp.Role.RoleId)
	}

	//Verify tags
	log.V(1).Info("Verifying Tags")
	_, err = i.VerifyTags(ctx, req)

	if err != nil {
		return &IAMRoleResponse{}, err
	}

	//Attach a tag
	log.V(1).Info("Attaching Tag")
	_, err = i.TagRole(ctx, req)

	if err != nil {
		return &IAMRoleResponse{}, err
	}

	//Add permission boundary
	log.V(1).Info("Attaching Permission Boundary")
	err = i.AddPermissionBoundary(ctx, req)

	if err != nil {
		return &IAMRoleResponse{}, err
	}

	//Attach managed role policy
	log.V(1).Info("Attaching Managed policies")
	for _, policy := range req.ManagedPolicies {
		if policy != "" {
			err = i.AttachManagedRolePolicy(ctx, policy, req.Name)
			if err != nil {
				log.Error(err, "Error while attaching managed policy", "policy", policy)
				return &IAMRoleResponse{}, err
			}
		}
	}

	log.V(1).Info("Attaching Inline role policies")

	_, err = i.UpdateRole(ctx, req)
	if err != nil {
		log.Error(err, "Error while updating role")
		return &IAMRoleResponse{}, err
	}

	return resp, nil
}

//VerifyTags function verifies the tags attached to the role
func (i *IAM) VerifyTags(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "VerifyTags")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	//Lets first list the tags and look for namespace and cluster tags

	listTags, err := i.Client.ListRoleTags(&iam.ListRoleTagsInput{
		RoleName: aws.String(req.Name),
	})

	if err != nil {
		return nil, err
	}

	flag := false
	for _, tag := range listTags.Tags {
		if aws.StringValue(tag.Key) == "Namespace" {
			if aws.StringValue(tag.Value) != req.Tags["Namespace"] {
				flag = true
				break
			}
		}
		if aws.StringValue(tag.Key) == "Cluster" {
			if aws.StringValue(tag.Value) != req.Tags["Cluster"] {
				flag = true
				break
			}
		}
	}

	if flag {
		return nil, fmt.Errorf("role name %s in AWS is not available. %s", req.Name, RoleAlreadyExistsError)
	}

	return &IAMRoleResponse{}, nil
}

//TagRole tags role with appropriate tags
func (i *IAM) TagRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "TagRole")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")

	//Attach the tags
	var tags []*iam.Tag
	for k, v := range req.Tags {
		tag := &iam.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		tags = append(tags, tag)
	}
	input := &iam.TagRoleInput{
		RoleName: aws.String(req.Name),
		Tags:     tags,
	}

	_, err := i.Client.TagRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeInvalidInputException:
				log.Error(err, iam.ErrCodeInvalidInputException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())

		}
		return &IAMRoleResponse{}, err
	}

	log.V(1).Info("Successfully completed TagRole call")
	return &IAMRoleResponse{}, nil
}

//AddPermissionBoundary adds permission boundary to the existing roles
func (i *IAM) AddPermissionBoundary(ctx context.Context, req IAMRoleRequest) error {
	log := log.Logger(ctx, "awsapi", "iam", "AddPermissionBoundary")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String(req.Name),
		PermissionsBoundary: aws.String(req.ManagedPermissionBoundaryPolicy),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return err
	}

	_, err := i.Client.PutRolePermissionsBoundary(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeInvalidInputException:
				log.Error(err, iam.ErrCodeInvalidInputException)
			case iam.ErrCodeUnmodifiableEntityException:
				log.Error(err, iam.ErrCodeUnmodifiableEntityException)
			case iam.ErrCodePolicyNotAttachableException:
				log.Error(err, iam.ErrCodePolicyNotAttachableException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())
		}
		return err
	}

	log.V(1).Info("Successfuly added permission boundary")
	return nil
}

//UpdateRole updates role
func (i *IAM) UpdateRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "UpdateRole")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.UpdateRoleInput{
		RoleName:           aws.String(req.Name),
		MaxSessionDuration: aws.Int64(req.SessionDuration),
		Description:        aws.String(req.Description),
	}
	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return nil, err
	}
	_, err := i.Client.UpdateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, "error in update roles")
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, "error in update role")
			//If access denied, one use case would be it is an existing role and we need to first attach permission boundary
		}
		return nil, err
	}
	//If it is already here means update role is successful, lets move on to Update Assume role policy
	//Lets double check this -- do we want to do this for every update?
	log.V(1).Info("Initiating api call", "api", "UpdateAssumeRolePolicy")

	inputPolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(req.Name),
		PolicyDocument: aws.String(req.TrustPolicy),
	}

	if err := inputPolicy.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return nil, err
	}

	_, err = i.Client.UpdateAssumeRolePolicy(inputPolicy)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())
		}
		return nil, err
	}

	//Attach the Inline policy
	log.V(1).Info("AssumeRole Policy is successfully updated")
	return i.AttachInlineRolePolicy(ctx, req)
}

//AttachInlineRolePolicy function attaches inline policy to the role
func (i *IAM) AttachInlineRolePolicy(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "AttachInlineRolePolicy")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.PutRolePolicyInput{
		RoleName:       aws.String(req.Name),
		PolicyName:     aws.String(req.PolicyName),
		PolicyDocument: aws.String(req.PermissionPolicy),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return nil, err
	}
	_, err := i.Client.PutRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeMalformedPolicyDocumentException:
				log.Error(err, iam.ErrCodeMalformedPolicyDocumentException)
			case iam.ErrCodeUnmodifiableEntityException:
				log.Error(err, iam.ErrCodeUnmodifiableEntityException)
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}
		return nil, err
	}
	log.V(1).Info("Successfully completed attaching InlineRolePolicy")
	return &IAMRoleResponse{}, nil
}

//GetRole gets the role from aws iam
func (i *IAM) GetRole(ctx context.Context, req IAMRoleRequest) (*iam.GetRoleOutput, error) {
	log := log.Logger(ctx, "awsapi", "iam", "GetRole")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	// First get the iam role policy on the AWS IAM side
	input := &iam.GetRoleInput{
		RoleName: aws.String(req.Name),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		//should log the error
		return nil, err
	}

	resp, err := i.Client.GetRole(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}

		return nil, err
	}
	log.V(1).Info("Successfully able to get the role")

	return resp, nil
}

//GetRolePolicy gets the role from aws iam
func (i *IAM) GetRolePolicy(ctx context.Context, req IAMRoleRequest) (*string, error) {
	log := log.Logger(ctx, "awsapi", "iam", "GetRolePolicy")
	log = log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	// First get the iam role policy on the AWS IAM side
	input := &iam.GetRolePolicyInput{
		PolicyName: aws.String(req.PolicyName),
		RoleName:   aws.String(req.Name),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		//should log the error
		return nil, err
	}

	resp, err := i.Client.GetRolePolicy(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}

		return nil, err
	}
	log.V(1).Info("Successfully able to get the policy")

	return resp.PolicyDocument, nil
}

// AttachManagedRolePolicy function attaches managed policy to the role
func (i *IAM) AttachManagedRolePolicy(ctx context.Context, policyArn string, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "AttachManagedRolePolicy")
	log = log.WithValues("roleName", roleName, "policyName", policyArn)
	log.V(1).Info("Initiating api call")

	_, err := i.Client.AttachRolePolicy(&iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.Error(err, iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeInvalidInputException:
				log.Error(err, iam.ErrCodeInvalidInputException)
			case iam.ErrCodeUnmodifiableEntityException:
				log.Error(err, iam.ErrCodeUnmodifiableEntityException)
			case iam.ErrCodePolicyNotAttachableException:
				log.Error(err, iam.ErrCodePolicyNotAttachableException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}
		return err
	}
	log.V(1).Info("Successfully attached managed policies")
	return nil
}

//DeleteRole function deletes the role in the account
func (i *IAM) DeleteRole(ctx context.Context, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "DeleteRole")
	log = log.WithValues("roleName", roleName)
	log.V(1).Info("Initiating api call")

	//Check if role exists

	managedPolicyList, err := i.Client.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NoSuchEntity") {
			log.Info("Role doesn't exist in the target account", "role_name", roleName)
			return nil
		} else {
			log.Error(err, "Unable to list attached managed policies for role")
			return err
		}
	}
	log.V(1).Info("listing attached policies", "policyList", managedPolicyList.AttachedPolicies)

	// Detach managed policies
	for _, policy := range managedPolicyList.AttachedPolicies {
		if err := i.DetachRolePolicy(ctx, aws.StringValue(policy.PolicyArn), roleName); err != nil {
			log.Error(err, "Unable to delete the policy", "policyName", aws.StringValue(policy.PolicyName))
			return err
		}
	}

	inlinePolicies, err := i.Client.ListRolePolicies(&iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		log.Error(err, "Unable to list inline policies for role")
		return err
	}

	// Delete inline policies
	for _, inlinePolicy := range inlinePolicies.PolicyNames {
		if err := i.DeleteInlinePolicy(ctx, aws.StringValue(inlinePolicy), roleName); err != nil {
			log.Error(err, "Unable to delete the policy", "policyName", aws.StringValue(inlinePolicy))
			return err
		}
	}

	input := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}

	_, err = i.Client.DeleteRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeNoSuchEntityException:
				//This is ok
				err = nil
				log.V(1).Info(iam.ErrCodeNoSuchEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		}
		return err
	}
	log.V(1).Info("Successfully deleted the role")
	return nil
}

//DeleteInlinePolicy function deletes inline policy
func (i *IAM) DeleteInlinePolicy(ctx context.Context, policyName string, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "DeleteInlinePolicy")
	log = log.WithValues("roleName", roleName, "policyName", policyName)
	log.V(1).Info("Initiating api call")

	input := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	}

	_, err := i.Client.DeleteRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.V(1).Info(iam.ErrCodeNoSuchEntityException)
				// This is ok
				return nil
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeUnmodifiableEntityException:
				log.Error(err, iam.ErrCodeUnmodifiableEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())
		}
		return err
	}

	log.V(1).Info("Successfully deleted inline policy")
	return nil
}

// DetachRolePolicy detaches a policy from role
func (i *IAM) DetachRolePolicy(ctx context.Context, policyArn string, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "DetachRolePolicy")
	log = log.WithValues("roleName", roleName, "policyArn", policyArn)
	log.V(1).Info("Initiating api call")

	_, err := i.Client.DetachRolePolicy(&iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				log.V(1).Info(iam.ErrCodeNoSuchEntityException)
				// This is ok
				return nil
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeUnmodifiableEntityException:
				log.Error(err, iam.ErrCodeUnmodifiableEntityException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())
		}
		return err
	}

	log.V(1).Info("Successfully detached policy")
	return nil
}

//CreateOIDCProvider creates OIDC IDP provider with AWS IAM
func (i *IAM) CreateOIDCProvider(ctx context.Context, url string, aud string, certThumpPrint string) error {
	log := log.Logger(ctx, "awsapi.iam", "CreateOIDCProvider")
	log = log.WithValues("url", url, "aud", aud)
	log.V(1).Info("Creating OIDC Provider with AWS IAM")

	input := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList:   []*string{aws.String(aud)},
		ThumbprintList: []*string{aws.String(certThumpPrint)},
		Url:            aws.String(url),
	}

	result, err := i.Client.CreateOpenIDConnectProvider(input)
	idpAlreadyExists := false
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeInvalidInputException:
				log.Error(err, iam.ErrCodeInvalidInputException)
			case iam.ErrCodeEntityAlreadyExistsException:
				log.Info("OIDC Provider already exists with this url. This should be okay")
				idpAlreadyExists = true
			case iam.ErrCodeLimitExceededException:
				log.Error(err, iam.ErrCodeLimitExceededException)
			case iam.ErrCodeServiceFailureException:
				log.Error(err, iam.ErrCodeServiceFailureException)
			default:
				log.Error(err, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Error(err, err.Error())
		}

		if !idpAlreadyExists {
			return err
		}
	}
	if idpAlreadyExists {
		log.Info("OIDC Provider already exists. skipping")
		return nil
	}

	//Print the ARN very first time
	log.Info("OIDC Provider created successfully", "Arn", aws.StringValue(result.OpenIDConnectProviderArn))

	return nil
}
