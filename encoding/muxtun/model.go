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
	"errors"
	"sync"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	iotagg "github.com/nabbar/golib/ioutils/aggregator"
)

const (
	strWriting = "error on writing packet to tunnel"
	strReading = "error on reading packet from tunnel"

	errPayTooBig = "payload error: too big"
	errCheckSum  = "invalid checksum"
)

var (
	ErrPayloadTooBig   = errors.New(errPayTooBig)
	ErrInvalidChecksum = errors.New(errCheckSum)
)

type mdl struct {
	w iotagg.Aggregator // aggregator for write with multiplexer sequence
	b sync.Pool         // pool for buffer writer & reader
	p sync.Pool         // pool for packet reader
	s strList           //  stream reader list

	x context.Context    // local context
	n context.CancelFunc // local cancel context

	// Ack Table & Control
	ao *ack         // Ack Table Outbound
	rs *atomic.Bool // RSYN Command Request
	vi *vStream     // Input Virtual Stream state
	vo *vStream     // Output Virtual Stream state

	// Logging Callbacks
	le libatm.Value[func(string, ...error)] // le (log-error): Atomic storage for the user-provided callback for internal error reporting.
	li libatm.Value[func(string, ...any)]   // li (log-info): Atomic storage for the user-provided callback for informational events.

	// counter
	cw *atomic.Uint64 // counter write size bytes
	cr *atomic.Uint64 // counter read size bytes
	cn *atomic.Int64  // counter time unix nano
	cs *atomic.Int64  // counter time unix second
	cl *atomic.Uint64 // counter for latency
}

func (o *mdl) Close() error {
	o.n()
	_ = o.w.Close()
	_ = o.s.Close()
	return nil
}

func (o *mdl) IsReaderOpen(id uint8) bool {
	if id == 0 {
		return o.Err() == nil
	}

	return o.vi.open(id)
}

func (o *mdl) IsWriterOpen(id uint8) bool {
	if id == 0 {
		return o.Err() == nil
	}

	return o.vo.open(id)
}
