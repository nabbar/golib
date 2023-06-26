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
	libsrv "github.com/nabbar/golib/server"
	libver "github.com/nabbar/golib/version"
)

type FuncWalk func(bindAddress string, srv libhtp.Server) bool

type Manage interface {
	Walk(fct FuncWalk) bool
	WalkLimit(fct FuncWalk, onlyBindAddress ...string) bool

	Clean()
	Load(bindAddress string) libhtp.Server
	Store(srv libhtp.Server)
	Delete(bindAddress string)

	StoreNew(cfg libhtp.Config, defLog liblog.FuncLog) error
	LoadAndDelete(bindAddress string) (val libhtp.Server, loaded bool)

	MonitorNames() []string
}

type Filter interface {
	Has(bindAddress string) bool
	Len() int
	List(fieldFilter, fieldReturn srvtps.FieldType, pattern, regex string) []string
	Filter(field srvtps.FieldType, pattern, regex string) Pool
}

type Pool interface {
	libsrv.Server

	Manage
	Filter

	Clone(ctx context.Context) Pool
	Merge(p Pool, def liblog.FuncLog) error
	Handler(fct srvtps.FuncHandler)
	Monitor(vrs libver.Version) ([]montps.Monitor, liberr.Error)
}

func New(ctx libctx.FuncContext, hdl srvtps.FuncHandler, srv ...libhtp.Server) Pool {
	p := &pool{
		m: sync.RWMutex{},
		p: libctx.NewConfig[string](ctx),
		h: hdl,
	}

	for _, s := range srv {
		if s != nil {
			p.Store(s)
		}
	}

	return p
}
