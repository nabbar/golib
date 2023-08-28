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

package policy

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

func (cli *client) List() (map[string]string, error) {
	out, err := cli.iam.ListPolicies(cli.GetContext(), &iam.ListPoliciesInput{})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make(map[string]string)

		for _, p := range out.Policies {
			res[*p.PolicyName] = *p.Arn
		}

		return res, nil
	}
}

func (cli *client) Get(arn string) (*types.Policy, error) {
	out, err := cli.iam.GetPolicy(cli.GetContext(), &iam.GetPolicyInput{
		PolicyArn: aws.String(arn),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil || out.Policy == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out.Policy, nil
	}
}

func (cli *client) Add(name, desc, policy string) (string, error) {
	out, err := cli.iam.CreatePolicy(cli.GetContext(), &iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		Description:    aws.String(desc),
		PolicyDocument: aws.String(policy),
	})

	if err != nil {
		return "", cli.GetError(err)
	} else {
		return *out.Policy.Arn, nil
	}
}

func (cli *client) Update(polArn, polContents string) error {
	var (
		pol *types.Policy
		lst map[string]string
		err error
	)

	if pol, err = cli.Get(polArn); err != nil {
		return err
	} else if lst, err = cli.VersionList(polArn, 0, false); err != nil {
		return err
	} else if len(lst) > 0 {
		for v := range lst {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if *pol.DefaultVersionId != v {
				if err = cli.VersionDel(polArn, v); err != nil {
					return err
				}
			}
		}
	}

	return cli.VersionAdd(polArn, polContents)
}

func (cli *client) Delete(polArn string) error {
	out, err := cli.iam.ListPolicyVersions(cli.GetContext(), &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(polArn),
	})

	if err != nil {
		return cli.GetError(err)
	} else {
		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if !v.IsDefaultVersion {
				_, _ = cli.iam.DeletePolicyVersion(cli.GetContext(), &iam.DeletePolicyVersionInput{
					PolicyArn: aws.String(polArn),
					VersionId: v.VersionId,
				})
			}
		}
	}

	if cli.GetContext().Err() != nil {
		return nil
	}

	_, err = cli.iam.DeletePolicy(cli.GetContext(), &iam.DeletePolicyInput{
		PolicyArn: aws.String(polArn),
	})

	return cli.GetError(err)
}
