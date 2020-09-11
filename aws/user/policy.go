package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyPut(policyDocument, policyName, username string) errors.Error {
	req := cli.iam.PutUserPolicyRequest(&iam.PutUserPolicyInput{
		PolicyDocument: aws.String(policyDocument),
		PolicyName:     aws.String(policyName),
		UserName:       aws.String(username),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) PolicyAttach(policyARN, username string) errors.Error {
	req := cli.iam.AttachUserPolicyRequest(&iam.AttachUserPolicyInput{
		PolicyArn: aws.String(policyARN),
		UserName:  aws.String(username),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
