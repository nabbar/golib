package configAws

import (
	"encoding/json"
	"net/http"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkcfg "github.com/aws/aws-sdk-go-v2/config"
	sdkcrd "github.com/aws/aws-sdk-go-v2/credentials"
	libaws "github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
)

func GetConfigModel() interface{} {
	return configModel{}
}

func NewConfigJsonUnmashal(p []byte) (libaws.Config, errors.Error) {
	c := configModel{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.ErrorParent(err)
	}

	return &awsModel{
		configModel: c,
		retryer:     nil,
	}, nil
}

func NewConfig(bucket, accessKey, secretKey, region string) libaws.Config {
	return &awsModel{
		configModel: configModel{
			Region:    region,
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket,
		},
		retryer: nil,
	}
}

func (c *awsModel) Clone() libaws.Config {
	return &awsModel{
		configModel: configModel{
			Region:    c.Region,
			AccessKey: c.AccessKey,
			SecretKey: c.SecretKey,
			Bucket:    c.Bucket,
		},
		retryer: c.retryer,
	}
}

func (c *awsModel) GetConfig(cli *http.Client) (*sdkaws.Config, errors.Error) {
	var (
		cfg sdkaws.Config
		err error
	)

	if cfg, err = sdkcfg.LoadDefaultConfig(); err != nil {
		return nil, ErrorConfigLoader.ErrorParent(err)
	}

	if c.AccessKey != "" && c.SecretKey != "" {
		cfg.Credentials = sdkcrd.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	}

	cfg.Retryer = c.retryer
	cfg.Region = c.Region

	if cli != nil {
		cfg.HTTPClient = cli
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
