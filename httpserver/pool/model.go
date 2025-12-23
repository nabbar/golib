/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package pool

import (
	"context"
	"sync"

	liblog "github.com/nabbar/golib/logger"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libhtp "github.com/nabbar/golib/httpserver"
	srvtps "github.com/nabbar/golib/httpserver/types"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

// pool is the internal implementation of the Pool interface.
// It uses sync.RWMutex for thread-safe operations and stores servers
// in a libctx.Config map indexed by bind address.
type pool struct {
	m sync.RWMutex          // mutex protects concurrent access
	p libctx.Config[string] // server map keyed by bind address
	h srvtps.FuncHandler    // optional shared handler function
}

func (o *pool) Clone(ctx context.Context) Pool {
	return &pool{
		m: sync.RWMutex{},
		p: o.p.Clone(ctx),
		h: o.h,
	}
}

func (o *pool) Merge(p Pool, def liblog.FuncLog) error {
	var err error

	p.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if s := o.Get(bindAddress); s == nil {
			o.Store(srv)
		} else if e := s.Merge(srv, def); e != nil {
			err = e
			return false
		} else {
			o.Store(s)
		}
		return true
	})

	return err
}

// Get retrieves a server by bind address.
// This is an internal helper method. Use Load for public access.
func (o *pool) Get(adr string) libhtp.Server {
	if i, l := o.p.Load(adr); !l {
		return nil
	} else if v, k := i.(libhtp.Server); !k {
		return nil
	} else {
		return v
	}
}

func (o *pool) Handler(fct srvtps.FuncHandler) {
	o.m.Lock()
	defer o.m.Unlock()
	o.h = fct
}

func (o *pool) MonitorNames() []string {
	var res = make([]string, 0)

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		res = append(res, srv.MonitorName())
		return true
	})

	return res
}

func (o *pool) Monitor(vrs libver.Version) ([]montps.Monitor, liberr.Error) {
	var (
		res = make([]montps.Monitor, 0)
		err = ErrorPoolMonitor.Error(nil)
	)

	o.Walk(func(bindAddress string, srv libhtp.Server) bool {
		if p, e := srv.Monitor(vrs); e != nil {
			err.Add(e)
		} else {
			res = append(res, p)
		}

		return true
	})

	if !err.HasParent() {
		err = nil
	}

	return res, err
}
