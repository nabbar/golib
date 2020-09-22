package helper

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/nabbar/golib/errors"
)

const (
	STATE_SUSPENDED = "Suspended"
	STATE_ENABLED   = "Enabled"
)

type Helper struct {
	ctx context.Context
	bkt string
}

func New(ctx context.Context, bucket string) Helper {
	return Helper{
		ctx: ctx,
		bkt: bucket,
	}
}

func (cli Helper) GetError(err error) errors.Error {
	if err == nil {
		return nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		return ErrorAws.Error(errors.NewError(0, fmt.Sprintf("(%s) %s", aerr.Code(), aerr.Message()), nil))
	}

	if aerr, ok := err.(errors.Error); ok {
		return ErrorAws.Error(aerr)
	}

	return ErrorAws.ErrorParent(err)
}

func (cli Helper) ErrorCode(err error) string {
	if aerr, ok := err.(awserr.Error); ok {
		return aerr.Code()
	}

	if aerr, ok := err.(errors.Error); ok {
		return aerr.CodeError("")
	}

	return ""
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

func (c *Helper) GetBucketName() string {
	return c.bkt
}

func (c *Helper) GetBucketAws() *string {
	return aws.String(c.bkt)
}

func (c *Helper) SetBucketName(bucket string) {
	c.bkt = bucket
}
