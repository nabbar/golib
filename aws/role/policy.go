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

func (cli *client) PolicyListAttached(roleName string) ([]types.AttachedPolicy, errors.Error) {
	out, err := cli.iam.ListAttachedRolePolicies(cli.GetContext(), &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out.AttachedPolicies, nil
	}
}

func (cli *client) PolicyAttach(policyARN, roleName string) errors.Error {
	_, err := cli.iam.AttachRolePolicy(cli.GetContext(), &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(policyARN, roleName string) errors.Error {
	_, err := cli.iam.DetachRolePolicy(cli.GetContext(), &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyARN),
		RoleName:  aws.String(roleName),
	})

	return cli.GetError(err)
}
