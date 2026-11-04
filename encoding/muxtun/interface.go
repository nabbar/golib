//go:build amd64 || arm64

/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package muxtun

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	iotagg "github.com/nabbar/golib/ioutils/aggregator"
	libsck "github.com/nabbar/golib/socket"
)

type MuxTunnel interface {
	io.Closer
	context.Context

	Writer(id uint8) io.Writer
	Reader(id uint8) io.ReadCloser

	Write(id uint8, pay []byte) (n int, err error)
	Read(id uint8, pay []byte) (n int, err error)

	// SetLoggerError registers a custom callback for internal error reporting.
	// This function is invoked for write failures, context errors, and recovered panics.
	// If the provided function is nil, a no-op function is internally assigned.
	// Thread-safety: Safely updatable during runtime using atomic storage.
	SetLoggerError(func(msg string, err ...error))

	// SetLoggerInfo registers a custom callback for lifecycle and operational events.
	// It is used to log startup progress, graceful shutdowns, and informational status.
	// If the provided function is nil, a no-op function is internally assigned.
	// Thread-safety: Safely updatable during runtime using atomic storage.
	SetLoggerInfo(func(msg string, arg ...any))

	Counter() (uint64, uint64, time.Duration, time.Duration)

	IsReaderOpen(id uint8) bool
	IsWriterOpen(id uint8) bool
}

func Socket(ctx libsck.Context, maxWriteQueue, maxReadQueue int) (MuxTunnel, error) {
	return New(ctx, ctx, ctx, maxWriteQueue, maxReadQueue)
}

func New(x context.Context, r io.Reader, w io.Writer, maxWriteQueue, maxReadQueue int) (MuxTunnel, error) {
	m := &mdl{
		w: nil,
		b: sync.Pool{},
		p: sync.Pool{},
		s: strList{},
		x: nil,
		n: nil,
		ao: &ack{
			m: sync.Mutex{},
			r: make(map[uint64]int64),
			s: nil,
		},
		rs: new(atomic.Bool),
		vi: &vStream{
			m: sync.Mutex{},
			v: nil,
		},
		vo: &vStream{
			m: sync.Mutex{},
			v: nil,
		},
		le: libatm.NewValue[func(string, ...error)](),
		li: libatm.NewValue[func(string, ...any)](),
		cw: new(atomic.Uint64),
		cr: new(atomic.Uint64),
		cn: new(atomic.Int64),
		cs: new(atomic.Int64),
		cl: new(atomic.Uint64),
	}

	m.x, m.n = context.WithCancel(x)
	m.ao.s = m.sendRSYN

	m.b.New = func() any {
		b := make([]byte, maxSizeBuffer)
		return &b
	}

	m.p.New = func() any {
		return &packet{
			i: 0,
			l: 0,
			c: 0,
			h: make([]byte, maxSizeHeader),
			p: make([]byte, maxSizePayload),
		}
	}

	for i := 0; i < maxWriteQueue+maxReadQueue; i++ {
		if i < maxWriteQueue {
			b := make([]byte, maxSizeBuffer)
			m.b.Put(&b)
		}

		m.p.Put(&packet{
			i: 0,
			l: 0,
			c: 0,
			h: make([]byte, maxSizeHeader),
			p: make([]byte, maxSizePayload),
		})
	}

	cfg := iotagg.Config{
		AsyncTimer: 0,
		AsyncMax:   0,
		AsyncFct:   nil,
		SyncTimer:  0,
		SyncFct:    nil,
		BufWriter:  maxWriteQueue,
		BufMaxSize: maxSizeBuffer,
		FctWriter:  w.Write,
	}

	agg, err := iotagg.New(m.x, cfg)
	if err != nil {
		return nil, err
	}

	if err = agg.Start(m.x); err != nil {
		_ = agg.Close()
		return nil, err
	}

	m.w = agg
	go m.readLoop(r)
	go m.ao.ticker(m.x, 250*time.Millisecond)
	go m.cntDelay(time.Second)

	return m, nil
}
