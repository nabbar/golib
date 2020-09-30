package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyList(groupName string) (map[string]string, errors.Error) {
	out, err := cli.iam.ListAttachedGroupPolicies(cli.GetContext(), &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make(map[string]string)

		for _, p := range out.AttachedPolicies {
			res[*p.PolicyName] = *p.PolicyArn
		}

		return res, nil
	}
}

func (cli *client) PolicyAttach(groupName, polArn string) errors.Error {
	_, err := cli.iam.AttachGroupPolicy(cli.GetContext(), &iam.AttachGroupPolicyInput{
		GroupName: aws.String(groupName),
		PolicyArn: aws.String(polArn),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(groupName, polArn string) errors.Error {
	_, err := cli.iam.DetachGroupPolicy(cli.GetContext(), &iam.DetachGroupPolicyInput{
		GroupName: aws.String(groupName),
		PolicyArn: aws.String(polArn),
	})

	return cli.GetError(err)
}
