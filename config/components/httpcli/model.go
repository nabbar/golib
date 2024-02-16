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

	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
	libhtc "github.com/nabbar/golib/httpcli"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

type componentHttpClient struct {
	x libctx.Config[uint8]
	c *atomic.Value // htcdns.Config
	d *atomic.Value // htcdns.DNSMapper
	f *atomic.Value // FuncDefaultCARoot
	s *atomic.Bool  // is Default at start / update
}

func (o *componentHttpClient) getRootCA() []string {
	if i := o.f.Load(); i == nil {
		return make([]string, 0)
	} else if v, k := i.(libtls.FctRootCA); !k {
		return make([]string, 0)
	} else if r := v(); len(r) < 1 {
		return make([]string, 0)
	} else {
		return r
	}
}

func (o *componentHttpClient) getDNSMapper() htcdns.DNSMapper {
	if i := o.d.Load(); i == nil {
		return nil
	} else if v, k := i.(htcdns.DNSMapper); !k {
		return nil
	} else {
		return v
	}
}

func (o *componentHttpClient) setDNSMapper(dns htcdns.DNSMapper) {
	if o.s.Load() {
		defer o.SetDefault()
	}

	if dns != nil {
		o.d.Store(dns)
	}
}

func (o *componentHttpClient) Config() htcdns.Config {
	if i := o.c.Load(); i == nil {
		return htcdns.Config{}
	} else if v, k := i.(*htcdns.Config); !k {
		return htcdns.Config{}
	} else {
		return *v
	}
}

func (o *componentHttpClient) setConfig(cfg htcdns.Config) {
	o.c.Store(&cfg)
}

func (o *componentHttpClient) SetDefault() {
	libhtc.SetDefaultDNSMapper(o.getDNSMapper())
}

func (o *componentHttpClient) SetAsDefaultHTTPClient(flag bool) {
	o.s.Store(flag)
}
