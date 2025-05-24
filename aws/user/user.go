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
	iamtps "github.com/aws/aws-sdk-go-v2/service/iam/types"
	awshlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) List() (map[string]string, error) {
	var (
		err error
		res = make(map[string]string, 0)
		fct = func(user iamtps.User) bool {
			if user.UserName == nil || len(*user.UserName) == 0 {
				return true
			} else if user.UserId == nil || len(*user.UserId) == 0 {
				return true
			}

			res[*user.UserName] = *user.UserId
			return true
		}
	)

	err = cli.Walk("", fct)
	return res, err
}

func (cli *client) Walk(prefix string, fct FuncWalkUsers) error {
	var (
		err error
		out *sdkiam.ListUsersOutput
		mrk *string
		trk = true
	)

	if fct == nil {
		fct = func(user iamtps.User) bool {
			return false
		}
	}

	for trk {
		var in = &sdkiam.ListUsersInput{}

		if len(prefix) > 0 {
			in.PathPrefix = sdkaws.String(prefix)
		}

		if mrk != nil && len(*mrk) > 0 {
			in.Marker = mrk
		}

		out, err = cli.iam.ListUsers(cli.GetContext(), in)

		if err != nil {
			return cli.GetError(err)
		} else if out == nil || len(out.Users) < 1 {
			return nil
		} else {
			trk = false
			mrk = nil
		}

		for i := range out.Users {
			if !fct(out.Users[i]) {
				return nil
			}
		}

		if out.IsTruncated && out.Marker != nil && len(*out.Marker) > 0 {
			trk = true
			mrk = out.Marker
		}
	}
	return nil
}

func (cli *client) Get(username string) (*iamtps.User, error) {
	out, err := cli.iam.GetUser(cli.GetContext(), &sdkiam.GetUserInput{
		UserName: sdkaws.String(username),
	})

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out.User, nil
}

func (cli *client) Create(username string) error {
	out, err := cli.iam.CreateUser(cli.GetContext(), &sdkiam.CreateUserInput{
		UserName: sdkaws.String(username),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out.User == nil {
		return awshlp.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(username string) error {
	_, err := cli.iam.DeleteUser(cli.GetContext(), &sdkiam.DeleteUserInput{
		UserName: sdkaws.String(username),
	})

	return cli.GetError(err)
}
