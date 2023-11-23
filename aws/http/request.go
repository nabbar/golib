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

package http

import (
	"bytes"
	"io"
	"net/http"
	"time"

	sdkcrd "github.com/aws/aws-sdk-go/aws/credentials"
	sdksv4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

func CopyReader(r io.Reader) (io.ReadSeekCloser, error) {
	var tmp = bytes.NewBuffer(make([]byte, 0))

	if _, err := io.Copy(tmp, r); err != nil {
		return nil, err
	} else {
		return &readerCloser{bytes.NewReader(tmp.Bytes())}, nil
	}
}

func NewReader(p []byte) io.ReadSeekCloser {
	return &readerCloser{bytes.NewReader(p)}
}

type readerCloser struct {
	io.ReadSeeker
}

func (r *readerCloser) Close() error {
	return nil
}

func Request(req *http.Request, cfg Config, service string) error {
	var (
		err error
		sig = sdksv4.NewSigner(sdkcrd.NewStaticCredentials(cfg.GetAccessKey(), cfg.GetSecretKey(), ""))
	)

	if req.Body == nil {
		_, err = sig.Sign(req, nil, service, cfg.GetRegion(), time.Now())
	} else if r, k := req.Body.(io.ReadSeekCloser); k {
		_, err = sig.Sign(req, r, service, cfg.GetRegion(), time.Now())
	} else if r, err = CopyReader(req.Body); err != nil {
		return err
	} else {
		req.Body = r
		_, err = sig.Sign(req, r, service, cfg.GetRegion(), time.Now())
	}

	return err
}
