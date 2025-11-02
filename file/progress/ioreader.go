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
	"math"
)

func (o *progress) Read(p []byte) (n int, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.analyze(o.f.Read(p))
}

func (o *progress) ReadAt(p []byte, off int64) (n int, err error) {
	if o == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	return o.analyze(o.f.ReadAt(p, off))
}

func (o *progress) ReadFrom(r io.Reader) (n int64, err error) {
	if o == nil || r == nil || o.f == nil {
		return 0, ErrorNilPointer.Error(nil)
	}

	var size = o.getBufferSize(0)

	if l, ok := r.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else if l.N > math.MaxInt {
			size = math.MaxInt
		} else {
			size = int(l.N)
		}
	}

	var bf = make([]byte, o.getBufferSize(size))

	for {
		var (
			nr int
			nw int
			er error
			ew error
		)

		// code from io.copy

		nr, er = r.Read(bf)
		if nr > 0 {
			nw, ew = o.Write(bf[:nr])
		}

		n += int64(nw)

		if er != nil && errors.Is(er, io.EOF) {
			o.finish()
			break
		} else if er != nil {
			err = er
			break
		} else if ew != nil {
			err = ew
			break
		} else if nw < 0 || nw < nr {
			err = errors.New("invalid write result")
			break
		} else if nr != nw {
			err = io.ErrShortWrite
			break
		}

		clear(bf)
	}

	return n, err
}
