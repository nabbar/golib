/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package mapCloser

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	libctx "github.com/nabbar/golib/context"
)

type Closer interface {
	Add(clo ...io.Closer)
	Get() []io.Closer
	Len() int

	Clean()
	Clone() Closer
	Close() error
}

func New(ctx context.Context) Closer {
	var (
		x, n = context.WithCancel(ctx)
		fx   = func() context.Context {
			return x
		}
	)

	c := &closer{
		f: n,
		i: new(atomic.Uint64),
		c: new(atomic.Bool),
		x: libctx.NewConfig[uint64](fx),
	}

	c.c.Store(false)
	c.i.Store(0)

	go func() {
		for !c.c.Load() {
			select {
			case <-c.x.Done():
				_ = c.Close()
				return
			default:
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	return c
}
