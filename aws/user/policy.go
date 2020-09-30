package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyPut(policyDocument, policyName, username string) errors.Error {
	_, err := cli.iam.PutUserPolicy(cli.GetContext(), &iam.PutUserPolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String(policyName),
		UserName:       aws.String(username),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyAttach(policyARN, username string) errors.Error {
	_, err := cli.iam.AttachUserPolicy(cli.GetContext(), &iam.AttachUserPolicyInput{
		PolicyArn: aws.String(policyARN),
		UserName:  aws.String(username),
	})

	return cli.GetError(err)
}
