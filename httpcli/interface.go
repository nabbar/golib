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
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	libtls "github.com/nabbar/golib/certificates"

	liberr "github.com/nabbar/golib/errors"
)

type FctHttpClient func() *http.Client

type RequestError interface {
	StatusCode() int
	Status() string
	Body() *bytes.Buffer
	Error() error
}

type Request interface {
	Clone() Request
	New() Request

	SetClient(fct FctHttpClient)
	UseClientPackage(ip string, tls libtls.TLSConfig, http2Tr bool, GlobalTimeout time.Duration)

	Endpoint(uri string) error
	SetUrl(u *url.URL)
	GetUrl() *url.URL
	AddPath(path string)
	AddParams(key, val string)

	AuthBearer(token string)
	AuthBasic(user, pass string)
	ContentType(content string)

	Header(key, value string)
	Method(mtd string)

	RequestJson(body interface{}) error
	RequestReader(body io.Reader)

	Error() RequestError

	Do(ctx context.Context) (*http.Response, liberr.Error)
	DoParse(ctx context.Context, model interface{}, validStatus ...int) liberr.Error
}

func New(fct FctHttpClient) Request {
	return &request{
		s: sync.Mutex{},
		f: fct,
		u: nil,
		h: make(url.Values),
		p: make(url.Values),
		b: bytes.NewBuffer(make([]byte, 0)),
		m: http.MethodGet,
		e: nil,
	}
}
