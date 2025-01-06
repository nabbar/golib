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
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
)

type ComponentHTTPClient interface {
	cfgtps.Component
	htcdns.DNSMapper

	Config() htcdns.Config
	SetDefault()
	SetAsDefaultHTTPClient(flag bool)
	SetFuncMessage(f htcdns.FuncMessage)
}

func GetRootCaCert(fct libtls.FctRootCA) tlscas.Cert {
	var res tlscas.Cert

	for _, c := range fct() {
		if res == nil {
			res, _ = tlscas.Parse(c)
		} else {
			_ = res.AppendString(c)
		}
	}

	return res
}

func New(ctx libctx.FuncContext, defCARoot libtls.FctRootCACert, isDeftHTTPClient bool, msg htcdns.FuncMessage) ComponentHTTPClient {
	c := &componentHttpClient{
		x: libctx.NewConfig[uint8](ctx),
		c: libatm.NewValue[*htcdns.Config](),
		d: libatm.NewValue[htcdns.DNSMapper](),
		f: libatm.NewValue[libtls.FctRootCACert](),
		m: libatm.NewValue[htcdns.FuncMessage](),
		s: new(atomic.Bool),
	}

	if defCARoot == nil {
		defCARoot = func() tlscas.Cert {
			return nil
		}
	}

	if msg == nil {
		msg = func(msg string) {}
	}

	c.f.Store(defCARoot)
	c.s.Store(isDeftHTTPClient)
	c.m.Store(msg)

	return c
}

func Register(cfg libcfg.Config, key string, cpt ComponentHTTPClient) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx libctx.FuncContext, cfg libcfg.Config, key string, defCARoot libtls.FctRootCACert, isDeftHTTPClient bool, msg htcdns.FuncMessage) {
	cfg.ComponentSet(key, New(ctx, defCARoot, isDeftHTTPClient, msg))
}

func Load(getCpt cfgtps.FuncCptGet, key string) ComponentHTTPClient {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentHTTPClient); !ok {
		return nil
	} else {
		return h
	}
}
