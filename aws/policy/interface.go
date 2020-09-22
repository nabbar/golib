package policy

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

type Policy interface {
	List() (map[string]string, errors.Error)
	Add(name, desc, policy string) (string, errors.Error)
	Update(polArn, polContents string) errors.Error
	Delete(polArn string) errors.Error
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) Policy {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
