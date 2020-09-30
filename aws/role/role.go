package role

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() ([]*types.Role, errors.Error) {
	out, err := cli.iam.ListRoles(cli.GetContext(), &iam.ListRolesInput{})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out.Roles, nil
	}
}

func (cli *client) Check(name string) (string, errors.Error) {
	out, err := cli.iam.GetRole(cli.GetContext(), &iam.GetRoleInput{
		RoleName: aws.String(name),
	})

	if err != nil {
		return "", cli.GetError(err)
	}

	return *out.Role.Arn, nil
}

func (cli *client) Add(name, role string) (string, errors.Error) {
	out, err := cli.iam.CreateRole(cli.GetContext(), &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(role),
		RoleName:                 aws.String(name),
	})

	if err != nil {
		return "", cli.GetError(err)
	} else {
		return *out.Role.Arn, nil
	}
}

func (cli *client) Delete(roleName string) errors.Error {
	_, err := cli.iam.DeleteRole(cli.GetContext(), &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})

	return cli.GetError(err)
}
