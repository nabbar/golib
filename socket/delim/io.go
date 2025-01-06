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
	if o == nil || o.r == nil {
		return 0, ErrInstance
	}

	b, e := o.r.ReadBytes(byte(o.d))

	if len(b) > 0 {
		if cap(p) < len(b) {
			p = append(p, make([]byte, len(b)-len(p))...)
		}
		copy(p, b)
	}

	return len(b), e
}

func (o *dlm) UnRead() ([]byte, error) {
	if o == nil || o.r == nil {
		return nil, ErrInstance
	}

	if s := o.r.Buffered(); s > 0 {
		b := make([]byte, s)
		_, e := o.r.Read(b)
		return b, e
	}

	return nil, nil
}

func (o *dlm) ReadBytes() ([]byte, error) {
	if o.r == nil {
		return nil, ErrInstance
	}

	return o.r.ReadBytes(byte(o.d))
}

func (o *dlm) Close() error {
	o.r.Reset(nil)
	o.r = nil

	return o.i.Close()
}

func (o *dlm) WriteTo(w io.Writer) (n int64, err error) {
	var (
		e error
		i int
		b []byte

		s = 1
		d = o.getDelimByte()
	)

	if o.r == nil {
		return 0, ErrInstance
	}

	for err == nil {
		b, err = o.r.ReadBytes(d)
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
