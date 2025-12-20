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
//
// The function is called after the wrapper's internal cleanup (flush, reset) but before
// returning from Close(). Any error returned by FuncClose is propagated to the caller.
//
// Common use cases:
//   - Closing underlying file handles or network connections
//   - Returning buffers to sync.Pool
//   - Updating metrics or logging
//   - Releasing external resources
//
// Example:
//
//	file, _ := os.Open("data.txt")
//	reader := NewReader(bufio.NewReader(file), file.Close)
//	defer reader.Close() // Closes both reader and file
type FuncClose func() error

// Buffer is a wrapper around bytes.Buffer that implements io.Closer.
// It provides all the standard buffer interfaces with automatic reset on close.
//
// The Buffer interface combines reading and writing capabilities with lifecycle management.
// When Close() is called, the underlying buffer is reset (all data cleared) and any
// custom close function is executed.
//
// All I/O operations are delegated directly to the underlying bytes.Buffer with zero
// overhead. The wrapper only adds the Close() method for lifecycle management.
//
// Thread safety: Not thread-safe. Concurrent access requires external synchronization.
type Buffer interface {
	io.Reader       // Read reads data from the buffer
	io.ReaderFrom   // ReadFrom reads data from a reader into the buffer
	io.ByteReader   // ReadByte reads a single byte
	io.RuneReader   // ReadRune reads a single UTF-8 encoded rune
	io.Writer       // Write writes data to the buffer
	io.WriterTo     // WriteTo writes buffer data to a writer
	io.ByteWriter   // WriteByte writes a single byte
	io.StringWriter // WriteString writes a string
	io.Closer       // Close resets the buffer and calls custom close function
}

// New creates a new Buffer from a bytes.Buffer without a custom close function.
//
// Deprecated: use NewBuffer instead of New. This function is maintained for
// backward compatibility but NewBuffer provides more flexibility with the
// optional FuncClose parameter.
func New(b *bytes.Buffer) Buffer {
	return NewBuffer(b, nil)
}

// NewBuffer creates a new Buffer from a bytes.Buffer and an optional FuncClose.
//
// Parameters:
//   - b: The underlying bytes.Buffer to wrap. If nil, a new empty buffer is created.
//   - fct: Optional custom close function. If not nil, called after buffer reset.
//
// The returned Buffer delegates all I/O operations to the underlying bytes.Buffer.
// On Close(), the buffer is reset (cleared) and then fct is called if provided.
//
// Nil handling: Passing nil for b creates a new empty buffer, allowing immediate use
// without additional initialization. This is useful for testing or when a buffer is
// conditionally needed.
//
// Example:
//
//	buf := NewBuffer(bytes.NewBuffer(nil), nil)
//	defer buf.Close()
//	buf.WriteString("data")
func NewBuffer(b *bytes.Buffer, fct FuncClose) Buffer {
	if b == nil {
		b = bytes.NewBuffer([]byte{})
	}
	return &buf{
		b: b,
		f: fct,
	}
}

// Reader is a wrapper around bufio.Reader that implements io.Closer.
// It provides read operations with automatic reset on close.
//
// The Reader interface provides buffered reading with lifecycle management.
// When Close() is called, the underlying reader is reset (buffered data released)
// and any custom close function is executed.
//
// Typical use case: Reading from files or network connections where you want
// to ensure both the buffer and the underlying resource are properly cleaned up.
//
// Thread safety: Not thread-safe. Concurrent access requires external synchronization.
type Reader interface {
	io.Reader   // Read reads data from the buffered reader
	io.WriterTo // WriteTo writes buffered data to a writer
	io.Closer   // Close resets the reader and calls custom close function
}

// NewReader creates a new Reader from a bufio.Reader and an optional FuncClose.
//
// Parameters:
//   - b: The underlying bufio.Reader to wrap. If nil, creates a reader from an empty source.
//   - fct: Optional custom close function. If not nil, called after reader reset.
//
// The returned Reader delegates all read operations to the underlying bufio.Reader.
// On Close(), the reader is reset (buffered data released) and then fct is called if provided.
//
// Nil handling: Passing nil for b creates a reader from an empty source that immediately
// returns io.EOF on any read operation. This is useful for testing or placeholder scenarios.
//
// Common pattern for file reading:
//
//	file, _ := os.Open("data.txt")
//	reader := NewReader(bufio.NewReader(file), file.Close)
//	defer reader.Close() // Closes both reader and file
func NewReader(b *bufio.Reader, fct FuncClose) Reader {
	if b == nil {
		b = bufio.NewReader(bytes.NewReader([]byte{}))
	}
	return &rdr{
		b: b,
		f: fct,
	}
}

// Writer is a wrapper around bufio.Writer that implements io.Closer.
// It provides write operations with automatic flush and reset on close.
//
// The Writer interface provides buffered writing with guaranteed flush on close.
// When Close() is called, buffered data is flushed, the writer is reset, and any
// custom close function is executed.
//
// Important: Data written to a Writer is buffered and may not be visible in the
// destination until Close() is called or the buffer is manually flushed.
//
// Typical use case: Writing to files or network connections where you want to
// ensure all buffered data is written and resources are properly cleaned up.
//
// Thread safety: Not thread-safe. Concurrent access requires external synchronization.
type Writer interface {
	io.Writer       // Write writes data to the buffered writer
	io.StringWriter // WriteString writes a string to the buffered writer
	io.ReaderFrom   // ReadFrom reads from a reader and writes to the buffered writer
	io.Closer       // Close flushes, resets the writer, and calls custom close function
}

// NewWriter creates a new Writer from a bufio.Writer and an optional FuncClose.
//
// Parameters:
//   - b: The underlying bufio.Writer to wrap. If nil, creates a writer to io.Discard.
//   - fct: Optional custom close function. If not nil, called after flush and reset.
//
// The returned Writer delegates all write operations to the underlying bufio.Writer.
// On Close(), buffered data is flushed (errors returned), the writer is reset, and
// then fct is called if provided.
//
// Nil handling: Passing nil for b creates a writer to io.Discard that accepts all
// writes without error but discards the data. This is useful for testing or when
// output needs to be silently ignored.
//
// Important: Close() now returns flush errors. Always check the error:
//
//	writer := NewWriter(bw, nil)
//	defer func() {
//	    if err := writer.Close(); err != nil {
//	        log.Printf("flush failed: %v", err)
//	    }
//	}()
func NewWriter(b *bufio.Writer, fct FuncClose) Writer {
	if b == nil {
		b = bufio.NewWriter(io.Discard)
	}
	return &wrt{
		b: b,
		f: fct,
	}
}

// ReadWriter is a wrapper around bufio.ReadWriter that implements io.Closer.
// It combines Reader and Writer interfaces with automatic flush on close.
//
// The ReadWriter interface provides bidirectional buffered I/O with lifecycle management.
// When Close() is called, buffered write data is flushed and any custom close function
// is executed.
//
// Limitation: Unlike Reader and Writer, ReadWriter cannot call Reset() on close because
// bufio.ReadWriter embeds both *Reader and *Writer, each with their own Reset() method,
// creating an ambiguous method call. This means the underlying readers/writers are not
// reset on close, only flushed.
//
// Typical use case: Network protocols or duplex communication channels where both
// reading and writing are needed with guaranteed flush on close.
//
// Thread safety: Not thread-safe. Concurrent access requires external synchronization.
type ReadWriter interface {
	Reader // Provides Read and WriteTo operations
	Writer // Provides Write, WriteString, and ReadFrom operations
}

// NewReadWriter creates a new ReadWriter from a bufio.ReadWriter and an optional FuncClose.
//
// Parameters:
//   - b: The underlying bufio.ReadWriter to wrap. If nil, creates a readwriter with
//     empty source (reads return EOF) and io.Discard destination (writes are discarded).
//   - fct: Optional custom close function. If not nil, called after flush.
//
// The returned ReadWriter delegates all I/O operations to the underlying bufio.ReadWriter.
// On Close(), buffered write data is flushed (errors returned) and then fct is called
// if provided.
//
// Limitation: Reset() is NOT called on close due to ambiguous methods in bufio.ReadWriter.
// This means the underlying readers/writers retain their state after close.
//
// Nil handling: Passing nil for b creates a readwriter where reads immediately return
// io.EOF and writes are silently discarded. This is useful for testing or placeholder
// scenarios.
//
// Common pattern for network connections:
//
//	conn, _ := net.Dial("tcp", "example.com:80")
//	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
//	wrapper := NewReadWriter(rw, conn.Close)
//	defer wrapper.Close() // Flushes and closes connection
func NewReadWriter(b *bufio.ReadWriter, fct FuncClose) ReadWriter {
	if b == nil {
		b = bufio.NewReadWriter(bufio.NewReader(bytes.NewReader([]byte{})), bufio.NewWriter(io.Discard))
	}
	return &rwt{
		b: b,
		f: fct,
	}
}
