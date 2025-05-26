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
	libmpu "github.com/nabbar/golib/aws/multipart"
	libsiz "github.com/nabbar/golib/size"
)

type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

type Metadata struct {
	Objects       int
	Versions      int
	DeleteMarkers int
}

type WalkFunc func(obj sdktps.Object) bool
type WalkFuncMetadata func(md Metadata) bool
type WalkFuncVersion func(obj sdktps.ObjectVersion) bool
type WalkFuncDelMak func(del sdktps.DeleteMarkerEntry) bool

type Object interface {
	Find(regex string) ([]string, error)
	Size(object string) (size int64, err error)

	List(continuationToken string) ([]sdktps.Object, string, int64, error)
	Walk(md WalkFuncMetadata, f WalkFunc) error

	ListPrefix(continuationToken string, prefix string) ([]sdktps.Object, string, int64, error)
	WalkPrefix(prefix string, md WalkFuncMetadata, f WalkFunc) error

	Head(object string) (*sdksss.HeadObjectOutput, error)
	Get(object string) (*sdksss.GetObjectOutput, error)
	Put(object string, body io.Reader) error
	Copy(source, destination string) error
	CopyBucket(bucketSource, source, bucketDestination, destination string) error
	Delete(check bool, object string) error
	DeleteAll(objects *sdktps.Delete) ([]sdktps.DeletedObject, error)
	GetAttributes(object, version string) (*sdksss.GetObjectAttributesOutput, error)

	MultipartList(keyMarker, markerId string) (uploads []sdktps.MultipartUpload, nextKeyMarker string, nextIdMarker string, count int64, e error)
	MultipartNew(partSize libsiz.Size, bucket, object string) libmpu.MultiPart
	MultipartCopy(partSize libsiz.Size, bucketSource, source, version, bucketDestination, destination string) error
	MultipartPut(object string, body io.Reader) error
	MultipartPutCustom(partSize libsiz.Size, object string, body io.Reader) error
	MultipartCancel(uploadId, key string) error

	UpdateMetadata(meta *sdksss.CopyObjectInput) error
	SetWebsite(object, redirect string) error

	VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err error)
	VersionWalk(md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error
	VersionWalkPrefix(prefix string, md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error

	VersionGet(object, version string) (*sdksss.GetObjectOutput, error)
	VersionHead(object, version string) (*sdksss.HeadObjectOutput, error)
	VersionSize(object, version string) (size int64, err error)
	VersionDelete(check bool, object, version string) error
	VersionCopy(source, version, destination string) error
	VersionCopyBucket(bucketSource, source, version, bucketDestination, destination string) error
	VersionDeleteLock(check bool, object, version string, byPassGovernance bool) error

	GetRetention(object, version string) (until time.Time, mode string, err error)
	SetRetention(object, version string, bypass bool, until time.Time, mode string) error
	GetLegalHold(object, version string) (sdktps.ObjectLockLegalHoldStatus, error)
	SetLegalHold(object, version string, flag sdktps.ObjectLockLegalHoldStatus) error

	GetTags(object, version string) ([]sdktps.Tag, error)
	SetTags(object, version string, tags ...sdktps.Tag) error
}

func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Object {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
