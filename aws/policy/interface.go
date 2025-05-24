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
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	iamtps "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libhlp "github.com/nabbar/golib/aws/helper"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}
type FuncWalkPolicy func(pol iamtps.Policy) bool

type Policy interface {
	List() (map[string]string, error)
	Get(arn string) (*iamtps.Policy, error)
	Add(name, desc, policy string) (string, error)
	Update(polArn, polContents string) error
	Delete(polArn string) error
	VersionList(arn string, maxItem int32, noDefaultVersion bool) (map[string]string, error)
	VersionGet(arn string, vers string) (*iamtps.PolicyVersion, error)
	VersionAdd(arn string, doc string) error
	VersionDel(arn string, vers string) error
	CompareUpdate(arn string, doc string) (upd bool, err error)

	Walk(prefix string, fct FuncWalkPolicy) error
	GetAllPolicies(prefix string) ([]string, error)
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Policy {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
