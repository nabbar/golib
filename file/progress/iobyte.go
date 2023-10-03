/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

func (o *progress) ReadByte() (byte, error) {
	var (
		p = make([]byte, 1)
		i int64
		n int
		e error
	)

	if i, e = o.fos.Seek(0, io.SeekCurrent); e != nil && !errors.Is(e, io.EOF) {
		return 0, e
	} else if n, e = o.fos.Read(p); e != nil && !errors.Is(e, io.EOF) {
		return 0, e
	} else if n > 1 {
		if _, e = o.fos.Seek(i+1, io.SeekStart); e != nil && !errors.Is(e, io.EOF) {
			return 0, e
		}
	} else if n == 0 {
		return 0, e
	}

	return p[0], nil
}

func (o *progress) WriteByte(c byte) error {
	var (
		p = []byte{0: c}
		i int64
		n int
		e error
	)

	if i, e = o.fos.Seek(0, io.SeekCurrent); e != nil && !errors.Is(e, io.EOF) {
		return e
	} else if n, e = o.fos.Write(p); e != nil && !errors.Is(e, io.EOF) {
		return e
	} else if n > 1 {
		if _, e = o.fos.Seek(i+1, io.SeekStart); e != nil && !errors.Is(e, io.EOF) {
			return e
		}
	}

	return nil
}
