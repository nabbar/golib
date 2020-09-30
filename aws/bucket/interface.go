package bucket

import (
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	ligerr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type Bucket interface {
	Check() ligerr.Error

	List() ([]*sdkstp.Bucket, ligerr.Error)
	Create() ligerr.Error
	Delete() ligerr.Error

	//FindObject(pattern string) ([]string, errors.Error)

	SetVersioning(state bool) ligerr.Error
	GetVersioning() (string, ligerr.Error)

	EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) ligerr.Error
	DeleteReplication() ligerr.Error
}

func New(ctx context.Context, bucket string, iam *sdkiam.Client, s3 *sdksss.Client) Bucket {
	return &client{
		Helper: libhlp.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
