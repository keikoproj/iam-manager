package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/log"
	"strings"
	"text/template"
)

//GetTrustPolicy constructs trust policy
func GetTrustPolicy(ctx context.Context, role *iammanagerv1alpha1.Iamrole) (string, error) {
	log := log.Logger(ctx, "internal.utils.utils", "GetTrustPolicy")
	tPolicy := role.Spec.AssumeRolePolicyDocument

	var statements []iammanagerv1alpha1.TrustPolicyStatement

	// Is it IRSA use case
	flag, saName := ParseIRSAAnnotation(ctx, role)

	//Construct AssumeRoleWithWebIdentity
	if flag {

		hostPath := fmt.Sprintf("%s", strings.TrimPrefix(config.Props.OIDCIssuerUrl(), "https://"))
		statement := iammanagerv1alpha1.TrustPolicyStatement{
			Effect: "Allow",
			Action: "sts:AssumeRoleWithWebIdentity",
			Principal: iammanagerv1alpha1.Principal{
				Federated: fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", config.Props.AWSAccountID(), hostPath),
			},
			Condition: &iammanagerv1alpha1.Condition{
				StringEquals: map[string]string{
					fmt.Sprintf("%s:sub", hostPath): fmt.Sprintf("system:serviceaccount:%s:%s", role.ObjectMeta.Namespace, saName),
				},
			},
		}
		statements = append(statements, statement)

	} else {
		// NON - IRSA which should cover AssumeRole usecase
		//For default use cases
		if tPolicy == nil || len(tPolicy.Statement) == 0 {
			trustPolicy, err := DefaultTrustPolicy(ctx, config.Props.DefaultTrustPolicy(), role.Namespace)
			if err != nil {
				msg := "unable to get the trust policy. It must follow v1alpha1.AssumeRolePolicyDocument syntax"
				log.Error(err, msg)
				return "", err
			}

			statements = append(statements, trustPolicy.Statement...)
		}
	}

	// If anything included in the request
	if tPolicy != nil && len(tPolicy.Statement) > 0 {
		statements = append(statements, role.Spec.AssumeRolePolicyDocument.Statement...)
	}
	tDoc := iammanagerv1alpha1.AssumeRolePolicyDocument{
		Version:   "2012-10-17",
		Statement: statements,
	}
	//Convert it to string

	output, err := json.Marshal(tDoc)
	if err != nil {
		msg := fmt.Sprintf("malformed trust policy document. unable to marshal it, err = %s", err.Error())
		err := errors.New(msg)
		log.Error(err, msg)
		return "", err
	}
	log.V(1).Info("trust policy generated successfully", "trust_policy", string(output))
	return string(output), nil
}

//Fields Template fields
type Fields struct {
	AccountID     string
	ClusterName   string
	NamespaceName string
	Region        string
}

//DefaultTrustPolicy converts the config map variable string to v1alpha1.AssumeRolePolicyDocument and executes Go Template if any
func DefaultTrustPolicy(ctx context.Context, trustPolicyDoc string, ns string) (*iammanagerv1alpha1.AssumeRolePolicyDocument, error) {
	log := log.Logger(ctx, "internal.utils.utils", "defaultTrustPolicy")
	if trustPolicyDoc == "" {
		msg := "default trust policy is not provided in the config map. Request must provide trust policy in the CR"
		err := errors.New(msg)
		log.Error(err, msg)
		return nil, err
	}
	fields := Fields{
		AccountID:     config.Props.AWSAccountID(),
		ClusterName:   config.Props.ClusterName(),
		NamespaceName: ns,
		Region:        config.Props.AWSRegion(),
	}

	t, err := template.New("trustTemplate").Parse(trustPolicyDoc)
	if err != nil {
		msg := "unable to create go template with default trust policy string"
		log.Error(err, msg)
		return nil, err
	}

	byteBuffer := bytes.NewBuffer([]byte{})
	if err = t.Execute(byteBuffer, fields); err != nil {
		msg := "unable to replace template values in default trust policy string"
		log.Error(err, msg)
		return nil, err
	}

	log.V(1).Info("Default trust policy from cm", "trust_policy", byteBuffer.String())

	var trustPolicy iammanagerv1alpha1.AssumeRolePolicyDocument
	if err := json.Unmarshal(byteBuffer.Bytes(), &trustPolicy); err != nil {
		log.Error(err, "unable to unmarshal default trust policy. It must follow v1alpha1.AssumeRolePolicyDocument syntax")
		return nil, err
	}

	return &trustPolicy, nil
}

// GenerateRoleName returns a roleName that should be created in IAM using
// the supplied iam.role.pattern. This pattern can be customized by the
// end-user.
func GenerateRoleName(ctx context.Context, iamRole iammanagerv1alpha1.Iamrole, props config.Properties) (string, error) {
	log := log.Logger(ctx, "internal.utils.utils", "GenerateRoleNam")
	tmpl, err := template.New("rolename").Parse(props.IamRolePattern())
	if err != nil {
		msg := "unable to parse supplied iam.role.pattern"
		log.Error(err, msg)
		return "", err
	}

	// Write the template output into a buffer and then grab that as a string.
	// There is no way in GoLang natively to do this.
	buf := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buf, "rolename", iamRole)
	if err != nil {
		msg := "unable to execute iam.role.pattern template against the iamrole object"
		log.Error(err, msg)
		return "", err
	}

	return buf.String(), nil
}
