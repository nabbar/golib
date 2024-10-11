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
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libsiz "github.com/nabbar/golib/size"
)

const (
	PartSizeMinimal       = 5 * libsiz.SizeMega
	PartSizeMaximal       = 5 * libsiz.SizeGiga
	MaxNumberPart   int32 = 10000
	MaxObjectSize         = 5 * libsiz.SizeTera
)

func (o *psh) getPartList() []sdktps.CompletedPart {
	if !o.IsReady() {
		return make([]sdktps.CompletedPart, 0)
	} else if i := o.prtList.Load(); i == nil {
		return make([]sdktps.CompletedPart, 0)
	} else if v, k := i.([]sdktps.CompletedPart); !k {
		return make([]sdktps.CompletedPart, 0)
	} else {
		return v
	}
}

func (o *psh) resetPartList() {
	if o.IsReady() {
		o.prtList.Store(make([]sdktps.CompletedPart, 0))
	}
}

func (o *psh) appendPartList(num *int32, res sdksss.UploadPartOutput) {
	if !o.IsReady() {
		return
	}

	l := o.getPartList()
	if len(l) < 1 {
		l = make([]sdktps.CompletedPart, 0)
	}

	o.prtList.Store(append(l, sdktps.CompletedPart{
		ChecksumCRC32:  res.ChecksumCRC32,
		ChecksumCRC32C: res.ChecksumCRC32C,
		ChecksumSHA1:   res.ChecksumSHA1,
		ChecksumSHA256: res.ChecksumSHA256,
		ETag:           res.ETag,
		PartNumber:     num,
	}))
}

func (o *psh) GetPartSize() libsiz.Size {
	if o == nil {
		return PartSizeMinimal
	} else {
		return o.cfg.getPartSize()
	}
}

func (o *psh) GetObjectSize() libsiz.Size {
	if o == nil {
		return 0
	} else {
		return libsiz.SizeFromInt64(o.objSize.Load())
	}
}

func (o *psh) GetObjectSizeLeft() libsiz.Size {
	return MaxObjectSize - o.GetObjectSize()
}

func (o *psh) GetObjectInfo() ObjectInfo {
	if o == nil {
		return ObjectInfo{}
	}

	if b := o.cfg.ObjectS3Options.Bucket; b == nil || len(*b) < 1 {
		return ObjectInfo{}
	} else if k := o.cfg.ObjectS3Options.Key; k == nil || len(*k) < 1 {
		return ObjectInfo{}
	} else {
		return ObjectInfo{
			Bucket:     *b,
			Object:     *k,
			IsMPU:      o.IsMPU(),
			TotalSize:  o.GetObjectSize(),
			NumberPart: o.Counter(),
		}
	}
}

func (o *psh) GetLastPartInfo() UploadInfo {
	if o == nil {
		return UploadInfo{}
	}

	var res = UploadInfo{
		IsMPU:      o.IsMPU(),
		PartNumber: o.nbrPart.Load(),
		UploadID:   "",
		Etag:       "",
		Checksum:   "",
	}

	if !o.IsReady() {
		return res
	} else if u, e := o.getUploadId(); e == nil && len(u) > 0 {
		res.UploadID = u
	}

	if l := o.getPartList(); len(l) > 0 {
		if p := l[len(l)-1]; p.ETag != nil && len(*p.ETag) > 0 {
			res.Etag = *p.ETag

			if p.ChecksumSHA256 != nil && len(*p.ChecksumSHA256) > 0 {
				res.Checksum = *p.ChecksumSHA256
			}
		}
	}

	return res
}

func (o *psh) IsReady() bool {
	if o == nil {
		return false
	} else {
		return !o.end.Load()
	}
}

func (o *psh) IsStarted() bool {
	if !o.IsReady() {
		return false
	} else if o.run.Load() {
		return true
	} else {
		u, e := o.getUploadId()
		return e == nil && len(u) > 0
	}
}

func (o *psh) IsMPU() bool {
	if !o.IsStarted() {
		return false
	} else {
		return len(o.getPartList()) > 0
	}
}

func (o *psh) Counter() int32 {
	return int32(len(o.getPartList()))
}

func (o *psh) CounterLeft() int32 {
	return MaxNumberPart - o.Counter()
}
