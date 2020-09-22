package role

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() ([]iam.Role, errors.Error) {
	req := cli.iam.ListRolesRequest(&iam.ListRolesInput{})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out.Roles, nil
	}
}

func (cli *client) Check(name string) (string, errors.Error) {
	req := cli.iam.GetRoleRequest(&iam.GetRoleInput{
		RoleName: aws.String(name),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return "", cli.GetError(err)
	}

	return *out.Role.Arn, nil
}

func (cli *client) Add(name, role string) (string, errors.Error) {
	req := cli.iam.CreateRoleRequest(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(role),
		RoleName:                 aws.String(name),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return "", cli.GetError(err)
	} else {
		return *out.Role.Arn, nil
	}
}

func (cli *client) Delete(roleName string) errors.Error {
	req := cli.iam.DeleteRoleRequest(&iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
