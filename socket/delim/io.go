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

package delim

import "io"

func (o *dlm) Reader() io.ReadCloser {
	return o
}

func (o *dlm) Copy(w io.Writer) (n int64, err error) {
	return o.WriteTo(w)
}

func (o *dlm) Read(p []byte) (n int, err error) {
	if r := o.getReader(); r == nil {
		return 0, ErrInstance
	} else {
		b, e := r.ReadBytes(o.getDelimByte())

		if len(b) > 0 {
			if cap(p) < len(b) {
				p = append(p, make([]byte, len(b)-len(p))...)
			}
			copy(p, b)
		}

		return len(b), e
	}
}

func (o *dlm) ReadBytes() ([]byte, error) {
	if r := o.getReader(); r == nil {
		return make([]byte, 0), ErrInstance
	} else {
		return r.ReadBytes(o.getDelimByte())
	}
}

func (o *dlm) Close() error {
	return o.getInput().Close()
}

func (o *dlm) WriteTo(w io.Writer) (n int64, err error) {
	var (
		i int
		s = 1
		e error
		b []byte
	)

	for err == nil && s > 0 {
		b, err = o.ReadBytes()
		s = len(b)

		if s > 0 {
			i, e = w.Write(b)
			n += int64(i)
		}

		b = b[:0]
		b = nil

		if err == nil && e != nil {
			err = e
		}
	}

	return n, err
}
