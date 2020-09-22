package aws

import (
	"context"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/logger"
)

type Config interface {
	Check(ctx context.Context) errors.Error
	Validate() errors.Error

	ResetRegionEndpoint()
	RegisterRegionEndpoint(region string, endpoint *url.URL) errors.Error
	RegisterRegionAws(endpoint *url.URL) errors.Error
	SetRegion(region string)
	GetRegion() string
	SetEndpoint(endpoint *url.URL)
	GetEndpoint() *url.URL

	ResolveEndpoint(service, region string) (aws.Endpoint, error)

	SetLogLevel(lvl logger.Level)
	SetAWSLogLevel(lvl aws.LogLevel)
	SetRetryer(retryer aws.Retryer)

	GetConfig(cli *http.Client) (aws.Config, errors.Error)
	JSON() ([]byte, error)
	Clone() Config

	GetBucketName() string
	SetBucketName(bucket string)
}

type AWS interface {
	Bucket() bucket.Bucket
	Group() group.Group
	Object() object.Object
	Policy() policy.Policy
	Role() role.Role
	User() user.User

	Clone() AWS
	Config() Config
	ForcePathStyle(enabled bool)

	GetBucketName() string
	SetBucketName(bucket string)
}

type client struct {
	p bool
	x context.Context
	c Config
	i *iam.Client
	s *s3.Client
}

func New(ctx context.Context, cfg Config, httpClient *http.Client) (AWS, errors.Error) {
	if cfg == nil {
		return nil, helper.ErrorConfigEmpty.Error(nil)
	}

	var (
		c aws.Config
		i *iam.Client
		s *s3.Client
		e errors.Error
	)

	if c, e = cfg.GetConfig(httpClient); e != nil {
		return nil, e
	}

	i = iam.New(c)
	s = s3.New(c)

	if httpClient != nil {
		i.HTTPClient = httpClient
		s.HTTPClient = httpClient
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &client{
		p: false,
		x: ctx,
		c: cfg,
		i: i,
		s: s,
	}, nil
}

func (c *client) getCliIAM() *iam.Client {
	i := iam.New(c.i.Config)
	i.HTTPClient = c.i.HTTPClient
	return i
}

func (c *client) getCliS3() *s3.Client {
	s := s3.New(c.s.Config)
	s.HTTPClient = c.s.HTTPClient
	s.ForcePathStyle = c.p
	return s
}

func (c *client) Clone() AWS {
	return &client{
		p: c.p,
		x: c.x,
		c: c.c.Clone(),
		i: c.getCliIAM(),
		s: c.getCliS3(),
	}
}
