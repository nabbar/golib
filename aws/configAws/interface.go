package configAws

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	aws2 "github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
)

func GetConfigModel() interface{} {
	return configModel{}
}

func NewConfigJsonUnmashal(p []byte) (aws2.Config, errors.Error) {
	c := configModel{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.ErrorParent(err)
	}

	return &awsModel{
		configModel: c,
		logLevel:    0,
		awsLevel:    0,
		retryer:     nil,
	}, nil
}

func NewConfig(bucket, accessKey, secretKey, region string) aws2.Config {
	return &awsModel{
		configModel: configModel{
			Region:    region,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket,
		},
		logLevel: 0,
		awsLevel: 0,
		retryer:  nil,
	}
}

func (c *awsModel) Clone() aws2.Config {
	return &awsModel{
		configModel: configModel{
			Region:    c.Region,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
			Bucket:    c.Bucket,
		},
		logLevel: c.logLevel,
		awsLevel: c.awsLevel,
		retryer:  c.retryer,
	}
}

func (c *awsModel) GetConfig(cli *http.Client) (aws.Config, errors.Error) {
	var (
		cfg aws.Config
		err error
	)

	if c.AccessKey != "" && c.SecretKey != "" {
		cfg = defaults.Config()
		cfg.Credentials = aws.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	} else if cfg, err = external.LoadDefaultAWSConfig(); err != nil {
		return cfg, ErrorConfigLoader.ErrorParent(err)
	}

	cfg.Logger = &awsLogger{c.logLevel}
	cfg.LogLevel = c.awsLevel
	cfg.Retryer = c.retryer
	cfg.EnableEndpointDiscovery = true
	cfg.Region = c.Region

	if cli != nil {
		cfg.HTTPClient = cli
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
