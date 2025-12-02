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

package progress

import (
	"errors"
	"io"
)

// ReadByte reads and returns a single byte from the file.
// It implements the io.ByteReader interface.
// The function preserves file position by seeking if more than one byte is read.
// Returns the byte and any error encountered, including io.EOF at end of file.
func (o *progress) ReadByte() (byte, error) {
	var (
		p = make([]byte, 1)
		i int64
		n int
		e error
	)

	if i, e = o.f.Seek(0, io.SeekCurrent); e != nil && !errors.Is(e, io.EOF) {
		return 0, e
	} else if n, e = o.f.Read(p); e != nil && !errors.Is(e, io.EOF) {
		return 0, e
	} else if n > 1 {
		if _, e = o.f.Seek(i+1, io.SeekStart); e != nil && !errors.Is(e, io.EOF) {
			return 0, e
		}
	} else if n == 0 {
		return 0, e
	}

	return p[0], nil
}

// WriteByte writes a single byte to the file.
// It implements the io.ByteWriter interface.
// The function preserves file position by seeking if more than one byte is written.
// Returns any error encountered during the write operation.
func (o *progress) WriteByte(c byte) error {
	var (
		p = []byte{0: c}
		i int64
		n int
		e error
	)

	if i, e = o.f.Seek(0, io.SeekCurrent); e != nil && !errors.Is(e, io.EOF) {
		return e
	} else if n, e = o.f.Write(p); e != nil && !errors.Is(e, io.EOF) {
		return e
	} else if n > 1 {
		if _, e = o.f.Seek(i+1, io.SeekStart); e != nil && !errors.Is(e, io.EOF) {
			return e
		}
	}

	return nil
}
