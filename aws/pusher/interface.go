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
	"context"
	"io"
	"sync/atomic"

	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	libsiz "github.com/nabbar/golib/size"
)

// ObjectInfo represents object information
type ObjectInfo struct {
	Bucket     string
	Object     string
	IsMPU      bool
	TotalSize  libsiz.Size
	NumberPart int32
}

// UploadInfo represents last part upload information
type UploadInfo struct {
	IsMPU      bool
	PartNumber int32
	UploadID   string
	Etag       string
	Checksum   string
}

// FuncClientS3 represents function to return an pure s3 client from aws s3 service sdk
type FuncClientS3 func() *sdksss.Client

// FuncOnUpload represents function to call when a part is uploaded or a complete object is uploaded
type FuncOnUpload func(upd UploadInfo, obj ObjectInfo, e error)

// FuncOnFinish represents function to call when upload is aborted or complete
type FuncOnFinish func(obj ObjectInfo, e error)

// Pusher is an helper interface to upload an object directly or with multiparts in s3
// this is a writer interface to be able to use it with io.Copy
// this interface allow to register callbacks and allow copy object from s3 in multiparts
type Pusher interface {
	io.WriteCloser
	io.ReaderFrom

	// Abort aborts the current multipart upload. This operation is irreversible.
	// This is same as calling close.
	//
	// This function stops the current multipart upload and releases any resources associated with it.
	// Return type: error
	Abort() error
	// Complete completes the current multipart upload.
	//
	// This function finalizes the multipart upload and returns any error that occurred during the process.
	// This will concatenate all the parts into a single object and checking the checksum if enabled.
	// Return type: error
	Complete() error

	CopyFromS3(bucket, object, versionId string) error

	// GetPartSize returns the size of each part for a multipart upload or the maximum size for a standard put object.
	//
	// Return type: libsiz.Size
	GetPartSize() libsiz.Size
	// GetObjectInfo returns information about the object being uploaded.
	//
	// This method provides details about the object, including its bucket, key, and total size.
	// Return type: ObjectInfo
	GetObjectInfo() ObjectInfo
	// GetLastPartInfo returns the last part upload information.
	//
	// Return type: UploadInfo
	GetLastPartInfo() UploadInfo
	// GetObjectSize returns the total size of the object currently uploaded.
	//
	// Return type: libsiz.Size
	GetObjectSize() libsiz.Size
	// GetObjectSizeLeft returns the remaining size of the object to be uploaded.
	//
	// This method provides the size left to upload for the current object.
	// Return type: libsiz.Size
	GetObjectSizeLeft() libsiz.Size

	// IsMPU indicates whether the current upload is a multipart upload.
	//
	// This method checks the type of the upload and returns true if it's a multipart upload, false otherwise.
	// Return type: bool
	IsMPU() bool
	// IsStarted returns a boolean indicating whether the creation to init a mpu or a put object has been done.
	//
	// Return type: bool
	IsStarted() bool
	// Counter returns the current number of parts uploaded.
	//
	// Return type: int32
	Counter() int32
	// CounterLeft returns the maximum number of parts left available to be uploaded.
	//
	// Return type: int32
	CounterLeft() int32
}

// New creates a new instance of Pusher.
// The Pusher is used to upload an object directly or with multipart in s3.
// The Pusher allows to register callbacks and allow copy object from s3 in multipart.
//
// The Pusher implements the io.WriteCloser and io.ReaderFrom interfaces. This allows it to be used with io.Copy.
// And so, Pusher is allowing to be used directly as an io.Writer or sending it an io.Reader.
//
// This function is a factory function named New that creates a new instance of Pusher.
// It takes a context.Context and a *Config as input, and returns a Pusher instance and an error.
// It checks if the cfg parameter is nil. If so, it returns an error ErrInvalidInstance.
// It creates a new instance of psh (a struct that implements Pusher) and initializes its fields.
// It calls an internal method on the cfg instance to get a working file and stores the result in the tmp field of the psh instance.
// If this internal method returns an error, it propagates it.
// If everything is successful, it returns the newly created psh instance as a Pusher and a nil error.
//
// Return type: Pusher, error
func New(ctx context.Context, cfg *Config) (Pusher, error) {
	if cfg == nil {
		return nil, ErrInvalidInstance
	}

	p := &psh{
		ctx:     ctx,
		run:     new(atomic.Bool),
		end:     new(atomic.Bool),
		tmp:     new(atomic.Value),
		cfg:     cfg,
		prtSha2: new(atomic.Value),
		prtMD5:  new(atomic.Value),
		objSha2: new(atomic.Value),
		updInfo: new(atomic.Value),
		nbrPart: new(atomic.Int32),
		prtList: new(atomic.Value),
		prtSize: new(atomic.Int64),
		objSize: new(atomic.Int64),
	}

	if i, e := cfg.getWorkingFile(); e != nil {
		return nil, e
	} else {
		p.tmp.Store(i)
	}

	return p, nil
}
