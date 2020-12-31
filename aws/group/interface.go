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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

type client struct {
	helper.Helper
	iam *iam.Client
	s3  *s3.Client
}

type Group interface {
	UserList(username string) ([]string, errors.Error)
	UserCheck(username, groupName string) (errors.Error, bool)
	UserAdd(username, groupName string) errors.Error
	UserRemove(username, groupName string) errors.Error

	List() (map[string]string, errors.Error)
	Add(groupName string) errors.Error
	Remove(groupName string) errors.Error

	PolicyList(groupName string) (map[string]string, errors.Error)
	PolicyAttach(groupName, polArn string) errors.Error
	PolicyDetach(groupName, polArn string) errors.Error
}

func New(ctx context.Context, bucket, region string, iam *iam.Client, s3 *s3.Client) Group {
	return &client{
		Helper: helper.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
