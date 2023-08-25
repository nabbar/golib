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

package helper

import (
	"context"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	STATE_SUSPENDED = "Suspended"
	STATE_ENABLED   = "Enabled"
)

type Helper struct {
	ctx context.Context
	bkt string
	reg string
}

func New(ctx context.Context, bucket, region string) Helper {
	return Helper{
		ctx: ctx,
		bkt: bucket,
		reg: region,
	}
}

func (cli Helper) GetError(err ...error) error {
	var e = ErrorAws.Error()
	e.Add(err...)
	if e.HasParent() {
		return e
	}
	return nil
}

func (cli *Helper) GetContext() context.Context {
	if cli.ctx == nil {
		cli.ctx = context.Background()
	}

	return cli.ctx
}

func (c *Helper) GetCloser(req *http.Request, rsp *http.Response) []io.Closer {
	res := make([]io.Closer, 0)

	if req != nil && req.Body != nil {
		res = append(res, req.Body)
	}

	if rsp != nil && rsp.Body != nil {
		res = append(res, rsp.Body)
	}

	return res
}

func (c *Helper) Close(req *http.Request, rsp *http.Response) {
	if req != nil && req.Body != nil {
		_ = req.Body.Close()
	}
	if rsp != nil && rsp.Body != nil {
		_ = rsp.Body.Close()
	}
}

func (c *Helper) GetRegion() string {
	return c.reg
}

func (c *Helper) GetRegionAws() *string {
	return aws.String(c.reg)
}

func (c *Helper) GetBucketName() string {
	return c.bkt
}

func (c *Helper) GetBucketAws() *string {
	return aws.String(c.bkt)
}

func (c *Helper) SetBucketName(bucket string) {
	c.bkt = bucket
}
