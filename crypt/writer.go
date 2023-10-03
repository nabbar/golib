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

type writer struct {
	f func(p []byte) (n int, err error)
}

func (r *writer) Write(p []byte) (n int, err error) {
	if r.f == nil {
		return 0, fmt.Errorf("invalid reader")
	} else {
		return r.f(p)
	}
}

func (o *crt) Writer(w io.Writer) io.Writer {
	fct := func(p []byte) (n int, err error) {
		n = len(p)
		if _, err = w.Write(o.Encode(p)); err != nil {
			return 0, err
		} else {
			return n, nil
		}
	}

	return &writer{
		f: fct,
	}
}

func (o *crt) WriterHex(w io.Writer) io.Writer {
	fct := func(p []byte) (n int, err error) {
		n = len(p)
		if _, err = w.Write(o.EncodeHex(p)); err != nil {
			return 0, err
		} else {
			return n, nil
		}
	}

	return &writer{
		f: fct,
	}
}
