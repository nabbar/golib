//go:build linux
// +build linux

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

package unix

import (
	"net"
	"os"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
)

var (
	closedChanStruct chan struct{}
)

func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

type srv struct {
	l net.Listener

	h *atomic.Value // handler
	c *atomic.Value // chan []byte
	s *atomic.Value // chan struct{}
	e *atomic.Value // function error
	i *atomic.Value // function info

	tr *atomic.Value // connection read timeout
	tw *atomic.Value // connection write timeout
	sr *atomic.Int32 // read buffer size
	sw *atomic.Int32 // write buffer size
	fs *atomic.Value // file unix socket
	fp *atomic.Int64 // file unix perm
}

func (o *srv) Done() <-chan struct{} {
	s := o.s.Load()
	if s != nil {
		return s.(chan struct{})
	}

	return closedChanStruct
}

func (o *srv) Shutdown() {
	if o == nil {
		return
	}

	s := o.s.Load()
	if s != nil {
		o.s.Store(nil)
	}
}

func (o *srv) RegisterFuncError(f libsck.FuncError) {
	if o == nil {
		return
	}

	o.e.Store(f)
}

func (o *srv) RegisterFuncInfo(f libsck.FuncInfo) {
	if o == nil {
		return
	}

	o.i.Store(f)
}

func (o *srv) SetReadTimeout(d time.Duration) {
	if o == nil {
		return
	}

	o.tr.Store(d)
}

func (o *srv) SetWriteTimeout(d time.Duration) {
	if o == nil {
		return
	}

	o.tw.Store(d)
}

func (o *srv) RegisterSocket(unixFile string, perm os.FileMode) {
	o.fs.Store(unixFile)
	o.fp.Store(int64(perm))
}

func (o *srv) fctError(e error) {
	if o == nil {
		return
	}

	v := o.e.Load()
	if v != nil {
		v.(libsck.FuncError)(e)
	}
}

func (o *srv) fctInfo(local, remote net.Addr, state libsck.ConnState) {
	if o == nil {
		return
	}

	v := o.i.Load()
	if v != nil {
		v.(libsck.FuncInfo)(local, remote, state)
	}
}

func (o *srv) handler() libsck.Handler {
	if o == nil {
		return nil
	}

	v := o.h.Load()
	if v != nil {
		return v.(libsck.Handler)
	}

	return nil
}
