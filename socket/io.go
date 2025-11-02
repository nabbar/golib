/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package socket

import (
	"fmt"
	"io"
)

var (
	// ErrInvalidInstance is returned when an operation is attempted on a nil or invalid instance.
	ErrInvalidInstance = fmt.Errorf("invalid instance")
	// closedChanStruct is a pre-closed channel returned by Done() when the instance is invalid.
	closedChanStruct chan struct{}
)

func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

// FctWriter is a function type that writes data to a destination.
// It follows the same signature as io.Writer.Write.
type FctWriter func(p []byte) (n int, err error)

// FctReader is a function type that reads data from a source.
// It follows the same signature as io.Reader.Read.
type FctReader func(p []byte) (n int, err error)

// FctClose is a function type that closes a resource.
// It follows the same signature as io.Closer.Close.
type FctClose func() error

// FctCheck is a function type that checks if a connection is active.
// It returns true if the connection is established, false otherwise.
type FctCheck func() bool

// FctDone is a function type that returns a channel for shutdown signaling.
// The returned channel is closed when the resource is closed or invalid.
type FctDone func() <-chan struct{}

// Reader extends io.ReadCloser with connection state tracking.
// It provides methods to check if the underlying connection is still active
// and to receive notification when the connection is closed.
type Reader interface {
	io.ReadCloser
	// IsConnected returns true if the underlying connection is active.
	IsConnected() bool
	// Done returns a channel that is closed when the reader is closed or becomes invalid.
	Done() <-chan struct{}
}

// Writer extends io.WriteCloser with connection state tracking.
// It provides methods to check if the underlying connection is still active
// and to receive notification when the connection is closed.
type Writer interface {
	io.WriteCloser
	// IsConnected returns true if the underlying connection is active.
	IsConnected() bool
	// Done returns a channel that is closed when the writer is closed or becomes invalid.
	Done() <-chan struct{}
}

// wrt is the internal implementation of the Writer interface.
// It wraps function callbacks to provide io.WriteCloser functionality
// with connection state tracking.
type wrt struct {
	w FctWriter // write function
	c FctClose  // close function
	d FctDone   // done channel function
	i FctCheck  // connection check function
}

// Write writes data using the configured write function.
// It returns ErrInvalidInstance if the instance is nil or unconfigured.
func (o *wrt) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInvalidInstance
	} else if o.w == nil {
		return 0, ErrInvalidInstance
	} else {
		return o.w(p)
	}
}

// Close closes the writer using the configured close function.
// It returns ErrInvalidInstance if the instance is nil or unconfigured.
func (o *wrt) Close() error {
	if o == nil {
		return ErrInvalidInstance
	} else if o.w == nil {
		return ErrInvalidInstance
	} else {
		return o.c()
	}
}

// IsConnected checks if the connection is active using the configured check function.
// It returns false if the instance is nil or unconfigured.
func (o *wrt) IsConnected() bool {
	if o == nil {
		return false
	} else if o.i == nil {
		return false
	} else {
		return o.i()
	}
}

// Done returns a channel that is closed when the writer is closed.
// If the instance is nil or unconfigured, it returns a pre-closed channel.
func (o *wrt) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	} else if o.d == nil {
		return closedChanStruct
	} else {
		return o.d()
	}
}

// rdr is the internal implementation of the Reader interface.
// It wraps function callbacks to provide io.ReadCloser functionality
// with connection state tracking.
type rdr struct {
	r FctReader // read function
	c FctClose  // close function
	d FctDone   // done channel function
	i FctCheck  // connection check function
}

// Read reads data using the configured read function.
// It returns ErrInvalidInstance if the instance is nil or unconfigured.
func (o *rdr) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInvalidInstance
	} else if o.r == nil {
		return 0, ErrInvalidInstance
	} else {
		return o.r(p)
	}
}

// Close closes the reader using the configured close function.
// It returns ErrInvalidInstance if the instance is nil or unconfigured.
func (o *rdr) Close() error {
	if o == nil {
		return ErrInvalidInstance
	} else if o.c == nil {
		return ErrInvalidInstance
	} else {
		return o.c()
	}
}

// IsConnected checks if the connection is active using the configured check function.
// It returns false if the instance is nil or unconfigured.
func (o *rdr) IsConnected() bool {
	if o == nil {
		return false
	} else if o.i == nil {
		return false
	} else {
		return o.i()
	}
}

// Done returns a channel that is closed when the reader is closed.
// If the instance is nil or unconfigured, it returns a pre-closed channel.
func (o *rdr) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	} else if o.d == nil {
		return closedChanStruct
	} else {
		return o.d()
	}
}

// NewReader creates a new Reader instance from the provided callback functions.
// This allows wrapping custom read/close logic into a Reader interface.
//
// Parameters:
//   - fctRead: function to read data
//   - fctClose: function to close the reader
//   - fctCheck: function to check connection status
//   - fctDone: function to get the done channel
//
// Returns a Reader that delegates to the provided functions.
func NewReader(fctRead FctReader, fctClose FctClose, fctCheck FctCheck, fctDone FctDone) Reader {
	return &rdr{
		r: fctRead,
		c: fctClose,
		d: fctDone,
		i: fctCheck,
	}
}

// NewWriter creates a new Writer instance from the provided callback functions.
// This allows wrapping custom write/close logic into a Writer interface.
//
// Parameters:
//   - fctWrite: function to write data
//   - fctClose: function to close the writer
//   - fctCheck: function to check connection status
//   - fctDone: function to get the done channel
//
// Returns a Writer that delegates to the provided functions.
func NewWriter(fctWrite FctWriter, fctClose FctClose, fctCheck FctCheck, fctDone FctDone) Writer {
	return &wrt{
		w: fctWrite,
		c: fctClose,
		d: fctDone,
		i: fctCheck,
	}
}
