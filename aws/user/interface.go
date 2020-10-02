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
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdkitp "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type User interface {
	List() (map[string]string, liberr.Error)
	Get(username string) (*sdkitp.User, liberr.Error)
	Create(username string) liberr.Error
	Delete(username string) liberr.Error

	PolicyPut(policyDocument, policyName, username string) liberr.Error
	PolicyAttach(policyARN, username string) liberr.Error

	LoginCheck(username string) liberr.Error
	LoginCreate(username, password string) liberr.Error
	LoginDelete(username string) liberr.Error

	AccessList(username string) (map[string]bool, liberr.Error)
	AccessCreate(username string) (string, string, liberr.Error)
	AccessDelete(username, accessKey string) liberr.Error
}

func New(ctx context.Context, bucket string, iam *sdkiam.Client, s3 *sdksss.Client) User {
	return &client{
		Helper: libhlp.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
