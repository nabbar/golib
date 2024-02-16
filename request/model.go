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
	"context"
	"io"
	"net/url"
	"sync"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	liblog "github.com/nabbar/golib/logger"
)

const (
	_ContentType         = "Content-Type"
	__ContentLength      = "Content-Length"
	_Authorization       = "Authorization"
	_AuthorizationBearer = "Bearer"
	_AuthorizationBasic  = "Basic"
)

type request struct {
	s sync.Mutex

	o *atomic.Value      // Options
	x libctx.FuncContext // Context function
	l liblog.FuncLog     // Default logger
	u *url.URL           // endpoint url
	h url.Values         // header values
	p url.Values         // parameters values
	b io.Reader          // body io reader
	m string             // method
	e *requestError      // Error pointer
	c *atomic.Value      // libhtc HTTPClient
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
		c = r.options()
	)

	if c == nil {
		c = &Options{}
	}

	if i, e := New(r.x, c, r.client()); e != nil {
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

	if r.l != nil {
		n.l = r.l
	}

	return n, nil
}

func (r *request) context() context.Context {
	if r.x != nil {
		if x := r.x(); x != nil {
			return x
		}
	}

	return context.Background()
}

func (r *request) client() libhtc.HttpClient {
	if i := r.c.Load(); i != nil {
		if c, k := i.(libhtc.HttpClient); k {
			return c
		}
	}

	c := libhtc.GetClient()
	r.c.Store(c)
	return c
}

func (r *request) Error() Error {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e
}

func (r *request) IsError() bool {
	r.s.Lock()
	defer r.s.Unlock()

	return r.e != nil && r.e.IsError()
}
