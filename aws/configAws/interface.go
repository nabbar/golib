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
	"encoding/json"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkcfg "github.com/aws/aws-sdk-go-v2/config"
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
		Model:   c,
		retryer: nil,
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenSupported,
			Response: sdkaws.ResponseChecksumValidationWhenSupported,
		},
	}, nil
}

func NewConfigStatusJsonUnmashal(p []byte) (libaws.Config, error) {
	c := ModelStatus{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.Error(err)
	}

	return &awsModel{
		Model:   c.Config,
		retryer: nil,
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenSupported,
			Response: sdkaws.ResponseChecksumValidationWhenSupported,
		},
	}, nil
}

func NewConfig(bucket, accessKey, secretKey, region string) libaws.Config {
	return &awsModel{
		Model: Model{
			Region:    region,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket,
		},
		retryer: nil,
		checksum: checksumOptions{
			Request:  sdkaws.RequestChecksumCalculationWhenSupported,
			Response: sdkaws.ResponseChecksumValidationWhenSupported,
		},
	}
}

func (c *awsModel) Clone() libaws.Config {
	return &awsModel{
		Model: Model{
			Region:    c.Region,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
			Bucket:    c.Bucket,
		},
		retryer: c.retryer,
		checksum: checksumOptions{
			Request:  c.checksum.Request,
			Response: c.checksum.Response,
		},
	}
}

func (c *awsModel) GetConfig(ctx context.Context, cli libhtc.HttpClient) (*sdkaws.Config, error) {
	var (
		cfg sdkaws.Config
		err error
	)

	if cfg, err = sdkcfg.LoadDefaultConfig(ctx); err != nil {
		return nil, ErrorConfigLoader.Error(err)
	}

	if len(c.AccessKey) < 1 || len(c.SecretKey) < 1 {
		cfg.Credentials = sdkaws.AnonymousCredentials{}
	} else {
		cfg.Credentials = sdkcrd.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	}

	cfg.Retryer = c.retryer
	cfg.Region = c.Region

	if cli != nil {
		cfg.HTTPClient = cli
	}

	if c.checksum.Request != sdkaws.RequestChecksumCalculationWhenRequired {
		cfg.RequestChecksumCalculation = sdkaws.RequestChecksumCalculationWhenSupported
	} else {
		cfg.RequestChecksumCalculation = sdkaws.RequestChecksumCalculationWhenRequired
	}

	if c.checksum.Response != sdkaws.ResponseChecksumValidationWhenRequired {
		cfg.ResponseChecksumValidation = sdkaws.ResponseChecksumValidationWhenSupported
	} else {
		cfg.ResponseChecksumValidation = sdkaws.ResponseChecksumValidationWhenRequired
	}

	return &cfg, nil
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
