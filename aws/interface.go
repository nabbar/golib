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
	"net/http"
	"net/url"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
	"github.com/nabbar/golib/errors"
)

type Config interface {
	Check(ctx context.Context) errors.Error
	Validate() errors.Error

	ResetRegionEndpoint()
	RegisterRegionEndpoint(region string, endpoint *url.URL) errors.Error
	RegisterRegionAws(endpoint *url.URL) errors.Error
	SetRegion(region string)
	GetRegion() string
	SetEndpoint(endpoint *url.URL)
	GetEndpoint() *url.URL

	IsHTTPs() bool
	ResolveEndpoint(service, region string) (sdkaws.Endpoint, error)
	SetRetryer(retryer func() sdkaws.Retryer)

	GetConfig(ctx context.Context, cli *http.Client) (*sdkaws.Config, errors.Error)
	JSON() ([]byte, error)
	Clone() Config

	GetBucketName() string
	SetBucketName(bucket string)
}

type AWS interface {
	Bucket() bucket.Bucket
	Group() group.Group
	Object() object.Object
	Policy() policy.Policy
	Role() role.Role
	User() user.User

	Clone(ctx context.Context) (AWS, errors.Error)
	Config() Config
	ForcePathStyle(ctx context.Context, enabled bool) errors.Error

	GetBucketName() string
	SetBucketName(bucket string)
}

type client struct {
	p bool
	x context.Context
	c Config
	i *sdkiam.Client
	s *sdksss.Client
	h *http.Client
}

func New(ctx context.Context, cfg Config, httpClient *http.Client) (AWS, errors.Error) {
	if cfg == nil {
		return nil, helper.ErrorConfigEmpty.Error(nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cli := &client{
		p: false,
		x: ctx,
		c: cfg,
		i: nil,
		s: nil,
		h: httpClient,
	}

	if i, e := cli.newClientIAM(ctx, httpClient); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli.newClientS3(ctx, httpClient); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}

func (cli *client) newClientIAM(ctx context.Context, httpClient *http.Client) (*sdkiam.Client, errors.Error) {
	var (
		c *sdkaws.Config
		i *sdkiam.Client
		e errors.Error
		r sdkaws.Retryer
	)

	if httpClient == nil {
		httpClient = cli.h
	}

	if c, e = cli.c.GetConfig(ctx, httpClient); e != nil {
		return nil, e
	}

	if c.Retryer != nil {
		r = c.Retryer()
	}

	i = sdkiam.New(sdkiam.Options{
		APIOptions:  c.APIOptions,
		Credentials: c.Credentials,
		EndpointOptions: sdkiam.EndpointResolverOptions{
			DisableHTTPS: !cli.c.IsHTTPs(),
		},
		EndpointResolver: cli.newIAMResolver(c),
		HTTPSignerV4:     sdksv4.NewSigner(),
		Region:           c.Region,
		Retryer:          r,
		HTTPClient:       httpClient,
	})

	return i, nil
}

func (cli *client) newClientS3(ctx context.Context, httpClient *http.Client) (*sdksss.Client, errors.Error) {
	var (
		c *sdkaws.Config
		s *sdksss.Client
		e errors.Error
		r sdkaws.Retryer
	)

	if httpClient == nil {
		httpClient = cli.h
	}

	if c, e = cli.c.GetConfig(ctx, httpClient); e != nil {
		return nil, e
	}

	if c.Retryer != nil {
		r = c.Retryer()
	}

	s = sdksss.New(sdksss.Options{
		APIOptions:  c.APIOptions,
		Credentials: c.Credentials,
		EndpointOptions: sdksss.EndpointResolverOptions{
			DisableHTTPS: !cli.c.IsHTTPs(),
		},
		EndpointResolver: cli.newS3Resolver(c),
		HTTPSignerV4:     sdksv4.NewSigner(),
		Region:           c.Region,
		Retryer:          r,
		HTTPClient:       httpClient,
		UsePathStyle:     cli.p,
	})

	return s, nil
}

func (c *client) Clone(ctx context.Context) (AWS, errors.Error) {
	cli := &client{
		p: false,
		x: c.x,
		c: c.c,
		i: nil,
		s: nil,
		h: c.h,
	}

	if i, e := cli.newClientIAM(ctx, c.h); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli.newClientS3(ctx, c.h); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}
