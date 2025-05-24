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
	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdktps "github.com/aws/aws-sdk-go-v2/service/iam/types"
)

/*
@DEPRECATED: PolicyAttachedList
*/
func (cli *client) PolicyListAttached(roleName string) ([]sdktps.AttachedPolicy, error) {
	out, _, err := cli.PolicyAttachedList(roleName, "")

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		return out, nil
	}
}

func (cli *client) PolicyAttach(policyARN, roleName string) error {
	_, err := cli.iam.AttachRolePolicy(cli.GetContext(), &sdkiam.AttachRolePolicyInput{
		PolicyArn: sdkaws.String(policyARN),
		RoleName:  sdkaws.String(roleName),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyDetach(policyARN, roleName string) error {
	_, err := cli.iam.DetachRolePolicy(cli.GetContext(), &sdkiam.DetachRolePolicyInput{
		PolicyArn: sdkaws.String(policyARN),
		RoleName:  sdkaws.String(roleName),
	})

	return cli.GetError(err)
}

func (cli *client) PolicyAttachedList(roleName, marker string) ([]sdktps.AttachedPolicy, string, error) {
	in := &sdkiam.ListAttachedRolePoliciesInput{
		RoleName: sdkaws.String(roleName),
		MaxItems: sdkaws.Int32(1000),
	}

	if marker != "" {
		in.Marker = sdkaws.String(marker)
	}

	lst, err := cli.iam.ListAttachedRolePolicies(cli.GetContext(), in)

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

func (cli *client) PolicyAttachedWalk(roleName string, fct FuncWalkPolicies) error {
	var m *string

	in := &sdkiam.ListAttachedRolePoliciesInput{
		RoleName: sdkaws.String(roleName),
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

		lst, err := cli.iam.ListAttachedRolePolicies(cli.GetContext(), in)

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

func (cli *client) PolicyDetachRoles(prefix string) ([]string, error) {
	var (
		e   error
		err error
		res = make([]string, 0)
		fct = func(role sdktps.Role) bool {
			if role.RoleName == nil || len(*role.RoleName) == 0 {
				return true
			} else if role.Arn == nil || len(*role.Arn) == 0 {
				return true
			} else if e = cli.detachPolicyDetachRoles(*role.RoleName); e != nil {
				return false
			}

			res = append(res, *role.RoleName)
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

func (cli *client) detachPolicyDetachRoles(roleName string) error {
	out, err := cli.iam.ListAttachedRolePolicies(cli.GetContext(), &sdkiam.ListAttachedRolePoliciesInput{
		RoleName: sdkaws.String(roleName),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil || out.AttachedPolicies == nil {
		return nil
	}

	for i := range out.AttachedPolicies {
		if out.AttachedPolicies[i].PolicyArn == nil || len(*out.AttachedPolicies[i].PolicyArn) == 0 {
			continue
		}

		_, err = cli.iam.DetachRolePolicy(cli.GetContext(), &sdkiam.DetachRolePolicyInput{
			RoleName:  sdkaws.String(roleName),
			PolicyArn: out.AttachedPolicies[i].PolicyArn,
		})

		if err != nil {
			return cli.GetError(err)
		}
	}

	return nil
}
