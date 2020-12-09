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
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
	libiou "github.com/nabbar/golib/ioutils"
)

const DefaultPartSize = 5 * libhlp.SizeMegaBytes

func (cli *client) MultipartPut(object string, body io.Reader) liberr.Error {
	return cli.MultipartPutCustom(DefaultPartSize, object, body)
}

func (cli *client) MultipartPutCustom(partSize libhlp.PartSize, object string, body io.Reader) liberr.Error {
	var (
		tmp libiou.FileProgress
		rio libhlp.ReaderPartSize
		upl *sdksss.CreateMultipartUploadOutput
		err error
	)

	defer func() {
		if tmp != nil {
			_ = tmp.Close()
		}
	}()

	upl, err = cli.s3.CreateMultipartUpload(cli.GetContext(), &sdksss.CreateMultipartUploadInput{
		Key:    sdkaws.String(object),
		Bucket: sdkaws.String(cli.GetBucketName()),
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
			return cli.multipartCancel(err, upl.UploadId, object)
		}

		_, err = io.Copy(tmp, rio)
		if err != nil {
			return cli.multipartCancel(err, upl.UploadId, object)
		}

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return cli.multipartCancel(err, upl.UploadId, object)
		}

		inf, err = tmp.FileStat()
		if err != nil {
			return cli.multipartCancel(err, upl.UploadId, object)
		}

		h := md5.New()
		if _, err := tmp.WriteTo(h); err != nil {
			return cli.multipartCancel(err, upl.UploadId, object)
		}

		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			return cli.multipartCancel(err, upl.UploadId, object)
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
			return cli.multipartCancel(err, upl.UploadId, object)
		} else if prt == nil || prt.ETag == nil || len(*prt.ETag) == 0 {
			return cli.multipartCancel(libhlp.ErrorResponse.Error(nil), upl.UploadId, object)
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
		return cli.multipartCancel(err, upl.UploadId, object)
	} else if prt == nil || prt.ETag == nil || len(*prt.ETag) == 0 {
		return cli.multipartCancel(libhlp.ErrorResponse.Error(nil), upl.UploadId, object)
	}

	return nil
}

func (cli *client) multipartCancel(err error, updIp *string, object string) liberr.Error {
	cnl, e := cli.s3.AbortMultipartUpload(cli.GetContext(), &sdksss.AbortMultipartUploadInput{
		Bucket:   sdkaws.String(cli.GetBucketName()),
		UploadId: updIp,
		Key:      sdkaws.String(object),
	})

	if e != nil {
		return cli.GetError(e, err)
	} else if cnl == nil {
		return libhlp.ErrorResponse.Error(cli.GetError(err))
	} else {
		return cli.GetError(err)
	}

}
