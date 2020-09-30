package user

import (
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdkitp "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type User interface {
	List() (map[string]string, liberr.Error)
	Get(username string) (*sdkitp.User, liberr.Error)
	Create(username string) liberr.Error
	Delete(username string) liberr.Error

	PolicyPut(policyDocument, policyName, username string) liberr.Error
	PolicyAttach(policyARN, username string) liberr.Error

	LoginCheck(username string) liberr.Error
	LoginCreate(username, password string) liberr.Error
	LoginDelete(username string) liberr.Error

	AccessList(username string) (map[string]bool, liberr.Error)
	AccessCreate(username string) (string, string, liberr.Error)
	AccessDelete(username, accessKey string) liberr.Error
}

func New(ctx context.Context, bucket string, iam *sdkiam.Client, s3 *sdksss.Client) User {
	return &client{
		Helper: libhlp.New(ctx, bucket),
		iam:    iam,
		s3:     s3,
	}
}
