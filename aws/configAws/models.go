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

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	libval "github.com/go-playground/validator/v10"
	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/httpcli"
)

type Model struct {
	Region    string `mapstructure:"region" json:"region" yaml:"region" toml:"region" validate:"printascii,required"`
	AccessKey string `mapstructure:"accesskey" json:"accesskey" yaml:"accesskey" toml:"accesskey" validate:"printascii,required"`
	SecretKey string `mapstructure:"secretkey" json:"secretkey" yaml:"secretkey" toml:"secretkey" validate:"printascii,required"`
	Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" toml:"bucket" validate:"printascii,omitempty"`
}

type awsModel struct {
	Model
	retryer func() sdkaws.Retryer
}

func (c *awsModel) Validate() errors.Error {
	err := ErrorConfigValidator.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.AddParent(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.StructNamespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}

func (c *awsModel) SetCredentials(accessKey, secretKey string) {
	c.AccessKey = accessKey
	c.SecretKey = secretKey
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

func (c *awsModel) ResolveEndpoint(service, region string) (sdkaws.Endpoint, error) {
	return sdkaws.Endpoint{}, ErrorEndpointInvalid.Error(nil)
}

func (c *awsModel) IsHTTPs() bool {
	return true
}

func (c *awsModel) SetRetryer(retryer func() sdkaws.Retryer) {
	c.retryer = retryer
}

func (c awsModel) Check(ctx context.Context) errors.Error {
	var (
		cfg *sdkaws.Config
		con net.Conn
		end sdkaws.Endpoint
		adr *url.URL
		err error
		e   errors.Error
	)

	if cfg, e = c.GetConfig(ctx, nil); e != nil {
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
		Timeout:   httpcli.ClientTimeout5Sec,
		KeepAlive: httpcli.ClientTimeout5Sec,
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
