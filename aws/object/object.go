package object

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List(continuationToken string) ([]s3.Object, string, int64, errors.Error) {
	in := s3.ListObjectsV2Input{
		Bucket: cli.GetBucketAws(),
	}

	if continuationToken != "" {
		in.ContinuationToken = aws.String(continuationToken)
	}

	req := cli.s3.ListObjectsV2Request(&in)

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, "", 0, cli.GetError(err)
	} else if *out.IsTruncated {
		return out.Contents, *out.NextContinuationToken, *out.KeyCount, nil
	} else {
		return out.Contents, "", *out.KeyCount, nil
	}
}

func (cli *client) Get(object string) (io.ReadCloser, []io.Closer, errors.Error) {
	req := cli.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    aws.String(object),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(nil, nil)

	if err != nil {
		cli.Close(req.HTTPRequest, req.HTTPResponse)
		return nil, nil, cli.GetError(err)
	} else if out.Body == nil {
		cli.Close(req.HTTPRequest, req.HTTPResponse)
		return nil, nil, helper.ErrorResponse.Error(nil)
	} else {
		return out.Body, cli.GetCloser(req.HTTPRequest, req.HTTPResponse), nil
	}
}

func (cli *client) Head(object string) (head map[string]interface{}, meta map[string]string, err errors.Error) {
	req := cli.s3.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    aws.String(object),
	})

	out, e := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if e != nil {
		return nil, nil, cli.GetError(e)
	} else if out.Metadata == nil {
		return nil, nil, helper.ErrorResponse.Error(nil)
	} else {
		res := make(map[string]interface{})
		if out.ContentType != nil {
			res["ContentType"] = *out.ContentType
		}
		if out.ContentDisposition != nil {
			res["ContentDisposition"] = *out.ContentDisposition
		}
		if out.ContentEncoding != nil {
			res["ContentEncoding"] = *out.ContentEncoding
		}
		if out.ContentLanguage != nil {
			res["ContentLanguage"] = *out.ContentLanguage
		}
		if out.ContentLength != nil {
			res["ContentLength"] = *out.ContentLength
		}

		return res, out.Metadata, nil
	}
}

func (cli *client) Size(object string) (size int64, err errors.Error) {
	var (
		h map[string]interface{}
		i interface{}
		j int64
		o bool
	)

	if h, _, err = cli.Head(object); err != nil {
		return
	} else if i, o = h["ContentLength"]; !o {
		return 0, nil
	} else if j, o = i.(int64); !o {
		return 0, nil
	} else {
		return j, nil
	}
}

func (cli *client) Put(object string, body *bytes.Reader) errors.Error {
	req := cli.s3.PutObjectRequest(&s3.PutObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    aws.String(object),
		Body:   body,
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	} else if out.ETag == nil {
		return helper.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(object string) errors.Error {
	if _, _, err := cli.Head(object); err != nil {
		return err
	}

	req := cli.s3.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    aws.String(object),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}
