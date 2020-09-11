package group

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

type Group interface {
	UserList(username string) ([]string, errors.Error)
	UserCheck(username, groupName string) (errors.Error, bool)
	UserAdd(username, groupName string) errors.Error
	UserRemove(username, groupName string) errors.Error

	List() (map[string]string, errors.Error)
	Add(groupName string) errors.Error
	Remove(groupName string) errors.Error

	PolicyList(groupName string) (map[string]string, errors.Error)
	PolicyAttach(groupName, polArn string) errors.Error
	PolicyDetach(groupName, polArn string) errors.Error
}

func New(ctx context.Context, bucket string, iam *iam.Client, s3 *s3.Client) Group {
	return &client{
		Helper: helper.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
