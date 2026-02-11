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
	/* #nosec */
	// #nosec nolint
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdktyp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

func (m *mpu) getPartList() []sdktyp.CompletedPart {
	if m == nil {
		return make([]sdktyp.CompletedPart, 0)
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if len(m.l) < 1 {
		return make([]sdktyp.CompletedPart, 0)
	}

	return m.l
}

func (m *mpu) Counter() int32 {
	if m == nil {
		return 0
	}

	m.m.RLock()
	defer m.m.RUnlock()

	return m.n
}

func (m *mpu) CounterLeft() int32 {
	if m == nil {
		return 0
	}

	m.m.RLock()
	defer m.m.RUnlock()

	if m.n >= MaxNumberPart {
		return 0
	}

	return MaxNumberPart - m.n
}

func (m *mpu) RegisterPart(etag string) {
	if m == nil {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	if len(m.l) < 1 {
		m.l = make([]sdktyp.CompletedPart, 0)
	}

	m.n++
	m.l = append(m.l, sdktyp.CompletedPart{
		ETag:       sdkaws.String(strings.Replace(etag, "\"", "", -1)), // nolint
		PartNumber: sdkaws.Int32(m.n),
	})
}

func (m *mpu) AddPart(r io.Reader) (n int64, e error) {
	if m == nil {
		return 0, ErrInvalidInstance
	}

	var (
		cli *sdksss.Client
		res *sdksss.UploadPartOutput
		tmp libfpg.Progress
		ctx = m.getContext()
		obj = m.getObject()
		bck = m.getBucket()
		mid = m.getMultipartID()
		hss string

		/* #nosec */
		// #nosec nolint
		hsh = md5.New()
	)

	defer func() {
		if tmp != nil {
			_ = tmp.CloseDelete()
		}
	}()

	if cli = m.getClient(); cli == nil {
		return 0, ErrInvalidClient
	} else if tmp, e = libfpg.Temp(""); e != nil {
		return 0, e
	} else if tmp == nil {
		return 0, ErrInvalidTMPFile
	}

	if n, e = io.Copy(tmp, r); e != nil && !errors.Is(e, io.EOF) {
		return n, e
	} else if n < 1 {
		return n, e
	} else if _, e = tmp.Seek(0, io.SeekStart); e != nil {
		return 0, e
	} else if _, e = tmp.WriteTo(hsh); e != nil && !errors.Is(e, io.EOF) {
		return 0, e
	} else if _, e = tmp.Seek(0, io.SeekStart); e != nil {
		return 0, e
	}

	hss = base64.StdEncoding.EncodeToString(hsh.Sum(nil))

	res, e = cli.UploadPart(ctx, &sdksss.UploadPartInput{
		Bucket:        sdkaws.String(bck),
		Key:           sdkaws.String(obj),
		UploadId:      sdkaws.String(mid),
		PartNumber:    sdkaws.Int32(m.Counter() + 1),
		ContentLength: sdkaws.Int64(n),
		Body:          tmp,
		RequestPayer:  sdktyp.RequestPayerRequester,
		ContentMD5:    sdkaws.String(hss),
	})

	if e != nil {
		m.callFuncOnPushPart("", e)
		return 0, e
	} else if res == nil || res.ETag == nil || len(*res.ETag) < 1 {
		m.callFuncOnPushPart("", ErrInvalidResponse)
		return 0, ErrInvalidResponse
	} else {
		t := *res.ETag
		m.callFuncOnPushPart(t, nil)
		m.RegisterPart(t)
	}

	return n, nil
}

func (m *mpu) AddToPart(p []byte) (n int, e error) {
	var (
		tmp libfpg.Progress
	)

	if tmp, e = m.getWorkingFile(); e != nil {
		return 0, e
	} else if tmp == nil {
		return 0, ErrInvalidTMPFile
	}

	for len(p) > 0 {
		var (
			r   []byte
			i   int
			s   int64
			siz = m.getPartSize().Int64()
		)

		if _, e = tmp.Seek(0, io.SeekStart); e != nil {
			return n, e
		} else if s, e = tmp.SizeEOF(); e != nil {
			return n, e
		} else if _, e = tmp.Seek(0, io.SeekEnd); e != nil {
			return n, e
		} else if s > 0 && s >= siz {
			if e = m.CheckSend(false, false); e != nil {
				return n, e
			}
			continue
		} else if s > 0 && s < siz {
			siz -= s
		}

		if int64(len(p)) > siz {
			r = p[:siz]
			p = p[siz:]
		} else {
			r = p
			p = nil
		}

		if i, e = tmp.Write(r); e != nil {
			return n, e
		} else if i != len(r) {
			return n, fmt.Errorf("write a wrong number of byte")
		} else if e = m.CheckSend(false, false); e != nil {
			return n, e
		} else {
			n += len(r)
		}
	}

	return n, nil
}

func (m *mpu) SendPart() error {
	return m.CheckSend(true, false)
}

func (m *mpu) CurrentSizePart() int64 {
	var (
		e   error
		s   int64
		tmp libfpg.Progress
	)

	if tmp, e = m.getWorkingFile(); e != nil {
		return 0
	} else if tmp == nil {
		return 0
	} else if _, e = tmp.Seek(0, io.SeekStart); e != nil {
		return 0
	} else {
		s, _ = tmp.SizeEOF()
		_, _ = tmp.Seek(0, io.SeekEnd)
		return s
	}
}

func (m *mpu) CheckSend(force, close bool) error {
	var (
		err error
		siz int64
		prt = m.getPartSize()
		tmp libfpg.Progress
	)

	if tmp, err = m.getWorkingFile(); err != nil {
		return err
	} else if tmp == nil {
		return ErrInvalidTMPFile
	} else if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		return err
	} else if siz, err = tmp.SizeEOF(); err != nil {
		return err
	} else if siz < prt.Int64() && !force {
		return nil
	} else if siz == 0 {
		return nil
	} else if siz > int64(MaxObjectSize) {
		return ErrWorkingPartFileExceedSize
	} else if close && m.Counter() < 1 && siz < DefaultPartSize.Int64() {
		return nil
	} else if _, err = m.sendPart(siz, tmp); err != nil {
		return err
	} else if err = tmp.Truncate(0); err != nil {
		return err
	} else if err = tmp.Sync(); err != nil {
		return err
	} else {
		return nil
	}
}

func (m *mpu) sendPart(siz int64, body io.Reader) (int64, error) {
	var (
		err error
		prt = m.getPartSize()
	)

	if prt, err = GetOptimalPartSize(libsiz.SizeFromInt64(siz), prt); err != nil {
		return 0, err
	} else if prt != m.getPartSize() {
		old := m.getPartSize()
		m.setPartSize(prt)
		defer func() {
			m.setPartSize(old)
		}()
	}

	return m.AddPart(body)
}
