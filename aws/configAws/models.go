package configAws

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-playground/validator/v10"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/httpcli"
	"github.com/nabbar/golib/logger"
)

type configModel struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"printascii,required"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"printascii,required"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"printascii,required"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"printascii,omitempty"`
}

type awsModel struct {
	configModel

	logLevel logger.Level
	awsLevel aws.LogLevel
	retryer  aws.Retryer
}

func (c *awsModel) Validate() errors.Error {
	val := validator.New()
	err := val.Struct(c)

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

	return nil
}

func (c *awsModel) ResetRegionEndpoint() {
}

func (c *awsModel) RegisterRegionEndpoint(region string, endpoint *url.URL) errors.Error {
	return nil
}

func (c *awsModel) RegisterRegionAws(endpoint *url.URL) errors.Error {
	return nil
}

func (c *awsModel) SetRegion(region string) {
	c.Region = region
}

func (c *awsModel) GetRegion() string {
	return c.Region
}

func (c *awsModel) SetEndpoint(endpoint *url.URL) {
}

func (c awsModel) GetEndpoint() *url.URL {
	return nil
}

func (c *awsModel) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
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
		end aws.Endpoint
		adr *url.URL
		err error
		e   errors.Error
	)

	if cfg, e = c.GetConfig(nil); e != nil {
		return e
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if end, err = cfg.EndpointResolver.ResolveEndpoint("s3", c.GetRegion()); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if adr, err = url.Parse(end.URL); err != nil {
		return ErrorEndpointInvalid.ErrorParent(err)
	}

	if _, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return ErrorCredentialsInvalid.ErrorParent(err)
	}

	d := net.Dialer{
		Timeout:   httpcli.TIMEOUT_5_SEC,
		KeepAlive: httpcli.TIMEOUT_5_SEC,
	}

	con, err = d.DialContext(ctx, "tcp", adr.Host)

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
