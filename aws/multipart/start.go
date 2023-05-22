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
	"mime"
	"path/filepath"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
)

func (m *mpu) IsStarted() bool {
	if cli := m.getClient(); cli == nil {
		return false
	} else if len(m.getMultipartID()) < 1 {
		return false
	}

	return true
}

func (m *mpu) getMimeType() string {
	if t := mime.TypeByExtension(filepath.Ext(m.getObject())); t == "" {
		return "application/octet-stream"
	} else {
		return t
	}
}

func (m *mpu) StartMPU() error {
	if m == nil {
		return ErrInvalidInstance
	}

	var (
		cli *sdksss.Client
		res *sdksss.CreateMultipartUploadOutput
		err error
		tpe = m.getMimeType()
		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
	)

	if cli = m.getClient(); cli == nil {
		return ErrInvalidClient
	}

	res, err = cli.CreateMultipartUpload(ctx, &sdksss.CreateMultipartUploadInput{
		Key:         sdkaws.String(obj),
		Bucket:      sdkaws.String(bck),
		ContentType: sdkaws.String(tpe),
	})

	if err != nil {
		return err
	} else if res == nil {
		return ErrInvalidResponse
	} else if res.UploadId == nil || len(*res.UploadId) < 1 {
		return ErrInvalidResponse
	}

	m.m.Lock()
	defer m.m.Unlock()
	m.i = *res.UploadId

	return nil
}
