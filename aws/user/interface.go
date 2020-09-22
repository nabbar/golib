package user

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

type User interface {
	List() (map[string]string, errors.Error)
	Get(username string) (*iam.User, errors.Error)
	Create(username string) errors.Error
	Delete(username string) errors.Error

	PolicyPut(policyDocument, policyName, username string) errors.Error
	PolicyAttach(policyARN, username string) errors.Error

	LoginCheck(username string) errors.Error
	LoginCreate(username, password string) errors.Error
	LoginDelete(username string) errors.Error

	AccessList(username string) (map[string]bool, errors.Error)
	AccessCreate(username string) (string, string, errors.Error)
	AccessDelete(username, accessKey string) errors.Error
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) User {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
