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

package configAws

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"regexp"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	libval "github.com/go-playground/validator/v10"
	libhtc "github.com/nabbar/golib/httpcli"
	libreq "github.com/nabbar/golib/request"
)

type Model struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"printascii,required"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"printascii,required"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"printascii,required"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"printascii,omitempty,bucket-s3"`
}

type ModelStatus struct {
	Config  Model                `json:"config" yaml:"config" toml:"config" mapstructure:"config" validate:"required"`
	Monitor libreq.OptionsHealth `json:"health" yaml:"health" toml:"health" mapstructure:"health" validate:""`
}

type checksumOptions struct {
	Request  sdkaws.RequestChecksumCalculation
	Response sdkaws.ResponseChecksumValidation
}

type awsModel struct {
	Model
	retryer  func() sdkaws.Retryer
	checksum checksumOptions
}

func validateBucketS3(fl libval.FieldLevel) bool {
	value := fl.Field().String()
	re := regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9]|\.[A-Za-z0-9]|-[A-Za-z0-9]){0,46}[A-Za-z0-9]$`)
	return re.MatchString(value)
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
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.StructNamespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
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
}

func (c *awsModel) RegisterRegionEndpoint(region string, endpoint *url.URL) error {
	return nil
}

func (c *awsModel) RegisterRegionAws(endpoint *url.URL) error {
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

func (c *awsModel) GetEndpoint() *url.URL {
	return nil
}

func (c *awsModel) ResolveEndpoint(service, region string) (sdkaws.Endpoint, error) { // nolint
	return sdkaws.Endpoint{}, ErrorEndpointInvalid.Error(nil) // nolint
}

func (c *awsModel) ResolveEndpointWithOptions(service, region string, options ...interface{}) (sdkaws.Endpoint, error) { // nolint
	return sdkaws.Endpoint{}, ErrorEndpointInvalid.Error(nil) // nolint
}

func (c *awsModel) GetDisableHTTPS() bool {
	return false
}

func (c *awsModel) GetResolvedRegion() string {
	return c.GetRegion()
}

func (c *awsModel) IsHTTPs() bool {
	return true
}

func (c *awsModel) SetRetryer(retryer func() sdkaws.Retryer) {
	c.retryer = retryer
}

func (c *awsModel) Check(ctx context.Context) error {
	var (
		cfg *sdkaws.Config
		con net.Conn
		end sdkaws.Endpoint // nolint
		adr *url.URL
		err error
		e   error
	)

	if cfg, e = c.GetConfig(ctx, nil); e != nil {
		return e
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// nolint
	if end, err = cfg.EndpointResolverWithOptions.ResolveEndpoint("s3", c.GetRegion()); err != nil { // nolint
		return ErrorEndpointInvalid.Error(err)
	}

	if adr, err = url.Parse(end.URL); err != nil {
		return ErrorEndpointInvalid.Error(err)
	}

	if _, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return ErrorCredentialsInvalid.Error(err)
	}

	d := net.Dialer{
		Timeout:   libhtc.ClientTimeout5Sec,
		KeepAlive: libhtc.ClientTimeout5Sec,
	}

	con, err = d.DialContext(ctx, "tcp", adr.Host)

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
