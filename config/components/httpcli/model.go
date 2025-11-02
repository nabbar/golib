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

package httpcli

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

type mod struct {
	x libctx.Config[uint8]

	c libatm.Value[*htcdns.Config]       // htcdns.Config
	d libatm.Value[htcdns.DNSMapper]     // htcdns.DNSMapper
	f libatm.Value[libtls.FctRootCACert] // FuncDefaultCARoot
	m libatm.Value[htcdns.FuncMessage]   // htcdns.FctMessage

	s *atomic.Bool // is Default at start / update
}

func (o *mod) getRootCA() tlscas.Cert {
	if i := o.f.Load(); i == nil {
		return nil
	} else if v := i(); v != nil && v.Len() < 1 {
		return nil
	} else {
		return v
	}
}

func (o *mod) getMessage() htcdns.FuncMessage {
	if i := o.m.Load(); i == nil {
		return nil
	} else {
		return i
	}
}

func (o *mod) SetFuncMessage(f htcdns.FuncMessage) {
	if f != nil {
		o.m.Store(f)
	}
}

func (o *mod) getDNSMapper() htcdns.DNSMapper {
	if i := o.d.Load(); i == nil {
		return nil
	} else {
		return i
	}
}

func (o *mod) setDNSMapper(dns htcdns.DNSMapper) {
	if o.s.Load() {
		defer o.SetDefault()
	}

	var old htcdns.DNSMapper

	if dns != nil {
		old = o.d.Swap(dns)
	}

	if old != nil {
		_ = old.Close()
	}
}

func (o *mod) Config() htcdns.Config {
	if i := o.c.Load(); i == nil {
		return htcdns.Config{}
	} else {
		return *i
	}
}

func (o *mod) setConfig(cfg htcdns.Config) {
	o.c.Store(&cfg)
}

func (o *mod) SetDefault() {
	libhtc.SetDefaultDNSMapper(o.getDNSMapper())
}

func (o *mod) SetAsDefaultHTTPClient(flag bool) {
	o.s.Store(flag)
}
