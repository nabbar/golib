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

package role

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() ([]types.Role, errors.Error) {
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
