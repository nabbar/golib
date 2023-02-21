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
	"io"
	"net/http"
	"net/url"
	"sync"

	liblog "github.com/nabbar/golib/logger"

	montps "github.com/nabbar/golib/monitor/types"

	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libver "github.com/nabbar/golib/version"
)

type Error interface {
	StatusCode() int
	Status() string
	Body() *bytes.Buffer
	Error() error

	IsError() bool
	IsStatusError() bool
	IsBodyError() bool

	ParseBody(i interface{}) bool
}

type Url interface {
	SetEndpoint(u string) error
	GetEndpoint() string

	SetPath(raw bool, path string)
	AddPath(raw bool, path ...string)

	SetMethod(mtd string)
	GetMethod() string

	CleanParams()
	DelParams(key string)
	SetParams(key, val string)
	AddParams(key, val string)

	GetFullUrl() *url.URL
	SetFullUrl(u *url.URL)
}

type Authorization interface {
	AuthBearer(token string)
	AuthBasic(user, pass string)
}

type Header interface {
	SetHeader(key, value string)
	AddHeader(key, value string)
	DelHeader(key string)

	CleanHeader()
	ContentType(mime string)
}

type Body interface {
	BodyJson(body interface{}) error
	BodyReader(body io.Reader, contentType string)
}

type Request interface {
	Url
	Authorization
	Header
	Body

	Clone() (Request, error)
	New() (Request, error)

	GetOption() *Options
	SetOption(opt *Options) error
	RegisterHTTPClient(fct libtls.FctHttpClient)
	RegisterDefaultLogger(fct liblog.FuncLog)
	RegisterContext(fct libctx.FuncContext)

	Error() Error
	IsError() bool

	Do() (*http.Response, liberr.Error)
	DoParse(model interface{}, validStatus ...int) liberr.Error

	Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error)
}

func New(ctx libctx.FuncContext, opt *Options) (Request, error) {
	r := &request{
		s: sync.Mutex{},
		o: nil,
		x: ctx,
		f: nil,
		u: nil,
		h: make(url.Values),
		p: make(url.Values),
		b: bytes.NewBuffer(make([]byte, 0)),
		m: http.MethodGet,
		e: nil,
	}

	if e := r.SetOption(opt); e != nil {
		return nil, e
	} else {
		return r, nil
	}
}
