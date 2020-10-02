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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) UserCheck(username, groupName string) (errors.Error, bool) {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
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

func (cli *client) UserList(username string) ([]string, errors.Error) {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make([]string, 0)

		for _, g := range out.Groups {
			res = append(res, *g.GroupName)
		}

		return res, nil
	}
}

func (cli *client) UserAdd(username, groupName string) errors.Error {
	_, err := cli.iam.AddUserToGroup(cli.GetContext(), &iam.AddUserToGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}

func (cli *client) UserRemove(username, groupName string) errors.Error {
	_, err := cli.iam.RemoveUserFromGroup(cli.GetContext(), &iam.RemoveUserFromGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}
