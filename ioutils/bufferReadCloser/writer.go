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

// wrt is the internal implementation of the Writer interface.
// It wraps a bufio.Writer with optional close functionality.
type wrt struct {
	b *bufio.Writer
	f FuncClose
}

// ReadFrom reads data from r until EOF and writes it to the writer.
func (b *wrt) ReadFrom(r io.Reader) (n int64, err error) {
	return b.b.ReadFrom(r)
}

// Write writes len(p) bytes from p to the underlying writer.
// The data is buffered and may not be immediately visible until flush.
func (b *wrt) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

// WriteString writes the contents of s to the writer.
// The data is buffered and may not be immediately visible until flush.
func (b *wrt) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}

// Close flushes any buffered data, resets the writer (releases resources),
// and calls the custom close function if provided.
// Returns any error from the custom close function.
func (b *wrt) Close() error {
	_ = b.b.Flush()
	b.b.Reset(nil)

	if b.f != nil {
		return b.f()
	}

	return nil
}
