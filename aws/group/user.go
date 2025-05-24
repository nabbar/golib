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

package group

import (
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) UserCheck(username, groupName string) (error, bool) {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &sdkiam.ListGroupsForUserInput{
		UserName: sdkaws.String(username),
	})

	if err != nil {
		return cli.GetError(err), false
	} else {
		for _, g := range out.Groups {
			if *g.GroupName == groupName {
				return nil, true
			}
		}
	}

	return nil, false
}

func (cli *client) UserList(username string) ([]string, error) {
	var (
		res = make([]string, 0)
		fct = func(grp types.Group) bool {
			if grp.GroupName != nil && len(*grp.GroupName) > 3 {
				res = append(res, *grp.GroupName)
			}
			return true
		}
	)

	err := cli.WalkGroupForUser(username, fct)
	return res, err
}

func (cli *client) WalkGroupForUser(username string, fct FuncWalkGroupForUser) error {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &sdkiam.ListGroupsForUserInput{
		UserName: sdkaws.String(username),
	})

	if fct == nil {
		fct = func(grp types.Group) bool {
			return false
		}
	}

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		return libhlp.ErrorAwsEmpty.Error(nil)
	} else {
		for i := range out.Groups {
			if !fct(out.Groups[i]) {
				return nil
			}
		}
	}

	return nil
}

func (cli *client) UserAdd(username, groupName string) error {
	_, err := cli.iam.AddUserToGroup(cli.GetContext(), &sdkiam.AddUserToGroupInput{
		UserName:  sdkaws.String(username),
		GroupName: sdkaws.String(groupName),
	})

	return cli.GetError(err)
}

func (cli *client) UserRemove(username, groupName string) error {
	_, err := cli.iam.RemoveUserFromGroup(cli.GetContext(), &sdkiam.RemoveUserFromGroupInput{
		UserName:  sdkaws.String(username),
		GroupName: sdkaws.String(groupName),
	})

	return cli.GetError(err)
}
