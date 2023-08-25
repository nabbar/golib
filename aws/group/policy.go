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
	sdktps "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (cli *client) PolicyList(groupName string) (map[string]string, error) {
	out, _, err := cli.PolicyAttachedList(groupName, "")

	if err != nil {
		return nil, err
	} else {
		var res = make(map[string]string)

		for _, p := range out {
			res[*p.PolicyName] = *p.PolicyArn
		}

		return res, nil
	}
}

func (cli *client) PolicyAttach(groupName, polArn string) error {
	_, err := cli.iam.AttachGroupPolicy(cli.GetContext(), &sdkiam.AttachGroupPolicyInput{
		GroupName: sdkaws.String(groupName),
		PolicyArn: sdkaws.String(polArn),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(groupName, polArn string) error {
	_, err := cli.iam.DetachGroupPolicy(cli.GetContext(), &sdkiam.DetachGroupPolicyInput{
		GroupName: sdkaws.String(groupName),
		PolicyArn: sdkaws.String(polArn),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyAttachedList(groupName, marker string) ([]sdktps.AttachedPolicy, string, error) {
	in := &sdkiam.ListAttachedGroupPoliciesInput{
		GroupName: sdkaws.String(groupName),
		MaxItems:  sdkaws.Int32(1000),
	}

	if marker != "" {
		in.Marker = sdkaws.String(marker)
	}

	lst, err := cli.iam.ListAttachedGroupPolicies(cli.GetContext(), in)

	if err != nil {
		return nil, "", cli.GetError(err)
	} else if lst == nil || lst.AttachedPolicies == nil {
		return nil, "", nil
	} else if lst.IsTruncated && lst.Marker != nil {
		return lst.AttachedPolicies, *lst.Marker, nil
	} else {
		return lst.AttachedPolicies, "", nil
	}
}

func (cli *client) PolicyAttachedWalk(groupName string, fct PoliciesWalkFunc) error {
	var m *string

	in := &sdkiam.ListAttachedGroupPoliciesInput{
		GroupName: sdkaws.String(groupName),
		MaxItems:  sdkaws.Int32(1000),
	}

	for {
		if m != nil {
			in.Marker = m
		} else {
			in.Marker = nil
		}

		lst, err := cli.iam.ListAttachedGroupPolicies(cli.GetContext(), in)

		if err != nil {
			return cli.GetError(err)
		} else if lst == nil || lst.AttachedPolicies == nil {
			return nil
		}

		var e error
		for _, p := range lst.AttachedPolicies {
			e = fct(e, p)
		}

		if e != nil {
			return e
		}

		if lst.IsTruncated && lst.Marker != nil {
			m = lst.Marker
		} else {
			return nil
		}
	}
}
