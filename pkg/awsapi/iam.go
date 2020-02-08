package awsapi

//go:generate mockgen -destination=mocks/mock_iamiface.go -package=mock_awsapi github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
//go:generate mockgen -destination=mocks/mock_iam.go -package=mock_awsapi github.com/keikoproj/iam-manager/pkg/awsapi IAMIface

import (
	"context"
	"fmt"
	"github.com/keikoproj/iam-manager/internal/config"

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
	iamTagKey   = "managedBy"
	iamTagValue = "iam-manager"
)

//IAMRoleRequest struct
type IAMRoleRequest struct {
	Name             string
	PolicyName       string
	Description      string
	SessionDuration  int64
	TrustPolicy      string
	PermissionPolicy string
}

type IAMRoleResponse struct {
}

type IAM struct {
	Client iamiface.IAMAPI
}

func New() *IAM {

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
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
	log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(req.TrustPolicy),
		RoleName:                 aws.String(req.Name),
		Description:              aws.String(req.Description),
		MaxSessionDuration:       aws.Int64(req.SessionDuration),
		PermissionsBoundary:      aws.String(config.Props.ManagedPermissionBoundaryPolicy()),
	}

	if err := input.Validate(); err != nil {
		log.Error(err, "input validation failed")
		return nil, err
	}

	roleAlreadyExists := false
	_, err := i.Client.CreateRole(input)
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
	for _, policy := range config.Props.ManagedPolicies() {
		err = i.AttachManagedRolePolicy(ctx, policy, req.Name)
		if err != nil {
			log.Error(err, "Error while attaching managed policy", "policy", policy)
			return &IAMRoleResponse{}, err
		}
	}

	log.V(1).Info("Attaching Inline role policies")
	return i.AttachInlineRolePolicy(ctx, req)
}

//TagRole tags role with appropriate tags
func (i *IAM) TagRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	log := log.Logger(ctx, "awsapi", "iam", "TagRole")
	log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.TagRoleInput{
		RoleName: aws.String(req.Name),
		Tags: []*iam.Tag{
			{
				Key:   aws.String(iamTagKey),
				Value: aws.String(iamTagValue),
			},
		},
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
	log.WithValues("roleName", req.Name)
	log.V(1).Info("Initiating api call")
	input := &iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String(req.Name),
		PermissionsBoundary: aws.String(config.Props.ManagedPermissionBoundaryPolicy()),
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
	log.WithValues("roleName", req.Name)
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
	log.WithValues("roleName", req.Name)
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

//GetRolePolicy gets the role from aws iam
func (i *IAM) GetRolePolicy(ctx context.Context, req IAMRoleRequest) (*string, error) {
	log := log.Logger(ctx, "awsapi", "iam", "GetRolePolicy")
	log.WithValues("roleName", req.Name)
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
func (i *IAM) AttachManagedRolePolicy(ctx context.Context, policyName string, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "AttachManagedRolePolicy")
	log.WithValues("roleName", roleName, "policyName", policyName)
	log.V(1).Info("Initiating api call")
	policyARN := aws.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s", config.Props.AWSAccountID(), policyName))

	_, err := i.Client.AttachRolePolicy(&iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: policyARN,
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
	log.WithValues("roleName", roleName)
	log.V(1).Info("Initiating api call")

	managedPolicyList, err := i.Client.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		log.Error(err, "Unable to list attached managed policies for role")
		return err
	}
	log.V(1).Info("Attached managed for role", "policyList", managedPolicyList.AttachedPolicies)

	// Detach managed policies
	for _, policy := range managedPolicyList.AttachedPolicies {
		if err := i.DetachRolePolicy(ctx, aws.StringValue(policy.PolicyName), roleName); err != nil {
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
	log.WithValues("roleName", roleName, "policyName", policyName)
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
func (i *IAM) DetachRolePolicy(ctx context.Context, policyName string, roleName string) error {
	log := log.Logger(ctx, "awsapi", "iam", "DetachRolePolicy")
	log.WithValues("roleName", roleName, "policyName", policyName)
	log.V(1).Info("Initiating api call")

	policyARN := aws.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s", config.Props.AWSAccountID(), policyName))

	_, err := i.Client.DetachRolePolicy(&iam.DetachRolePolicyInput{
		PolicyArn: policyARN,
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
