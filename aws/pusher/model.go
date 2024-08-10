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
	"sync/atomic"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
)

type psh struct {
	ctx context.Context
	run *atomic.Bool  // is running / starting
	end *atomic.Bool  // is closed
	tmp *atomic.Value // working *os.File
	cfg *Config       // configuration

	prtSha2 *atomic.Value // part sha256
	prtMD5  *atomic.Value // part md5
	objSha2 *atomic.Value // object sha256

	updInfo *atomic.Value // sdksss.CreateMultipartUploadOutput
	nbrPart *atomic.Int32 // number of part
	prtList *atomic.Value // []sdksss.CompletedPart - part list

	prtSize *atomic.Int64 // current part size
	objSize *atomic.Int64 // object size
}

func (o *psh) Abort() error {
	if !o.IsReady() {
		return ErrInvalidInstance
	} else if !o.IsMPU() {
		return nil
	} else if u, e := o.getUploadId(); e != nil {
		return e
	} else if len(u) < 1 {
		return nil
	}

	return o.abortUpload()
}

func (o *psh) Complete() error {
	if !o.IsReady() {
		return ErrInvalidInstance
	}

	if o.prtSize.Load() > 0 {
		if e := o.pushObject(); e != nil {
			return e
		}
	}

	if !o.IsMPU() {
		return nil
	} else if u, e := o.getUploadId(); e != nil {
		return e
	} else if len(u) < 1 {
		return nil
	}

	return o.completeUpload()
}

func (o *psh) CopyFromS3(bucket, object, versionId string) error {

	var (
		e error
		c *sdksss.Client
		u string
		l []*sdksss.UploadPartCopyInput
		r *sdksss.UploadPartCopyOutput
	)

	if !o.IsReady() {
		return ErrInvalidInstance
	} else if c = o.cfg.getClientS3(); c == nil {
		return ErrInvalidClient
	} else if u, e = o.getUploadId(); e != nil {
		return e
	} else if l, e = o.getCopyObjectPart(bucket, object, versionId); e != nil {
		return e
	} else if len(l) < 1 {
		return nil
	} else {
		if len(u) < 1 {
			if u, e = o.createUploadId(); e != nil {
				return e
			}
		}

		if len(u) < 1 {
			return ErrInvalidUploadID
		} else if !o.IsStarted() {
			return ErrInvalidUploadID
		}

		for _, p := range l {
			// update uploadId & part number
			p.UploadId = sdkaws.String(u)
			p.PartNumber = sdkaws.Int32(o.nbrPart.Add(1))

			if r, e = c.UploadPartCopy(o.ctx, p); e != nil {
				o.nbrPart.Add(-1)
				return e
			} else if r == nil || r.CopyPartResult == nil || r.CopyPartResult.ETag == nil || len(*r.CopyPartResult.ETag) < 1 {
				o.nbrPart.Add(-1)
				return ErrInvalidResponse
			} else {
				o.appendPartList(p.PartNumber, sdksss.UploadPartOutput{
					ChecksumCRC32:  r.CopyPartResult.ChecksumCRC32,
					ChecksumCRC32C: r.CopyPartResult.ChecksumCRC32C,
					ChecksumSHA1:   r.CopyPartResult.ChecksumSHA1,
					ChecksumSHA256: r.CopyPartResult.ChecksumSHA256,
					ETag:           r.CopyPartResult.ETag,
				})
				o.cfg.onUpload(o.GetLastPartInfo(), o.GetObjectInfo(), nil)
			}
		}

		return o.Complete()
	}
}
