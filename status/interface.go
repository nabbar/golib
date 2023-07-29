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

package status

import (
	"context"
	"sync"
	"sync/atomic"

	monsts "github.com/nabbar/golib/monitor/status"

	ginsdk "github.com/gin-gonic/gin"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

type Route interface {
	Expose(ctx context.Context)
	MiddleWare(c *ginsdk.Context)
	SetErrorReturn(f func() liberr.ReturnGin)
}

type Info interface {
	SetInfo(name, release, hash string)
	SetVersion(vers libver.Version)
}

type Pool interface {
	montps.PoolStatus
	RegisterPool(fct montps.FuncPool)
}

type Status interface {
	Route
	Info
	Pool

	SetConfig(cfg Config)

	IsHealthy(name ...string) bool
	IsStrictlyHealthy(name ...string) bool

	IsCacheHealthy() bool
	IsCacheStrictlyHealthy() bool
}

func New(ctx libctx.FuncContext) Status {
	s := &sts{
		m: sync.RWMutex{},
		p: nil,
		r: nil,
		x: libctx.NewConfig[string](ctx),
		c: ch{
			m: new(atomic.Int32),
			t: new(atomic.Value),
			c: new(atomic.Int64),
			f: nil,
		},
		fn: nil,
		fr: nil,
		fh: nil,
		fd: nil,
	}

	s.c.f = func() monsts.Status {
		r, _ := s.getStatus()
		return r
	}

	return s
}
