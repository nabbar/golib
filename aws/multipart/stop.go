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

package multipart

import (
	"io"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libfpg "github.com/nabbar/golib/file/progress"
)

func (m *mpu) StopMPU(abort bool) error {
	if m == nil {
		return ErrInvalidInstance
	}

	var (
		err error
		lst = m.getPartList()
	)

	defer func() {
		m.m.Lock()
		if m.w != nil {
			if i, e := m.w.Stat(); e == nil && i.Size() < 1 {
				_ = m.w.CloseDelete()
				m.w = nil
			}
		}
		m.m.Unlock()
	}()

	if !abort {
		if err = m.CheckSend(true, true); err != nil {
			return err
		}
	}

	if abort || len(lst) < 1 {
		err = m.abortMPU()
	} else {
		err = m.completeMPU()
	}

	if !abort && err == nil && len(lst) < 1 && m.CurrentSizePart() > 0 {
		err = m.SendObject()
	}

	m.callFuncOnComplete(abort, len(lst), m.getObject(), err)

	if err != nil {
		_ = m.closeWorkingFile()
	}

	m.cleanMPU()
	return nil
}

func (m *mpu) cleanMPU() {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	m.i = ""
	m.o = ""
	m.b = ""
	m.l = nil
	m.n = 0
}

func (m *mpu) SendObject() error {
	var (
		err error
		cli *sdksss.Client
		res *sdksss.PutObjectOutput
		tmp libfpg.Progress

		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
		tpe = m.getMimeType()
	)

	if cli = m.getClient(); cli == nil {
		return ErrInvalidClient
	} else if m.CurrentSizePart() < 1 {
		return nil
	} else if tmp, err = m.getWorkingFile(); err != nil {
		return err
	} else if tmp == nil {
		return ErrInvalidTMPFile
	} else if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	res, err = cli.PutObject(ctx, &sdksss.PutObjectInput{
		Bucket:      sdkaws.String(bck),
		Key:         sdkaws.String(obj),
		Body:        tmp,
		ContentType: sdkaws.String(tpe),
	})

	if err == nil {
		if res == nil {
			err = ErrInvalidResponse
		} else if res.ETag == nil || len(*res.ETag) < 1 {
			err = ErrInvalidResponse
		}
	}

	return err
}

func (m *mpu) abortMPU() error {
	var (
		cli *sdksss.Client
		err error
		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
		mid = m.getMultipartID()
		mod = &sdksss.AbortMultipartUploadInput{
			Bucket:   sdkaws.String(bck),
			Key:      sdkaws.String(obj),
			UploadId: sdkaws.String(mid),
		}
	)

	if cli = m.getClient(); cli == nil {
		return ErrInvalidClient
	} else if len(mid) < 1 {
		return nil
	} else if _, err = cli.AbortMultipartUpload(ctx, mod); err != nil {
		return err
	}

	return nil
}

func (m *mpu) completeMPU() error {
	var (
		cli *sdksss.Client
		res *sdksss.CompleteMultipartUploadOutput
		err error
		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
		mid = m.getMultipartID()
		lst = m.getPartList()
		mod = &sdksss.CompleteMultipartUploadInput{
			Bucket:   sdkaws.String(bck),
			Key:      sdkaws.String(obj),
			UploadId: sdkaws.String(mid),
			MultipartUpload: &sdktyp.CompletedMultipartUpload{
				Parts: lst,
			},
			RequestPayer: sdktyp.RequestPayerRequester,
		}
	)

	if cli = m.getClient(); cli == nil {
		return ErrInvalidClient
	} else if len(mid) < 1 {
		return ErrInvalidUploadID
	} else if res, err = cli.CompleteMultipartUpload(ctx, mod); err != nil {
		return err
	} else if res.Key == nil || len(*res.Key) < 1 {
		return ErrInvalidResponse
	}

	return nil
}
