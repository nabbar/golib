package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) LoginCheck(username string) errors.Error {
	req := cli.iam.GetLoginProfileRequest(&iam.GetLoginProfileInput{
		UserName: aws.String(username),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) LoginCreate(username, password string) errors.Error {
	req := cli.iam.CreateLoginProfileRequest(&iam.CreateLoginProfileInput{
		UserName:              aws.String(username),
		Password:              aws.String(password),
		PasswordResetRequired: aws.Bool(false),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	} else if out.LoginProfile == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) LoginDelete(username string) errors.Error {
	req := cli.iam.DeleteLoginProfileRequest(&iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
