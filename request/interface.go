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
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
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
	BodyReader(body io.Reader, contentType string) error
}

type BodyRetryer interface {
	Retry() io.Reader
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
	RegisterHTTPClient(cli libhtc.HttpClient)
	RegisterDefaultLogger(fct liblog.FuncLog)
	RegisterContext(fct libctx.FuncContext)

	Error() Error
	IsError() bool

	Do() (*http.Response, error)
	DoParse(model interface{}, validStatus ...int) error
	DoParseRetry(retry int, model interface{}, validStatus ...int) error

	Monitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error)
	HealthCheck(ctx context.Context) error
}

func New(ctx libctx.FuncContext, opt *Options, cli libhtc.HttpClient) (Request, error) {
	r := &request{
		mux: sync.Mutex{},
		opt: new(atomic.Value),
		ctx: new(atomic.Value),
		log: new(atomic.Value),
		uri: nil,
		hdr: sync.Map{},
		prm: make(url.Values),
		bdr: nil,
		mth: http.MethodGet,
		err: new(atomic.Value),
		cli: new(atomic.Value),
	}

	if cli != nil {
		r.cli.Store(cli)
	} else {
		r.cli.Store(libhtc.GetClient())
	}

	if e := r.SetOption(opt); e != nil {
		return nil, e
	} else {
		return r, nil
	}
}
