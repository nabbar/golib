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

// rdr implements the Reader interface by wrapping an io.ReadCloser
// with progress tracking capabilities using atomic operations for thread safety.
type rdr struct {
	r  io.ReadCloser                     // underlying reader
	cr *atomic.Int64                     // cumulative byte counter (thread-safe)
	fi libatm.Value[libfpg.FctIncrement] // increment callback (thread-safe)
	fe libatm.Value[libfpg.FctEOF]       // EOF callback (thread-safe)
	fr libatm.Value[libfpg.FctReset]     // reset callback (thread-safe)
}

// Read implements io.Reader by delegating to the underlying reader
// and invoking the increment callback with the number of bytes read.
// If EOF is encountered, the EOF callback is also invoked.
func (r *rdr) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.inc(n)

	if errors.Is(err, io.EOF) {
		r.finish()
	}

	return n, err
}

// Close implements io.Closer by closing the underlying reader.
func (r *rdr) Close() error {
	return r.r.Close()
}

// RegisterFctIncrement implements Progress by storing the increment callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (r *rdr) RegisterFctIncrement(fct libfpg.FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	r.fi.Store(fct)
}

// RegisterFctReset implements Progress by storing the reset callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (r *rdr) RegisterFctReset(fct libfpg.FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	r.fr.Store(fct)
}

// RegisterFctEOF implements Progress by storing the EOF callback.
// Nil callbacks are converted to no-op functions to prevent nil pointer panics.
// This method is thread-safe and can be called concurrently.
func (r *rdr) RegisterFctEOF(fct libfpg.FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	r.fe.Store(fct)
}

// inc atomically increments the cumulative byte counter and invokes the
// increment callback with the number of bytes read in this operation.
// The callback is invoked even if n is 0 or if the read operation failed.
func (r *rdr) inc(n int) {
	if r == nil {
		return
	}

	r.cr.Add(int64(n))

	f := r.fi.Load()
	if f != nil {
		f(int64(n))
	}
}

// finish invokes the EOF callback when the underlying reader reaches end-of-file.
// This is called automatically by Read() when io.EOF is encountered.
func (r *rdr) finish() {
	if r == nil {
		return
	}

	f := r.fe.Load()
	if f != nil {
		f()
	}
}

// Reset implements Progress by invoking the reset callback with the provided
// maximum size and the current cumulative byte count.
// This is useful for multi-stage operations or progress bar updates.
func (r *rdr) Reset(max int64) {
	if r == nil {
		return
	}

	f := r.fr.Load()
	if f != nil {
		f(max, r.cr.Load())
	}
}
