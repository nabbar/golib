package object

import (
	"io"
	"os"

	"github.com/nabbar/golib/ioutils"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

const DefaultPartSize = 5 * libhlp.SizeMegaBytes

func (cli *client) MultipartPut(object string, body io.Reader) liberr.Error {
	return cli.MultipartPutCustom(DefaultPartSize, object, body)
}

func (cli *client) MultipartPutCustom(partSize libhlp.PartSize, object string, body io.Reader) liberr.Error {
	var (
		tmp ioutils.FileProgress
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

		tmp, err = ioutils.NewFileProgressTemp()
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

		prt, err = cli.s3.UploadPart(cli.GetContext(), &sdksss.UploadPartInput{
			Bucket:        sdkaws.String(cli.GetBucketName()),
			Body:          tmp,
			PartNumber:    sdkaws.Int32(rio.CurrPart()),
			UploadId:      upl.UploadId,
			Key:           sdkaws.String(object),
			ContentLength: sdkaws.Int64(inf.Size()),
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
		UploadId:        upl.UploadId,
		MultipartUpload: rio.CompPart(),
		Bucket:          sdkaws.String(cli.GetBucketName()),
		Key:             sdkaws.String(object),
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
