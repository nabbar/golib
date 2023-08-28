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

func (cli *client) PolicyAttachedList(username, marker string) ([]sdktps.AttachedPolicy, string, error) {
	in := &sdkiam.ListAttachedUserPoliciesInput{
		UserName: sdkaws.String(username),
		MaxItems: sdkaws.Int32(1000),
	}

	if marker != "" {
		in.Marker = sdkaws.String(marker)
	}

	lst, err := cli.iam.ListAttachedUserPolicies(cli.GetContext(), in)

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

func (cli *client) PolicyAttachedWalk(username string, fct PoliciesWalkFunc) error {
	var m *string

	in := &sdkiam.ListAttachedUserPoliciesInput{
		UserName: sdkaws.String(username),
		MaxItems: sdkaws.Int32(1000),
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
