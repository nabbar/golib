package role

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyListAttached(roleName string) ([]*types.AttachedPolicy, errors.Error) {
	out, err := cli.iam.ListAttachedRolePolicies(cli.GetContext(), &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out.AttachedPolicies, nil
	}
}

func (cli *client) PolicyAttach(policyARN, roleName string) errors.Error {
	_, err := cli.iam.AttachRolePolicy(cli.GetContext(), &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(policyARN, roleName string) errors.Error {
	_, err := cli.iam.DetachRolePolicy(cli.GetContext(), &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	return cli.GetError(err)
}
