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

	"github.com/nabbar/golib/size"
)

// Dlm is an alias to the internal dlm struct, exposed for testing purposes.
type Dlm = dlm

// NewInternal creates a new Dlm with specific internal state for testing.
func NewInternal(r io.ReadCloser, s size.Size, buffer []byte) *Dlm {
	return &dlm{
		i: r,
		s: s,
		b: buffer,
	}
}

// Fill exposes the private fill method for testing.
func (o *Dlm) Fill() error {
	return o.fill()
}

// ReadBuf exposes the private readBuf method for testing.
func (o *Dlm) ReadBuf(p []byte) (int, error) {
	return o.readBuf(p)
}
