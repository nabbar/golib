package role

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

type Role interface {
	List() ([]iam.Role, errors.Error)
	Check(name string) (string, errors.Error)
	Add(name, role string) (string, errors.Error)
	Delete(roleName string) errors.Error

	PolicyAttach(policyARN, roleName string) errors.Error
	PolicyDetach(policyARN, roleName string) errors.Error

	PolicyListAttached(roleName string) ([]iam.AttachedPolicy, errors.Error)
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) Role {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
