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
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
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

func (cli *client) Add(name, desc, policy string) (string, errors.Error) {
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

func (cli *client) Update(polArn, polContents string) errors.Error {
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

	_, err = cli.iam.CreatePolicyVersion(cli.GetContext(), &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(polArn),
		PolicyDocument: aws.String(polContents),
		SetAsDefault:   true,
	})

	return cli.GetError(err)
}

func (cli *client) Delete(polArn string) errors.Error {
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
