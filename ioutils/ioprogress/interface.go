/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package ioprogress

import (
	"io"
	"sync/atomic"

	libfpg "github.com/nabbar/golib/file/progress"
)

type Progress interface {
	RegisterFctIncrement(fct libfpg.FctIncrement)
	RegisterFctReset(fct libfpg.FctReset)
	RegisterFctEOF(fct libfpg.FctEOF)
	Reset(max int64)
}

type Reader interface {
	io.ReadCloser
	Progress
}

type Writer interface {
	io.WriteCloser
	Progress
}

func NewReadCloser(r io.ReadCloser) Reader {
	return &rdr{
		r:  r,
		cr: new(atomic.Int64),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}
}

func NewWriteCloser(w io.WriteCloser) Writer {
	return &wrt{
		w:  w,
		cr: new(atomic.Int64),
		fi: new(atomic.Value),
		fe: new(atomic.Value),
		fr: new(atomic.Value),
	}
}
