/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package crypt

import (
	"fmt"
	"io"
)

type reader struct {
	f func(p []byte) (n int, err error)
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.f == nil {
		return 0, fmt.Errorf("invalid reader")
	} else {
		return r.f(p)
	}
}

func (o *crt) Reader(r io.Reader) io.Reader {
	fct := func(p []byte) (n int, err error) {
		var (
			a = make([]byte, 0, cap(p))
			b = make([]byte, cap(p)+o.a.Overhead())
		)

		if n, err = r.Read(b); err != nil {
			return 0, err
		} else if a, err = o.Decode(b[:n]); err != nil {
			return 0, err
		} else {
			copy(p, a)
			n = len(a)
			clear(a)
			clear(b)
			return n, nil
		}
	}

	return &reader{
		f: fct,
	}
}

func (o *crt) ReaderHex(r io.Reader) io.Reader {
	fct := func(p []byte) (n int, err error) {
		var (
			a = make([]byte, 0, cap(p))
			b = make([]byte, (cap(p)+o.a.Overhead())*2)
		)

		if n, err = r.Read(b); err != nil {
			return 0, err
		} else if a, err = o.DecodeHex(b[:n]); err != nil {
			return 0, err
		} else {
			copy(p, a)
			n = len(a)
			clear(a)
			clear(b)
			return n, nil
		}
	}

	return &reader{
		f: fct,
	}
}
