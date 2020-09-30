package object

import (
	"bytes"
	"context"
	"io"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

type client struct {
	helper.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type Object interface {
	Find(pattern string) ([]string, errors.Error)
	Size(object string) (size int64, err errors.Error)

	List(continuationToken string) ([]*sdktps.Object, string, int64, errors.Error)
	Head(object string) (*sdksss.HeadObjectOutput, errors.Error)
	Get(object string) (*sdksss.GetObjectOutput, errors.Error)
	Put(object string, body *bytes.Reader) errors.Error
	Delete(object string) errors.Error

	MultipartPut(object string, body io.Reader) errors.Error
	MultipartPutCustom(partSize helper.PartSize, object string, body io.Reader) errors.Error
}

func New(ctx context.Context, bucket string, iam *sdkiam.Client, s3 *sdksss.Client) Object {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
