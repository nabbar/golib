/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package policy

import (
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

const maxItemList int32 = 1000

func (cli *client) VersionList(arn string, maxItem int32) (map[string]string, liberr.Error) {
	if arn == "" {
		//nolint #goerr113
		return nil, libhlp.ErrorParamsEmpty.ErrorParent(fmt.Errorf("arn is empty"))
	}

	if maxItem < 1 {
		maxItem = maxItemList
	}

	var (
		marker = ""
		res    = make(map[string]string)
	)

	for {
		in := iam.ListPolicyVersionsInput{
			PolicyArn: aws.String(arn),
			MaxItems:  aws.Int32(maxItem),
		}

		if marker != "" {
			in.Marker = aws.String(marker)
		}

		out, err := cli.iam.ListPolicyVersions(cli.GetContext(), &in)

		if err != nil {
			return nil, cli.GetError(err)
		} else if out == nil || out.Versions == nil {
			return nil, libhlp.ErrorResponse.Error(nil)
		} else if len(out.Versions) < 1 {
			return res, nil
		}

		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil, nil
			}

			if v.VersionId == nil || len(*v.VersionId) < 1 {
				continue
			}

			if v.Document == nil || len(*v.Document) < 1 {
				res[*v.VersionId] = ""
			} else {
				res[*v.VersionId] = *v.Document
			}
		}

		if out.IsTruncated && out.Marker != nil && len(*out.Marker) > 0 {
			marker = *out.Marker
		} else {
			break
		}
	}

	return res, nil
}

func (cli *client) VersionAdd(arn string, doc string) liberr.Error {
	out, err := cli.iam.CreatePolicyVersion(cli.GetContext(), &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(arn),
		PolicyDocument: aws.String(doc),
		SetAsDefault:   true,
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		return libhlp.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) VersionGet(arn string, vers string) (*types.PolicyVersion, liberr.Error) {
	out, err := cli.iam.GetPolicyVersion(cli.GetContext(), &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(vers),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil || out.PolicyVersion == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out.PolicyVersion, nil
	}
}

func (cli *client) VersionDel(arn string, vers string) liberr.Error {
	out, err := cli.iam.DeletePolicyVersion(cli.GetContext(), &iam.DeletePolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(vers),
	})

	if err != nil {
		return cli.GetError(err)
	} else if out == nil {
		return libhlp.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) CompareUpdate(arn string, doc string) (upd bool, err liberr.Error) {
	var (
		e   error
		pol *types.Policy
		pvs *types.PolicyVersion
		vrs string
		dec string
	)

	if pol, err = cli.Get(arn); err != nil {
		return false, err
	} else if pol == nil {
		return false, libhlp.ErrorResponse.Error(nil)
	} else {
		vrs = *pol.DefaultVersionId
	}

	if pvs, err = cli.VersionGet(arn, vrs); err != nil {
		return false, err
	} else if pvs == nil {
		return false, libhlp.ErrorResponse.Error(nil)
	} else if *pvs.Document == doc {
		return false, nil
	} else if dec, e = url.QueryUnescape(*pvs.Document); e != nil {
		dec = *pvs.Document
	}

	if dec == doc {
		return false, nil
	} else if err = cli.Update(*pol.Arn, doc); err != nil {
		return true, err
	} else {
		return true, nil
	}
}
