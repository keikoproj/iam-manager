package awsapi

//go:generate mockgen -destination=mocks/mock_eksiface.go -package=mock_awsapi github.com/aws/aws-sdk-go/service/eks/eksiface EKSAPI

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"

	"github.com/keikoproj/iam-manager/pkg/logging"
)

type EKSIface interface {
	DescribeCluster(ctx context.Context, clusterName string)
}

type EKS struct {
	Client eksiface.EKSAPI
}

func NewEKS(region string) *EKS {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		panic(err)
	}
	return &EKS{
		Client: eks.New(sess),
	}
}

//DescribeCluster function provides cluster info
func (e *EKS) DescribeCluster(ctx context.Context, clusterName string) (*eks.DescribeClusterOutput, error) {
	log := logging.Logger(ctx, "awsapi", "eks", "DescribeCluster")
	log.WithValues("clusterName", clusterName)
	log.V(1).Info("Initiating api call")

	input := &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	}
	resp, err := e.Client.DescribeCluster(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case eks.ErrCodeResourceNotFoundException:
				log.Error(err, eks.ErrCodeResourceNotFoundException)
			case eks.ErrCodeClientException:
				log.Error(err, eks.ErrCodeClientException)
			case eks.ErrCodeServerException:
				log.Error(err, eks.ErrCodeServerException)
			case eks.ErrCodeServiceUnavailableException:
				log.Error(err, eks.ErrCodeServiceUnavailableException)
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

	log.Info("Successfully retrieved cluster info")
	return resp, nil
}
