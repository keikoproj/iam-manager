package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/log"
	"strings"
)

//TrustPolicy struct
type TrustPolicy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

//Statement struct
type Statement struct {
	Effect    Effect `json:"Effect"`
	Action    string `json:"Action"`
	Principal iammanagerv1alpha1.Principal
}

// Effect describes whether to allow or deny the specific action
// Allowed values are
// - "Allow" : allows the specific action on resources
// - "Deny" : denies the specific action on resources
// +kubebuilder:validation:Enum=Allow;Deny
type Effect string

//Lets use template

//GetTrustPolicy constructs trust policy
func GetTrustPolicy(ctx context.Context, tPolicy *iammanagerv1alpha1.TrustPolicy) (string, error) {
	log := log.Logger(ctx, "internal.utils.utils", "GetTrustPolicy")
	trustPolicy := &TrustPolicy{
		Version: "2012-10-17",
		Statement: []Statement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
			},
		},
	}
	//Same trust policy for all the roles use case
	//Retrieve the trust policy role arn from config map
	if tPolicy == nil || (len(tPolicy.Principal.AWS) == 0 && tPolicy.Principal.Service == "") {
		if len(config.Props.TrustPolicyARNs()) < 1 {
			msg := "default trust policy is not provided in the config map. Request must provide trust policy in the CR"
			err := errors.New(msg)
			log.Error(err, msg)
			return "", err
		}
		var aws []string
		for _, arn := range config.Props.TrustPolicyARNs() {
			aws = append(aws, arn)
		}
		trustPolicy.Statement[0].Principal.AWS = aws
	} else {
		if len(tPolicy.Principal.AWS) != 0 {
			var aws []string
			for _, arn := range tPolicy.Principal.AWS {
				aws = append(aws, arn)
			}
			trustPolicy.Statement[0].Principal.AWS = aws
		}
		if tPolicy.Principal.Service != "" {
			//May be lets validate that service must end with .amazonaws.com
			if !strings.HasSuffix(tPolicy.Principal.Service, "amazonaws.com") {
				msg := fmt.Sprintf("service %s must end with amazonaws.com in TrustPolicy", tPolicy.Principal.Service)
				err := errors.New(msg)
				log.Error(err, msg)
				return "", err
			}
			trustPolicy.Statement[0].Principal.Service = tPolicy.Principal.Service
		}
	}

	//Convert it to string

	output, err := json.Marshal(trustPolicy)
	if err != nil {
		msg := fmt.Sprintf("malformed trust policy document. unable to marshal it, err = %s", err.Error())
		err := errors.New(msg)
		log.Error(err, msg)
		return "", err
	}
	log.V(1).Info("trust policy generated successfully", "trust_policy", string(output))
	return string(output), nil
}
