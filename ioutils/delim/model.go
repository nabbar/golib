/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package delim

import (
	"io"
	"sync"

	libsiz "github.com/nabbar/golib/size"
)

// dlm is the internal implementation of the BufferDelim interface.
// It wraps an io.ReadCloser with a buffered reader and tracks the delimiter character.
//
// Fields:
//   - i: The underlying io.ReadCloser that provides the input stream
//   - r: The delimiter rune used to separate data chunks
//   - b: The internal byte buffer used for reading chunks
//   - s: The maximum size of the buffer
//   - d: Flag indicating whether to discard data on buffer overflow
//
// The struct is not exported to maintain encapsulation and allow future implementation changes
// without breaking the public API.
type dlm struct {
	m sync.Mutex
	i io.ReadCloser // input io.ReadCloser
	r byte          // delimiter rune character
	b []byte        // buffer
	s libsiz.Size   // size of buffer
	d bool          // if max size is reached, discard overflow or return error
}

// Delim returns the delimiter rune configured for this BufferDelim instance.
// This value is set during construction via New() and remains constant for the lifetime of the instance.
func (o *dlm) Delim() rune {
	return rune(o.r)
}
