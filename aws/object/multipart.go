package object

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

const buffSize = 64 * 1024 // double buff of io.copyBuffer

func (cli *client) MultipartPut(object string, body io.Reader) errors.Error {
	return cli.MultipartPutCustom(helper.SetSizeInt64(s3manager.MinUploadPartSize), object, body, 0)
}

func (cli *client) MultipartPutCustom(partSize helper.PartSize, object string, body io.Reader, concurrent int) errors.Error {
	uploader := s3manager.NewUploaderWithClient(cli.s3)

	if partSize > 0 {
		uploader.PartSize = partSize.Int64()
	} else {
		uploader.PartSize = helper.SetSizeInt64(s3manager.MinUploadPartSize).Int64()
	}

	if concurrent > 0 {
		uploader.Concurrency = concurrent
	}

	// Set Buffer size to 64Kb (this is the min size available)
	uploader.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(buffSize)

	_, err := uploader.UploadWithContext(cli.GetContext(), &s3manager.UploadInput{
		Bucket: cli.GetBucketAws(),
		Key:    aws.String(object),
		Body:   body,
	})

	return cli.GetError(err)
}
