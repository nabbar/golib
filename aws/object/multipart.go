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

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	libmpu "github.com/nabbar/golib/aws/multipart"
	liberr "github.com/nabbar/golib/errors"
	libsiz "github.com/nabbar/golib/size"
)

// MultipartList implement the ListMultipartUploads.
// See docs for more infos : https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListMultipartUploads.html
func (cli *client) MultipartList(keyMarker, markerId string) (uploads []sdktyp.MultipartUpload, nextKeyMarker string, nextIdMarker string, count int64, e liberr.Error) {
	in := &sdksss.ListMultipartUploadsInput{
		Bucket:     sdkaws.String(cli.GetBucketName()),
		MaxUploads: 1000,
	}

	if keyMarker != "" && markerId != "" {
		in.KeyMarker = sdkaws.String(keyMarker)
		in.UploadIdMarker = sdkaws.String(markerId)
	}

	out, err := cli.s3.ListMultipartUploads(cli.GetContext(), in)

	if err != nil {
		return nil, "", "", 0, cli.GetError(err)
	} else if out.IsTruncated {
		return out.Uploads, *out.NextKeyMarker, *out.NextUploadIdMarker, int64(out.MaxUploads), nil
	} else {
		return out.Uploads, "", "", int64(out.MaxUploads), nil
	}
}

func (cli *client) MultipartNew(partSize libsiz.Size, object string) libmpu.MultiPart {
	m := libmpu.New(partSize, object, cli.GetBucketName())
	m.RegisterContext(cli.GetContext)
	m.RegisterClientS3(func() *sdksss.Client {
		return cli.s3
	})

	return m
}

func (cli *client) MultipartPut(object string, body io.Reader) liberr.Error {
	return cli.MultipartPutCustom(libmpu.DefaultPartSize, object, body)
}

func (cli *client) MultipartPutCustom(partSize libsiz.Size, object string, body io.Reader) liberr.Error {
	var (
		e error
		m = cli.MultipartNew(partSize, object)
	)

	defer func() {
		if m != nil {
			_ = m.Close()
		}
	}()

	if e = m.StartMPU(); e != nil {
		return cli.GetError(e)
	} else if _, e = io.Copy(m, body); e != nil {
		return cli.GetError(e)
	} else if e = m.StopMPU(false); e != nil {
		return cli.GetError(e)
	} else {
		m = nil
	}

	return nil
}

func (cli *client) MultipartCancel(uploadId, key string) liberr.Error {
	res, err := cli.s3.AbortMultipartUpload(cli.GetContext(), &sdksss.AbortMultipartUploadInput{
		Bucket:   sdkaws.String(cli.GetBucketName()),
		UploadId: sdkaws.String(uploadId),
		Key:      sdkaws.String(key),
	})

	if err != nil {
		return cli.GetError(err)
	} else if res == nil {
		return libhlp.ErrorResponse.Error(nil)
	} else {
		return nil
	}
}
