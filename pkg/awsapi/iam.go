package awsapi

//go:generate mockgen -destination=mocks/mock_iamiface.go -package=mock_awsapi github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
//go:generate mockgen -destination=mocks/mock_iam.go -package=mock_awsapi github.com/keikoproj/iam-manager/pkg/awsapi IAMIface

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

//IAMIface defines interface methods
type IAMIface interface {
	CreateRole(ctx context.Context, req IAMRoleRequest)
	GetRole(ctx context.Context, roleName string)
	UpdateRole(ctx context.Context, req IAMRoleRequest)
	DeleteRole(ctx context.Context, roleName string)
	AttachInlineRolePolicy(ctx context.Context, req IAMRoleRequest)
	AddPermissionBoundary(ctx context.Context, req IAMRoleRequest)
}

const (
	iamTagKey   = "managedBy"
	iamTagValue = "iam-manager"
)

var (
	IamManagedPermissionBoundaryPolicy = "arn:aws:iam::%s:policy/iam-manager-permission-boundary"
	ManagedPolicies                    []string
	AwsAccountId                       string
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

	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(req.TrustPolicy),
		RoleName:                 aws.String(req.Name),
		Description:              aws.String(req.Description),
		MaxSessionDuration:       aws.Int64(req.SessionDuration),
		PermissionsBoundary:      aws.String(IamManagedPermissionBoundaryPolicy),
	}

	if err := input.Validate(); err != nil {
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
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		if !roleAlreadyExists {
			return nil, err
		}
	}

	//Attach a tag
	_, err = i.TagRole(ctx, req)

	if err != nil {
		return &IAMRoleResponse{}, err
	}

	//Attach inline role policy
	err = i.AddPermissionBoundary(ctx, req)

	if err != nil {
		return &IAMRoleResponse{}, err
	}

	//Attach managed role policy
	for _, policy := range ManagedPolicies {
		err = i.AttachManagedRolePolicy(ctx, policy, req.Name)
		if err != nil {
			fmt.Printf("Error while attaching managed policy %s: %v", policy, err)
			return &IAMRoleResponse{}, err
		}
	}

	return i.AttachInlineRolePolicy(ctx, req)
}

//TagRole tags role with appropriate tags
func (i *IAM) TagRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	input := &iam.TagRoleInput{
		RoleName: aws.String(req.Name),
		Tags: []*iam.Tag{
			{
				Key:   aws.String(iamTagKey),
				Value: aws.String(iamTagValue),
			},
		},
	}

	result, err := i.Client.TagRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())

		}
		return &IAMRoleResponse{}, err
	}

	fmt.Println(result)
	return &IAMRoleResponse{}, nil
}

//AddPermissionBoundary adds permission boundary to the existing roles
func (i *IAM) AddPermissionBoundary(ctx context.Context, req IAMRoleRequest) error {
	input := &iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String(req.Name),
		PermissionsBoundary: aws.String(IamManagedPermissionBoundaryPolicy),
	}

	if err := input.Validate(); err != nil {
		return err
	}

	_, err := i.Client.PutRolePermissionsBoundary(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeInvalidInputException:
				fmt.Println(iam.ErrCodeInvalidInputException, aerr.Error())
			case iam.ErrCodeUnmodifiableEntityException:
				fmt.Println(iam.ErrCodeUnmodifiableEntityException, aerr.Error())
			case iam.ErrCodePolicyNotAttachableException:
				fmt.Println(iam.ErrCodePolicyNotAttachableException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	return nil
}

//UpdateRole updates role
func (i *IAM) UpdateRole(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	input := &iam.UpdateRoleInput{
		RoleName:           aws.String(req.Name),
		MaxSessionDuration: aws.Int64(req.SessionDuration),
		Description:        aws.String(req.Description),
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	_, err := i.Client.UpdateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println("error in update roles" + err.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println("error in update role" + err.Error())
			//If access denied, one use case would be it is an existing role and we need to first attach permission boundary
		}
		return nil, err
	}
	//If it is already here means update role is successful, lets move on to Update Assume role policy
	//Lets double check this -- do we want to do this for every update?
	inputPolicy := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(req.Name),
		PolicyDocument: aws.String(req.TrustPolicy),
	}

	if err := inputPolicy.Validate(); err != nil {
		return nil, err
	}

	_, err = i.Client.UpdateAssumeRolePolicy(inputPolicy)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	//Attach the Inline policy
	return i.AttachInlineRolePolicy(ctx, req)
}

//AttachInlineRolePolicy function attaches inline policy to the role
func (i *IAM) AttachInlineRolePolicy(ctx context.Context, req IAMRoleRequest) (*IAMRoleResponse, error) {
	input := &iam.PutRolePolicyInput{
		RoleName:       aws.String(req.Name),
		PolicyName:     aws.String(req.PolicyName),
		PolicyDocument: aws.String(req.PermissionPolicy),
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}
	_, err := i.Client.PutRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return nil, err
	}
	return &IAMRoleResponse{}, nil
}

// AttachManagedRolePolicy function attaches managed policy to the role
func (i *IAM) AttachManagedRolePolicy(ctx context.Context, policyName string, roleName string) error {

	policyARN := aws.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s", AwsAccountId, policyName))

	_, err := i.Client.AttachRolePolicy(&iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: policyARN,
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				fmt.Println(iam.ErrCodeEntityAlreadyExistsException, aerr.Error())
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return err
	}
	return nil
}

//DeleteRole function deletes the role in the account
func (i *IAM) DeleteRole(ctx context.Context, roleName string) error {

	for _, policy := range ManagedPolicies {
		if err := i.DetachRolePolicy(ctx, policy, roleName); err != nil {
			fmt.Printf("Unable to detach the policy %s", policy)
			return err
		}
	}

	//Lets first delete inline policy
	policyName := fmt.Sprintf("%s-policy", roleName)
	if err := i.DeleteInlinePolicy(ctx, policyName, roleName); err != nil {
		fmt.Println("Unable to delete the policy")
		return err
	}

	/*// Detach remaining policies
	attachedPolicies, err := i.Client.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		fmt.Printf("Unable to list attached the policies for role %s", roleName)
		return err
	}
	fmt.Println("Attached policies are: ", attachedPolicies.AttachedPolicies)*/

	input := &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	}

	_, err := i.Client.DeleteRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeNoSuchEntityException:
				//This is ok
				err = nil
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return err
	}
	return nil
}

//DeleteInlinePolicy function deletes inline policy
func (i *IAM) DeleteInlinePolicy(ctx context.Context, policyName string, roleName string) error {
	input := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	}

	result, err := i.Client.DeleteRolePolicy(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
				// This is ok
				return nil
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeUnmodifiableEntityException:
				fmt.Println(iam.ErrCodeUnmodifiableEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	fmt.Println(result)
	return nil
}

// DetachRolePolicy detaches a policy from role
func (i *IAM) DetachRolePolicy(ctx context.Context, policyName string, roleName string) error {

	fmt.Println("Detaching role policy: ", policyName)
	policyARN := aws.String(fmt.Sprintf("arn:aws:iam::%s:policy/%s", AwsAccountId, policyName))

	result, err := i.Client.DetachRolePolicy(&iam.DetachRolePolicyInput{
		PolicyArn: policyARN,
		RoleName:  aws.String(roleName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeNoSuchEntityException:
				fmt.Println(iam.ErrCodeNoSuchEntityException, aerr.Error())
				// This is ok
				return nil
			case iam.ErrCodeLimitExceededException:
				fmt.Println(iam.ErrCodeLimitExceededException, aerr.Error())
			case iam.ErrCodeUnmodifiableEntityException:
				fmt.Println(iam.ErrCodeUnmodifiableEntityException, aerr.Error())
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return err
	}

	fmt.Println(result)
	return nil
}
