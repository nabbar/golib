package user

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) AccessList(username string) (map[string]bool, errors.Error) {
	var req = &iam.ListAccessKeysInput{}

	if username != "" {
		req = &iam.ListAccessKeysInput{
			UserName: aws.String(username),
		}
	}

	out, err := cli.iam.ListAccessKeys(cli.GetContext(), req)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.AccessKeyMetadata == nil {
		return nil, helper.ErrorResponse.Error(nil)
	} else {
		var res = make(map[string]bool)

		for _, a := range out.AccessKeyMetadata {
			switch a.Status {
			case types.StatusTypeActive:
				res[*a.AccessKeyId] = true
			case types.StatusTypeInactive:
				res[*a.AccessKeyId] = false
			}
		}

		return res, nil
	}
}

func (cli *client) AccessCreate(username string) (string, string, errors.Error) {
	var req = &iam.CreateAccessKeyInput{}

	if username != "" {
		req = &iam.CreateAccessKeyInput{
			UserName: aws.String(username),
		}
	}

	out, err := cli.iam.CreateAccessKey(cli.GetContext(), req)

	if err != nil {
		return "", "", cli.GetError(err)
	} else if out.AccessKey == nil {
		return "", "", helper.ErrorResponse.Error(nil)
	} else {
		return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
	}
}

func (cli *client) AccessDelete(username, accessKey string) errors.Error {
	var req = &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(accessKey),
	}

	if username != "" {
		req = &iam.DeleteAccessKeyInput{
			AccessKeyId: aws.String(accessKey),
			UserName:    aws.String(username),
		}
	}

	_, err := cli.iam.DeleteAccessKey(cli.GetContext(), req)

	return cli.GetError(err)
}
