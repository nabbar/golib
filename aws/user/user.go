package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	req := cli.iam.ListUsersRequest(&iam.ListUsersInput{})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

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

func (cli *client) Get(username string) (*iam.User, errors.Error) {
	req := cli.iam.GetUserRequest(&iam.GetUserInput{
		UserName: aws.String(username),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out.User, nil
}

func (cli *client) Create(username string) errors.Error {
	req := cli.iam.CreateUserRequest(&iam.CreateUserInput{
		UserName: aws.String(username),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	} else if out.User == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(username string) errors.Error {
	req := cli.iam.DeleteUserRequest(&iam.DeleteUserInput{
		UserName: aws.String(username),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
