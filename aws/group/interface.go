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

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdktps "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awshlp "github.com/nabbar/golib/aws/helper"
)

type client struct {
	awshlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type PoliciesWalkFunc func(err error, pol sdktps.AttachedPolicy) error

type Group interface {
	UserList(username string) ([]string, error)
	UserCheck(username, groupName string) (error, bool)
	UserAdd(username, groupName string) error
	UserRemove(username, groupName string) error

	List() (*sdkiam.ListGroupsOutput, error)
	Add(groupName string) error
	Remove(groupName string) error

	PolicyList(groupName string) (map[string]string, error)
	PolicyAttach(groupName, polArn string) error
	PolicyDetach(groupName, polArn string) error
	PolicyAttachedList(groupName, marker string) ([]sdktps.AttachedPolicy, string, error)
	PolicyAttachedWalk(groupName string, fct PoliciesWalkFunc) error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Group {
	return &client{
		Helper: awshlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
