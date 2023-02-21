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

package run

import (
	"context"
	"net"
	"net/http"
	"sync"

	liblog "github.com/nabbar/golib/logger"
)

type sRun struct {
	m   sync.RWMutex
	err error
	chn chan struct{}
	ctx context.Context
	cnl context.CancelFunc
	log liblog.FuncLog
	srv *http.Server
	run bool
	tls bool
}

func (o *sRun) logger() liblog.Logger {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.log == nil {
		return liblog.GetDefault()
	} else if l := o.log(); l == nil {
		return liblog.GetDefault()
	} else {
		return l
	}
}

func (o *sRun) SetServer(srv *http.Server, log liblog.FuncLog, tls bool) {
	o.m.Lock()
	defer o.m.Unlock()

	srv.BaseContext = func(listener net.Listener) context.Context {
		return o.getContext()
	}

	o.srv = srv
	o.log = log
	o.tls = tls
}

func (o *sRun) getServer() *http.Server {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.srv
}

func (o *sRun) delServer() {
	o.m.Lock()
	defer o.m.Unlock()
	if o.srv != nil {
		o.logger().Entry(liblog.ErrorLevel, "closing server").ErrorAdd(true, o.srv.Close()).Check(liblog.NilLevel)
		o.srv = nil
	}
}

func (o *sRun) getContext() context.Context {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.ctx
}

func (o *sRun) getCancel() context.CancelFunc {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.cnl
}

func (o *sRun) isTLS() bool {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.tls
}

func (o *sRun) GetError() error {
	o.m.RLock()
	defer o.m.RUnlock()
	return o.err
}

func (o *sRun) setError(err error) {
	o.m.Lock()
	defer o.m.Unlock()
	o.err = err
}
