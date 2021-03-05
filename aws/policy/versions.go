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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) GetVersion(arn string, vers string) (*types.PolicyVersion, liberr.Error) {
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

func (cli *client) CompareUpdate(arn string, doc string) (upd bool, err liberr.Error) {
	var (
		pol *types.Policy
		pvs *types.PolicyVersion
		vrs string
	)

	if pol, err = cli.Get(arn); err != nil {
		return false, err
	} else if pol == nil {
		return false, libhlp.ErrorResponse.Error(nil)
	} else {
		vrs = *pol.DefaultVersionId
	}

	if pvs, err = cli.GetVersion(arn, vrs); err != nil {
		return false, err
	} else if pvs == nil {
		return false, libhlp.ErrorResponse.Error(nil)
	} else if *pvs.Document == doc {
		return false, nil
	} else if err = cli.Update(*pol.PolicyName, doc); err != nil {
		return true, err
	} else {
		return true, nil
	}
}
