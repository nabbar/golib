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
	//nolint #nosec
	/* #nosec */
	"crypto/md5"
	"encoding/base64"
	"io"
	"mime"
	"path/filepath"

	libsiz "github.com/nabbar/golib/size"

	//nolint #gci
	"os"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
	libiou "github.com/nabbar/golib/ioutils"
)

const DefaultPartSize = 5 * libsiz.SizeMega

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

func (cli *client) MultipartPut(object string, body io.Reader) liberr.Error {
	return cli.MultipartPutCustom(DefaultPartSize, object, body)
}

func (cli *client) MultipartPutCustom(partSize libsiz.Size, object string, body io.Reader) liberr.Error {
	var (
		tmp libiou.FileProgress
		rio libhlp.ReaderPartSize
		upl *sdksss.CreateMultipartUploadOutput
		err error
		tpe *string
	)

	defer func() {
		if tmp != nil {
			_ = tmp.Close()
		}
	}()

	if t := mime.TypeByExtension(filepath.Ext(object)); t == "" {
		tpe = sdkaws.String("application/octet-stream")
	} else {
		tpe = sdkaws.String(t)
	}

	upl, err = cli.s3.CreateMultipartUpload(cli.GetContext(), &sdksss.CreateMultipartUploadInput{
		Key:         sdkaws.String(object),
		Bucket:      sdkaws.String(cli.GetBucketName()),
		ContentType: tpe,
	})

	if err != nil {
		return cli.GetError(err)
	} else if upl == nil {
		return libhlp.ErrorResponse.Error(nil)
	}

	rio = libhlp.NewReaderPartSize(body, partSize)

	for !rio.IeOEF() {
		var (
			inf os.FileInfo
			prt *sdksss.UploadPartOutput
		)

		tmp, err = libiou.NewFileProgressTemp()
		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		}

		_, err = io.Copy(tmp, rio)
		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		}

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		}

		inf, err = tmp.FileStat()
		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		}

		/* #nosec */
		h := md5.New()
		if _, e := tmp.WriteTo(h); e != nil {
			return cli._MultipartCancel(e, upl.UploadId, object)
		}

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		}

		prt, err = cli.s3.UploadPart(cli.GetContext(), &sdksss.UploadPartInput{
			Bucket:        sdkaws.String(cli.GetBucketName()),
			Body:          tmp,
			PartNumber:    rio.CurrPart(),
			UploadId:      upl.UploadId,
			Key:           sdkaws.String(object),
			ContentLength: inf.Size(),
			RequestPayer:  sdktyp.RequestPayerRequester,
			ContentMD5:    sdkaws.String(base64.StdEncoding.EncodeToString(h.Sum(nil))),
		})

		_ = tmp.Close()
		tmp = nil

		if err != nil {
			return cli._MultipartCancel(err, upl.UploadId, object)
		} else if prt == nil || prt.ETag == nil || len(*prt.ETag) == 0 {
			return cli._MultipartCancel(libhlp.ErrorResponse.Error(nil), upl.UploadId, object)
		}

		rio.NextPart(prt.ETag)
	}

	var prt *sdksss.CompleteMultipartUploadOutput
	prt, err = cli.s3.CompleteMultipartUpload(cli.GetContext(), &sdksss.CompleteMultipartUploadInput{
		Bucket:          sdkaws.String(cli.GetBucketName()),
		Key:             sdkaws.String(object),
		UploadId:        upl.UploadId,
		MultipartUpload: rio.CompPart(),
		RequestPayer:    sdktyp.RequestPayerRequester,
	})

	if err != nil {
		return cli._MultipartCancel(err, upl.UploadId, object)
	} else if prt == nil || prt.ETag == nil || len(*prt.ETag) == 0 {
		return cli._MultipartCancel(libhlp.ErrorResponse.Error(nil), upl.UploadId, object)
	}

	return nil
}

func (cli *client) _MultipartCancel(err error, updIp *string, object string) liberr.Error {
	cnl, e := cli.s3.AbortMultipartUpload(cli.GetContext(), &sdksss.AbortMultipartUploadInput{
		Bucket:   sdkaws.String(cli.GetBucketName()),
		UploadId: updIp,
		Key:      sdkaws.String(object),
	})

	if e != nil {
		return cli.GetError(e, err)
	} else if cnl == nil {
		return libhlp.ErrorResponse.Error(cli.GetError(err))
	} else if err != nil {
		return cli.GetError(err)
	} else {
		return nil
	}
}

func (cli *client) MultipartCancel(uploadId, key string) liberr.Error {
	return cli._MultipartCancel(nil, sdkaws.String(uploadId), key)
}
