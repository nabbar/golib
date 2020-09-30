package helper

import (
	"context"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	liberr "github.com/nabbar/golib/errors"
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

func (cli Helper) GetError(err ...error) liberr.Error {
	var er = ErrorAws.Error(nil)

	for _, e := range err {
		if e == nil {
			continue
		}
		if n, ok := e.(liberr.Error); ok {
			er.AddParentError(n)
		} else {
			er.AddParent(e)
		}
	}

	if er.HasParent() {
		return er
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

func (c *Helper) GetBucketName() string {
	return c.bkt
}

func (c *Helper) GetBucketAws() *string {
	return aws.String(c.bkt)
}

func (c *Helper) SetBucketName(bucket string) {
	c.bkt = bucket
}
