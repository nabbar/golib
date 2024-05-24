/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package configCustom

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net"
	"net/url"
	"regexp"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	libval "github.com/go-playground/validator/v10"
	libhtc "github.com/nabbar/golib/httpcli"
	libreq "github.com/nabbar/golib/request"
)

type Model struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"required,hostname"`
	Endpoint  string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint" toml:"endpoint" validate:"url"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"omitempty,printascii"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"omitempty,printascii"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"omitempty,bucket-s3"`
}

type ModelStatus struct {
	Config  Model                `json:"config" yaml:"config" toml:"config" mapstructure:"config" validate:"required"`
	Monitor libreq.OptionsHealth `json:"health" yaml:"health" toml:"health" mapstructure:"health" validate:""`
}

type awsModel struct {
	Model

	retryer   func() sdkaws.Retryer
	endpoint  *url.URL
	mapRegion map[string]*url.URL
}

func containsInvalidSequences(input string) bool {
	for i := 1; i < len(input); i++ {
		if (input[i] == '.' && (input[i-1] == '.' || input[i-1] == '-')) ||
			(input[i] == '-' && (input[i-1] == '-' || input[i-1] == '.')) {
			return true
		}
	}
	return false
}

func validateBucketS3(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	re := regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*[A-Za-z0-9]$`)
	if !re.MatchString(value) {
		return false
	}
	if containsInvalidSequences(value) {
		return false
	}
	return true
}

func (c *awsModel) Validate() error {
	err := ErrorConfigValidator.Error(nil)

	validate := libval.New()
	valErr := validate.RegisterValidation("bucket-s3", validateBucketS3)
	if valErr != nil {
		err.Add(valErr)
	}

	if er := validate.Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.StructNamespace(), e.ActualTag()))
		}
	}

	if c.Endpoint != "" && c.endpoint == nil {
		var e error
		if c.endpoint, e = url.Parse(c.Endpoint); e != nil {
			err.Add(e)
		} else if er := c.RegisterRegionAws(c.endpoint); er != nil {
			err.Add(er)
		}
	} else if !err.HasParent() && c.endpoint != nil && c.Endpoint == "" {
		c.Endpoint = c.endpoint.String()
	}

	if !err.HasParent() && c.endpoint != nil && c.Region != "" {
		if e := c.RegisterRegionEndpoint("", c.endpoint); e != nil {
			err.Add(e)
		}
	}

	if !err.HasParent() {
		err = nil
	}

	return err
}

func (c *awsModel) GetAccessKey() string {
	return c.AccessKey
}

func (c *awsModel) GetSecretKey() string {
	return c.SecretKey
}

func (c *awsModel) SetCredentials(accessKey, secretKey string) {
	c.AccessKey = accessKey
	c.SecretKey = secretKey
}

func (c *awsModel) ResetRegionEndpoint() {
	c.mapRegion = make(map[string]*url.URL)
}

func (c *awsModel) RegisterRegionEndpoint(region string, endpoint *url.URL) error {
	if endpoint == nil && c.endpoint != nil {
		endpoint = c.endpoint
	} else if endpoint == nil && c.Endpoint != "" {
		var err error
		if endpoint, err = url.Parse(c.Endpoint); err != nil {
			return ErrorEndpointInvalid.Error(err)
		}
	}

	if endpoint == nil {
		return ErrorEndpointInvalid.Error(nil)
	}

	if region == "" && c.Region != "" {
		region = c.Region
	}

	val := libval.New()

	if err := val.Var(endpoint, "url,required"); err != nil {
		return ErrorEndpointInvalid.Error(err)
	} else if err := val.Var(region, "printascii,required"); err != nil {
		return ErrorRegionInvalid.Error(err)
	}

	if c.mapRegion == nil {
		c.mapRegion = make(map[string]*url.URL)
	}

	c.mapRegion[region] = endpoint

	return nil
}

func (c *awsModel) RegisterRegionAws(endpoint *url.URL) error {
	if endpoint == nil && c.endpoint != nil {
		endpoint = c.endpoint
	} else if endpoint == nil && c.Endpoint != "" {
		var err error
		if endpoint, err = url.Parse(c.Endpoint); err != nil {
			return ErrorEndpointInvalid.Error(err)
		}
	}

	if endpoint == nil {
		return ErrorEndpointInvalid.Error(nil)
	}

	val := libval.New()
	if err := val.Var(endpoint, "url,required"); err != nil {
		return ErrorEndpointInvalid.Error(err)
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

func (c *awsModel) GetEndpoint() *url.URL {
	return c.endpoint
}

func (c *awsModel) ResolveEndpoint(service, region string) (sdkaws.Endpoint, error) {
	return c.ResolveEndpointWithOptions(service, region)
}

func (c *awsModel) ResolveEndpointWithOptions(service, region string, options ...interface{}) (sdkaws.Endpoint, error) {
	if e, ok := c.mapRegion[region]; ok {
		return sdkaws.Endpoint{
			URL:           strings.TrimSuffix(e.String(), "/"),
			SigningRegion: region,
			SigningName:   service,
		}, nil
	}

	if c.Endpoint != "" {
		return sdkaws.Endpoint{
			URL:           strings.TrimSuffix(c.Endpoint, "/"),
			SigningRegion: region,
			SigningName:   service,
		}, nil
	}

	return sdkaws.Endpoint{}, ErrorEndpointInvalid.Error(nil)
}

func (c *awsModel) GetDisableHTTPS() bool {
	return false
}

func (c *awsModel) GetResolvedRegion() string {
	return c.GetRegion()
}

func (c *awsModel) IsHTTPs() bool {
	return strings.HasSuffix(strings.ToLower(c.endpoint.Scheme), "s")
}

func (c *awsModel) SetRetryer(retryer func() sdkaws.Retryer) {
	c.retryer = retryer
}

func (c *awsModel) Check(ctx context.Context) error {
	var (
		cfg *sdkaws.Config
		con net.Conn
		err error
		e   error
	)

	if cfg, e = c.GetConfig(ctx, nil); e != nil {
		return e
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if _, err = cfg.EndpointResolverWithOptions.ResolveEndpoint("s3", c.GetRegion()); err != nil {
		return ErrorEndpointInvalid.Error(err)
	}

	if _, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return ErrorCredentialsInvalid.Error(err)
	}

	d := net.Dialer{
		Timeout:   libhtc.ClientTimeout5Sec,
		KeepAlive: libhtc.ClientTimeout5Sec,
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
		return ErrorEndpointInvalid.Error(err)
	}

	return nil
}
