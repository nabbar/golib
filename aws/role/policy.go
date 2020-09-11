package role

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyListAttached(roleName string) ([]iam.AttachedPolicy, errors.Error) {
	req := cli.iam.ListAttachedRolePoliciesRequest(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out.AttachedPolicies, nil
	}
}

func (cli *client) PolicyAttach(policyARN, roleName string) errors.Error {
	req := cli.iam.AttachRolePolicyRequest(&iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(policyARN, roleName string) errors.Error {
	req := cli.iam.DetachRolePolicyRequest(&iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
