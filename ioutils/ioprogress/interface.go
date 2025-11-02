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
	"io"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libfpg "github.com/nabbar/golib/file/progress"
)

type Progress interface {
	// RegisterFctIncrement registers a function to be called when an increment of size n is done.
	//
	// The function takes one argument, the size of the increment in bytes.
	// If the function is nil, it will be ignored.
	//
	// The function will be called immediately after the increment is done.
	// The function will be called on the same goroutine as the Reader or Writer.
	// The function will be called with the size of the increment, even if the Read or Write failed.
	// The function will be called until it returns an error.
	//
	// The function must not block.
	RegisterFctIncrement(fct libfpg.FctIncrement)
	// RegisterFctReset registers a function to be called when a reset is done.
	//
	// The function takes two arguments, the maximum size of the progress in bytes and the current size of the progress in bytes.
	// If the function is nil, it will be ignored.
	//
	// The function will be called immediately after the reset is done.
	// The function will be called on the same goroutine as the Reader or Writer.
	// The function will be called until it returns an error.
	//
	// The function must not block.
	RegisterFctReset(fct libfpg.FctReset)
	// RegisterFctEOF registers a function to be called when the end of the file is reached.
	//
	// The function takes no argument.
	// If the function is nil, it will be ignored.
	//
	// The function will be called immediately after the end of the file is reached.
	// The function will be called on the same goroutine as the Reader or Writer.
	// The function will be called until it returns an error.
	//
	// The function must not block.
	RegisterFctEOF(fct libfpg.FctEOF)
	// Reset resets the progress to the given maximum size.
	//
	// The function takes one argument, the maximum size of the progress in bytes.
	//
	// The function will be called immediately after the reset is done.
	// The function will be called on the same goroutine as the Reader or Writer.
	// The function will be called until it returns an error.
	//
	// The function must not block.
	Reset(max int64)
}

type Reader interface {
	io.ReadCloser
	Progress
}

type Writer interface {
	io.WriteCloser
	Progress
}

// NewReadCloser creates a new io.ReadCloser from the given io.ReadCloser, and
// returns a Reader compatible with the ioprogress interface.
//
// The returned Reader implements the io.ReadCloser interface, and
// the Progress interface.
//
// The returned Reader is a wrapper around the given io.ReadCloser.
// It keeps track of the current progress, and can register functions to
// be called when an increment of size n is done, when the end of the file
// is reached, and when the progress is reset.
//
// The returned Reader is safe to use concurrently.
//
// The returned Reader is not a clone of the given io.ReadCloser. It is
// a wrapper around the given io.ReadCloser. This means that any
// operations done on the returned Reader will affect the given io.ReadCloser.
//
// The returned Reader is valid until the given io.ReadCloser is closed.
// Once the given io.ReadCloser is closed, the returned Reader is invalid.
func NewReadCloser(r io.ReadCloser) Reader {
	o := &rdr{
		r:  r,
		cr: new(atomic.Int64),
		fi: libatm.NewValue[libfpg.FctIncrement](),
		fe: libatm.NewValue[libfpg.FctEOF](),
		fr: libatm.NewValue[libfpg.FctReset](),
	}
	o.RegisterFctIncrement(nil)
	o.RegisterFctEOF(nil)
	o.RegisterFctReset(nil)
	return o
}

// NewWriteCloser creates a new io.WriteCloser from the given io.WriteCloser, and
// returns a Writer compatible with the ioprogress interface.
//
// The returned Writer implements the io.WriteCloser interface, and
// the Progress interface.
//
// The returned Writer is a wrapper around the given io.WriteCloser.
// It keeps track of the current progress, and can register functions to
// be called when an increment of size n is done, when the end of the file
// is reached, and when the progress is reset.
//
// The returned Writer is safe to use concurrently.
//
// The returned Writer is not a clone of the given io.WriteCloser. It is
// a wrapper around the given io.WriteCloser. This means that any
// operations done on the returned Writer will affect the given io.WriteCloser.
//
// The returned Writer is valid until the given io.WriteCloser is closed.
// Once the given io.WriteCloser is closed, the returned Writer is invalid.
func NewWriteCloser(w io.WriteCloser) Writer {
	o := &wrt{
		w:  w,
		cr: new(atomic.Int64),
		fi: libatm.NewValue[libfpg.FctIncrement](),
		fe: libatm.NewValue[libfpg.FctEOF](),
		fr: libatm.NewValue[libfpg.FctReset](),
	}
	o.RegisterFctIncrement(nil)
	o.RegisterFctEOF(nil)
	o.RegisterFctReset(nil)
	return o
}
