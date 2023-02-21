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

	liberr "github.com/nabbar/golib/errors"

	montps "github.com/nabbar/golib/monitor/types"

	libctx "github.com/nabbar/golib/context"

	libver "github.com/nabbar/golib/version"

	ginsdk "github.com/gin-gonic/gin"
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
	montps.Pool
	RegisterPool(fct montps.FuncPool)
}

type Status interface {
	Route
	Info
	Pool

	SetConfig(cfg Config)
	IsHealthy(name ...string) bool
}

func New(ctx libctx.FuncContext) Status {
	return &sts{
		m:  sync.RWMutex{},
		p:  nil,
		x:  libctx.NewConfig[string](ctx),
		fn: nil,
		fr: nil,
		fh: nil,
		fd: nil,
	}
}
