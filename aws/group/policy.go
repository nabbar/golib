package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) PolicyList(groupName string) (map[string]string, errors.Error) {
	req := cli.iam.ListAttachedGroupPoliciesRequest(&iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if out, err := req.Send(cli.GetContext()); err != nil {
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
	req := cli.iam.AttachGroupPolicyRequest(&iam.AttachGroupPolicyInput{
		GroupName: aws.String(groupName),
		PolicyArn: aws.String(polArn),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(groupName, polArn string) errors.Error {
	req := cli.iam.DetachGroupPolicyRequest(&iam.DetachGroupPolicyInput{
		GroupName: aws.String(groupName),
		PolicyArn: aws.String(polArn),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}
