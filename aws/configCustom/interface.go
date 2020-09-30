package configCustom

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkcrd "github.com/aws/aws-sdk-go-v2/credentials"
	libaws "github.com/nabbar/golib/aws"
	"github.com/nabbar/golib/errors"
)

func GetConfigModel() interface{} {
	return Model{}
}

func NewConfigJsonUnmashal(p []byte) (libaws.Config, errors.Error) {
	c := Model{}
	if err := json.Unmarshal(p, &c); err != nil {
		return nil, ErrorConfigJsonUnmarshall.ErrorParent(err)
	}

	return &awsModel{
		Model:     c,
		retryer:   nil,
		mapRegion: nil,
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
	}
}

func (c *awsModel) GetConfig(cli *http.Client) (*sdkaws.Config, errors.Error) {

	cfg := sdkaws.NewConfig()

	cfg.Credentials = sdkcrd.NewStaticCredentialsProvider(c.AccessKey, c.SecretKey, "")
	cfg.Retryer = c.retryer
	cfg.EndpointResolver = sdkaws.EndpointResolverFunc(c.ResolveEndpoint)
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
