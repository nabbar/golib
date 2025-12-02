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
	"io"
)

// rwt is the internal implementation of the ReadWriter interface.
// It wraps a bufio.ReadWriter with optional close functionality.
type rwt struct {
	b *bufio.ReadWriter
	f FuncClose
}

// Read reads up to len(p) bytes from the reader.
func (b *rwt) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

// WriteTo writes data to w until the reader is drained or an error occurs.
func (b *rwt) WriteTo(w io.Writer) (n int64, err error) {
	return b.b.WriteTo(w)
}

// ReadFrom reads data from r until EOF and writes it to the writer.
func (b *rwt) ReadFrom(r io.Reader) (n int64, err error) {
	return b.b.ReadFrom(r)
}

// Write writes len(p) bytes from p to the underlying writer.
// The data is buffered and may not be immediately visible until flush.
func (b *rwt) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

// WriteString writes the contents of s to the writer.
// The data is buffered and may not be immediately visible until flush.
func (b *rwt) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}

// Close flushes any buffered write data and calls the custom close function if provided.
// Note: Reset is not called because bufio.ReadWriter has ambiguous Reset methods
// (both Reader and Writer have Reset, which one would be called?).
// Returns any error from flush or the custom close function.
func (b *rwt) Close() error {
	e := b.b.Flush()

	// Cannot call Reset => ambiguous method (Reader.Reset or Writer.Reset?)
	// b.b.Reset(nil)

	if b.f != nil {
		return b.f()
	}

	return e
}
