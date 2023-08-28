/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package request

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	liberr "github.com/nabbar/golib/errors"
)

func (r *request) _MakeRequest(ctx context.Context, u *url.URL, mtd string, body io.Reader, head url.Values, params url.Values) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequestWithContext(ctx, mtd, u.String(), body)

	if err != nil {
		return nil, ErrorCreateRequest.Error(err)
	}

	if len(head) > 0 {
		for k := range head {
			req.Header.Set(k, head.Get(k))
		}
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k := range params {
			q.Add(k, params.Get(k))
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func (r *request) _CheckResponse(rsp *http.Response, validStatus ...int) (*bytes.Buffer, error) {
	var (
		e error
		b = bytes.NewBuffer(make([]byte, 0))
	)

	defer func() {
		if rsp != nil && !rsp.Close && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp == nil {
		return b, ErrorResponseInvalid.Error(nil)
	}

	if rsp.Body != nil {
		if _, e = io.Copy(b, rsp.Body); e != nil {
			return b, ErrorResponseLoadBody.Error(e)
		}
	}

	if !r._IsValidCode(validStatus, rsp.StatusCode) {
		return b, ErrorResponseStatus.Error(nil)
	}

	return b, nil
}

func (r *request) _IsValidCode(listValid []int, statusCode int) bool {
	if len(listValid) < 1 {
		return true
	}

	for _, c := range listValid {
		if c == statusCode {
			return true
		}
	}

	return false
}

func (r *request) _IsValidContents(contains []string, buf *bytes.Buffer) bool {
	if len(contains) < 1 {
		return true
	} else if buf.Len() < 1 {
		return false
	}

	for _, c := range contains {
		if strings.Contains(buf.String(), c) {
			return true
		}
	}

	return false
}

func (r *request) Do() (*http.Response, error) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.m == "" || r.u == nil || r.u.String() == "" {
		return nil, ErrorParamInvalid.Error(nil)
	}

	var (
		e   error
		req *http.Request
		rsp *http.Response
		err error
	)

	r.e = &requestError{
		c:  0,
		s:  "",
		se: false,
		b:  bytes.NewBuffer(make([]byte, 0)),
		be: false,
		e:  nil,
	}

	req, err = r._MakeRequest(r.context(), r.u, r.m, r.b, r.h, r.p)
	if err != nil {
		r.e.e = err
		return nil, err
	}

	rsp, e = r.client().Do(req)

	if e != nil {
		r.e.e = e
		return nil, ErrorSendRequest.Error(e)
	}

	return rsp, nil
}

func (r *request) DoParse(model interface{}, validStatus ...int) error {
	var (
		e error
		b = bytes.NewBuffer(make([]byte, 0))

		err error
		rsp *http.Response
	)

	r.e = &requestError{
		c:  0,
		s:  "",
		se: false,
		b:  bytes.NewBuffer(make([]byte, 0)),
		be: false,
		e:  nil,
	}

	if rsp, err = r.Do(); err != nil {
		return err
	} else if rsp == nil {
		return ErrorResponseInvalid.Error(nil)
	} else {
		r.e.c = rsp.StatusCode
		r.e.s = rsp.Status
	}

	b, err = r._CheckResponse(rsp, validStatus...)
	r.e.b = b

	if er := liberr.Get(err); er != nil && er.HasCode(ErrorResponseStatus) {
		r.e.se = true
	} else if err != nil {
		r.e.e = err
		return err
	}

	if b.Len() > 0 {
		if e = json.Unmarshal(b.Bytes(), model); e != nil {
			r.e.be = true
			r.e.e = e
			return ErrorResponseUnmarshall.Error(e)
		}
	}

	return nil
}

func (r *request) DoParseRetry(retry int, model interface{}, validStatus ...int) error {
	var e error

	for i := 0; i < retry; i++ {
		if e = r.DoParse(model, validStatus...); e != nil {
			continue
		} else if r.IsError() {
			continue
		} else {
			return nil
		}
	}

	return e
}
