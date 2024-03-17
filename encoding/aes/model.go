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

package aes

import (
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidBufferSize = errors.New("invalid buffer size")
)

type crt struct {
	a cipher.AEAD
	n []byte
}

func (o *crt) Encode(p []byte) []byte {
	if len(p) < 1 {
		return make([]byte, 0)
	}
	return o.a.Seal(nil, o.n, p, nil)
}

func (o *crt) Decode(p []byte) ([]byte, error) {
	if len(p) < 1 {
		return make([]byte, 0), nil
	}
	return o.a.Open(nil, o.n, p, nil)
}

func (o *crt) EncodeReader(r io.Reader) io.Reader {
	fct := func(p []byte) (n int, err error) {
		var b []byte

		if cap(p) < 1+o.a.Overhead() {
			return 0, ErrInvalidBufferSize
		} else {
			b = make([]byte, cap(p)-o.a.Overhead())
		}

		if n, err = r.Read(b); err != nil {
			return 0, err
		} else {
			b = o.Encode(b[:n])
			n = len(b)

			if n > cap(p) {
				return 0, ErrInvalidBufferSize
			} else {
				copy(p, b)
			}

			clear(b)
			b = b[:0]

			return n, err
		}
	}

	return &reader{
		f: fct,
	}
}

func (o *crt) DecodeReader(r io.Reader) io.Reader {
	fct := func(p []byte) (n int, err error) {
		var b = make([]byte, cap(p)+o.a.Overhead())

		if n, err = r.Read(b); err != nil {
			return 0, err
		} else {
			b, err = o.Decode(b[:n])
			n = len(b)

			if n > cap(p) {
				return 0, ErrInvalidBufferSize
			} else {
				copy(p, b)
			}

			clear(b)
			b = b[:0]

			return n, err
		}
	}

	return &reader{
		f: fct,
	}
}

func (o *crt) EncodeWriter(w io.Writer) io.Writer {
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

func (o *crt) DecodeWriter(w io.Writer) io.Writer {
	fct := func(p []byte) (n int, err error) {
		n = len(p)
		if p, err = o.Decode(p); err != nil {
			return 0, err
		} else if _, err = w.Write(p); err != nil {
			return 0, err
		} else {
			return n, nil
		}
	}

	return &writer{
		f: fct,
	}
}

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
