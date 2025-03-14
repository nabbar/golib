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

package hexa

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidBufferSize = errors.New("invalid buffer size")
)

type crt struct{}

func (o *crt) Encode(p []byte) []byte {
	if len(p) < 1 {
		return make([]byte, 0)
	}

	var d = make([]byte, hex.EncodedLen(len(p)))

	hex.Encode(d, p)

	return d
}

func (o *crt) Decode(p []byte) ([]byte, error) {
	if len(p) < 1 {
		return make([]byte, 0), nil
	}

	var d = make([]byte, hex.DecodedLen(len(p)))

	_, e := hex.Decode(d, p)

	return d, e
}

func (o *crt) EncodeReader(r io.Reader) io.ReadCloser {
	f := func(p []byte) (n int, err error) {
		var b []byte

		if cap(p) < 2 {
			return 0, ErrInvalidBufferSize
		} else {
			b = make([]byte, hex.DecodedLen(cap(p)))
		}

		n, err = r.Read(b)

		if n > 0 {
			b = o.Encode(b[:n])
			n = len(b)

			if n > cap(p) {
				clear(b)
				b = b[:0]
				err = ErrInvalidBufferSize
				n = 0
			} else {
				copy(p, b)
				clear(b)
				b = b[:0]
			}
		}

		return n, err
	}

	c := func() error {
		if rc, ok := r.(io.Closer); ok {
			return rc.Close()
		}

		return nil
	}

	return &reader{f: f, c: c}
}

func (o *crt) DecodeReader(r io.Reader) io.ReadCloser {
	var h = hex.NewDecoder(r)
	f := func(p []byte) (n int, err error) {
		return h.Read(p)
	}
	c := func() error {
		if rc, ok := r.(io.Closer); ok {
			return rc.Close()
		}
		return nil
	}
	return &reader{f: f, c: c}
}

func (o *crt) EncodeWriter(w io.Writer) io.WriteCloser {
	var h = hex.NewEncoder(w)
	f := func(p []byte) (n int, err error) {
		return h.Write(p)
	}
	c := func() error {
		if wc, ok := w.(io.Closer); ok {
			return wc.Close()
		}
		return nil
	}
	return &writer{f: f, c: c}
}

func (o *crt) DecodeWriter(w io.Writer) io.WriteCloser {
	f := func(p []byte) (n int, err error) {
		var b []byte
		n = len(p)

		if b, err = o.Decode(p); err != nil {
			return 0, err
		} else {
			_, err = w.Write(b)

			clear(b)
			b = b[:0]

			if err != nil {
				return 0, err
			} else {
				return n, nil
			}
		}
	}
	c := func() error {
		if wc, ok := w.(io.Closer); ok {
			return wc.Close()
		}
		return nil
	}
	return &writer{f: f, c: c}
}

func (o *crt) Reset() {
}

type reader struct {
	f func(p []byte) (n int, err error)
	c func() error
}

func (r *reader) Read(p []byte) (n int, err error) {
	if r.f == nil {
		return 0, fmt.Errorf("invalid reader")
	} else {
		return r.f(p)
	}
}

func (r *reader) Close() error {
	if r.c == nil {
		return fmt.Errorf("invalid closer")
	} else {
		return r.c()
	}
}

type writer struct {
	f func(p []byte) (n int, err error)
	c func() error
}

func (r *writer) Write(p []byte) (n int, err error) {
	if r.f == nil {
		return 0, fmt.Errorf("invalid reader")
	} else {
		return r.f(p)
	}
}

func (r *writer) Close() error {
	if r.c == nil {
		return fmt.Errorf("invalid closer")
	} else {
		return r.c()
	}
}
