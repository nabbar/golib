package configCustom

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	aws2 "github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
)

func GetConfigModel() interface{} {
	return Model{}
}

func NewConfigJsonUnmashal(p []byte) (aws2.Config, errors.Error) {
	c := Model{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.ErrorParent(err)
	}

	return &awsModel{
		Model:     c,
		logLevel:  0,
		awsLevel:  0,
		retryer:   nil,
		mapRegion: nil,
	}, nil
}

func NewConfig(bucket, accessKey, secretKey string, endpoint *url.URL, region string) aws2.Config {
	return &awsModel{
		Model: Model{
			Region:    region,
			Endpoint:  strings.TrimSuffix(endpoint.String(), "/"),
			AccessKey: accessKey,
			SecretKey: secretKey,
			Bucket:    bucket,
		},
		endpoint:  endpoint,
		logLevel:  0,
		awsLevel:  0,
		retryer:   nil,
		mapRegion: make(map[string]*url.URL),
	}
}

func (c *awsModel) Clone() aws2.Config {
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
		logLevel:  c.logLevel,
		awsLevel:  c.awsLevel,
		retryer:   c.retryer,
		endpoint:  c.endpoint,
		mapRegion: m,
	}
}

func (c *awsModel) GetConfig(cli *http.Client) (aws.Config, errors.Error) {
	cfg := defaults.Config()
	cfg.Credentials = aws.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	cfg.Logger = &awsLogger{c.logLevel}
	cfg.LogLevel = c.awsLevel
	cfg.Retryer = c.retryer
	cfg.EnableEndpointDiscovery = false
	cfg.DisableEndpointHostPrefix = true
	cfg.EndpointResolver = aws.EndpointResolverFunc(c.ResolveEndpoint)
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
