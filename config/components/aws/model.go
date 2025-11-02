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

package aws

import (
	"sync/atomic"

	libaws "github.com/nabbar/golib/aws"
	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	montps "github.com/nabbar/golib/monitor/types"
	libreq "github.com/nabbar/golib/request"
)

type mod struct {
	x libctx.Config[uint8]
	d ConfigDriver
	p *atomic.Value // montps.FuncPool
	c *atomic.Value // libhtc.HTTPClient
	a *atomic.Value // libaws.AWS
	r *atomic.Value // libreq.Request
	s *atomic.Bool  // status running == true
}

func (o *mod) RegisterHTTPClient(cli libhtc.HttpClient) {
	if cli == nil {
		cli = libhtc.GetClient()
	}

	o.c.Store(cli)
}

func (o *mod) GetAws() libaws.AWS {
	if o.s.Load() {
		return o.getAws()
	}

	return nil
}

func (o *mod) getAws() libaws.AWS {
	if i := o.a.Load(); i == nil {
		return nil
	} else if v, k := i.(libaws.AWS); !k {
		return nil
	} else {
		return v
	}
}

func (o *mod) SetAws(cli libaws.AWS) {
	if cli != nil {
		o.a.Store(cli)
	}
}

func (o *mod) getPool() montps.Pool {
	if i := o.p.Load(); i == nil {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if p := v(); p == nil {
		return nil
	} else {
		return p
	}
}

func (o *mod) RegisterMonitorPool(fct montps.FuncPool) {
	if fct == nil {
		fct = func() montps.Pool {
			return nil
		}
	}

	o.p.Store(fct)
}

func (o *mod) getClient() libhtc.HttpClient {
	if i := o.c.Load(); i == nil {
		return libhtc.GetClient()
	} else if v, k := i.(libhtc.HttpClient); !k {
		return libhtc.GetClient()
	} else {
		return v
	}
}

func (o *mod) getRequest() libreq.Request {
	if i := o.r.Load(); i == nil {
		return nil
	} else if v, k := i.(libreq.Request); !k {
		return nil
	} else {
		return v
	}
}

func (o *mod) setRequest(req libreq.Request) {
	if req != nil {
		o.r.Store(req)
	}
}
