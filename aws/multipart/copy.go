/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package multipart

import (
	"fmt"
	"path"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (m *mpu) Copy(fromBucket, fromObject, fromVersionId string) error {
	if m == nil {
		return ErrInvalidInstance
	}

	var (
		err error
		cli *sdksss.Client
		src string
		res *sdksss.UploadPartCopyOutput
		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
		mid = m.getMultipartID()
	)

	if cli = m.getClient(); cli == nil {
		return ErrInvalidClient
	}

	src = path.Join(fromBucket, fromObject)
	if len(fromVersionId) > 0 {
		src += "?versionID=" + fromVersionId
	}

	for _, p := range m.getCopyPart(fromBucket, fromObject, fromVersionId) {
		res, err = cli.UploadPartCopy(ctx, &sdksss.UploadPartCopyInput{
			Bucket:          sdkaws.String(bck),
			CopySource:      sdkaws.String(src),
			Key:             sdkaws.String(obj),
			PartNumber:      sdkaws.Int32(m.Counter() + 1),
			UploadId:        sdkaws.String(mid),
			CopySourceRange: sdkaws.String("bytes=" + p),
			RequestPayer:    sdktyp.RequestPayerRequester,
		})

		if err != nil {
			m.callFuncOnPushPart("", err)
			return err
		} else if res == nil || res.CopyPartResult == nil || res.CopyPartResult.ETag == nil || len(*res.CopyPartResult.ETag) < 1 {
			m.callFuncOnPushPart("", ErrInvalidResponse)
			return ErrInvalidResponse
		} else {
			t := *res.CopyPartResult.ETag
			m.callFuncOnPushPart(t, nil)
			m.RegisterPart(t)
		}
	}

	return nil
}

func (m *mpu) getCopyPart(fromBucket, fromObject, fromVersionId string) []string {
	if m == nil {
		return make([]string, 0)
	}

	var (
		err error
		res = make([]string, 0)
		cli *sdksss.Client
		hdo *sdksss.HeadObjectOutput
		ctx = m.getContext()
		prt = m.s.Int64() - 1
	)

	if cli = m.getClient(); cli == nil {
		return res
	}

	inp := &sdksss.HeadObjectInput{
		Bucket:       sdkaws.String(fromBucket),
		Key:          sdkaws.String(fromObject),
		RequestPayer: sdktyp.RequestPayerRequester,
	}

	if len(fromVersionId) > 0 {
		inp.VersionId = sdkaws.String(fromVersionId)
	}

	hdo, err = cli.HeadObject(ctx, inp)

	if err != nil {
		return res
	} else if hdo == nil || hdo.ETag == nil || len(*hdo.ETag) < 1 {
		return res
	} else if s := hdo.ContentLength; s == nil {
		return res
	} else if size := *s; size < 1 {
		return res
	} else {
		var i int64 = 0
		for i < size {
			j := i + prt
			if j > size {
				j = size
			}

			res = append(res, fmt.Sprintf("%d-%d", i, j))
			i = j + 1
		}
	}

	return res
}
