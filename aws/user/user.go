package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	out, err := cli.iam.ListUsers(cli.GetContext(), &iam.ListUsersInput{})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.Users == nil {
		return nil, helper.ErrorResponse.Error(nil)
	} else {
		var res = make(map[string]string)

		for _, u := range out.Users {
			res[*u.UserId] = *u.UserName
		}

		return res, nil
	}
}

func (cli *client) Get(username string) (*types.User, errors.Error) {
	out, err := cli.iam.GetUser(cli.GetContext(), &iam.GetUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out.User, nil
}

func (cli *client) Create(username string) errors.Error {
	out, err := cli.iam.CreateUser(cli.GetContext(), &iam.CreateUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out.User == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(username string) errors.Error {
	_, err := cli.iam.DeleteUser(cli.GetContext(), &iam.DeleteUserInput{
		UserName: aws.String(username),
	})

	return cli.GetError(err)
}
