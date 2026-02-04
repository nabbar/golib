/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package hooksyslog

import (
	"errors"
	"io"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"
)

// Write sends a byte slice to the underlying aggregator. This method implements
// the io.Writer interface for the hook.
//
// This method contains a critical recovery mechanism. If a write fails because the
// underlying aggregator has been closed (e.g., due to a network error that couldn't
// be resolved by the socket's internal reconnect logic), it will attempt to
// re-initialize the aggregator by calling `setAgg`. This makes the hook resilient
// to prolonged connection losses.
//
// A mutex ensures that only one goroutine attempts to re-initialize the writer at a time.
func (o *hks) Write(p []byte) (n int, err error) {
	n, err = o.w.Write(p)

	// If the write was successful or the error is not a "closed resources" error, return.
	if err == nil || !errors.Is(err, iotagg.ErrClosedResources) {
		return n, err
	}

	// If we reach here, the aggregator's writer is closed.
	// Acquire a lock to ensure only one goroutine attempts to recover.
	o.m.Lock()
	defer o.m.Unlock()

	// After acquiring the lock, another goroutine might have already fixed the writer.
	// Retry the write to check if recovery is still necessary.
	n, err = o.w.Write(p)
	if err == nil || !errors.Is(err, iotagg.ErrClosedResources) {
		return n, err
	}

	// Recovery is necessary. Re-initialize the aggregator.
	var (
		e error
		a io.Writer
		l bool
	)

	a, l, e = setAgg(o.o.network, o.o.endpoint)
	if e != nil {
		return n, e
	}

	o.w = a
	o.l.Store(l)

	// Attempt the write one last time with the new writer.
	if n, err = o.w.Write(p); err != nil {
		return n, err
	}

	// Log a message to the new writer to indicate that a recovery has occurred.
	_, _ = o.w.Write([]byte(time.Now().Format(time.RFC3339) + " recovered closed resources, maybe some implementation error - info : " + o.getSyslogInfo() + "\n"))

	return n, err
}

// Close marks the hook as closed and decrements the reference count on the shared
// aggregator. If this hook is the last user of the aggregator, the aggregator's
// resources (including the network connection) will be released.
// This method implements the io.Closer interface.
func (o *hks) Close() error {
	if o.r.CompareAndSwap(true, false) {
		delAgg(o.o.network, o.o.endpoint)
	}
	return nil
}
