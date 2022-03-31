/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package httpcli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	libtls "github.com/nabbar/golib/certificates"

	liberr "github.com/nabbar/golib/errors"
)

type request struct {
	s sync.Mutex

	f FctHttpClient
	u *url.URL
	h url.Values
	p url.Values
	b io.Reader
	m string
	e *requestError
}

func (r *request) Clone() Request {
	n := &request{
		s: sync.Mutex{},
		f: r.f,
		u: &url.URL{
			Scheme:      r.u.Scheme,
			Opaque:      r.u.Opaque,
			User:        r.u.User,
			Host:        r.u.Host,
			Path:        r.u.Path,
			RawPath:     r.u.RawPath,
			ForceQuery:  r.u.ForceQuery,
			RawQuery:    r.u.RawQuery,
			Fragment:    r.u.Fragment,
			RawFragment: r.u.RawFragment,
		},
		h: make(url.Values),
		p: make(url.Values),
		b: r.b,
		m: http.MethodGet,
		e: nil,
	}

	for k, v := range r.h {
		n.h[k] = v
	}

	for k, v := range r.p {
		n.p[k] = v
	}

	return n
}

func (r *request) New() Request {
	return &request{
		s: sync.Mutex{},
		f: r.f,
		u: nil,
		h: make(url.Values),
		p: make(url.Values),
		b: bytes.NewBuffer(make([]byte, 0)),
		m: http.MethodGet,
		e: nil,
	}
}

func (r *request) _GetClient() *http.Client {
	if r.f != nil {
		if c := r.f(); c != nil {
			return c
		}
	}

	return &http.Client{}
}

func (r *request) SetClient(fct FctHttpClient) {
	r.s.Lock()
	defer r.s.Unlock()
	r.f = fct
}

func (r *request) UseClientPackage(ip string, tls libtls.TLSConfig, http2Tr bool, GlobalTimeout time.Duration) {
	r.s.Lock()
	defer r.s.Unlock()

	if tls == nil {
		tls = libtls.Default
	}

	if len(ip) > 0 {
		r.f = func() *http.Client {
			if h, e := GetClientTlsForceIp(NetworkTCP, ip, r.u.Hostname(), tls, http2Tr, GlobalTimeout); e == nil {
				return h
			} else if h, e = GetClientTimeout(r.u.Hostname(), false, GlobalTimeout); e == nil {
				return h
			}

			return &http.Client{}
		}
	} else {
		r.f = func() *http.Client {
			if h, e := GetClientTls(r.u.Hostname(), tls, http2Tr, GlobalTimeout); e == nil {
				return h
			} else if h, e = GetClientTls(r.u.Hostname(), tls, false, GlobalTimeout); e == nil {
				return h
			}

			return &http.Client{}
		}
	}
}

func (r *request) Endpoint(uri string) error {
	if u, e := url.Parse(uri); e != nil {
		return e
	} else {
		r.s.Lock()
		defer r.s.Unlock()
		r.u = u
	}

	return nil
}

func (r *request) SetUrl(u *url.URL) {
	r.s.Lock()
	defer r.s.Unlock()
	r.u = u
}

func (r *request) GetUrl() *url.URL {
	r.s.Lock()
	defer r.s.Unlock()
	return r.u
}

func (r *request) AddPath(path string) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.u == nil {
		return
	}

	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}

	if strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	r.u.Path = filepath.Join(r.u.Path, path)
}

func (r *request) AddParams(key, val string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.p) < 1 {
		r.p = make(url.Values)
	}

	r.p.Set(key, val)
}

func (r *request) AuthBearer(token string) {
	r.Header("Authorization", "Bearer "+token)
}

func (r *request) AuthBasic(user, pass string) {
	r.Header("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
}

func (r *request) ContentType(content string) {
	r.Header("Content-Type", content)
}

func (r *request) Header(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Set(key, value)
}

func (r *request) Method(mtd string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.m = mtd
}

func (r *request) RequestJson(body interface{}) error {
	if p, e := json.Marshal(body); e != nil {
		return e
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		r.b = bytes.NewBuffer(p)
	}

	r.ContentType("application/json")
	return nil
}

func (r *request) RequestReader(body io.Reader) {
	r.s.Lock()
	defer r.s.Unlock()

	r.b = body
}

func (r *request) Error() RequestError {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e
}

func (r *request) Do(ctx context.Context) (*http.Response, liberr.Error) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.m == "" || r.u == nil || r.u.String() == "" {
		return nil, ErrorParamsInvalid.Error(nil)
	}

	var (
		e   error
		req *http.Request
		rsp *http.Response
		err liberr.Error
	)

	r.e = nil

	req, err = r._MakeRequest(ctx)
	if err != nil {
		return nil, err
	}

	rsp, e = r._GetClient().Do(req)

	if e != nil {
		r.e = &requestError{
			c: 0,
			s: "",
			b: nil,
			e: e,
		}
		return nil, ErrorSendRequest.ErrorParent(e)
	}

	return rsp, nil
}

func (r *request) _MakeRequest(ctx context.Context) (*http.Request, liberr.Error) {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequestWithContext(ctx, r.m, r.u.String(), r.b)

	if err != nil {
		return nil, ErrorCreateRequest.ErrorParent(err)
	}

	if len(r.h) > 0 {
		for k := range r.h {
			req.Header.Set(k, r.h.Get(k))
		}
	}

	q := req.URL.Query()
	for k := range r.p {
		q.Add(k, r.p.Get(k))
	}
	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (r *request) DoParse(ctx context.Context, model interface{}, validStatus ...int) liberr.Error {
	var (
		e error
		b = bytes.NewBuffer(make([]byte, 0))

		err liberr.Error
		rsp *http.Response
	)

	if rsp, err = r.Do(ctx); err != nil {
		return err
	} else if rsp == nil {
		return ErrorResponseInvalid.Error(nil)
	}

	defer func() {
		if !rsp.Close && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	if rsp.Body != nil {
		if _, e = io.Copy(b, rsp.Body); e != nil {
			r.e = &requestError{
				c: rsp.StatusCode,
				s: rsp.Status,
				b: b,
				e: e,
			}
			return ErrorResponseLoadBody.ErrorParent(e)
		}
	}

	if !r._IsValidCode(validStatus, rsp.StatusCode) {
		r.e = &requestError{
			c: rsp.StatusCode,
			s: rsp.Status,
			b: b,
			e: nil,
		}
		return ErrorResponseStatus.Error(nil)
	}

	if e = json.Unmarshal(b.Bytes(), model); e != nil {
		r.e = &requestError{
			c: rsp.StatusCode,
			s: rsp.Status,
			b: b,
			e: e,
		}
		return ErrorResponseUnmarshall.ErrorParent(e)
	}

	return nil
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
