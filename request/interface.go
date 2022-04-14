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
	"io"
	"net/http"
	"net/url"
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libsts "github.com/nabbar/golib/status"
)

type FctHttpClient func(def libtls.TLSConfig, servername string) *http.Client
type FctTLSDefault func() libtls.TLSConfig

type RequestError interface {
	StatusCode() int
	Status() string
	Body() *bytes.Buffer
	Error() error
}

type Request interface {
	Clone() (Request, error)
	New() (Request, error)

	GetOption() *Options
	SetOption(opt *Options) error
	SetClient(fct FctHttpClient)
	SetContext(fct libcfg.FuncContext)

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

	AuthBearer(token string)
	AuthBasic(user, pass string)
	ContentType(content string)

	CleanHeader()
	DelHeader(key string)
	SetHeader(key, value string)
	AddHeader(key, value string)

	BodyJson(body interface{}) error
	BodyReader(body io.Reader, contentType string)

	Error() RequestError
	IsError() bool

	Do() (*http.Response, liberr.Error)
	DoParse(model interface{}, validStatus ...int) liberr.Error

	StatusRegister(sts libsts.RouteStatus, prefix string)
	StatusRegisterInfo(fct libsts.FctInfo)
}

func New(ctx libcfg.FuncContext, cli FctHttpClient, opt Options) (Request, error) {
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

	r.SetClient(cli)

	if e := r.SetOption(&opt); e != nil {
		return nil, e
	} else {
		return r, nil
	}
}
