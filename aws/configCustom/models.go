package configCustom

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-playground/validator/v10"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/httpcli"
	"github.com/nabbar/golib/logger"
)

type configModel struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"printascii,required"`
	Endpoint  string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint" toml:"endpoint" validate:"url,required"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"printascii,required"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"printascii,required"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"printascii,omitempty"`
}

type awsModel struct {
	configModel

	logLevel  logger.Level
	awsLevel  aws.LogLevel
	retryer   aws.Retryer
	endpoint  *url.URL
	mapRegion map[string]*url.URL
}

func (c *awsModel) Validate() errors.Error {
	val := validator.New()
	err := val.Struct(c)

	if err != nil {
		if e, ok := err.(*validator.InvalidValidationError); ok {
			return ErrorConfigValidator.ErrorParent(e)
		}

		out := ErrorConfigValidator.Error(nil)

		for _, e := range err.(validator.ValidationErrors) {
			//nolint goerr113
			out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
		}

		if out.HasParent() {
			return out
		}
	}

	if c.Endpoint != "" && c.endpoint == nil {
		if c.endpoint, err = url.Parse(c.Endpoint); err != nil {
			return ErrorEndpointInvalid.ErrorParent(err)
		}

		if e := c.RegisterRegionAws(c.endpoint); e != nil {
			return e
		}
	} else if c.endpoint != nil && c.Endpoint == "" {
		c.Endpoint = c.endpoint.String()
	}

	if c.endpoint != nil && c.Region != "" {
		if e := c.RegisterRegionEndpoint("", c.endpoint); e != nil {
			return e
		}
	}

	return nil
}

func (c *awsModel) ResetRegionEndpoint() {
	c.mapRegion = make(map[string]*url.URL)
}

func (c *awsModel) RegisterRegionEndpoint(region string, endpoint *url.URL) errors.Error {
	if endpoint == nil && c.endpoint != nil {
		endpoint = c.endpoint
	} else if endpoint == nil && c.Endpoint != "" {
		var err error
		if endpoint, err = url.Parse(c.Endpoint); err != nil {
			return ErrorEndpointInvalid.ErrorParent(err)
		}
	}

	if endpoint == nil {
		return ErrorEndpointInvalid.Error(nil)
	}

	if region == "" && c.Region != "" {
		region = c.Region
	}

	val := validator.New()

	if err := val.Var(endpoint, "url,required"); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	} else if err := val.Var(region, "printascii,required"); err != nil {
		return ErrorRegionInvalid.ErrorParent(err)
	}

	if c.mapRegion == nil {
		c.mapRegion = make(map[string]*url.URL)
	}

	c.mapRegion[region] = endpoint

	return nil
}

func (c *awsModel) RegisterRegionAws(endpoint *url.URL) errors.Error {
	if endpoint == nil && c.endpoint != nil {
		endpoint = c.endpoint
	} else if endpoint == nil && c.Endpoint != "" {
		var err error
		if endpoint, err = url.Parse(c.Endpoint); err != nil {
			return ErrorEndpointInvalid.ErrorParent(err)
		}
	}

	if endpoint == nil {
		return ErrorEndpointInvalid.Error(nil)
	}

	val := validator.New()
	if err := val.Var(endpoint, "url,required"); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if c.Region == "" {
		c.SetRegion("us-east-1")
	}

	if c.mapRegion == nil {
		c.mapRegion = make(map[string]*url.URL)
	}

	for _, r := range []string{
		"af-south-1",
		"ap-east-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ca-central-1",
		"cn-north-1",
		"cn-northwest-1",
		"eu-central-1",
		"eu-north-1",
		"eu-south-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"me-south-1",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-gov-east-1",
		"us-gov-west-1",
		"us-west-1",
		"us-west-2",
	} {
		c.mapRegion[r] = endpoint
	}

	return nil
}

func (c *awsModel) SetRegion(region string) {
	c.Region = region
}

func (c *awsModel) GetRegion() string {
	return c.Region
}

func (c *awsModel) SetEndpoint(endpoint *url.URL) {
	c.endpoint = endpoint
	c.Endpoint = strings.TrimSuffix(c.endpoint.String(), "/")
}

func (c awsModel) GetEndpoint() *url.URL {
	return c.endpoint
}

func (c *awsModel) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	if e, ok := c.mapRegion[region]; ok {
		return aws.Endpoint{
			URL: strings.TrimSuffix(e.String(), "/"),
		}, nil
	}

	if c.Endpoint != "" {
		return aws.Endpoint{
			URL: strings.TrimSuffix(c.Endpoint, "/"),
		}, nil
	}

	logger.DebugLevel.Logf("Called ResolveEndpoint for service '%s' / region '%s' with nil endpoint", service, region)
	return aws.Endpoint{}, ErrorEndpointInvalid.Error(nil)
}

func (c *awsModel) SetLogLevel(lvl logger.Level) {
	c.logLevel = lvl
}

func (c *awsModel) SetAWSLogLevel(lvl aws.LogLevel) {
	c.awsLevel = lvl
}

func (c *awsModel) SetRetryer(retryer aws.Retryer) {
	c.retryer = retryer
}

func (c awsModel) Check(ctx context.Context) errors.Error {
	var (
		cfg aws.Config
		con net.Conn
		err error
		e   errors.Error
	)

	if cfg, e = c.GetConfig(nil); e != nil {
		return e
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if _, err = cfg.EndpointResolver.ResolveEndpoint("s3", c.GetRegion()); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if _, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return ErrorCredentialsInvalid.ErrorParent(err)
	}

	d := net.Dialer{
		Timeout:   httpcli.TIMEOUT_5_SEC,
		KeepAlive: httpcli.TIMEOUT_5_SEC,
	}

	if c.endpoint.Port() == "" && c.endpoint.Scheme == "http" {
		con, err = d.DialContext(ctx, "tcp", c.endpoint.Hostname()+":80")
	} else if c.endpoint.Port() == "" && c.endpoint.Scheme == "https" {
		con, err = d.DialContext(ctx, "tcp", c.endpoint.Hostname()+":443")
	} else {
		con, err = d.DialContext(ctx, "tcp", c.endpoint.Host)
	}

	defer func() {
		if con != nil {
			_ = con.Close()
		}
	}()

	if err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	return nil
}