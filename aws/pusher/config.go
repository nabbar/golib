/*
 *  MIT License
 *
 *  Copyright (c) 2024 Nicolas JUHEL
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

package pusher

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libsiz "github.com/nabbar/golib/size"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

type ConfigObjectOptions struct {
	// The name of the bucket where the multipart upload is initiated and where the
	// object is uploaded.
	//
	// Directory buckets - When you use this operation with a directory bucket, you
	// must use virtual-hosted-style requests in the format
	// Bucket_name.s3express-az_id.region.amazonaws.com . Path-style requests are not
	// supported. Directory bucket names must be unique in the chosen Availability
	// Zone. Bucket names must follow the format bucket_base_name--az-id--x-s3 (for
	// example, DOC-EXAMPLE-BUCKET--usw2-az1--x-s3 ). For information about bucket
	// naming restrictions, see [Directory bucket naming rules]in the Amazon S3 User Guide.
	//
	// Access points - When you use this action with an access point, you must provide
	// the alias of the access point in place of the bucket name or specify the access
	// point ARN. When using the access point ARN, you must direct requests to the
	// access point hostname. The access point hostname takes the form
	// AccessPointName-AccountId.s3-accesspoint.Region.amazonaws.com. When using this
	// action with an access point through the Amazon Web Services SDKs, you provide
	// the access point ARN in place of the bucket name. For more information about
	// access point ARNs, see [Using access points]in the Amazon S3 User Guide.
	//
	// Access points and Object Lambda access points are not supported by directory
	// buckets.
	//
	// S3 on Outposts - When you use this action with Amazon S3 on Outposts, you must
	// direct requests to the S3 on Outposts hostname. The S3 on Outposts hostname
	// takes the form
	// AccessPointName-AccountId.outpostID.s3-outposts.Region.amazonaws.com . When you
	// use this action with S3 on Outposts through the Amazon Web Services SDKs, you
	// provide the Outposts access point ARN in place of the bucket name. For more
	// information about S3 on Outposts ARNs, see [What is S3 on Outposts?]in the Amazon S3 User Guide.
	//
	// [Directory bucket naming rules]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-naming-rules.html
	// [What is S3 on Outposts?]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/S3onOutposts.html
	// [Using access points]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-access-points.html
	//
	// This member is required.
	Bucket *string

	// Object key for which the multipart upload is to be initiated.
	//
	// This member is required.
	Key *string

	// The canned ACL to apply to the object. For more information, see [Canned ACL] in the Amazon
	// S3 User Guide.
	//
	// When adding a new object, you can use headers to grant ACL-based permissions to
	// individual Amazon Web Services accounts or to predefined groups defined by
	// Amazon S3. These permissions are then added to the ACL on the object. By
	// default, all objects are private. Only the owner has full access control. For
	// more information, see [Access Control List (ACL) Overview]and [Managing ACLs Using the REST API] in the Amazon S3 User Guide.
	//
	// If the bucket that you're uploading objects to uses the bucket owner enforced
	// setting for S3 Object Ownership, ACLs are disabled and no longer affect
	// permissions. Buckets that use this setting only accept PUT requests that don't
	// specify an ACL or PUT requests that specify bucket owner full control ACLs, such
	// as the bucket-owner-full-control canned ACL or an equivalent form of this ACL
	// expressed in the XML format. PUT requests that contain other ACLs (for example,
	// custom grants to certain Amazon Web Services accounts) fail and return a 400
	// error with the error code AccessControlListNotSupported . For more information,
	// see [Controlling ownership of objects and disabling ACLs]in the Amazon S3 User Guide.
	//
	//   - This functionality is not supported for directory buckets.
	//
	//   - This functionality is not supported for Amazon S3 on Outposts.
	//
	// [Managing ACLs Using the REST API]: https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-using-rest-api.html
	// [Access Control List (ACL) Overview]: https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html
	// [Canned ACL]: https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL
	// [Controlling ownership of objects and disabling ACLs]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/about-object-ownership.html
	ACL sdktps.ObjectCannedACL

	// Can be used to specify caching behavior along the request/reply chain. For more
	// information, see [http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9].
	//
	// [http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9]: http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9
	CacheControl *string

	// Specifies presentational information for the object. For more information, see [https://www.rfc-editor.org/rfc/rfc6266#section-4].
	//
	// [https://www.rfc-editor.org/rfc/rfc6266#section-4]: https://www.rfc-editor.org/rfc/rfc6266#section-4
	ContentDisposition *string

	// Specifies what content encodings have been applied to the object and thus what
	// decoding mechanisms must be applied to obtain the media-type referenced by the
	// Content-Type header field. For more information, see [https://www.rfc-editor.org/rfc/rfc9110.html#field.content-encoding].
	//
	// [https://www.rfc-editor.org/rfc/rfc9110.html#field.content-encoding]: https://www.rfc-editor.org/rfc/rfc9110.html#field.content-encoding
	ContentEncoding *string

	// The language the content is in.
	ContentLanguage *string

	// A standard MIME type describing the format of the contents. For more
	// information, see [https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type].
	//
	// [https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type]: https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type
	ContentType *string

	// The date and time at which the object is no longer cacheable. For more
	// information, see [https://www.rfc-editor.org/rfc/rfc7234#section-5.3].
	//
	// [https://www.rfc-editor.org/rfc/rfc7234#section-5.3]: https://www.rfc-editor.org/rfc/rfc7234#section-5.3
	Expires *time.Time

	// Gives the grantee READ, READ_ACP, and WRITE_ACP permissions on the object.
	//
	//   - This functionality is not supported for directory buckets.
	//
	//   - This functionality is not supported for Amazon S3 on Outposts.
	GrantFullControl *string

	// Allows grantee to read the object data and its metadata.
	//
	//   - This functionality is not supported for directory buckets.
	//
	//   - This functionality is not supported for Amazon S3 on Outposts.
	GrantRead *string

	// Allows grantee to read the object ACL.
	//
	//   - This functionality is not supported for directory buckets.
	//
	//   - This functionality is not supported for Amazon S3 on Outposts.
	GrantReadACP *string

	// Allows grantee to write the ACL for the applicable object.
	//
	//   - This functionality is not supported for directory buckets.
	//
	//   - This functionality is not supported for Amazon S3 on Outposts.
	GrantWriteACP *string

	// A map of metadata to store with the object in S3.
	Metadata map[string]string

	// Specifies whether a legal hold will be applied to this object. For more
	// information about S3 Object Lock, see [Object Lock]in the Amazon S3 User Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Object Lock]: https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock.html
	ObjectLockLegalHoldStatus sdktps.ObjectLockLegalHoldStatus

	// The Object Lock mode that you want to apply to this object.
	//
	// This functionality is not supported for directory buckets.
	ObjectLockMode sdktps.ObjectLockMode

	// The date and time when you want this object's Object Lock to expire. Must be
	// formatted as a timestamp parameter.
	//
	// This functionality is not supported for directory buckets.
	ObjectLockRetainUntilDate *time.Time

	// By default, Amazon S3 uses the STANDARD Storage Class to store newly created
	// objects. The STANDARD storage class provides high durability and high
	// availability. Depending on performance needs, you can specify a different
	// Storage Class. For more information, see [Storage Classes]in the Amazon S3 User Guide.
	//
	//   - For directory buckets, only the S3 Express One Zone storage class is
	//   supported to store newly created objects.
	//
	//   - Amazon S3 on Outposts only uses the OUTPOSTS Storage Class.
	//
	// [Storage Classes]: https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html
	StorageClass sdktps.StorageClass

	// The tag-set for the object. The tag-set must be encoded as URL Query
	// parameters. (For example, "Key1=Value1")
	//
	// This functionality is not supported for directory buckets.
	Tagging *string

	// If the bucket is configured as a website, redirects requests for this object to
	// another object in the same bucket or to an external URL. Amazon S3 stores the
	// value of this header in the object metadata. For information about object
	// metadata, see [Object Key and Metadata]in the Amazon S3 User Guide.
	//
	// In the following example, the request header sets the redirect to an object
	// (anotherPage.html) in the same bucket:
	//
	//     x-amz-website-redirect-location: /anotherPage.html
	//
	// In the following example, the request header sets the object redirect to
	// another website:
	//
	//     x-amz-website-redirect-location: http://www.example.com/
	//
	// For more information about website hosting in Amazon S3, see [Hosting Websites on Amazon S3] and [How to Configure Website Page Redirects] in the
	// Amazon S3 User Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [How to Configure Website Page Redirects]: https://docs.aws.amazon.com/AmazonS3/latest/dev/how-to-page-redirect.html
	// [Hosting Websites on Amazon S3]: https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteHosting.html
	// [Object Key and Metadata]: https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html
	WebsiteRedirectLocation *string
}

type Config struct {
	// FuncGetClientS3 is a function type that represents a function that returns a pointer to
	//   an *github.com/aws/aws-sdk-go-v2/service/s3.Client object.
	FuncGetClientS3 FuncClientS3
	// FuncCallOnUpload is a callback function to be executed when a part is uploaded or a complete object is uploaded.
	//
	// fct: is a function type that represents a function that takes three parameters:
	//   upd: UploadInfo that contains information about the last part uploaded,
	//   obj: ObjectInfo that contains information about the object being uploaded,
	//   e: error that contains any error that occurred during the upload.
	FuncCallOnUpload FuncOnUpload
	// FuncCallOnAbort is a callback function to be executed when the upload is aborted or close is called.
	//
	// fct: is a function type that represents a function that takes two parameters:
	//   obj: ObjectInfo that contains information about the object being uploaded,
	//   e: error that contains any error that occurred during the upload.
	FuncCallOnAbort FuncOnFinish
	// FuncCallOnComplete is a callback function to be executed when the upload is completed.
	//
	// fct: is a function type that represents a function that takes two parameters:
	//   obj: ObjectInfo that contains information about the object being uploaded,
	//   e: error that contains any error that occurred during the upload.
	FuncCallOnComplete FuncOnFinish
	// WorkingPath represents the working path for the part file. If the part file already exists, it will be overwritten / truncated.
	// When content of uploaded part written, it will be appended to this working file. This file will grow up to the size of a part.
	// If the size of a part has been reached, the current part is uploaded and the file is truncated. When the first part is uploaded, the mpu is created.
	// If the complete method is called before the size of part has been reached, the upload is converted to a standard put object and mpu is set to false.
	//
	// WorkingPath can be a relative or absolute path, directory or file.
	// If a directory is specified, the part file will be created in that directory with an unique name based on the bucket name and unique random part. .
	WorkingPath string
	// PartSize defines the size of each part for a multipart upload or the maximum size for a standard put object.
	// The part size is a size that is at least of 5 MB and at most of 5 GB.
	PartSize libsiz.Size
	// BufferSize defines the buffer size for the pusher.
	// This buffer size is useful if the readerFrom method is used.
	BufferSize int
	// CheckSum enables or disables the checksum for the pusher.
	// This is a integrity check of the uploaded part and object.
	// The checksum method used will be sha256 if true.
	CheckSum bool

	// ObjectS3Options defines the options for the object s3.
	ObjectS3Options ConfigObjectOptions
}

func (o *Config) getClientS3() *sdksss.Client {
	if o == nil {
		return nil
	} else if f := o.FuncGetClientS3; f == nil {
		return nil
	} else {
		return f()
	}
}

func (o *Config) onUpload(upd UploadInfo, obj ObjectInfo, e error) {
	if o == nil {
		return
	} else if f := o.FuncCallOnUpload; f != nil {
		f(upd, obj, e)
	}
}

func (o *Config) onAbort(obj ObjectInfo, e error) {
	if o == nil {
		return
	} else if f := o.FuncCallOnAbort; f != nil {
		f(obj, e)
	}
}

func (o *Config) onComplete(obj ObjectInfo, e error) {
	if o == nil {
		return
	} else if f := o.FuncCallOnComplete; f != nil {
		f(obj, e)
	}
}

func (o *Config) getPartSize() libsiz.Size {
	if o == nil {
		return PartSizeMinimal
	}

	if o.PartSize < PartSizeMinimal {
		o.PartSize = PartSizeMinimal
	} else if o.PartSize > PartSizeMaximal {
		o.PartSize = PartSizeMaximal
	}

	return o.PartSize
}

func (o *Config) getBufferSize() int {
	if s := o.BufferSize; s < 512 {
		return 512
	} else if s > 1024*1024 {
		return 1024 * 1024
	} else {
		return s
	}
}

func (o *Config) isCheckSum() bool {
	return o.CheckSum
}

func (o *Config) getUploadPartInput() *sdksss.UploadPartInput {
	var chk sdktps.ChecksumAlgorithm
	if o.CheckSum {
		chk = sdktps.ChecksumAlgorithmSha256
	}

	return &sdksss.UploadPartInput{
		Bucket:            o.ObjectS3Options.Bucket,
		Key:               o.ObjectS3Options.Key,
		ChecksumAlgorithm: chk,
	}
}

func (o *Config) getPutObjectInput() *sdksss.PutObjectInput {
	var chk sdktps.ChecksumAlgorithm
	if o.CheckSum {
		chk = sdktps.ChecksumAlgorithmSha256
	}

	return &sdksss.PutObjectInput{
		Bucket:                    o.ObjectS3Options.Bucket,
		Key:                       o.ObjectS3Options.Key,
		ACL:                       o.ObjectS3Options.ACL,
		CacheControl:              o.ObjectS3Options.CacheControl,
		ChecksumAlgorithm:         chk,
		ContentDisposition:        o.ObjectS3Options.ContentDisposition,
		ContentEncoding:           o.ObjectS3Options.ContentEncoding,
		ContentLanguage:           o.ObjectS3Options.ContentLanguage,
		ContentType:               o.ObjectS3Options.ContentType,
		Expires:                   o.ObjectS3Options.Expires,
		GrantFullControl:          o.ObjectS3Options.GrantFullControl,
		GrantRead:                 o.ObjectS3Options.GrantRead,
		GrantReadACP:              o.ObjectS3Options.GrantReadACP,
		GrantWriteACP:             o.ObjectS3Options.GrantWriteACP,
		Metadata:                  o.ObjectS3Options.Metadata,
		ObjectLockLegalHoldStatus: o.ObjectS3Options.ObjectLockLegalHoldStatus,
		ObjectLockMode:            o.ObjectS3Options.ObjectLockMode,
		ObjectLockRetainUntilDate: o.ObjectS3Options.ObjectLockRetainUntilDate,
		StorageClass:              o.ObjectS3Options.StorageClass,
		Tagging:                   o.ObjectS3Options.Tagging,
		WebsiteRedirectLocation:   o.ObjectS3Options.WebsiteRedirectLocation,
	}
}

func (o *Config) getCreateMultipartUploadInput() *sdksss.CreateMultipartUploadInput {
	var chk sdktps.ChecksumAlgorithm
	if o.CheckSum {
		chk = sdktps.ChecksumAlgorithmSha256
	}

	return &sdksss.CreateMultipartUploadInput{
		Bucket:                    o.ObjectS3Options.Bucket,
		Key:                       o.ObjectS3Options.Key,
		ACL:                       o.ObjectS3Options.ACL,
		CacheControl:              o.ObjectS3Options.CacheControl,
		ChecksumAlgorithm:         chk,
		ContentDisposition:        o.ObjectS3Options.ContentDisposition,
		ContentEncoding:           o.ObjectS3Options.ContentEncoding,
		ContentLanguage:           o.ObjectS3Options.ContentLanguage,
		ContentType:               o.ObjectS3Options.ContentType,
		Expires:                   o.ObjectS3Options.Expires,
		GrantFullControl:          o.ObjectS3Options.GrantFullControl,
		GrantRead:                 o.ObjectS3Options.GrantRead,
		GrantReadACP:              o.ObjectS3Options.GrantReadACP,
		GrantWriteACP:             o.ObjectS3Options.GrantWriteACP,
		Metadata:                  o.ObjectS3Options.Metadata,
		ObjectLockLegalHoldStatus: o.ObjectS3Options.ObjectLockLegalHoldStatus,
		ObjectLockMode:            o.ObjectS3Options.ObjectLockMode,
		ObjectLockRetainUntilDate: o.ObjectS3Options.ObjectLockRetainUntilDate,
		StorageClass:              o.ObjectS3Options.StorageClass,
		Tagging:                   o.ObjectS3Options.Tagging,
		WebsiteRedirectLocation:   o.ObjectS3Options.WebsiteRedirectLocation,
	}
}

func (o *Config) getCompleteMultipartUploadInput() *sdksss.CompleteMultipartUploadInput {
	return &sdksss.CompleteMultipartUploadInput{
		Bucket: o.ObjectS3Options.Bucket,
		Key:    o.ObjectS3Options.Key,
		MultipartUpload: &sdktps.CompletedMultipartUpload{
			Parts: make([]sdktps.CompletedPart, 0),
		},
	}
}

func (o *Config) getAbortMultipartUploadInput() *sdksss.AbortMultipartUploadInput {
	return &sdksss.AbortMultipartUploadInput{
		Bucket: o.ObjectS3Options.Bucket,
		Key:    o.ObjectS3Options.Key,
	}
}

func (o *Config) getUploadPartCopyInput(src, srcRange string) *sdksss.UploadPartCopyInput {
	return &sdksss.UploadPartCopyInput{
		Bucket:          o.ObjectS3Options.Bucket,
		CopySource:      sdkaws.String(src),
		Key:             o.ObjectS3Options.Key,
		PartNumber:      nil,
		UploadId:        nil,
		CopySourceRange: sdkaws.String("bytes=" + srcRange),
	}
}

func (o *Config) getWorkingFile() (*os.File, error) {
	wrk, inf, err := o.getWorkingPath()
	if err != nil {
		return nil, err
	} else if inf.IsDir() {
		return o.createTempWorkingFile(wrk)
	}

	r, e := os.OpenRoot(path.Dir(wrk))
	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if e != nil {
		return nil, e
	} else {
		return r.Create(path.Base(wrk))
	}
}

func (o *Config) getWorkingPath() (string, os.FileInfo, error) {
	if o.WorkingPath == "" {
		o.WorkingPath = os.TempDir()
	}

	var (
		err error
		inf os.FileInfo
	)

	if inf, err = os.Stat(o.WorkingPath); err != nil && os.IsNotExist(err) {
		if _, err = os.Stat(filepath.Dir(o.WorkingPath)); err != nil && os.IsNotExist(err) {
			return o.WorkingPath, nil, err
		} else if err != nil {
			return o.WorkingPath, nil, err
		} else if h, e := os.Create(o.WorkingPath); e != nil {
			return o.WorkingPath, nil, e
		} else {
			if e = h.Close(); e != nil {
				return o.WorkingPath, nil, e
			} else if inf, err = os.Stat(o.WorkingPath); err != nil {
				return o.WorkingPath, nil, err
			}
		}
	} else if err != nil {
		return o.WorkingPath, nil, err
	}

	if inf == nil {
		return o.WorkingPath, nil, os.ErrInvalid
	} else if !inf.IsDir() && !inf.Mode().IsRegular() {
		return o.WorkingPath, nil, os.ErrInvalid
	} else if inf.IsDir() {
		if h, e := os.CreateTemp(o.WorkingPath, "chk_*"); e != nil {
			return o.WorkingPath, nil, e
		} else {
			n := h.Name()
			if e = h.Close(); e != nil {
				return o.WorkingPath, nil, e
			} else if e = os.Remove(n); e != nil { // #nosec nolint
				return o.WorkingPath, nil, e
			}
		}
	}

	return o.WorkingPath, inf, nil
}

func (o *Config) createTempWorkingFile(workingPath string) (*os.File, error) {
	var pfx string

	if o.ObjectS3Options.Bucket != nil && len(*o.ObjectS3Options.Bucket) > 0 {
		pfx = clearString(*o.ObjectS3Options.Bucket) + "_*"
	} else if o.ObjectS3Options.Key != nil && len(*o.ObjectS3Options.Key) > 0 {
		pfx = clearString(*o.ObjectS3Options.Key) + "_*"
	} else {
		pfx = "obj_*"
	}

	return os.CreateTemp(workingPath, pfx)
}
