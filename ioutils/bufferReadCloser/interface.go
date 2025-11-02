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

package bufferReadCloser

import (
	"bufio"
	"bytes"
	"io"
)

// FuncClose is an optional custom close function that is called when a wrapper is closed.
// It allows for additional cleanup logic beyond the default reset behavior.
type FuncClose func() error

// Buffer is a wrapper around bytes.Buffer that implements io.Closer.
// It provides all the standard buffer interfaces with automatic reset on close.
type Buffer interface {
	io.Reader
	io.ReaderFrom
	io.ByteReader
	io.RuneReader
	io.Writer
	io.WriterTo
	io.ByteWriter
	io.StringWriter
	io.Closer
}

// New creates a new Buffer from a bytes.Buffer without a custom close function.
// Deprecated: use NewBuffer instead of New.
func New(b *bytes.Buffer) Buffer {
	return NewBuffer(b, nil)
}

// NewBuffer creates a new Buffer from a bytes.Buffer and an optional
// FuncClose. If FuncClose is not nil, it is called when the Buffer is
// closed.
func NewBuffer(b *bytes.Buffer, fct FuncClose) Buffer {
	return &buf{
		b: b,
		f: fct,
	}
}

// Reader is a wrapper around bufio.Reader that implements io.Closer.
// It provides read operations with automatic reset on close.
type Reader interface {
	io.Reader
	io.WriterTo
	io.Closer
}

// NewReader creates a new Reader from a bufio.Reader and an optional
// FuncClose. If FuncClose is not nil, it is called when the Reader is
// closed.
func NewReader(b *bufio.Reader, fct FuncClose) Reader {
	return &rdr{
		b: b,
		f: fct,
	}
}

// Writer is a wrapper around bufio.Writer that implements io.Closer.
// It provides write operations with automatic flush and reset on close.
type Writer interface {
	io.Writer
	io.StringWriter
	io.ReaderFrom
	io.Closer
}

// NewWriter creates a new Writer from a bufio.Writer and an optional
// FuncClose. If FuncClose is not nil, it is called when the Writer is
// closed.
func NewWriter(b *bufio.Writer, fct FuncClose) Writer {
	return &wrt{
		b: b,
		f: fct,
	}
}

// ReadWriter is a wrapper around bufio.ReadWriter that implements io.Closer.
// It combines Reader and Writer interfaces with automatic flush on close.
// Note: Reset is not called on close due to ambiguous method in bufio.ReadWriter.
type ReadWriter interface {
	Reader
	Writer
}

// NewReadWriter creates a new ReadWriter from a bufio.ReadWriter and an optional
// FuncClose. If FuncClose is not nil, it is called when the ReadWriter is closed.
//
// The ReadWriter implements both Reader and Writer interfaces, providing bidirectional
// buffered I/O with automatic flush on close. Note that Reset cannot be called on
// close due to the ambiguous Reset method in bufio.ReadWriter (both Reader and Writer
// have Reset methods).
func NewReadWriter(b *bufio.ReadWriter, fct FuncClose) ReadWriter {
	return &rwt{
		b: b,
		f: fct,
	}
}
