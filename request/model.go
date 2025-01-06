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
	"net/url"
	"sync"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"

	libhtc "github.com/nabbar/golib/httpcli"
)

const (
	_ContentType         = "Content-Type"
	__ContentLength      = "Content-Length"
	_Authorization       = "Authorization"
	_AuthorizationBearer = "Bearer"
	_AuthorizationBasic  = "Basic"
)

type request struct {
	mux sync.Mutex

	opt *atomic.Value // Options
	ctx *atomic.Value // Context function: libctx.FuncContext
	log *atomic.Value // Default logger : liblog.FuncLog
	uri *url.URL      // endpoint url
	hdr sync.Map      // header values
	prm url.Values    // parameters values
	bdr BodyRetryer   // body io reader
	mth string        // method
	err *atomic.Value // Error pointer: *requestError
	cli *atomic.Value // libhtc HTTPClient
}

func (r *request) Clone() (Request, error) {
	if n, e := r.New(); e != nil {
		return nil, e
	} else {
		n.CleanHeader()

		r.hdr.Range(func(key, value any) bool {
			if u, l := key.(string); !l {
				return true
			} else if v, k := value.([]string); !k {
				return true
			} else {
				for _, w := range v {
					n.AddHeader(u, w)
				}
			}

			return true
		})

		n.CleanParams()
		for k := range r.prm {
			n.SetParams(k, r.prm.Get(k))
		}

		return n, nil
	}
}

func (r *request) New() (Request, error) {
	var (
		n *request
		c = r.options()
	)

	if c == nil {
		c = &Options{}
	}

	if i, e := New(r.getFuncContext(), c, r.client()); e != nil {
		return nil, e
	} else {
		n = i.(*request)
	}

	if r.uri != nil {
		n.uri = &url.URL{
			Scheme:      r.uri.Scheme,
			Opaque:      r.uri.Opaque,
			User:        r.uri.User,
			Host:        r.uri.Host,
			Path:        r.uri.Path,
			RawPath:     r.uri.RawPath,
			ForceQuery:  r.uri.ForceQuery,
			RawQuery:    r.uri.RawQuery,
			Fragment:    r.uri.Fragment,
			RawFragment: r.uri.RawFragment,
		}
	}

	if r.log != nil {
		n.log = r.log
	}

	return n, nil
}

func (r *request) Free() {
	if i := r.err.Load(); i != nil {
		if v, k := i.(*requestError); !k {
			v.Free()
			r.err.Store(&requestError{})
		}
	}
}

func (r *request) RegisterDefaultLogger(fct liblog.FuncLog) {
	if fct == nil {
		fct = func() liblog.Logger {
			return nil
		}
	}

	r.log.Store(fct)
}

func (r *request) _getDefaultLogger() liblog.Logger {
	if i := r.log.Load(); i == nil {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else {
		return v()
	}
}

func (r *request) RegisterContext(fct libctx.FuncContext) {
	if fct == nil {
		fct = context.Background
	}

	r.ctx.Store(fct)
}

func (r *request) getFuncContext() libctx.FuncContext {
	if i := r.ctx.Load(); i == nil {
		return nil
	} else if v, k := i.(libctx.FuncContext); !k {
		return nil
	} else {
		return v
	}
}

func (r *request) context() context.Context {
	if f := r.getFuncContext(); f == nil {
		return context.Background()
	} else {
		return f()
	}
}

func (r *request) RegisterHTTPClient(cli libhtc.HttpClient) {
	if cli == nil {
		cli = libhtc.GetClient()
	}

	r.cli.Store(cli)
}

func (r *request) client() libhtc.HttpClient {
	if i := r.cli.Load(); i != nil {
		if c, k := i.(libhtc.HttpClient); k {
			return c
		}
	}

	return libhtc.GetClient()
}

func (r *request) Error() Error {
	return r.getError()
}

func (r *request) IsError() bool {
	err := r.getError()
	return err.IsError()
}
