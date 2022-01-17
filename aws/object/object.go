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

package object

import (
	"io"
	"mime"
	"path/filepath"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) List(continuationToken string) ([]sdktps.Object, string, int64, liberr.Error) {
	in := sdksss.ListObjectsV2Input{
		Bucket: cli.GetBucketAws(),
	}

	if continuationToken != "" {
		in.ContinuationToken = sdkaws.String(continuationToken)
	}

	out, err := cli.s3.ListObjectsV2(cli.GetContext(), &in)

	if err != nil {
		return nil, "", 0, cli.GetError(err)
	} else if out.IsTruncated {
		return out.Contents, *out.NextContinuationToken, int64(out.KeyCount), nil
	} else {
		return out.Contents, "", int64(out.KeyCount), nil
	}
}

func (cli *client) Get(object string) (*sdksss.GetObjectOutput, liberr.Error) {
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
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}
}

func (cli *client) Head(object string) (*sdksss.HeadObjectOutput, liberr.Error) {
	out, e := cli.s3.HeadObject(cli.GetContext(), &sdksss.HeadObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	})

	if e != nil {
		return nil, cli.GetError(e)
	} else if out.ETag == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out, nil
	}
}

func (cli *client) Size(object string) (size int64, err liberr.Error) {
	var (
		h *sdksss.HeadObjectOutput
	)

	if h, err = cli.Head(object); err != nil {
		return
	} else {
		return h.ContentLength, nil
	}
}

func (cli *client) Put(object string, body io.Reader) liberr.Error {
	var tpe *string

	if t := mime.TypeByExtension(filepath.Ext(object)); t == "" {
		tpe = sdkaws.String("application/octet-stream")
	} else {
		tpe = sdkaws.String(t)
	}

	out, err := cli.s3.PutObject(cli.GetContext(), &sdksss.PutObjectInput{
		Bucket:      cli.GetBucketAws(),
		Key:         sdkaws.String(object),
		Body:        body,
		ContentType: tpe,
	})

	if err != nil {
		return cli.GetError(err)
	} else if out.ETag == nil {
		return libhlp.ErrorResponse.Error(nil)
	}

	return nil
}

func (cli *client) Delete(object string) liberr.Error {
	if _, err := cli.Head(object); err != nil {
		return err
	}

	_, err := cli.s3.DeleteObject(cli.GetContext(), &sdksss.DeleteObjectInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	})

	return cli.GetError(err)
}

func (cli *client) UpdateMetadata(meta *sdksss.CopyObjectInput) liberr.Error {
	_, err := cli.s3.CopyObject(cli.GetContext(), meta)

	return cli.GetError(err)
}

func (cli *client) SetWebsite(object, redirect string) liberr.Error {
	var err error

	_, err = cli.s3.PutObjectAcl(cli.GetContext(), &sdksss.PutObjectAclInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
		ACL:    sdktps.ObjectCannedACLPublicRead,
	})

	if err != nil {
		return cli.GetError(err)
	}

	if redirect == "" {
		return nil
	}

	meta := &sdksss.CopyObjectInput{
		Bucket:                  cli.GetBucketAws(),
		CopySource:              sdkaws.String(cli.GetBucketName() + "/" + object),
		Key:                     sdkaws.String(object),
		WebsiteRedirectLocation: sdkaws.String(redirect),
	}

	return cli.UpdateMetadata(meta)
}
