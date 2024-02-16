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

package httpserver

import (
	"sync"

	libctx "github.com/nabbar/golib/context"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libsrv "github.com/nabbar/golib/server"
	libver "github.com/nabbar/golib/version"
)

type Info interface {
	GetName() string
	GetBindable() string
	GetExpose() string

	IsDisable() bool
	IsTLS() bool
}

type Server interface {
	libsrv.Server

	Info

	Handler(h srvtps.FuncHandler)
	Merge(s Server, def liblog.FuncLog) error

	GetConfig() *Config
	SetConfig(cfg Config, defLog liblog.FuncLog) error

	Monitor(vrs libver.Version) (montps.Monitor, error)
	MonitorName() string
}

func New(cfg Config, defLog liblog.FuncLog) (Server, error) {
	s := &srv{
		m: sync.RWMutex{},
		r: nil,
		c: libctx.NewConfig[string](cfg.getParentContext),
	}

	s.Handler(cfg.getHandlerFunc)

	if e := s.SetConfig(cfg, defLog); e != nil {
		return nil, e
	}

	return s, nil
}
