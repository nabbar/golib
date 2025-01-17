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

package tls

import (
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
)

type ComponentTlS interface {
	cfgtps.Component
	Config() *libtls.Config
	GetTLS() libtls.TLSConfig
	SetTLS(tls libtls.TLSConfig)
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

func New(ctx libctx.FuncContext, defCARoot libtls.FctRootCACert) ComponentTlS {
	return &componentTls{
		m: sync.RWMutex{},
		x: libctx.NewConfig[uint8](ctx),
		t: nil,
		c: nil,
		f: defCARoot,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentTlS) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx libctx.FuncContext, cfg libcfg.Config, key string, defCARoot libtls.FctRootCACert) {
	cfg.ComponentSet(key, New(ctx, defCARoot))
}

func Load(getCpt cfgtps.FuncCptGet, key string) ComponentTlS {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentTlS); !ok {
		return nil
	} else {
		return h
	}
}
