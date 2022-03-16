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

package natsServer

import (
	"sync"

	libsts "github.com/nabbar/golib/status"

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
	libnat "github.com/nabbar/golib/nats"
)

const (
	DefaultTlsKey = "tls"
	ComponentType = "natsServer"
)

type ComponentNats interface {
	libcfg.Component

	SetTLSKey(tlsKey string)
	GetServer() (libnat.Server, liberr.Error)
	SetStatusRouter(sts libsts.RouteStatus, prefix string)
}

func New(tlsKey string) ComponentNats {
	if tlsKey == "" {
		tlsKey = DefaultTlsKey
	}

	return &componentNats{
		ctx: nil,
		get: nil,
		fsa: nil,
		fsb: nil,
		fra: nil,
		frb: nil,
		m:   sync.Mutex{},
		t:   tlsKey,
		n:   nil,
	}
}

func Register(cfg libcfg.Config, key string, cpt ComponentNats) {
	cfg.ComponentSet(key, cpt)
}

func RegisterNew(cfg libcfg.Config, key, tlsKey string) {
	cfg.ComponentSet(key, New(tlsKey))
}

func Load(getCpt libcfg.FuncComponentGet, key string) ComponentNats {
	if c := getCpt(key); c == nil {
		return nil
	} else if h, ok := c.(ComponentNats); !ok {
		return nil
	} else {
		return h
	}
}
