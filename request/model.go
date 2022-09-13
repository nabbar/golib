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

package request

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libsts "github.com/nabbar/golib/status"
)

const (
	_ContentType         = "Content-Type"
	_Authorization       = "Authorization"
	_AuthorizationBearer = "Bearer"
	_AuthorizationBasic  = "Basic"
)

type request struct {
	s sync.Mutex

	o *atomic.Value      // Options
	x libcfg.FuncContext // Context function
	f FctHttpClient      // Http client func
	u *url.URL           // endpoint url
	h url.Values         // header values
	p url.Values         // parameters values
	b io.Reader          // body io reader
	m string             // method
	i libsts.FctInfo     // Status Info func
	e *requestError      // Error pointer
}

func (r *request) _StatusInfo() (name string, release string, build string) {
	edp := r.GetFullUrl().Hostname()

	r.s.Lock()
	defer r.s.Unlock()

	if r.i != nil {
		name, release, build = r.i()
	}

	if name == "" {
		name = fmt.Sprintf("%s", edp)
	}

	if release == "" {
		release = strings.TrimLeft(runtime.Version(), "go")
		release = strings.TrimLeft(release, "Go")
		release = strings.TrimLeft(release, "GO")
	}

	return name, release, build
}

func (r *request) _StatusHealth() error {
	opts := r.GetOption()
	ednp := r.GetFullUrl()

	r.s.Lock()
	defer r.s.Unlock()

	head := make(url.Values, 0)
	if v := r.h.Get(_Authorization); v != "" {
		head.Set(_Authorization, v)
	}

	if !opts.Health.Enable {
		return nil
	}

	if opts.Health.Endpoint != "" {
		if u, e := url.Parse(opts.Health.Endpoint); e == nil {
			ednp = u
		}
	}

	if opts.Health.Auth.Basic.Enable {
		head.Set(_Authorization, _AuthorizationBasic+" "+base64.StdEncoding.EncodeToString([]byte(opts.Health.Auth.Basic.Username+":"+opts.Health.Auth.Basic.Password)))
	} else if opts.Health.Auth.Bearer.Enable {
		head.Set(_Authorization, _AuthorizationBearer+" "+opts.Health.Auth.Bearer.Token)
	}

	var (
		e error

		err liberr.Error
		buf *bytes.Buffer
		req *http.Request
		rsp *http.Response
	)

	req, err = r._MakeRequest(ednp, http.MethodGet, nil, head, nil)

	if err != nil {
		return err
	}

	rsp, e = r._GetClient().Do(req)

	if e != nil {
		return ErrorSendRequest.ErrorParent(e)
	}

	if buf, err = r._CheckResponse(rsp); err != nil {
		return err
	}

	if len(opts.Health.Result.ValidHTTPCode) > 0 {
		if !r._IsValidCode(opts.Health.Result.ValidHTTPCode, rsp.StatusCode) {
			return ErrorResponseStatus.ErrorParent(fmt.Errorf("status: %s", rsp.Status))
		}
	} else if len(opts.Health.Result.InvalidHTTPCode) > 0 {
		if r._IsValidCode(opts.Health.Result.InvalidHTTPCode, rsp.StatusCode) {
			return ErrorResponseStatus.ErrorParent(fmt.Errorf("status: %s", rsp.Status))
		}
	}

	if len(opts.Health.Result.Contain) > 0 {
		if !r._IsValidContents(opts.Health.Result.Contain, buf) {
			return ErrorResponseContainsNotFound.Error(nil)
		}
	} else if len(opts.Health.Result.NotContain) > 0 {
		if r._IsValidContents(opts.Health.Result.NotContain, buf) {
			return ErrorResponseNotContainsFound.Error(nil)
		}
	}

	return nil
}

func (r *request) _GetContext() context.Context {
	if r.x != nil {
		if x := r.x(); x != nil {
			return x
		}
	}

	return context.Background()
}

func (r *request) _GetOption() *Options {
	if r.o == nil {
		return nil
	} else if i := r.o.Load(); i == nil {
		return nil
	} else if o, ok := i.(*Options); !ok {
		return nil
	} else {
		return o
	}
}

func (r *request) _GetDefaultTLS() libtls.TLSConfig {
	if cfg := r._GetOption(); cfg != nil {
		return cfg._GetDefaultTLS()
	}

	return nil
}

func (r *request) _GetClient() *http.Client {
	var h string

	if r.u != nil {
		h = r.u.Hostname()
	}

	if r.f != nil {
		if c := r.f(r._GetDefaultTLS(), h); c != nil {
			return c
		}
	}

	if cfg := r._GetOption(); cfg != nil {
		return cfg.GetClientHTTP(h)
	}

	return &http.Client{}
}

func (r *request) _MakeRequest(u *url.URL, mtd string, body io.Reader, head url.Values, params url.Values) (*http.Request, liberr.Error) {
	var (
		req *http.Request
		err error
	)

	req, err = http.NewRequestWithContext(r._GetContext(), mtd, u.String(), body)

	if err != nil {
		return nil, ErrorCreateRequest.ErrorParent(err)
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

func (r *request) _CheckResponse(rsp *http.Response, validStatus ...int) (*bytes.Buffer, liberr.Error) {
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
			return b, ErrorResponseLoadBody.ErrorParent(e)
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

func (r *request) Clone() (Request, error) {
	if n, e := r.New(); e != nil {
		return nil, e
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		n.CleanHeader()
		for k := range r.h {
			n.SetHeader(k, r.h.Get(k))
		}

		n.CleanParams()
		for k := range r.p {
			n.SetParams(k, r.p.Get(k))
		}

		return n, nil
	}
}

func (r *request) New() (Request, error) {
	r.s.Lock()
	defer r.s.Unlock()

	var (
		n *request
		c = r._GetOption()
	)

	if c == nil {
		c = &Options{}
	}

	if i, e := New(r.x, r.f, *c); e != nil {
		return nil, e
	} else {
		n = i.(*request)
	}

	if r.u != nil {
		n.u = &url.URL{
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
		}
	}

	return n, nil
}

func (r *request) GetOption() *Options {
	r.s.Lock()
	defer r.s.Unlock()

	return r._GetOption()
}

func (r *request) SetOption(opt *Options) error {
	if e := r.SetEndpoint(opt.Endpoint); e != nil {
		return e
	}

	if opt.Auth.Basic.Enable {
		r.AuthBasic(opt.Auth.Basic.Username, opt.Auth.Basic.Password)
	} else if opt.Auth.Bearer.Enable {
		r.AuthBearer(opt.Auth.Bearer.Token)
	}

	r.s.Lock()
	defer r.s.Unlock()

	if r.o == nil {
		r.o = new(atomic.Value)
	}

	r.o.Store(opt)
	return nil
}

func (r *request) SetClient(fct FctHttpClient) {
	r.s.Lock()
	defer r.s.Unlock()

	r.f = fct
}

func (r *request) SetContext(fct libcfg.FuncContext) {
	r.s.Lock()
	defer r.s.Unlock()

	r.x = fct
}

func (r *request) SetEndpoint(u string) error {
	if uri, err := url.Parse(u); err != nil {
		return err
	} else {
		r.s.Lock()
		defer r.s.Unlock()

		r.u = uri
		return nil
	}
}

func (r *request) GetEndpoint() string {
	r.s.Lock()
	defer r.s.Unlock()

	return r.u.String()
}

func (r *request) SetPath(raw bool, path string) {
	r.s.Lock()
	defer r.s.Unlock()

	if raw {
		r.u.RawPath = path
	} else {
		r.u.Path = path
	}
}

func (r *request) AddPath(raw bool, path ...string) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.u == nil {
		return
	}

	for i := range path {
		if raw && strings.HasSuffix(r.u.RawPath, "/") && strings.HasPrefix(path[i], "/") {
			path[i] = strings.TrimPrefix(path[i], "/")
		} else if !raw && strings.HasSuffix(r.u.Path, "/") && strings.HasPrefix(path[i], "/") {
			path[i] = strings.TrimPrefix(path[i], "/")
		}

		if strings.HasSuffix(path[i], "/") {
			path[i] = strings.TrimSuffix(path[i], "/")
		}

		if raw {
			r.u.RawPath = filepath.Join(r.u.RawPath, path[i])
		} else {
			r.u.Path = filepath.Join(r.u.Path, path[i])
		}
	}
}

func (r *request) SetMethod(method string) {
	r.s.Lock()
	defer r.s.Unlock()

	switch strings.ToUpper(method) {
	case http.MethodGet:
		r.m = http.MethodGet
	case http.MethodHead:
		r.m = http.MethodHead
	case http.MethodPost:
		r.m = http.MethodPost
	case http.MethodPut:
		r.m = http.MethodPut
	case http.MethodPatch:
		r.m = http.MethodPatch
	case http.MethodDelete:
		r.m = http.MethodDelete
	case http.MethodConnect:
		r.m = http.MethodConnect
	case http.MethodOptions:
		r.m = http.MethodOptions
	case http.MethodTrace:
		r.m = http.MethodTrace
	default:
		r.m = strings.ToUpper(method)
	}

	if r.m == "" {
		r.m = http.MethodGet
	}
}

func (r *request) GetMethod() string {
	r.s.Lock()
	defer r.s.Unlock()

	return r.m
}

func (r *request) CleanParams() {
	r.s.Lock()
	defer r.s.Unlock()

	r.p = make(url.Values)
}

func (r *request) DelParams(key string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.p.Del(key)
}

func (r *request) SetParams(key, val string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.p) < 1 {
		r.p = make(url.Values)
	}

	r.p.Set(key, val)
}

func (r *request) AddParams(key, val string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.p) < 1 {
		r.p = make(url.Values)
	}

	r.p.Set(key, val)
}

func (r *request) GetFullUrl() *url.URL {
	r.s.Lock()
	defer r.s.Unlock()

	return r.u
}

func (r *request) SetFullUrl(u *url.URL) {
	r.s.Lock()
	defer r.s.Unlock()

	r.u = u
}

func (r *request) AuthBearer(token string) {
	r.SetHeader(_Authorization, _AuthorizationBearer+" "+token)
}

func (r *request) AuthBasic(user, pass string) {
	r.SetHeader(_Authorization, _AuthorizationBasic+" "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
}

func (r *request) ContentType(content string) {
	r.SetHeader(_ContentType, content)
}

func (r *request) CleanHeader() {
	r.s.Lock()
	defer r.s.Unlock()

	r.h = make(url.Values)
}

func (r *request) DelHeader(key string) {
	r.s.Lock()
	defer r.s.Unlock()

	r.h.Del(key)
}

func (r *request) SetHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Set(key, value)
}

func (r *request) AddHeader(key, value string) {
	r.s.Lock()
	defer r.s.Unlock()

	if len(r.h) < 1 {
		r.h = make(url.Values)
	}

	r.h.Add(key, value)
}

func (r *request) BodyJson(body interface{}) error {
	if p, e := json.Marshal(body); e != nil {
		return e
	} else {
		r._BodyReader(bytes.NewBuffer(p))
	}

	r.ContentType("application/json")
	return nil
}

func (r *request) BodyReader(body io.Reader, contentType string) {
	r._BodyReader(body)

	if contentType != "" {
		r.ContentType(contentType)
	}
}

func (r *request) _BodyReader(body io.Reader) {
	r.s.Lock()
	defer r.s.Unlock()

	r.b = body
}

func (r *request) Error() RequestError {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e
}

func (r *request) IsError() bool {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e != nil && r.e.IsError()
}

func (r *request) Do() (*http.Response, liberr.Error) {
	r.s.Lock()
	defer r.s.Unlock()

	if r.m == "" || r.u == nil || r.u.String() == "" {
		return nil, ErrorParamInvalid.Error(nil)
	}

	var (
		e   error
		req *http.Request
		rsp *http.Response
		err liberr.Error
	)

	r.e = &requestError{
		c:  0,
		s:  "",
		se: false,
		b:  bytes.NewBuffer(make([]byte, 0)),
		be: false,
		e:  nil,
	}

	req, err = r._MakeRequest(r.u, r.m, r.b, r.h, r.p)
	if err != nil {
		r.e.e = err
		return nil, err
	}

	rsp, e = r._GetClient().Do(req)

	if e != nil {
		r.e.e = e
		return nil, ErrorSendRequest.ErrorParent(e)
	}

	return rsp, nil
}

func (r *request) DoParse(model interface{}, validStatus ...int) liberr.Error {
	var (
		e error
		b = bytes.NewBuffer(make([]byte, 0))

		err liberr.Error
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

	if err != nil && err.HasCodeError(ErrorResponseStatus) {
		r.e.se = true
	} else if err != nil {
		r.e.e = err
		return err
	}

	if b.Len() > 0 {
		if e = json.Unmarshal(b.Bytes(), model); e != nil {
			r.e.be = true
			r.e.e = e
			return ErrorResponseUnmarshall.ErrorParent(e)
		}
	}

	return nil
}

func (r *request) StatusRegister(sts libsts.RouteStatus, prefix string) {
	opts := r.GetOption()

	r.s.Lock()
	defer r.s.Unlock()

	if len(prefix) > 0 {
		prefix = fmt.Sprintf("%s %s", prefix, r.u.Hostname())
	} else {
		prefix = fmt.Sprintf("%s %s", "HTTP Request", r.u.Hostname())
	}

	opts.Health.Status.RegisterStatus(sts, prefix, r._StatusInfo, r._StatusHealth)
}

func (r *request) StatusRegisterInfo(fct libsts.FctInfo) {
	r.s.Lock()
	defer r.s.Unlock()

	r.i = fct
}
