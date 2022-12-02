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
	"sync"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awsbck "github.com/nabbar/golib/aws/bucket"
	awsgrp "github.com/nabbar/golib/aws/group"
	awsobj "github.com/nabbar/golib/aws/object"
	awspol "github.com/nabbar/golib/aws/policy"
	awsrol "github.com/nabbar/golib/aws/role"
	awsusr "github.com/nabbar/golib/aws/user"
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	m sync.Mutex
	p bool
	o []func(signer *sdksv4.SignerOptions)
	x context.Context
	c Config
	i *sdkiam.Client
	s *sdksss.Client
	h *http.Client
}

func (c *client) _NewClientIAM(ctx context.Context, httpClient *http.Client) (*sdkiam.Client, liberr.Error) {
	var (
		cfg *sdkaws.Config
		iam *sdkiam.Client
		err liberr.Error
		ret sdkaws.Retryer
		sig *sdksv4.Signer
	)

	if httpClient == nil {
		httpClient = c.h
	}

	if cfg, err = c.c.GetConfig(ctx, httpClient); err != nil {
		return nil, err
	}

	if cfg.Retryer != nil {
		ret = cfg.Retryer()
	}

	if len(c.o) > 0 {
		sig = sdksv4.NewSigner(c.o...)
	} else {
		sig = sdksv4.NewSigner()
	}

	iam = sdkiam.New(sdkiam.Options{
		APIOptions:  cfg.APIOptions,
		Credentials: cfg.Credentials,
		EndpointOptions: sdkiam.EndpointResolverOptions{
			DisableHTTPS: !c.c.IsHTTPs(),
		},
		EndpointResolver: c._NewIAMResolver(cfg),
		HTTPSignerV4:     sig,
		Region:           cfg.Region,
		Retryer:          ret,
		HTTPClient:       httpClient,
	})

	return iam, nil
}

func (c *client) _NewClientS3(ctx context.Context, httpClient *http.Client) (*sdksss.Client, liberr.Error) {
	var (
		sss *sdksss.Client
		err liberr.Error
		ret sdkaws.Retryer
		cfg *sdkaws.Config
		sig *sdksv4.Signer
	)

	if httpClient == nil {
		httpClient = c.h
	}

	if cfg, err = c.c.GetConfig(ctx, httpClient); err != nil {
		return nil, err
	}

	if cfg.Retryer != nil {
		ret = cfg.Retryer()
	}

	if len(c.o) > 0 {
		sig = sdksv4.NewSigner(c.o...)
	} else {
		sig = sdksv4.NewSigner()
	}

	sss = sdksss.New(sdksss.Options{
		APIOptions:  cfg.APIOptions,
		Credentials: cfg.Credentials,
		EndpointOptions: sdksss.EndpointResolverOptions{
			DisableHTTPS: !c.c.IsHTTPs(),
		},
		EndpointResolver: c._NewS3Resolver(cfg),
		HTTPSignerV4:     sig,
		Region:           cfg.Region,
		Retryer:          ret,
		HTTPClient:       httpClient,
		UsePathStyle:     c.p,
	})

	return sss, nil
}

func (c *client) NewForConfig(ctx context.Context, cfg Config) (AWS, liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	n := &client{
		m: sync.Mutex{},
		p: c.p,
		x: c.x,
		c: cfg,
		o: c.o,
		i: nil,
		s: nil,
		h: c.h,
	}

	if i, e := n._NewClientIAM(ctx, c.h); e != nil {
		return nil, e
	} else {
		n.i = i
	}

	if s, e := n._NewClientS3(ctx, c.h); e != nil {
		return nil, e
	} else {
		n.s = s
	}

	return n, nil
}

func (c *client) Clone(ctx context.Context) (AWS, liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	n := &client{
		m: sync.Mutex{},
		p: c.p,
		x: c.x,
		c: c.c.Clone(),
		o: c.o,
		i: nil,
		s: nil,
		h: c.h,
	}

	if i, e := n._NewClientIAM(ctx, c.h); e != nil {
		return nil, e
	} else {
		n.i = i
	}

	if s, e := n._NewClientS3(ctx, c.h); e != nil {
		return nil, e
	} else {
		n.s = s
	}

	return n, nil
}

func (c *client) ForcePathStyle(ctx context.Context, enabled bool) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	c.p = enabled

	if s, e := c._NewClientS3(ctx, nil); e != nil {
		return e
	} else {
		c.s = s
	}

	return nil
}

func (c *client) ForceSignerOptions(ctx context.Context, fct ...func(signer *sdksv4.SignerOptions)) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	c.o = fct

	if i, e := c._NewClientIAM(ctx, nil); e != nil {
		return e
	} else {
		c.i = i
	}

	if s, e := c._NewClientS3(ctx, nil); e != nil {
		return e
	} else {
		c.s = s
	}

	return nil
}

func (c *client) Config() Config {
	c.m.Lock()
	defer c.m.Unlock()

	return c.c
}

func (c *client) HTTPCli() *http.Client {
	c.m.Lock()
	defer c.m.Unlock()

	return c.h
}

func (c *client) Bucket() awsbck.Bucket {
	c.m.Lock()
	defer c.m.Unlock()

	return awsbck.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Group() awsgrp.Group {
	c.m.Lock()
	defer c.m.Unlock()

	return awsgrp.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Object() awsobj.Object {
	c.m.Lock()
	defer c.m.Unlock()

	return awsobj.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Policy() awspol.Policy {
	c.m.Lock()
	defer c.m.Unlock()

	return awspol.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Role() awsrol.Role {
	c.m.Lock()
	defer c.m.Unlock()

	return awsrol.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) User() awsusr.User {
	c.m.Lock()
	defer c.m.Unlock()

	return awsusr.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) GetBucketName() string {
	c.m.Lock()
	defer c.m.Unlock()

	return c.c.GetBucketName()
}

func (c *client) SetBucketName(bucket string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.c.SetBucketName(bucket)
}

func (c *client) GetClientS3() *sdksss.Client {
	c.m.Lock()
	defer c.m.Unlock()

	return c.s
}

func (c *client) SetClientS3(aws *sdksss.Client) {
	c.m.Lock()
	defer c.m.Unlock()

	c.s = aws
}

func (c *client) GetClientIam() *sdkiam.Client {
	c.m.Lock()
	defer c.m.Unlock()

	return c.i
}

func (c *client) SetClientIam(aws *sdkiam.Client) {
	c.m.Lock()
	defer c.m.Unlock()

	c.i = aws
}
