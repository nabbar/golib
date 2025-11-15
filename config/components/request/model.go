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
	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	montps "github.com/nabbar/golib/monitor/types"
	libreq "github.com/nabbar/golib/request"
)

type mod struct {
	x libctx.Config[uint8]
	r libatm.Value[libreq.Request]    // libreq.Request
	c libatm.Value[libhtc.HttpClient] // libhtc.HttpClient
}

func (o *mod) SetHTTPClient(cli libhtc.HttpClient) {
	o.setClient(cli)
}

func (o *mod) Request() libreq.Request {
	return o.getRequest()
}

func (o *mod) RegisterMonitorPool(fct montps.FuncPool) {
	o.setPool(fct)
}

func (o *mod) getRequest() libreq.Request {
	if i := o.r.Load(); i == nil {
		return nil
	} else {
		return i
	}
}

func (o *mod) setRequest(req libreq.Request) {
	if req == nil {
		return
	}
	o.r.Store(req)
}

func (o *mod) getClient() libhtc.HttpClient {
	if i := o.c.Load(); i == nil {
		return libhtc.GetClient()
	} else {
		return i
	}
}

func (o *mod) setClient(cli libhtc.HttpClient) {
	if cli == nil {
		return
	}

	o.c.Store(cli)
}

func (o *mod) getPool() montps.Pool {
	if i, l := o.x.Load(keyFctMonitorPool); !l || i == nil {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if p := v(); p == nil {
		return nil
	} else {
		return p
	}
}

func (o *mod) setPool(fct montps.FuncPool) {
	if fct == nil {
		fct = func() montps.Pool {
			return nil
		}
	}

	o.x.Store(keyFctMonitorPool, fct)
}
