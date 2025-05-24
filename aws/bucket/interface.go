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

package bucket

import (
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type WalkFunc func(bucket sdkstp.Bucket) bool

type Bucket interface {
	Check() error

	List() ([]sdkstp.Bucket, error)
	Walk(f WalkFunc) error
	Create(RegionConstraint string) error
	CreateWithLock(RegionConstraint string) error
	Delete() error

	//FindObject(pattern string) ([]string, errors.Error)

	SetVersioning(state bool) error
	GetVersioning() (string, error)

	LoadReplication() (*sdkstp.ReplicationConfiguration, error)
	EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) error
	DeleteReplication() error

	PutWebsite(index, error string) error
	GetWebsite() (*sdksss.GetBucketWebsiteOutput, error)

	SetCORS(cors []sdkstp.CORSRule) error
	GetCORS() ([]sdkstp.CORSRule, error)

	GetACL() (*sdkstp.AccessControlPolicy, error)
	SetACL(ACP *sdkstp.AccessControlPolicy, cannedACL sdkstp.BucketCannedACL, header ACLHeaders) error
	SetACLPolicy(ACP *sdkstp.AccessControlPolicy) error
	SetACLHeader(cannedACL sdkstp.BucketCannedACL, header ACLHeaders) error

	GetLifeCycle() ([]sdkstp.LifecycleRule, error)
	SetLifeCycle(rules ...sdkstp.LifecycleRule) error
	GetLock() (*sdkstp.ObjectLockConfiguration, error)
	SetLock(cfg sdkstp.ObjectLockConfiguration, token string) error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Bucket {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
