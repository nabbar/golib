/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package iowrapper

import (
	"io"

	libatm "github.com/nabbar/golib/atomic"
)

// iow is the internal implementation of IOWrapper.
// It stores the underlying object and atomic values for custom functions.
// All fields use atomic operations to ensure thread-safe concurrent access.
type iow struct {
	i any                     // Underlying I/O object
	r libatm.Value[FuncRead]  // Atomic custom read function
	w libatm.Value[FuncWrite] // Atomic custom write function
	s libatm.Value[FuncSeek]  // Atomic custom seek function
	c libatm.Value[FuncClose] // Atomic custom close function
}

// SetRead sets a custom read function for this wrapper.
// If read is nil, resets to default behavior (delegates to underlying io.Reader).
// This method is thread-safe and can be called concurrently with Read operations.
func (o *iow) SetRead(read FuncRead) {
	if read == nil {
		read = o.fakeRead
	}

	o.r.Store(read)
}

// SetWrite sets a custom write function for this wrapper.
// If write is nil, resets to default behavior (delegates to underlying io.Writer).
// This method is thread-safe and can be called concurrently with Write operations.
func (o *iow) SetWrite(write FuncWrite) {
	if write == nil {
		write = o.fakeWrite
	}

	o.w.Store(write)
}

// SetSeek sets a custom seek function for this wrapper.
// If seek is nil, resets to default behavior (delegates to underlying io.Seeker).
// This method is thread-safe and can be called concurrently with Seek operations.
func (o *iow) SetSeek(seek FuncSeek) {
	if seek == nil {
		seek = o.fakeSeek
	}

	o.s.Store(seek)
}

// SetClose sets a custom close function for this wrapper.
// If close is nil, resets to default behavior (delegates to underlying io.Closer).
// This method is thread-safe and can be called concurrently with Close operations.
func (o *iow) SetClose(close FuncClose) {
	if close == nil {
		close = o.fakeClose
	}

	o.c.Store(close)
}

// Read implements io.Reader by delegating to the custom function or underlying reader.
// Returns the number of bytes read and any error.
//
// Behavior:
//   - If custom function returns nil: returns (0, io.ErrUnexpectedEOF)
//   - If custom function returns empty slice: returns (0, nil)
//   - If custom function returns data: copies to p and returns (len, nil)
//   - If no custom function and no underlying io.Reader: returns (0, io.ErrUnexpectedEOF)
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (o *iow) Read(p []byte) (n int, err error) {
	var r []byte

	// Load custom function atomically
	fn := o.r.Load()
	if fn != nil {
		r = fn(p)
	} else {
		r = o.fakeRead(p)
	}

	// nil indicates error or EOF
	if r == nil {
		return 0, io.ErrUnexpectedEOF
	}

	// Copy returns the number of elements copied, which is min(len(p), len(r))
	n = copy(p, r)
	return n, nil
}

// Write implements io.Writer by delegating to the custom function or underlying writer.
// Returns the number of bytes written and any error.
//
// Behavior:
//   - If custom function returns nil: returns (0, io.ErrUnexpectedEOF)
//   - If custom function returns data: returns (len(data), nil)
//   - If no custom function and no underlying io.Writer: returns (0, io.ErrUnexpectedEOF)
//
// The returned byte count is the length of the slice returned by the custom function,
// which should represent the bytes successfully written.
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (o *iow) Write(p []byte) (n int, err error) {
	var r []byte

	// Load custom function atomically
	fn := o.w.Load()
	if fn != nil {
		r = fn(p)
	} else {
		r = o.fakeWrite(p)
	}

	// nil indicates error
	if r == nil {
		return 0, io.ErrUnexpectedEOF
	}

	// For Write, we return the length of the data we processed (r)
	// The custom function should return the bytes that were written
	return len(r), nil
}

// Seek implements io.Seeker by delegating to the custom function or underlying seeker.
// Returns the new position and any error.
//
// Behavior:
//   - If custom function is set: returns result from custom function
//   - If no custom function and underlying implements io.Seeker: delegates to it
//   - Otherwise: returns (0, io.ErrUnexpectedEOF)
//
// The whence parameter should be one of io.SeekStart, io.SeekCurrent, or io.SeekEnd.
//
// This method is thread-safe and can be called concurrently from multiple goroutines.
func (o *iow) Seek(offset int64, whence int) (int64, error) {
	// Load custom function atomically
	fn := o.s.Load()
	if fn != nil {
		return fn(offset, whence)
	}
	return o.fakeSeek(offset, whence)
}

// Close implements io.Closer by delegating to the custom function or underlying closer.
// Returns any error from the close operation.
//
// Behavior:
//   - If custom function is set: returns result from custom function
//   - If no custom function and underlying implements io.Closer: delegates to it
//   - Otherwise: returns nil (graceful degradation for non-closeable objects)
//
// This method is thread-safe and can be called concurrently from multiple goroutines,
// though calling Close multiple times may have undefined behavior depending on the
// underlying object.
func (o *iow) Close() error {
	// Load custom function atomically
	fn := o.c.Load()
	if fn != nil {
		return fn()
	}
	return o.fakeClose()
}

// fakeRead is the default read function that delegates to the underlying io.Reader.
// Returns nil if the underlying object doesn't implement io.Reader, which will cause
// Read() to return io.ErrUnexpectedEOF.
//
// If the underlying reader returns an error (including io.EOF), the error is encapsulated
// by returning the partial data read (p[:n]). The wrapper's Read method will return this
// data with no error if n > 0.
func (o *iow) fakeRead(p []byte) []byte {
	if r, k := o.i.(io.Reader); k {
		// Delegate to underlying reader; errors are encapsulated in the slice length
		n, _ := r.Read(p)
		return p[:n]
	}
	// No underlying reader available
	return nil
}

// fakeWrite is the default write function that delegates to the underlying io.Writer.
// Returns nil if the underlying object doesn't implement io.Writer, which will cause
// Write() to return io.ErrUnexpectedEOF.
//
// If the underlying writer returns an error, the error is encapsulated by returning
// the partial data written (p[:n]). The wrapper's Write method will return len(p[:n])
// with no error if n > 0.
func (o *iow) fakeWrite(p []byte) []byte {
	if r, k := o.i.(io.Writer); k {
		// Delegate to underlying writer; errors are encapsulated in the slice length
		n, _ := r.Write(p)
		return p[:n]
	}
	// No underlying writer available
	return nil
}

// fakeSeek is the default seek function that delegates to the underlying io.Seeker.
// Returns io.ErrUnexpectedEOF if the underlying object doesn't implement io.Seeker.
func (o *iow) fakeSeek(offset int64, whence int) (int64, error) {
	if r, k := o.i.(io.Seeker); k {
		// Delegate to underlying seeker
		return r.Seek(offset, whence)
	}
	// No underlying seeker available
	return 0, io.ErrUnexpectedEOF
}

// fakeClose is the default close function that delegates to the underlying io.Closer.
// Returns nil if the underlying object doesn't implement io.Closer (no error on non-closeable objects).
func (o *iow) fakeClose() error {
	if r, k := o.i.(io.Closer); k {
		// Delegate to underlying closer
		return r.Close()
	}
	// No underlying closer available; graceful degradation
	return nil
}
