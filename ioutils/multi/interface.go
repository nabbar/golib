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

package multi

import (
	"io"
	"sync"
	"sync/atomic"
)

// Multi is a thread-safe I/O multiplexer that broadcasts writes to multiple
// destinations and manages a single input source.
//
// The Multi interface extends io.ReadWriteCloser and io.StringWriter with
// methods for managing multiple write destinations and an input source.
// All methods are safe for concurrent use.
//
// Write operations (Write, WriteString) are broadcast to all registered
// writers via io.MultiWriter. Read operations are performed on the currently
// set input source.
//
// Example usage:
//
//	m := multi.New()
//	var buf1, buf2 bytes.Buffer
//	m.AddWriter(&buf1, &buf2)
//	m.Write([]byte("data")) // written to both buf1 and buf2
type Multi interface {
	io.ReadWriteCloser
	io.StringWriter

	// Clean removes all registered writers and resets the multiplexer to
	// use io.Discard. After calling Clean, subsequent writes will be discarded
	// until new writers are added via AddWriter.
	Clean()

	// AddWriter adds one or more io.Writer destinations to the multiplexer.
	// All subsequent Write and WriteString operations will be broadcast to
	// these writers in addition to any previously added writers.
	// Nil writers are silently skipped.
	//
	// AddWriter is safe to call concurrently.
	AddWriter(w ...io.Writer)

	// SetInput sets or replaces the input source for read operations.
	// If i is nil, a DiscardCloser is used as the default.
	//
	// SetInput is safe to call concurrently, though note that the underlying
	// io.ReadCloser itself may not be safe for concurrent reads.
	SetInput(i io.ReadCloser)

	// Reader returns the current input source.
	// Returns the io.ReadCloser set via SetInput, or DiscardCloser if none was set.
	Reader() io.ReadCloser

	// Writer returns the current write destination.
	// This is typically an io.MultiWriter wrapping all registered writers,
	// or io.Discard if no writers have been added.
	Writer() io.Writer

	// Copy copies data from the input source (Reader) to all registered
	// write destinations (Writer). It's a convenience wrapper around io.Copy.
	//
	// Returns the number of bytes copied and any error encountered.
	// The copy stops at EOF or the first error.
	Copy() (n int64, err error)
}

// New creates and initializes a new Multi instance.
//
// The returned Multi is ready for use and safe for concurrent operations.
// It is initialized with:
//   - DiscardCloser as the default input source (reads return 0 bytes)
//   - io.Discard as the default output destination (writes are discarded)
//
// You should add writers via AddWriter and optionally set an input via
// SetInput before performing I/O operations.
//
// Example:
//
//	m := multi.New()
//	var output bytes.Buffer
//	m.AddWriter(&output)
//	m.Write([]byte("hello"))
func New() Multi {
	m := &mlt{
		i: new(atomic.Value),
		d: new(atomic.Value),
		c: new(atomic.Int64),
		w: sync.Map{},
	}
	// Initialize with safe defaults to prevent panics
	m.i.Store(&readerWrapper{ReadCloser: DiscardCloser{}})
	m.d.Store(io.MultiWriter(io.Discard))
	return m
}
