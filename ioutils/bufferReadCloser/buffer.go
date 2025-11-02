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
	"bytes"
	"io"
)

// buf is the internal implementation of the Buffer interface.
// It wraps a bytes.Buffer with optional close functionality.
type buf struct {
	b *bytes.Buffer
	f FuncClose
}

// Read reads up to len(p) bytes from the buffer.
func (b *buf) Read(p []byte) (n int, err error) {
	return b.b.Read(p)
}

// ReadFrom reads data from r until EOF and appends it to the buffer.
func (b *buf) ReadFrom(r io.Reader) (n int64, err error) {
	return b.b.ReadFrom(r)
}

// ReadByte reads and returns the next byte from the buffer.
func (b *buf) ReadByte() (byte, error) {
	return b.b.ReadByte()
}

// ReadRune reads and returns the next UTF-8 encoded Unicode character from the buffer.
func (b *buf) ReadRune() (r rune, size int, err error) {
	return b.b.ReadRune()
}

// Write appends the contents of p to the buffer.
func (b *buf) Write(p []byte) (n int, err error) {
	return b.b.Write(p)
}

// WriteString appends the contents of s to the buffer.
func (b *buf) WriteString(s string) (n int, err error) {
	return b.b.WriteString(s)
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
func (b *buf) WriteTo(w io.Writer) (n int64, err error) {
	return b.b.WriteTo(w)
}

// WriteByte appends the byte c to the buffer.
func (b *buf) WriteByte(c byte) error {
	return b.b.WriteByte(c)
}

// Close resets the buffer (clears all data) and calls the custom close function if provided.
// Returns any error from the custom close function.
func (b *buf) Close() error {
	b.b.Reset()

	if b.f != nil {
		return b.f()
	}

	return nil
}
