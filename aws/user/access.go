/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package user

import (
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdktps "github.com/aws/aws-sdk-go-v2/service/iam/types"
	awshlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) AccessListAll() ([]sdktps.AccessKeyMetadata, liberr.Error) {
	var req = &sdkiam.ListAccessKeysInput{}

	out, err := cli.iam.ListAccessKeys(cli.GetContext(), req)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.AccessKeyMetadata == nil {
		return nil, awshlp.ErrorResponse.Error(nil)
	} else {
		return out.AccessKeyMetadata, nil
	}
}

func (cli *client) AccessList(username string) (map[string]bool, liberr.Error) {
	var req = &sdkiam.ListAccessKeysInput{}

	if username != "" {
		req = &sdkiam.ListAccessKeysInput{
			UserName: sdkaws.String(username),
		}
	}

	out, err := cli.iam.ListAccessKeys(cli.GetContext(), req)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.AccessKeyMetadata == nil {
		return nil, awshlp.ErrorResponse.Error(nil)
	} else {
		var res = make(map[string]bool)

		for _, a := range out.AccessKeyMetadata {
			switch a.Status {
			case sdktps.StatusTypeActive:
				res[*a.AccessKeyId] = true
			case sdktps.StatusTypeInactive:
				res[*a.AccessKeyId] = false
			}
		}

		return res, nil
	}
}

func (cli *client) AccessCreate(username string) (string, string, liberr.Error) {
	var req = &sdkiam.CreateAccessKeyInput{}

	if username != "" {
		req = &sdkiam.CreateAccessKeyInput{
			UserName: sdkaws.String(username),
		}
	}

	out, err := cli.iam.CreateAccessKey(cli.GetContext(), req)

	if err != nil {
		return "", "", cli.GetError(err)
	} else if out.AccessKey == nil {
		return "", "", awshlp.ErrorResponse.Error(nil)
	} else {
		return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
	}
}

func (cli *client) AccessDelete(username, accessKey string) liberr.Error {
	var req = &sdkiam.DeleteAccessKeyInput{
		AccessKeyId: sdkaws.String(accessKey),
	}

	if username != "" {
		req = &sdkiam.DeleteAccessKeyInput{
			AccessKeyId: sdkaws.String(accessKey),
			UserName:    sdkaws.String(username),
		}
	}

	_, err := cli.iam.DeleteAccessKey(cli.GetContext(), req)

	return cli.GetError(err)
}
