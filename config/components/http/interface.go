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
	"net/http"
	"sync"

	libcfg "github.com/nabbar/golib/config"
	libhts "github.com/nabbar/golib/httpserver"
)

const (
	DefaultTlsKey = "tls"
	DefaultLogKey = "log"
	ComponentType = "http"
)

type ComponentHttp interface {
	libcfg.Component

	SetTLSKey(tlsKey string)
	SetLOGKey(logKey string)
	SetHandler(handler map[string]http.Handler)

	GetPool() libhts.PoolServer
	SetPool(pool libhts.PoolServer)
}

func New(tlsKey, logKey string, handler map[string]http.Handler) ComponentHttp {
	if tlsKey == "" {
		tlsKey = DefaultTlsKey
	}

	if logKey == "" {
		logKey = DefaultLogKey
	}

	return &componentHttp{
		m:    sync.Mutex{},
		tls:  tlsKey,
		log:  logKey,
		run:  false,
		hand: handler,
		pool: nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentHttp) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key string, tlsKey, logKey string, handler map[string]http.Handler) {
	cfg.ComponentSet(key, New(tlsKey, logKey, handler))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentHttp {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentHttp); !ok {
		return nil
	} else {
		return h
	}
}
