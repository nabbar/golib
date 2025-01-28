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

package aws

import (
	"context"
	"net/url"
	"sync"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awsbck "github.com/nabbar/golib/aws/bucket"
	awsgrp "github.com/nabbar/golib/aws/group"
	awshlp "github.com/nabbar/golib/aws/helper"
	awsobj "github.com/nabbar/golib/aws/object"
	awspol "github.com/nabbar/golib/aws/policy"
	awsrol "github.com/nabbar/golib/aws/role"
	awsusr "github.com/nabbar/golib/aws/user"
	libhtc "github.com/nabbar/golib/httpcli"
)

type Config interface {
	Check(ctx context.Context) error
	Validate() error

	GetAccessKey() string
	GetSecretKey() string
	SetCredentials(accessKey, secretKey string)
	ResetRegionEndpoint()
	RegisterRegionEndpoint(region string, endpoint *url.URL) error
	RegisterRegionAws(endpoint *url.URL) error
	SetRegion(region string)
	GetRegion() string
	SetEndpoint(endpoint *url.URL)
	GetEndpoint() *url.URL

	IsHTTPs() bool
	ResolveEndpoint(service, region string) (sdkaws.Endpoint, error)
	ResolveEndpointWithOptions(service, region string, options ...interface{}) (sdkaws.Endpoint, error)
	GetDisableHTTPS() bool
	GetResolvedRegion() string
	SetRetryer(retryer func() sdkaws.Retryer)

	GetConfig(ctx context.Context, cli libhtc.HttpClient) (*sdkaws.Config, error)
	JSON() ([]byte, error)
	Clone() Config

	GetBucketName() string
	SetBucketName(bucket string)

	SetChecksumValidation(req sdkaws.RequestChecksumCalculation, rsp sdkaws.ResponseChecksumValidation)
}

type AWS interface {
	Bucket() awsbck.Bucket
	Group() awsgrp.Group
	Object() awsobj.Object
	Policy() awspol.Policy
	Role() awsrol.Role
	User() awsusr.User

	Config() Config
	HTTPCli() libhtc.HttpClient
	Clone(ctx context.Context) (AWS, error)
	NewForConfig(ctx context.Context, cfg Config) (AWS, error)
	ForcePathStyle(ctx context.Context, enabled bool) error
	ForceSignerOptions(ctx context.Context, fct ...func(signer *sdksv4.SignerOptions)) error

	GetBucketName() string
	SetBucketName(bucket string)
	SetHTTPTimeout(dur time.Duration) error
	GetHTTPTimeout() time.Duration
	GetClientS3() *sdksss.Client
	SetClientS3(aws *sdksss.Client)
	GetClientIam() *sdkiam.Client
	SetClientIam(aws *sdkiam.Client)
}

func New(ctx context.Context, cfg Config, httpClient libhtc.HttpClient) (AWS, error) {
	if cfg == nil {
		return nil, awshlp.ErrorConfigEmpty.Error(nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cli := &client{
		m: sync.Mutex{},
		p: false,
		o: make([]func(signer *sdksv4.SignerOptions), 0),
		x: ctx,
		c: cfg,
		i: nil,
		s: nil,
		h: httpClient,
	}

	if i, e := cli._NewClientIAM(ctx, httpClient, nil); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli._NewClientS3(ctx, httpClient, nil); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}
