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

package socket

import (
	"net"
	"sync/atomic"
	"time"
)

const (
	defaultTimeoutRead  = time.Second
	defaultTimeoutWrite = 5 * time.Second
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
	f *atomic.Value // function error

	tr *atomic.Value // connection read timeout
	tw *atomic.Value // connection write timeout
	sr *atomic.Int32 // read buffer size
	sw *atomic.Int32 // write buffer size
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

func (o *srv) RegisterFuncError(f FuncError) {
	if o == nil {
		return
	}

	o.f.Store(f)
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

func (o *srv) fctError(e error) {
	if o == nil {
		return
	}

	v := o.f.Load()
	if v != nil {
		v.(FuncError)(e)
	}
}

func (o *srv) handler() Handler {
	if o == nil {
		return nil
	}

	v := o.h.Load()
	if v != nil {
		return v.(Handler)
	}

	return nil
}
