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
	"context"
	"io"
	"os"
	"sync/atomic"
	"time"
)

type FuncError func(e error)
type Handler func(request io.Reader, response io.Writer)

type Server interface {
	RegisterFuncError(f FuncError)
	SetReadTimeout(d time.Duration)
	SetWriteTimeout(d time.Duration)

	Listen(ctx context.Context, unixFile string, perm os.FileMode)
	Shutdown()
	Done() <-chan struct{}
}

func New(h Handler, sizeBuffRead, sizeBuffWrite int32) Server {
	c := new(atomic.Value)
	c.Store(make(chan []byte))

	s := new(atomic.Value)
	s.Store(make(chan struct{}))

	f := new(atomic.Value)
	f.Store(h)

	sr := new(atomic.Int32)
	sr.Store(sizeBuffRead)

	sw := new(atomic.Int32)
	sw.Store(sizeBuffWrite)

	return &srv{
		l:  nil,
		h:  f,
		c:  c,
		s:  s,
		f:  new(atomic.Value),
		tr: new(atomic.Value),
		tw: new(atomic.Value),
		sr: sr,
		sw: sw,
	}
}
