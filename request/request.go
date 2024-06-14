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

func (r *request) makeRequest(ctx context.Context, u *url.URL, mtd string, body BodyRetryer, head http.Header, params url.Values) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, mtd, u.String(), body.Retry())
	} else {
		req, err = http.NewRequestWithContext(ctx, mtd, u.String(), nil)
	}

	if err != nil {
		return nil, ErrorCreateRequest.Error(err)
	}

	req.Header = head

	if len(params) > 0 {
		q := req.URL.Query()
		for k := range params {
			q.Add(k, params.Get(k))
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func (r *request) checkResponse(rsp *http.Response, validStatus ...int) (*bytes.Buffer, error) {
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

	if !r.isValidCode(validStatus, rsp.StatusCode) {
		return b, ErrorResponseStatus.Error(nil)
	}

	return b, nil
}

func (r *request) isValidCode(listValid []int, statusCode int) bool {
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

func (r *request) isValidContents(contains []string, buf *bytes.Buffer) bool {
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
	r.mux.Lock()
	defer r.mux.Unlock()

	if r.mth == "" || r.uri == nil || r.uri.String() == "" {
		return nil, ErrorParamInvalid.Error(nil)
	}

	var (
		e   error
		req *http.Request
		rer *requestError
		rsp *http.Response
		err error
	)

	r.newError()
	rer = r.getError()

	req, err = r.makeRequest(r.context(), r.uri, r.mth, r.bdr, r.httpHeader(), r.prm)

	if err != nil {
		rer.err = err
		r.setError(rer)
		return nil, ErrorCreateRequest.Error(err)
	}

	rsp, e = r.client().Do(req)

	if e != nil {
		rer.err = e
		r.setError(rer)
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
		rer *requestError
	)

	if rsp, err = r.Do(); err != nil {
		return err
	} else if rsp == nil {
		return ErrorResponseInvalid.Error(nil)
	} else {
		rer = r.getError()
		rer.code = rsp.StatusCode
		rer.status = rsp.Status
	}

	b, err = r.checkResponse(rsp, validStatus...)
	rer.bufBody = b

	if er := liberr.Get(err); er != nil && er.HasCode(ErrorResponseStatus) {
		rer.statusErr = true
	} else if err != nil {
		rer.err = err
		r.setError(rer)
		return err
	}

	if b.Len() > 0 {
		if e = json.Unmarshal(b.Bytes(), model); e != nil {
			rer.bodyErr = true
			rer.err = e
			r.setError(rer)
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
