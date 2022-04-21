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
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type Bucket interface {
	Check() liberr.Error

	List() ([]sdkstp.Bucket, liberr.Error)
	Create(RegionConstraint string) liberr.Error
	Delete() liberr.Error

	//FindObject(pattern string) ([]string, errors.Error)

	SetVersioning(state bool) liberr.Error
	GetVersioning() (string, liberr.Error)

	LoadReplication() (*sdkstp.ReplicationConfiguration, liberr.Error)
	EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) liberr.Error
	DeleteReplication() liberr.Error

	PutWebsite(index, error string) liberr.Error
	GetWebsite() (*sdksss.GetBucketWebsiteOutput, liberr.Error)

	SetCORS(cors []sdkstp.CORSRule) liberr.Error
	GetCORS() ([]sdkstp.CORSRule, liberr.Error)

	GetACL() (*sdkstp.AccessControlPolicy, liberr.Error)
	SetACL(ACP *sdkstp.AccessControlPolicy, cannedACL sdkstp.BucketCannedACL, header ACLHeaders) liberr.Error
	SetACLPolicy(ACP *sdkstp.AccessControlPolicy) liberr.Error
	SetACLHeader(cannedACL sdkstp.BucketCannedACL, header ACLHeaders) liberr.Error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Bucket {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
