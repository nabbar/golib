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

package progress

import (
	"errors"
	"io"
)

// RegisterFctIncrement registers a callback function that is called after each successful
// read or write operation. The callback receives the number of bytes processed in that operation.
// If fct is nil, a no-op function is registered.
// The callback is stored atomically and can be safely called from concurrent goroutines.
func (o *progress) RegisterFctIncrement(fct FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	o.fi.Store(fct)
}

// RegisterFctReset registers a callback function that is called when the file position
// is reset (e.g., via Seek or Truncate operations). The callback receives two parameters:
//   - size: the maximum size of the file
//   - current: the current position after the reset
//
// If fct is nil, a no-op function is registered.
// The callback is stored atomically and can be safely called from concurrent goroutines.
func (o *progress) RegisterFctReset(fct FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	o.fr.Store(fct)
}

// RegisterFctEOF registers a callback function that is called when end-of-file (EOF)
// is reached during a read operation. This signals completion of reading the entire file.
// If fct is nil, a no-op function is registered.
// The callback is stored atomically and can be safely called from concurrent goroutines.
func (o *progress) RegisterFctEOF(fct FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	o.fe.Store(fct)
}

// SetRegisterProgress propagates all registered callbacks from this Progress instance
// to another Progress instance. This is useful for chaining progress tracking across
// multiple file operations (e.g., copying from one file to another).
// Only non-nil callbacks are propagated.
func (o *progress) SetRegisterProgress(f Progress) {
	i := o.fi.Load()
	if i != nil {
		f.RegisterFctIncrement(i.(FctIncrement))
	}

	i = o.fr.Load()
	if i != nil {
		f.RegisterFctReset(i.(FctReset))
	}

	i = o.fe.Load()
	if i != nil {
		f.RegisterFctEOF(i.(FctEOF))
	}
}

// inc invokes the increment callback with the specified byte count.
// This is called internally after each successful read/write operation.
// The callback is invoked only if registered and instance is not nil.
func (o *progress) inc(n int64) {
	if o == nil {
		return
	}

	f := o.fi.Load()
	if f != nil {
		f.(FctIncrement)(n)
	}
}

// finish invokes the EOF callback to signal end of file reached.
// This is called internally when io.EOF is detected during read operations.
// The callback is invoked only if registered and instance is not nil.
func (o *progress) finish() {
	if o == nil {
		return
	}

	f := o.fe.Load()
	if f != nil {
		f.(FctEOF)()
	}
}

// reset invokes the reset callback with auto-detected file size.
// This is called internally after seek operations and truncation.
func (o *progress) reset() {
	o.Reset(0)
}

// Reset invokes the reset callback with the specified maximum size and current position.
// If max is less than 1, it is automatically detected from file statistics.
// The callback receives the file size and current position from beginning of file.
// This method is public to allow manual reset triggering if needed.
func (o *progress) Reset(max int64) {
	if o == nil {
		return
	}

	f := o.fr.Load()

	if f != nil {
		if max < 1 {
			if i, e := o.Stat(); e != nil {
				return
			} else {
				max = i.Size()
			}
		}

		if s, e := o.SizeBOF(); e != nil {
			return
		} else if s >= 0 {
			f.(FctReset)(max, s)
		}
	}
}

// analyze processes the result of an I/O operation by invoking appropriate callbacks.
// It calls the increment callback if bytes were processed (i != 0).
// It calls the EOF callback if an EOF error is detected.
// This method wraps I/O results to provide transparent progress tracking.
func (o *progress) analyze(i int, e error) (n int, err error) {
	if o == nil {
		return i, e
	}

	if i != 0 {
		o.inc(int64(i))
	}

	if e != nil && errors.Is(e, io.EOF) {
		o.finish()
	}

	return i, e
}
