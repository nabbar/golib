package object

import (
	"bytes"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List(continuationToken string) ([]*sdktps.Object, string, int64, errors.Error) {
	in := sdksss.ListObjectsV2Input{
		Bucket: cli.GetBucketAws(),
	}

	if continuationToken != "" {
		in.ContinuationToken = sdkaws.String(continuationToken)
	}

	out, err := cli.s3.ListObjectsV2(cli.GetContext(), &in)

	if err != nil {
		return nil, "", 0, cli.GetError(err)
	} else if *out.IsTruncated {
		return out.Contents, *out.NextContinuationToken, int64(*out.KeyCount), nil
	} else {
		return out.Contents, "", int64(*out.KeyCount), nil
	}
}

func (cli *client) Get(object string) (*sdksss.GetObjectOutput, errors.Error) {
	out, err := cli.s3.GetObject(cli.GetContext(), &sdksss.GetObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	})

	if err != nil {
		defer func() {
			if out != nil && out.Body != nil {
				_ = out.Body.Close()
			}
		}()
		return nil, cli.GetError(err)
	} else if out.Body == nil {
		return nil, helper.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}
}

func (cli *client) Head(object string) (*sdksss.HeadObjectOutput, errors.Error) {
	out, e := cli.s3.HeadObject(cli.GetContext(), &sdksss.HeadObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	})

	if e != nil {
		return nil, cli.GetError(e)
	} else if out.ETag == nil {
		return nil, helper.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}
}

func (cli *client) Size(object string) (size int64, err errors.Error) {
	var (
		h *sdksss.HeadObjectOutput
	)

	if h, err = cli.Head(object); err != nil {
		return
	} else {
		return *h.ContentLength, nil
	}
}

func (cli *client) Put(object string, body *bytes.Reader) errors.Error {
	out, err := cli.s3.PutObject(cli.GetContext(), &sdksss.PutObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		Body:   body,
	})

	if err != nil {
		return cli.GetError(err)
	} else if out.ETag == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(object string) errors.Error {
	if _, err := cli.Head(object); err != nil {
		return err
	}

	_, err := cli.s3.DeleteObject(cli.GetContext(), &sdksss.DeleteObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	})

	return cli.GetError(err)
}
