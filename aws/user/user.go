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
	"fmt"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamtps "github.com/aws/aws-sdk-go-v2/service/iam/types"
	awshlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) List() (map[string]string, error) {
	out, err := cli.iam.ListUsers(cli.GetContext(), &sdkiam.ListUsersInput{})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out.Users == nil {
		return nil, awshlp.ErrorResponse.Error(nil)
	} else {
		var res = make(map[string]string)

		for _, u := range out.Users {
			res[*u.UserId] = *u.UserName
		}

		return res, nil
	}
}
func (cli *client) detachUserFromGroupsAndPolicies(username string) error {
	groups, err := cli.iam.ListGroupsForUser(cli.GetContext(), &sdkiam.ListGroupsForUserInput{
		UserName: sdkaws.String(username),
	})
	if err != nil {
		return cli.GetError(err)
	}

	for _, group := range groups.Groups {
		_, err := cli.iam.RemoveUserFromGroup(cli.GetContext(), &sdkiam.RemoveUserFromGroupInput{
			UserName:  sdkaws.String(username),
			GroupName: group.GroupName,
		})
		if err != nil {
			return cli.GetError(err)
		}
	}

	attachedPoliciesOutput, err := cli.iam.ListAttachedUserPolicies(cli.GetContext(), &sdkiam.ListAttachedUserPoliciesInput{
		UserName: sdkaws.String(username),
	})
	if err != nil {
		return cli.GetError(err)
	}

	for _, policy := range attachedPoliciesOutput.AttachedPolicies {
		_, err := cli.iam.DetachUserPolicy(cli.GetContext(), &sdkiam.DetachUserPolicyInput{
			UserName:  sdkaws.String(username),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return cli.GetError(err)
		}
	}

	return nil
}

func (cli *client) Walk(prefix string, fct UserFunc) error {
	var (
		err error
		out *sdkiam.ListUsersOutput
		mrk *string
		trk = true
	)

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

		for _, u := range out.Users {
			if !fct(u) {
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

func (cli *client) DetachUsers(prefix string) ([]string, error) {
	var detachedUsernames []string

	err := cli.Walk(prefix, func(user types.User) bool {
		if err := cli.detachUserFromGroupsAndPolicies(*user.UserName); err != nil {
			fmt.Println(err)
			return false
		}
		detachedUsernames = append(detachedUsernames, *user.UserName)
		return true
	})

	if err != nil {
		return nil, err
	}

	return detachedUsernames, nil
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
