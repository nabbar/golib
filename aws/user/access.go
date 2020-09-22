package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) AccessList(username string) (map[string]bool, errors.Error) {
	var req iam.ListAccessKeysRequest

	if username != "" {
		req = cli.iam.ListAccessKeysRequest(&iam.ListAccessKeysInput{
			UserName: aws.String(username),
		})
	} else {
		req = cli.iam.ListAccessKeysRequest(&iam.ListAccessKeysInput{})
	}

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.AccessKeyMetadata == nil {
		return nil, helper.ErrorResponse.Error(nil)
	} else {
		var res = make(map[string]bool)

		for _, a := range out.AccessKeyMetadata {
			switch a.Status {
			case iam.StatusTypeActive:
				res[*a.AccessKeyId] = true
			case iam.StatusTypeInactive:
				res[*a.AccessKeyId] = false
			}
		}

		return res, nil
	}
}

func (cli *client) AccessCreate(username string) (string, string, errors.Error) {
	var req iam.CreateAccessKeyRequest

	if username != "" {
		req = cli.iam.CreateAccessKeyRequest(&iam.CreateAccessKeyInput{
			UserName: aws.String(username),
		})
	} else {
		req = cli.iam.CreateAccessKeyRequest(&iam.CreateAccessKeyInput{})
	}

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return "", "", cli.GetError(err)
	} else if out.AccessKey == nil {
		return "", "", helper.ErrorResponse.Error(nil)
	} else {
		return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
	}
}

func (cli *client) AccessDelete(username, accessKey string) errors.Error {
	var req iam.DeleteAccessKeyRequest

	if username != "" {
		req = cli.iam.DeleteAccessKeyRequest(&iam.DeleteAccessKeyInput{
			AccessKeyId: aws.String(accessKey),
			UserName:    aws.String(username),
		})
	} else {
		req = cli.iam.DeleteAccessKeyRequest(&iam.DeleteAccessKeyInput{
			AccessKeyId: aws.String(accessKey),
		})
	}

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
