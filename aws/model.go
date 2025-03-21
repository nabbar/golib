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
	"fmt"
	"net/http"
	"sync"
	"time"

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
	libhtc "github.com/nabbar/golib/httpcli"
)

type client struct {
	m sync.Mutex
	p bool
	o []func(signer *sdksv4.SignerOptions)
	x context.Context
	c Config
	i *sdkiam.Client
	s *sdksss.Client
	h libhtc.HttpClient
}

func (c *client) SetHTTPTimeout(dur time.Duration) error {
	c.m.Lock()
	defer c.m.Unlock()

	var h libhtc.HttpClient

	if c.h == nil {
		return fmt.Errorf("missing http client")
	} else if cli, ok := c.h.(*http.Client); !ok {
		return fmt.Errorf("not a standard http client, cannot change timeout")
	} else {
		h = &http.Client{
			Transport:     cli.Transport,
			CheckRedirect: cli.CheckRedirect,
			Jar:           cli.Jar,
			Timeout:       dur,
		}
	}

	if cli, err := c._NewClientS3(c.x, h, c.s); err != nil {
		return err
	} else {
		c.s = cli
	}

	if cli, err := c._NewClientIAM(c.x, h, c.i); err != nil {
		return err
	} else {
		c.i = cli
	}

	return nil
}

func (c *client) GetHTTPTimeout() time.Duration {
	c.m.Lock()
	defer c.m.Unlock()

	if c.h == nil {
		return 0
	} else if cli, ok := c.h.(*http.Client); !ok {
		return 0
	} else {
		return cli.Timeout
	}
}

func (c *client) _NewClientIAM(ctx context.Context, httpClient libhtc.HttpClient, cli *sdkiam.Client) (*sdkiam.Client, error) {
	var (
		cfg *sdkaws.Config
		iam *sdkiam.Client
		err error
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

	if cli == nil {
		iam = sdkiam.New(sdkiam.Options{
			APIOptions:  cfg.APIOptions,
			Credentials: cfg.Credentials,
			EndpointOptions: sdkiam.EndpointResolverOptions{
				DisableHTTPS: !c.c.IsHTTPs(),
			},
			BaseEndpoint:       sdkaws.String(c.c.GetEndpoint().String()),
			EndpointResolver:   c._NewIAMResolver(cfg),
			EndpointResolverV2: c._NewIAMResolverV2(c.c),
			HTTPSignerV4:       sig,
			Region:             cfg.Region,
			Retryer:            ret,
			HTTPClient:         httpClient,
		})
	} else {
		opt := cli.Options()
		opt.HTTPClient = httpClient
		opt.HTTPSignerV4 = sig
		iam = sdkiam.New(opt)
	}

	return iam, nil
}

func (c *client) _NewClientS3(ctx context.Context, httpClient libhtc.HttpClient, cli *sdksss.Client) (*sdksss.Client, error) {
	var (
		sss *sdksss.Client
		err error
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

	if cli == nil {
		sss = sdksss.New(sdksss.Options{
			APIOptions:  cfg.APIOptions,
			Credentials: cfg.Credentials,
			EndpointOptions: sdksss.EndpointResolverOptions{
				DisableHTTPS: !c.c.IsHTTPs(),
			},
			BaseEndpoint:       sdkaws.String(c.c.GetEndpoint().String()),
			EndpointResolver:   c._NewS3Resolver(cfg),
			EndpointResolverV2: c._NewS3ResolverV2(c.c),
			HTTPSignerV4:       sig,
			Region:             cfg.Region,
			Retryer:            ret,
			HTTPClient:         httpClient,
			UsePathStyle:       c.p,
		}, c.updateConfigS3)
	} else {
		opt := cli.Options()
		opt.HTTPClient = httpClient
		opt.HTTPSignerV4 = sig
		opt.UsePathStyle = c.p
		sss = sdksss.New(opt, c.updateConfigS3)
	}

	return sss, nil
}

func (c *client) updateConfigIAM(opt *sdkiam.Options) {

}

func (c *client) updateConfigS3(opt *sdksss.Options) {
	req, rsp := c.c.GetChecksumValidation()
	opt.RequestChecksumCalculation = req
	opt.ResponseChecksumValidation = rsp
}

func (c *client) NewForConfig(ctx context.Context, cfg Config) (AWS, error) {
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

	if i, e := n._NewClientIAM(ctx, c.h, nil); e != nil {
		return nil, e
	} else {
		n.i = i
	}

	if s, e := n._NewClientS3(ctx, c.h, nil); e != nil {
		return nil, e
	} else {
		n.s = s
	}

	return n, nil
}

func (c *client) Clone(ctx context.Context) (AWS, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if ctx == nil {
		ctx = c.x
	}

	n := &client{
		m: sync.Mutex{},
		p: c.p,
		x: ctx,
		c: c.c.Clone(),
		o: c.o,
		i: nil,
		s: nil,
		h: c.h,
	}

	if i, e := n._NewClientIAM(ctx, c.h, nil); e != nil {
		return nil, e
	} else {
		n.i = i
	}

	if s, e := n._NewClientS3(ctx, c.h, nil); e != nil {
		return nil, e
	} else {
		n.s = s
	}

	return n, nil
}

func (c *client) ForcePathStyle(ctx context.Context, enabled bool) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.p = enabled

	if s, e := c._NewClientS3(ctx, nil, c.s); e != nil {
		return e
	} else {
		c.s = s
	}

	if i, e := c._NewClientIAM(ctx, nil, c.i); e != nil {
		return e
	} else {
		c.i = i
	}

	return nil
}

func (c *client) ForceSignerOptions(ctx context.Context, fct ...func(signer *sdksv4.SignerOptions)) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.o = fct

	if i, e := c._NewClientIAM(ctx, nil, c.i); e != nil {
		return e
	} else {
		c.i = i
	}

	if s, e := c._NewClientS3(ctx, nil, c.s); e != nil {
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

func (c *client) HTTPCli() libhtc.HttpClient {
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
