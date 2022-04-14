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
	"sync"

	libcfg "github.com/nabbar/golib/config"
	libreq "github.com/nabbar/golib/request"
)

const (
	ComponentType = "request"
)

type ComponentRequest interface {
	libcfg.Component

	SetHTTPClient(fct libreq.FctHttpClient)
	SetDefaultTLS(key string)

	Request() (libreq.Request, error)
}

func New(tls string, cli libreq.FctHttpClient) ComponentRequest {
	return &componentRequest{
		m: sync.Mutex{},
		r: nil,
		t: tls,
		c: cli,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentRequest) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key, tls string, cli libreq.FctHttpClient) {
	cfg.ComponentSet(key, New(tls, cli))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentRequest {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentRequest); !ok {
		return nil
	} else {
		return h
	}
}
