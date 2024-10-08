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
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktps "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (o *psh) getUploadId() (string, error) {
	if !o.IsReady() {
		return "", ErrInvalidInstance
	} else if i := o.updInfo.Load(); i == nil {
		return "", nil
	} else if v, k := i.(*sdksss.CreateMultipartUploadOutput); !k {
		return "", nil
	} else if v.UploadId == nil {
		return "", nil
	} else {
		return *(v.UploadId), nil
	}
}

func (o *psh) resetUploadId() {
	if o.IsReady() {
		o.updInfo.Store(&sdksss.CreateMultipartUploadOutput{})
	}
}

func (o *psh) createUploadId() (string, error) {
	if !o.IsReady() {
		return "", ErrInvalidInstance
	} else if c := o.cfg.getClientS3(); c == nil {
		return "", ErrInvalidClient
	} else if i := o.cfg.getCreateMultipartUploadInput(); i == nil {
		return "", ErrInvalidInstance
	} else if out, err := c.CreateMultipartUpload(o.ctx, i); err != nil {
		return "", err
	} else if out == nil {
		return "", ErrInvalidResponse
	} else if out.UploadId == nil || len(*out.UploadId) < 1 {
		return "", ErrInvalidResponse
	} else {
		o.updInfo.Store(out)
		o.run.Store(true)
		return *(out.UploadId), nil
	}
}

func (o *psh) getUploadPartInput() (*sdksss.UploadPartInput, error) {
	var (
		e error
		f io.Reader // part io reader
		s int64     // part size
		h []byte    // checksum result
		u string    // upload id
		i *sdksss.UploadPartInput
	)

	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else {
		i = o.cfg.getUploadPartInput()
	}

	if u, e = o.getUploadId(); e != nil {
		return nil, e
	}

	if len(u) < 1 {
		if u, e = o.createUploadId(); e != nil {
			return nil, e
		}
	}

	if len(u) < 1 {
		return nil, ErrInvalidUploadID
	} else if !o.IsStarted() {
		return nil, ErrInvalidUploadID
	} else {
		i.UploadId = sdkaws.String(u)
	}

	if e = o.fileReset(); e != nil {
		return nil, e
	} else if f, e = o.getFile(); e != nil {
		return nil, e
	} else {
		i.Body = f
	}

	if s = o.prtSize.Load(); s > 0 {
		i.ContentLength = sdkaws.Int64(s)
	} else {
		return nil, ErrEmptyContents
	}

	if h, e = o.md5Checksum(); e != nil {
		return nil, e
	} else if len(h) < 1 {
		return nil, ErrInvalidChecksum
	} else {
		i.ContentMD5 = sdkaws.String(base64.StdEncoding.EncodeToString(h))
	}

	if h, e = o.shaPartChecksum(); e == nil && len(h) > 0 {
		i.ChecksumAlgorithm = sdktps.ChecksumAlgorithmSha256
		i.ChecksumSHA256 = sdkaws.String(base64.StdEncoding.EncodeToString(h))
	}

	i.PartNumber = sdkaws.Int32(o.nbrPart.Add(1))

	return i, nil
}

func (o *psh) getPutObjectInput() (*sdksss.PutObjectInput, error) {
	var (
		e error
		i *sdksss.PutObjectInput
		f io.Reader // part io reader
		s int64     // part size
		h []byte    // checksum result
	)

	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if o.IsStarted() {
		return nil, ErrInvalidInstance
	}

	i = o.cfg.getPutObjectInput()

	if e = o.fileReset(); e != nil {
		return nil, e
	} else if f, e = o.getFile(); e != nil {
		return nil, e
	} else {
		i.Body = f
	}

	if s = o.prtSize.Load(); s > 0 {
		i.ContentLength = sdkaws.Int64(s)
	} else {
		return nil, ErrEmptyContents
	}

	if h, e = o.md5Checksum(); e != nil {
		return nil, e
	} else if len(h) < 1 {
		return nil, ErrInvalidChecksum
	} else {
		i.ContentMD5 = sdkaws.String(base64.StdEncoding.EncodeToString(h))
	}

	if h, e = o.shaObjChecksum(); e == nil && len(h) > 0 {
		i.ChecksumAlgorithm = sdktps.ChecksumAlgorithmSha256
		i.ChecksumSHA256 = sdkaws.String(base64.StdEncoding.EncodeToString(h))
	}

	return i, nil
}

func (o *psh) pushObject() error {
	var (
		err error
		ret bool
		fct = o.pushMPUObject
	)

	if !o.IsReady() {
		return ErrInvalidInstance
	} else if !o.IsMPU() && o.prtSize.Load() < o.cfg.PartSize.Int64() {
		fct = o.pushSingleObject
	}

	defer func() {
		if o.IsMPU() {
			o.cfg.onUpload(o.GetLastPartInfo(), o.GetObjectInfo(), err)
		} else {
			o.cfg.onComplete(o.GetObjectInfo(), err)
		}
	}()

	for i := 0; i < 10; i++ {
		if err, ret = fct(); err == nil {
			return nil
		} else if ret {
			return err
		}

		time.Sleep(10 * time.Second)
	}

	return err
}

func (o *psh) pushMPUObject() (error, bool) {
	if !o.IsReady() {
		return ErrInvalidInstance, false
	}

	var (
		e error
		c *sdksss.Client
		i *sdksss.UploadPartInput
		r *sdksss.UploadPartOutput
	)

	if c = o.cfg.getClientS3(); c == nil {
		return ErrInvalidClient, false
	} else if i, e = o.getUploadPartInput(); e != nil {
		return e, false
	} else if r, e = c.UploadPart(o.ctx, i); e == nil && (r == nil || r.ETag == nil || len(*r.ETag) < 1) {
		e = ErrInvalidResponse
	} else if e == nil {
		o.appendPartList(i.PartNumber, *r)
	}

	if e != nil {
		o.nbrPart.Add(-1)
		return e, false
	}

	_ = o.fileReset()
	_ = o.fileTruncate()
	_ = o.md5Reset()
	_ = o.shaPartReset()

	return nil, true
}

func (o *psh) pushSingleObject() (error, bool) {
	if !o.IsReady() {
		return ErrInvalidInstance, false
	} else if o.IsStarted() {
		return ErrInvalidInstance, false
	} else if o.IsMPU() {
		return ErrInvalidInstance, false
	}

	var (
		e error
		c *sdksss.Client
		i *sdksss.PutObjectInput
		r *sdksss.PutObjectOutput
	)

	if i, e = o.getPutObjectInput(); e != nil {
		return e, false
	} else if c = o.cfg.getClientS3(); c == nil {
		return ErrInvalidClient, false
	} else if r, e = c.PutObject(o.ctx, i); e == nil && (r == nil || r.ETag == nil || len(*r.ETag) < 1) {
		e = ErrInvalidResponse
	}

	if e != nil {
		return e, false
	}

	_ = o.fileReset()
	_ = o.fileTruncate()
	_ = o.fileRemove()
	_ = o.md5Reset()
	_ = o.shaPartReset()

	o.run.Store(true)
	o.end.Store(true)

	return nil, true
}

func (o *psh) abortUpload() error {
	var (
		e error
		u string         // upload id
		c *sdksss.Client // s3 client

		i *sdksss.AbortMultipartUploadInput
		r *sdksss.AbortMultipartUploadOutput
	)

	if !o.IsReady() {
		return ErrInvalidInstance
	} else if c = o.cfg.getClientS3(); c == nil {
		return ErrInvalidClient
	} else if u, e = o.getUploadId(); e != nil {
		return e
	} else if len(u) < 1 {
		return nil
	} else {
		i = o.cfg.getAbortMultipartUploadInput()
		i.UploadId = sdkaws.String(u)
	}

	defer func() {
		o.cfg.onAbort(o.GetObjectInfo(), e)
	}()

	if r, e = c.AbortMultipartUpload(o.ctx, i); e == nil && r == nil {
		e = ErrInvalidResponse
	}

	_ = o.fileReset()
	_ = o.fileTruncate()
	_ = o.fileRemove()
	o.end.Store(true)

	return e
}

func (o *psh) completeUpload() error {
	var (
		e error
		u string                 // upload id
		c *sdksss.Client         // s3 client
		l []sdktps.CompletedPart // part list

		i *sdksss.CompleteMultipartUploadInput
		r *sdksss.CompleteMultipartUploadOutput
	)

	if !o.IsReady() {
		return ErrInvalidInstance
	} else if c = o.cfg.getClientS3(); c == nil {
		return ErrInvalidClient
	} else if l = o.getPartList(); len(l) < 1 {
		return ErrEmptyContents
	} else if u, e = o.getUploadId(); e != nil {
		return e
	} else if len(u) < 1 {
		return ErrInvalidUploadID
	} else {
		i = o.cfg.getCompleteMultipartUploadInput()
		i.UploadId = sdkaws.String(u)

		for _, p := range l {
			i.MultipartUpload.Parts = append(i.MultipartUpload.Parts, sdktps.CompletedPart{
				ETag:       p.ETag,
				PartNumber: p.PartNumber,
			})
		}

		if chk, err := o.shaObjChecksum(); err == nil && len(chk) > 0 {
			i.ChecksumSHA256 = sdkaws.String(base64.StdEncoding.EncodeToString(chk))
		}
	}

	defer func() {
		o.cfg.onComplete(o.GetObjectInfo(), e)
	}()

	if r, e = c.CompleteMultipartUpload(o.ctx, i); e == nil && r == nil {
		e = ErrInvalidResponse
	}

	_ = o.fileReset()
	_ = o.fileTruncate()
	_ = o.fileRemove()
	o.end.Store(true)

	return e
}

func (o *psh) getHeadObject(bucket, object, versionID string) (*sdksss.HeadObjectOutput, error) {
	var (
		c  *sdksss.Client
		in *sdksss.HeadObjectInput
	)

	if !o.IsReady() {
		return nil, ErrInvalidInstance
	} else if c = o.cfg.getClientS3(); c == nil {
		return nil, ErrInvalidClient
	} else {
		in = &sdksss.HeadObjectInput{
			Bucket: sdkaws.String(bucket),
			Key:    sdkaws.String(object),
		}
	}

	if len(versionID) > 0 {
		in.VersionId = &versionID
	}

	if out, err := c.HeadObject(o.ctx, in); err != nil {
		return nil, err
	} else if out == nil || out.ETag == nil || len(*out.ETag) < 1 || out.ContentLength == nil || *out.ContentLength < 1 {
		return nil, ErrInvalidResponse
	} else {
		return out, nil
	}
}

func (o *psh) getCopyObjectPart(bucket, object, versionID string) ([]*sdksss.UploadPartCopyInput, error) {
	if !o.IsReady() {
		return nil, ErrInvalidInstance
	}

	var (
		e error
		h *sdksss.HeadObjectOutput    // head object
		i int64                       // current size cursor
		s int64                       // total size of object
		f = path.Join(bucket, object) // from source

		p = o.cfg.getPartSize().Int64()
		r = make([]*sdksss.UploadPartCopyInput, 0)
	)

	if len(versionID) > 0 {
		f += "?versionID=" + versionID
	}

	if h, e = o.getHeadObject(bucket, object, versionID); e != nil {
		return nil, e
	} else if h == nil || h.ETag == nil || len(*h.ETag) < 1 || h.ContentLength == nil || *h.ContentLength < 1 {
		return nil, ErrInvalidResponse
	} else {
		s = *h.ContentLength
	}

	// for current size cursor less than total size of object
	for i < s {

		// end of part if current size cursor + size of a part
		j := i + p

		// if the end size of the part is greater than total size of object, then end part will be total size of object
		if j > s {
			j = s
		}

		// append the part range of size : from size to end size
		r = append(r, o.cfg.getUploadPartCopyInput(f, fmt.Sprintf("%d-%d", i, j)))

		// moving cursor of size to end size + 1
		i = j + 1
	}

	return r, nil
}
