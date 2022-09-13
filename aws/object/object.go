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

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

func (cli *client) List(continuationToken string) ([]sdktps.Object, string, int64, liberr.Error) {
	return cli.ListPrefix(continuationToken, "")
}

func (cli *client) ListPrefix(continuationToken string, prefix string) ([]sdktps.Object, string, int64, liberr.Error) {
	in := sdksss.ListObjectsV2Input{
		Bucket: cli.GetBucketAws(),
	}

	if continuationToken != "" {
		in.ContinuationToken = sdkaws.String(continuationToken)
	}

	if prefix != "" {
		in.Prefix = sdkaws.String(prefix)
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

func (cli *client) Walk(f WalkFunc) liberr.Error {
	return cli.WalkPrefix("", f)
}

func (cli *client) WalkPrefix(prefix string, f WalkFunc) liberr.Error {
	in := sdksss.ListObjectsV2Input{
		Bucket: cli.GetBucketAws(),
	}

	if prefix != "" {
		in.Prefix = sdkaws.String(prefix)
	}

	var (
		e liberr.Error
		t = sdkaws.String("")
	)

	for {
		if len(*t) > 0 {
			in.ContinuationToken = t
		}

		out, err := cli.s3.ListObjectsV2(cli.GetContext(), &in)

		if err != nil {
			return cli.GetError(err)
		}

		for _, o := range out.Contents {
			if o.Key == nil || len(*o.Key) < 3 {
				continue
			}

			if f != nil {
				e = f(e, o)
			}
		}

		if out.IsTruncated {
			t = out.NextContinuationToken
		} else {
			return e
		}
	}
}

func (cli *client) Get(object string) (*sdksss.GetObjectOutput, liberr.Error) {
	return cli.VersionGet(object, "")
}

func (cli *client) Head(object string) (*sdksss.HeadObjectOutput, liberr.Error) {
	return cli.VersionHead(object, "")
}

func (cli *client) Size(object string) (size int64, err liberr.Error) {
	return cli.VersionSize(object, "")
}

func (cli *client) Delete(check bool, object string) liberr.Error {
	return cli.VersionDelete(check, object, "")
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

func (cli *client) DeleteAll(objects *sdktps.Delete) ([]sdktps.DeletedObject, liberr.Error) {
	in := sdksss.DeleteObjectsInput{
		Bucket: cli.GetBucketAws(),
		Delete: objects,
	}

	out, err := cli.s3.DeleteObjects(cli.GetContext(), &in)

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else {
		return out.Deleted, nil
	}
}

func (cli *client) GetAttributes(object, version string) (*sdksss.GetObjectAttributesOutput, liberr.Error) {
	in := sdksss.GetObjectAttributesInput{
		Bucket: cli.GetBucketAws(),
		Key:    sdkaws.String(object),
	}

	if version != "" {
		in.VersionId = sdkaws.String(version)
	}

	out, err := cli.s3.GetObjectAttributes(cli.GetContext(), &in)

	if err != nil {
		return nil, cli.GetError(err)
	}

	return out, nil
}
