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

// client implements the Object interface for S3 object operations.
type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

// Metadata contains statistics about objects, versions, and delete markers.
type Metadata struct {
	Objects       int // Number of current objects
	Versions      int // Number of object versions
	DeleteMarkers int // Number of delete markers
}

// WalkFunc is a callback function for iterating over objects.
// Return true to continue iteration, false to stop.
type WalkFunc func(obj sdktps.Object) bool

// WalkFuncMetadata is a callback function called with metadata statistics.
// Return true to continue iteration, false to stop.
type WalkFuncMetadata func(md Metadata) bool

// WalkFuncVersion is a callback function for iterating over object versions.
// Return true to continue iteration, false to stop.
type WalkFuncVersion func(obj sdktps.ObjectVersion) bool

// WalkFuncDelMak is a callback function for iterating over delete markers.
// Return true to continue iteration, false to stop.
type WalkFuncDelMak func(del sdktps.DeleteMarkerEntry) bool

// Object provides operations for S3 object management including:
// - CRUD operations (create, read, update, delete)
// - Multipart upload for large files
// - Object versioning and lifecycle
// - Metadata and tagging
// - Object locking and retention
// - Legal hold management
type Object interface {
	// Find searches for objects matching the given regex pattern.
	Find(regex string) ([]string, error)

	// Size returns the size of the specified object in bytes.
	Size(object string) (size int64, err error)

	// List returns objects in the bucket, with pagination support.
	// Returns: objects, nextContinuationToken, count, error
	List(continuationToken string) ([]sdktps.Object, string, int64, error)

	// Walk iterates over all objects, calling md with metadata and f for each object.
	Walk(md WalkFuncMetadata, f WalkFunc) error

	// ListPrefix returns objects with the specified prefix, with pagination.
	// Returns: objects, nextContinuationToken, count, error
	ListPrefix(continuationToken string, prefix string) ([]sdktps.Object, string, int64, error)

	// WalkPrefix iterates over objects with the specified prefix.
	WalkPrefix(prefix string, md WalkFuncMetadata, f WalkFunc) error

	// Head retrieves metadata for an object without downloading it.
	Head(object string) (*sdksss.HeadObjectOutput, error)

	// Get retrieves an object's content and metadata.
	// Caller must close the returned Body ReadCloser.
	Get(object string) (*sdksss.GetObjectOutput, error)

	// Put uploads an object from an io.Reader.
	// For large files (>5MB), consider using MultipartPut.
	Put(object string, body io.Reader) error

	// Copy copies an object within the same bucket.
	Copy(source, destination string) error

	// CopyBucket copies an object between different buckets.
	CopyBucket(bucketSource, source, bucketDestination, destination string) error

	// Delete removes an object. If check is true, verifies object exists first.
	Delete(check bool, object string) error

	// DeleteAll removes multiple objects in a single request.
	DeleteAll(objects *sdktps.Delete) ([]sdktps.DeletedObject, error)

	// GetAttributes retrieves object attributes for a specific version.
	GetAttributes(object, version string) (*sdksss.GetObjectAttributesOutput, error)

	// MultipartList lists in-progress multipart uploads with pagination.
	MultipartList(keyMarker, markerId string) (uploads []sdktps.MultipartUpload, nextKeyMarker string, nextIdMarker string, count int64, e error)

	// MultipartNew creates a new multipart upload session.
	// Use for manual control of multipart upload process.
	MultipartNew(partSize libsiz.Size, bucket, object string) libmpu.MultiPart

	// MultipartCopy copies an object using multipart upload.
	// Useful for large objects (>5GB).
	MultipartCopy(partSize libsiz.Size, bucketSource, source, version, bucketDestination, destination string) error

	// MultipartPut uploads an object using multipart upload with default part size (10MB).
	MultipartPut(object string, body io.Reader) error

	// MultipartPutCustom uploads an object using multipart upload with custom part size.
	MultipartPutCustom(partSize libsiz.Size, object string, body io.Reader) error

	// MultipartCancel aborts a multipart upload session.
	MultipartCancel(uploadId, key string) error

	// UpdateMetadata updates object metadata by copying it to itself.
	UpdateMetadata(meta *sdksss.CopyObjectInput) error

	// SetWebsite configures website redirect for an object.
	SetWebsite(object, redirect string) error

	// VersionList lists all versions of objects with pagination.
	VersionList(prefix, keyMarker, markerId string) (version []sdktps.ObjectVersion, delMarker []sdktps.DeleteMarkerEntry, nextKeyMarker, nextMarkerId string, count int64, err error)

	// VersionWalk iterates over all object versions and delete markers.
	VersionWalk(md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error

	// VersionWalkPrefix iterates over versions with the specified prefix.
	VersionWalkPrefix(prefix string, md WalkFuncMetadata, fv WalkFuncVersion, fd WalkFuncDelMak) error

	// VersionGet retrieves a specific version of an object.
	VersionGet(object, version string) (*sdksss.GetObjectOutput, error)

	// VersionHead retrieves metadata for a specific version.
	VersionHead(object, version string) (*sdksss.HeadObjectOutput, error)

	// VersionSize returns the size of a specific version in bytes.
	VersionSize(object, version string) (size int64, err error)

	// VersionDelete deletes a specific version of an object.
	VersionDelete(check bool, object, version string) error

	// VersionCopy copies a specific version within the same bucket.
	VersionCopy(source, version, destination string) error

	// VersionCopyBucket copies a specific version between buckets.
	VersionCopyBucket(bucketSource, source, version, bucketDestination, destination string) error

	// VersionDeleteLock deletes a locked version, optionally bypassing governance mode.
	VersionDeleteLock(check bool, object, version string, byPassGovernance bool) error

	// GetRetention retrieves the retention configuration for an object version.
	// Returns: retainUntilDate, mode, error
	GetRetention(object, version string) (until time.Time, mode string, err error)

	// SetRetention configures retention for an object version.
	// bypass: bypass governance mode restrictions
	// until: retain until this time
	// mode: GOVERNANCE or COMPLIANCE
	SetRetention(object, version string, bypass bool, until time.Time, mode string) error

	// GetLegalHold retrieves the legal hold status for an object version.
	GetLegalHold(object, version string) (sdktps.ObjectLockLegalHoldStatus, error)

	// SetLegalHold enables or disables legal hold on an object version.
	// Legal hold prevents deletion regardless of retention period.
	SetLegalHold(object, version string, flag sdktps.ObjectLockLegalHoldStatus) error

	// GetTags retrieves tags for an object or specific version.
	GetTags(object, version string) ([]sdktps.Tag, error)

	// SetTags sets tags for an object or specific version.
	SetTags(object, version string, tags ...sdktps.Tag) error
}

// New creates a new Object client for the specified bucket.
// The client provides access to all object management operations.
func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Object {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
