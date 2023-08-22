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

package gzipreader

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
)

type gzr struct {
	r  io.Reader
	b  *bytes.Buffer
	z  *gzip.Writer
	nc int64
	nu int64
}

func (o *gzr) Read(p []byte) (n int, err error) {
	var (
		s []byte

		er error
		nr int

		ew error
		nw int
	)

	if i := cap(p); i > o.b.Cap() || i < 1 {
		s = make([]byte, 0, o.b.Cap())
	} else {
		s = make([]byte, 0, i)
	}

	nr, er = o.r.Read(s)
	o.nu += int64(nr)

	if er != nil && !errors.Is(er, io.EOF) {
		return 0, err
	} else if nr > 0 {
		nw, ew = o.z.Write(s)
	}

	if ew != nil {
		return 0, ew
	} else if nw != nr {
		return 0, errors.New("invalid write buffer")
	} else if er != nil && errors.Is(er, io.EOF) {
		if ew = o.z.Close(); ew != nil {
			return 0, ew
		}
	}

	copy(p, o.b.Bytes())
	o.b.Reset()
	o.nc += int64(len(p))

	return len(p), er
}

func (o *gzr) LenCompressed() int64 {
	return o.nc
}

func (o *gzr) LenUnCompressed() int64 {
	return o.nu
}

func (o *gzr) Rate() float64 {
	return 1 - float64(o.nc/o.nu)
}
