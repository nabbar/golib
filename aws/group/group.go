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
)

func (cli *client) List() (map[string]string, error) {
	if out, err := cli.iam.ListGroups(cli.GetContext(), &sdkiam.ListGroupsInput{}); err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make(map[string]string)

		for _, g := range out.Groups {
			res[*g.GroupId] = *g.GroupName
		}

		return res, nil
	}
}

func (cli *client) Walk(prefix string, fct GroupFunc) error {
	var (
		err error
		out *sdkiam.ListGroupsOutput
		mrk *string
		trk = true
	)

	for trk {
		var in = &sdkiam.ListGroupsInput{}
		if len(prefix) > 0 {
			in.PathPrefix = sdkaws.String(prefix)
		}

		if mrk != nil && len(*mrk) > 0 {
			in.Marker = mrk
		}

		out, err = cli.iam.ListGroups(cli.GetContext(), in)

		if err != nil {
			return cli.GetError(err)
		} else if out == nil || len(out.Groups) < 1 {
			return nil
		} else {
			trk = false
			mrk = nil
		}

		for _, g := range out.Groups {
			if !fct(g) {
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

func (cli *client) detachPoliciesFromGroup(groupName string) error {
	attachedPoliciesOutput, err := cli.iam.ListAttachedGroupPolicies(cli.GetContext(),
		&sdkiam.ListAttachedGroupPoliciesInput{
			GroupName: sdkaws.String(groupName),
		})
	if err != nil {
		return cli.GetError(err)
	}

	for _, policy := range attachedPoliciesOutput.AttachedPolicies {
		_, err := cli.iam.DetachGroupPolicy(cli.GetContext(),
			&sdkiam.DetachGroupPolicyInput{
				GroupName: sdkaws.String(groupName),
				PolicyArn: policy.PolicyArn,
			})
		if err != nil {
			return cli.GetError(err)
		}
	}

	return nil
}

func (cli *client) DetachGroups(prefix string) ([]string, error) {
	var detachedGroupNames []string
	err := cli.Walk(prefix, func(group types.Group) bool {
		if *group.GroupName == "FullAccessGroup" || *group.GroupName == "ReadOnlyGroup" {
			return true
		}

		if err := cli.detachPoliciesFromGroup(*group.GroupName); err != nil {
			return false
		}

		detachedGroupNames = append(detachedGroupNames, *group.GroupName)
		return true
	})

	if err != nil {
		return nil, err
	}

	return detachedGroupNames, nil
}

func (cli *client) Add(groupName string) error {
	_, err := cli.iam.CreateGroup(cli.GetContext(), &sdkiam.CreateGroupInput{
		GroupName: sdkaws.String(groupName),
	})

	return cli.GetError(err)
}

func (cli *client) Remove(groupName string) error {
	_, err := cli.iam.DeleteGroup(cli.GetContext(), &sdkiam.DeleteGroupInput{
		GroupName: sdkaws.String(groupName),
	})

	return cli.GetError(err)
}
