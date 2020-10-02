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
