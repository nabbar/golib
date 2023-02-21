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

package http

import (
	"sync"

	libcfg "github.com/nabbar/golib/config"
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	htpool "github.com/nabbar/golib/httpserver/pool"
	srvtps "github.com/nabbar/golib/httpserver/types"
)

const (
	DefaultTlsKey = "t"
)

type ComponentHttp interface {
	cfgtps.Component

	SetTLSKey(tlsKey string)
	SetHandler(fct srvtps.FuncHandler)

	GetPool() htpool.Pool
	SetPool(pool htpool.Pool)
}

func New(ctx libctx.FuncContext, tlsKey string, hdl srvtps.FuncHandler) ComponentHttp {
	if tlsKey == "" {
		tlsKey = DefaultTlsKey
	}

	return &componentHttp{
		m: sync.RWMutex{},
		x: libctx.NewConfig[uint8](ctx),
		t: tlsKey,
		h: hdl,
		s: nil,
		p: nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentHttp) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(ctx libctx.FuncContext, cfg libcfg.Config, key string, tlsKey string, hdl srvtps.FuncHandler) {
	cfg.ComponentSet(key, New(ctx, tlsKey, hdl))
}

func Load(getCpt cfgtps.FuncCptGet, key string) ComponentHttp {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentHttp); !ok {
		return nil
	} else {
		return h
	}
}
