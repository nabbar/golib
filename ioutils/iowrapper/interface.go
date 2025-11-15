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

// Package iowrapper provides a flexible I/O wrapper that enables customization of read, write, seek,
// and close operations on any underlying I/O object. It implements standard Go interfaces (io.Reader,
// io.Writer, io.Seeker, io.Closer) and allows intercepting, transforming, or monitoring I/O operations
// without modifying the underlying implementation.
package iowrapper

import (
	"io"

	libatm "github.com/nabbar/golib/atomic"
)

// FuncRead is a custom read function that receives a buffer and returns the data read.
// Return nil to signal EOF or error, empty slice for 0 bytes read, or a slice with the data.
type FuncRead func(p []byte) []byte

// FuncWrite is a custom write function that receives data to write and returns bytes written.
// Return nil to signal error, or a slice representing the bytes that were written.
type FuncWrite func(p []byte) []byte

// FuncSeek is a custom seek function that repositions the offset.
// It receives the offset and whence parameters and returns the new position and any error.
type FuncSeek func(offset int64, whence int) (int64, error)

// FuncClose is a custom close function that performs cleanup operations.
// It returns an error if the close operation fails.
type FuncClose func() error

// IOWrapper is an interface that wraps basic I/O operations with customizable behavior.
// It implements all standard Go I/O interfaces and provides methods to set custom
// functions for each operation. All operations are thread-safe.
type IOWrapper interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer

	// SetRead sets a custom read function for this wrapper.
	// The function will be called on every Read operation. Pass nil to reset to default behavior
	// (delegates to underlying io.Reader if available, or returns io.ErrUnexpectedEOF).
	//
	// Custom function behavior:
	//   - Return nil to signal EOF/error (Read returns io.ErrUnexpectedEOF)
	//   - Return empty slice for 0 bytes read
	//   - Return slice with data (will be copied to caller's buffer)
	//
	// Thread-safe: Can be called concurrently with Read operations.
	SetRead(read FuncRead)

	// SetWrite sets a custom write function for this wrapper.
	// The function will be called on every Write operation. Pass nil to reset to default behavior
	// (delegates to underlying io.Writer if available, or returns io.ErrUnexpectedEOF).
	//
	// Custom function behavior:
	//   - Return nil to signal error (Write returns io.ErrUnexpectedEOF)
	//   - Return slice representing bytes written (len determines bytes written count)
	//
	// Thread-safe: Can be called concurrently with Write operations.
	SetWrite(write FuncWrite)

	// SetSeek sets a custom seek function for this wrapper.
	// The function will be called on every Seek operation. Pass nil to reset to default behavior
	// (delegates to underlying io.Seeker if available, or returns io.ErrUnexpectedEOF).
	//
	// Thread-safe: Can be called concurrently with Seek operations.
	SetSeek(seek FuncSeek)

	// SetClose sets a custom close function for this wrapper.
	// The function will be called on Close operation. Pass nil to reset to default behavior
	// (delegates to underlying io.Closer if available, or returns nil).
	//
	// Thread-safe: Can be called concurrently with Close operations.
	SetClose(close FuncClose)
}

// New creates a new IOWrapper that wraps the given object.
//
// The wrapper will delegate I/O operations to the underlying object if it implements
// the corresponding interfaces (io.Reader, io.Writer, io.Seeker, io.Closer).
// If the underlying object doesn't implement an interface, the operation will return
// an appropriate error (io.ErrUnexpectedEOF for Read/Write/Seek, nil for Close).
//
// Parameters:
//   - in: Any object to wrap (can be nil, io.Reader, io.Writer, io.Seeker, io.Closer, or any combination)
//
// Returns:
//   - IOWrapper: A new wrapper instance with default behavior
//
// The wrapper is fully thread-safe. Custom functions can be set using SetRead, SetWrite,
// SetSeek, and SetClose methods to intercept and customize I/O operations.
//
// Example:
//
//	wrapper := iowrapper.New(bytes.NewBuffer([]byte("data")))
//	wrapper.SetRead(func(p []byte) []byte {
//	    // Custom read logic
//	    return []byte("custom data")
//	})
func New(in any) IOWrapper {
	return &iow{
		i: in,
		r: libatm.NewValue[FuncRead](),
		w: libatm.NewValue[FuncWrite](),
		s: libatm.NewValue[FuncSeek](),
		c: libatm.NewValue[FuncClose](),
	}
}
