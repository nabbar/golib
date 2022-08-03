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
	"context"
	"io"
	"time"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type Object interface {
	Find(regex string) ([]string, liberr.Error)
	Size(object string) (size int64, err liberr.Error)

	List(continuationToken string) ([]sdktps.Object, string, int64, liberr.Error)
	ListPrefix(continuationToken string, prefix string) ([]sdktps.Object, string, int64, liberr.Error)

	Head(object string) (*sdksss.HeadObjectOutput, liberr.Error)
	Get(object string) (*sdksss.GetObjectOutput, liberr.Error)
	Put(object string, body io.Reader) liberr.Error
	Delete(check bool, object string) liberr.Error
	DeleteAll(objects *sdktps.Delete) ([]sdktps.DeletedObject, liberr.Error)
	GetAttributes(object, version string) (*sdksss.GetObjectAttributesOutput, liberr.Error)

	MultipartList(keyMarker, markerId string) (uploads []sdktps.MultipartUpload, nextKeyMarker string, nextIdMarker string, count int64, e liberr.Error)
	MultipartPut(object string, body io.Reader) liberr.Error
	MultipartPutCustom(partSize libhlp.PartSize, object string, body io.Reader) liberr.Error
	MultipartCancel(uploadId, key string) liberr.Error

	UpdateMetadata(meta *sdksss.CopyObjectInput) liberr.Error
	SetWebsite(object, redirect string) liberr.Error

	VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err liberr.Error)
	VersionGet(object, version string) (*sdksss.GetObjectOutput, liberr.Error)
	VersionHead(object, version string) (*sdksss.HeadObjectOutput, liberr.Error)
	VersionSize(object, version string) (size int64, err liberr.Error)
	VersionDelete(check bool, object, version string) liberr.Error

	GetRetention(object, version string) (*sdktps.ObjectLockRetention, liberr.Error)
	SetRetention(object, version string, retentionUntil time.Time) liberr.Error

	GetTags(object, version string) ([]sdktps.Tag, liberr.Error)
	SetTags(object, version string, tags ...sdktps.Tag) liberr.Error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Object {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
