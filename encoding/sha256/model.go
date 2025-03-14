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

package sha256

import (
	"fmt"
	"hash"
	"io"
)

type crt struct {
	hsh hash.Hash
}

func (o *crt) Encode(p []byte) []byte {
	if o.hsh == nil {
		return make([]byte, 0)
	}

	if len(p) > 0 {
		if _, e := o.hsh.Write(p); e != nil {
			return make([]byte, 0)
		}
	}

	if q := o.hsh.Sum(nil); len(q) < 1 {
		return make([]byte, 0)
	} else {
		return q[:]
	}
}

func (o *crt) Decode(p []byte) ([]byte, error) {
	return nil, fmt.Errorf("unexpected call function")
}

func (o *crt) EncodeReader(r io.Reader) io.ReadCloser {
	f := func(p []byte) (n int, err error) {
		n, err = r.Read(p)

		if n > 0 && o.hsh != nil {
			o.hsh.Write(p[0:n])
		}

		return n, err
	}

	c := func() error {
		if rc, ok := r.(io.Closer); ok {
			return rc.Close()
		} else {
			return nil
		}
	}

	return &reader{f: f, c: c}
}

func (o *crt) DecodeReader(r io.Reader) io.ReadCloser {
	return nil
}

func (o *crt) EncodeWriter(w io.Writer) io.WriteCloser {
	f := func(p []byte) (n int, err error) {
		n, err = w.Write(p)

		if n > 0 && o.hsh != nil {
			o.hsh.Write(p[0:n])
		}

		return n, err
	}

	c := func() error {
		if wc, ok := w.(io.Closer); ok {
			return wc.Close()
		} else {
			return nil
		}
	}

	return &writer{f: f, c: c}
}

func (o *crt) DecodeWriter(w io.Writer) io.WriteCloser {
	return nil
}

func (o *crt) Reset() {
	if o.hsh != nil {
		o.hsh.Reset()
	}
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
		return 0, fmt.Errorf("invalid writer")
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
