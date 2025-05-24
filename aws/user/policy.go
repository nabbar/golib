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
)

func (cli *client) PolicyPut(policyDocument, policyName, username string) error {
	_, err := cli.iam.PutUserPolicy(cli.GetContext(), &sdkiam.PutUserPolicyInput{
		PolicyDocument: sdkaws.String(policyDocument),
		PolicyName:     sdkaws.String(policyName),
		UserName:       sdkaws.String(username),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyAttachedList(username string) ([]sdktps.AttachedPolicy, error) {
	var (
		err error
		res = make([]sdktps.AttachedPolicy, 0)
		fct = func(pol sdktps.AttachedPolicy) bool {
			res = append(res, pol)
			return true
		}
	)

	err = cli.PolicyAttachedWalk(username, fct)
	return res, err
}

func (cli *client) PolicyAttachedWalk(username string, fct FuncWalkPolicies) error {
	var m *string

	in := &sdkiam.ListAttachedUserPoliciesInput{
		UserName: sdkaws.String(username),
		MaxItems: sdkaws.Int32(1000),
	}

	if fct == nil {
		fct = func(pol sdktps.AttachedPolicy) bool {
			return false
		}
	}

	for {
		if m != nil {
			in.Marker = m
		} else {
			in.Marker = nil
		}

		lst, err := cli.iam.ListAttachedUserPolicies(cli.GetContext(), in)

		if err != nil {
			return cli.GetError(err)
		} else if lst == nil || lst.AttachedPolicies == nil {
			return nil
		}

		for i := range lst.AttachedPolicies {
			if !fct(lst.AttachedPolicies[i]) {
				return nil
			}
		}

		if lst.IsTruncated && lst.Marker != nil {
			m = lst.Marker
		} else {
			return nil
		}
	}
}

func (cli *client) PolicyAttach(policyARN, username string) error {
	_, err := cli.iam.AttachUserPolicy(cli.GetContext(), &sdkiam.AttachUserPolicyInput{
		PolicyArn: sdkaws.String(policyARN),
		UserName:  sdkaws.String(username),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(policyARN, username string) error {
	_, err := cli.iam.DetachUserPolicy(cli.GetContext(), &sdkiam.DetachUserPolicyInput{
		PolicyArn: sdkaws.String(policyARN),
		UserName:  sdkaws.String(username),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetachUsers(prefix string) ([]string, error) {
	var (
		e   error
		err error
		res = make([]string, 0)
		fct = func(user sdktps.User) bool {
			if user.UserName == nil || len(*user.UserName) == 0 {
				return true
			}

			if e = cli.removeGroupsPolicyDetachUsers(*user.UserName); e != nil {
				return false
			} else if e = cli.detachPolicyDetachUsers(*user.UserName); e != nil {
				return false
			}

			res = append(res, *user.UserName)
			return true
		}
	)

	err = cli.Walk(prefix, fct)
	if err != nil {
		return nil, err
	} else if e != nil {
		return nil, e
	} else {
		return res, nil
	}
}

func (cli *client) removeGroupsPolicyDetachUsers(username string) error {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &sdkiam.ListGroupsForUserInput{
		UserName: sdkaws.String(username),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || len(out.Groups) < 1 {
		return nil
	}

	for i := range out.Groups {
		if out.Groups[i].GroupName == nil || len(*out.Groups[i].GroupName) == 0 {
			continue
		}

		_, err = cli.iam.RemoveUserFromGroup(cli.GetContext(), &sdkiam.RemoveUserFromGroupInput{
			UserName:  sdkaws.String(username),
			GroupName: out.Groups[i].GroupName,
		})

		if err != nil {
			return cli.GetError(err)
		}
	}

	return nil
}

func (cli *client) detachPolicyDetachUsers(username string) error {
	out, err := cli.iam.ListAttachedUserPolicies(cli.GetContext(), &sdkiam.ListAttachedUserPoliciesInput{
		UserName: sdkaws.String(username),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || len(out.AttachedPolicies) < 1 {
		return nil
	}

	for i := range out.AttachedPolicies {
		if out.AttachedPolicies[i].PolicyArn == nil || len(*out.AttachedPolicies[i].PolicyArn) == 0 {
			continue
		}

		_, err = cli.iam.DetachUserPolicy(cli.GetContext(), &sdkiam.DetachUserPolicyInput{
			UserName:  sdkaws.String(username),
			PolicyArn: out.AttachedPolicies[i].PolicyArn,
		})

		if err != nil {
			return cli.GetError(err)
		}
	}

	return nil
}
