/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package ioprogress

import (
	"errors"
	"io"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libfpg "github.com/nabbar/golib/file/progress"
)

// wrt implements the Writer interface by wrapping an io.WriteCloser
// with progress tracking capabilities using atomic operations for thread safety.
type wrt struct {
	w  io.WriteCloser                    // underlying writer
	cr *atomic.Int64                     // cumulative byte counter (thread-safe)
	fi libatm.Value[libfpg.FctIncrement] // increment callback (thread-safe)
	fe libatm.Value[libfpg.FctEOF]       // EOF callback (thread-safe)
	fr libatm.Value[libfpg.FctReset]     // reset callback (thread-safe)
}

// Write implements io.Writer by delegating to the underlying writer
// and invoking the increment callback with the number of bytes written.
// If EOF is encountered (rare for writes), the EOF callback is also invoked.
func (w *wrt) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.inc(n)

	if errors.Is(err, io.EOF) {
		w.finish()
	}

	return n, err
}

// Close implements io.Closer by closing the underlying writer.
func (w *wrt) Close() error {
	return w.w.Close()
}

// RegisterFctIncrement implements Progress by storing the increment callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (w *wrt) RegisterFctIncrement(fct libfpg.FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	w.fi.Store(fct)
}

// RegisterFctReset implements Progress by storing the reset callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (w *wrt) RegisterFctReset(fct libfpg.FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	w.fr.Store(fct)
}

// RegisterFctEOF implements Progress by storing the EOF callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (w *wrt) RegisterFctEOF(fct libfpg.FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	w.fe.Store(fct)
}

// inc atomically increments the cumulative byte counter and invokes the
// increment callback with the number of bytes written in this operation.
// The callback is invoked even if n is 0 or if the write operation failed.
func (w *wrt) inc(n int) {
	if w == nil {
		return
	}

	w.cr.Add(int64(n))

	f := w.fi.Load()
	if f != nil {
		f(int64(n))
	}
}

// finish invokes the EOF callback when EOF is encountered (rare for write operations).
// This is called automatically by Write() when io.EOF is encountered.
func (w *wrt) finish() {
	if w == nil {
		return
	}

	f := w.fe.Load()
	if f != nil {
		f()
	}
}

// Reset implements Progress by invoking the reset callback with the provided
// maximum size and the current cumulative byte count.
// This is useful for multi-stage operations or progress bar updates.
func (w *wrt) Reset(max int64) {
	if w == nil {
		return
	}

	f := w.fr.Load()
	if f != nil {
		f(max, w.cr.Load())
	}
}
