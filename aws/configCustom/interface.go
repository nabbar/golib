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
	"encoding/json"
	"net/url"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkcrd "github.com/aws/aws-sdk-go-v2/credentials"
	libaws "github.com/nabbar/golib/aws"
	libhtc "github.com/nabbar/golib/httpcli"
)

func GetConfigModel() interface{} {
	return Model{}
}

func NewConfigJsonUnmashal(p []byte) (libaws.Config, error) {
	c := Model{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.Error(err)
	}

	return &awsModel{
		Model:     c,
		retryer:   nil,
		endpoint:  nil,
		mapRegion: nil,
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenRequired,
			Response: sdkaws.ResponseChecksumValidationWhenRequired,
		},
	}, nil
}

func NewConfigStatusJsonUnmashal(p []byte) (libaws.Config, error) {
	c := ModelStatus{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.Error(err)
	}

	return &awsModel{
		Model:     c.Config,
		retryer:   nil,
		mapRegion: nil,
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenRequired,
			Response: sdkaws.ResponseChecksumValidationWhenRequired,
		},
	}, nil
}

func NewConfig(bucket, accessKey, secretKey string, endpoint *url.URL, region string) libaws.Config {
	return &awsModel{
		Model: Model{
			Region:    region,
			Endpoint:  strings.TrimSuffix(endpoint.String(), "/"),
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket,
		},
		endpoint:  endpoint,
		retryer:   nil,
		mapRegion: make(map[string]*url.URL),
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenRequired,
			Response: sdkaws.ResponseChecksumValidationWhenRequired,
		},
	}
}

func (c *awsModel) Clone() libaws.Config {
	m := make(map[string]*url.URL)

	for r, e := range c.mapRegion {
		m[r] = e
	}

	return &awsModel{
		Model: Model{
			Region:    c.Region,
			Endpoint:  c.Endpoint,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
			Bucket:    c.Bucket,
		},
		retryer:   c.retryer,
		endpoint:  c.endpoint,
		mapRegion: m,
		checksum: checksumOptions{
			Request:  c.checksum.Request,
			Response: c.checksum.Response,
		},
	}
}

func (c *awsModel) GetConfig(ctx context.Context, cli libhtc.HttpClient) (*sdkaws.Config, error) {

	cfg := sdkaws.NewConfig()

	if len(c.AccessKey) < 1 || len(c.SecretKey) < 1 {
		cfg.Credentials = sdkaws.AnonymousCredentials{}
	} else {
		cfg.Credentials = sdkcrd.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	}

	cfg.Retryer = c.retryer
	cfg.EndpointResolver = sdkaws.EndpointResolverFunc(c.ResolveEndpoint)                                  // nolint
	cfg.EndpointResolverWithOptions = sdkaws.EndpointResolverWithOptionsFunc(c.ResolveEndpointWithOptions) // nolint
	cfg.Region = c.Region

	if cli != nil {
		cfg.HTTPClient = cli
	}

	if c.checksum.Request != sdkaws.RequestChecksumCalculationWhenSupported {
		cfg.RequestChecksumCalculation = sdkaws.RequestChecksumCalculationWhenRequired
	} else {
		cfg.RequestChecksumCalculation = sdkaws.RequestChecksumCalculationWhenSupported
	}

	if c.checksum.Response != sdkaws.ResponseChecksumValidationWhenSupported {
		cfg.ResponseChecksumValidation = sdkaws.ResponseChecksumValidationWhenRequired
	} else {
		cfg.ResponseChecksumValidation = sdkaws.ResponseChecksumValidationWhenSupported
	}

	return cfg, nil
}

func (c *awsModel) GetBucketName() string {
	return c.Bucket
}

func (c *awsModel) SetBucketName(bucket string) {
	c.Bucket = bucket
}

func (c *awsModel) JSON() ([]byte, error) {
	return json.MarshalIndent(c, "", " ")
}

func (c *awsModel) SetChecksumValidation(req sdkaws.RequestChecksumCalculation, rsp sdkaws.ResponseChecksumValidation) {
	c.checksum = checksumOptions{
		Request:  req,
		Response: rsp,
	}
}

func (c *awsModel) GetChecksumValidation() (req sdkaws.RequestChecksumCalculation, rsp sdkaws.ResponseChecksumValidation) {
	return c.checksum.Request, c.checksum.Response
}
