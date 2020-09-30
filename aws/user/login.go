package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) LoginCheck(username string) errors.Error {
	_, err := cli.iam.GetLoginProfile(cli.GetContext(), &iam.GetLoginProfileInput{
		UserName: aws.String(username),
	})

	return cli.GetError(err)
}

func (cli *client) LoginCreate(username, password string) errors.Error {
	out, err := cli.iam.CreateLoginProfile(cli.GetContext(), &iam.CreateLoginProfileInput{
		UserName:              aws.String(username),
		Password:              aws.String(password),
		PasswordResetRequired: aws.Bool(false),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out.LoginProfile == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) LoginDelete(username string) errors.Error {
	_, err := cli.iam.DeleteLoginProfile(cli.GetContext(), &iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	})

	return cli.GetError(err)
}
