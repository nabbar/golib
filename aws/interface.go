package aws

import (
	"context"
	"net/http"
	"net/url"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
	"github.com/nabbar/golib/errors"
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

	IsHTTPs() bool
	ResolveEndpoint(service, region string) (sdkaws.Endpoint, error)
	SetRetryer(retryer sdkaws.Retryer)

	GetConfig(cli *http.Client) (*sdkaws.Config, errors.Error)
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

	Clone() (AWS, errors.Error)
	Config() Config
	ForcePathStyle(enabled bool) errors.Error

	GetBucketName() string
	SetBucketName(bucket string)
}

type client struct {
	p bool
	x context.Context
	c Config
	i *sdkiam.Client
	s *sdksss.Client
	h *http.Client
}

func New(ctx context.Context, cfg Config, httpClient *http.Client) (AWS, errors.Error) {
	if cfg == nil {
		return nil, helper.ErrorConfigEmpty.Error(nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cli := &client{
		p: false,
		x: ctx,
		c: cfg,
		i: nil,
		s: nil,
		h: httpClient,
	}

	if i, e := cli.newClientIAM(httpClient); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli.newClientS3(httpClient); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}

func (cli *client) newClientIAM(httpClient *http.Client) (*sdkiam.Client, errors.Error) {
	var (
		c *sdkaws.Config
		i *sdkiam.Client
		e errors.Error
	)

	if httpClient == nil {
		httpClient = cli.h
	}

	if c, e = cli.c.GetConfig(httpClient); e != nil {
		return nil, e
	}

	i = sdkiam.New(sdkiam.Options{
		APIOptions:  c.APIOptions,
		Credentials: c.Credentials,
		EndpointOptions: sdkiam.ResolverOptions{
			DisableHTTPS: cli.c.IsHTTPs(),
		},
		EndpointResolver: sdkiam.WithEndpointResolver(c.EndpointResolver, nil),
		HTTPSignerV4:     sdksv4.NewSigner(),
		Region:           c.Region,
		Retryer:          c.Retryer,
		HTTPClient:       httpClient,
	})

	return i, nil
}

func (cli *client) newClientS3(httpClient *http.Client) (*sdksss.Client, errors.Error) {
	var (
		c *sdkaws.Config
		s *sdksss.Client
		e errors.Error
	)

	if httpClient == nil {
		httpClient = cli.h
	}

	if c, e = cli.c.GetConfig(httpClient); e != nil {
		return nil, e
	}

	s = sdksss.New(sdksss.Options{
		APIOptions:  c.APIOptions,
		Credentials: c.Credentials,
		EndpointOptions: sdksss.ResolverOptions{
			DisableHTTPS: cli.c.IsHTTPs(),
		},
		EndpointResolver: sdksss.WithEndpointResolver(c.EndpointResolver, nil),
		HTTPSignerV4:     sdksv4.NewSigner(),
		Region:           c.Region,
		Retryer:          c.Retryer,
		HTTPClient:       httpClient,
		UsePathStyle:     cli.p,
	})

	return s, nil
}

func (c *client) Clone() (AWS, errors.Error) {
	cli := &client{
		p: false,
		x: c.x,
		c: c.c,
		i: nil,
		s: nil,
		h: c.h,
	}

	if i, e := cli.newClientIAM(c.h); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli.newClientS3(c.h); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}
