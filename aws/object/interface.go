package object

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

type client struct {
	helper.Helper
	iam *iam.Client
	s3  *s3.Client
}

type Object interface {
	Find(pattern string) ([]string, errors.Error)
	Size(object string) (size int64, err errors.Error)

	List(continuationToken string) ([]s3.Object, string, int64, errors.Error)
	Head(object string) (head map[string]interface{}, meta map[string]string, err errors.Error)
	Get(object string) (io.ReadCloser, []io.Closer, errors.Error)
	Put(object string, body *bytes.Reader) errors.Error
	Delete(object string) errors.Error

	MultipartPut(object string, body io.Reader) errors.Error
	MultipartPutCustom(partSize helper.PartSize, object string, body io.Reader, concurrent int) errors.Error
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) Object {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
