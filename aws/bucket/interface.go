package bucket

import (
	"context"

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

type Bucket interface {
	Check() errors.Error

	List() ([]s3.Bucket, errors.Error)
	Create() errors.Error
	Delete() errors.Error

	//FindObject(pattern string) ([]string, errors.Error)

	SetVersioning(state bool) errors.Error
	GetVersioning() (string, errors.Error)

	EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) errors.Error
	DeleteReplication() errors.Error
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) Bucket {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
