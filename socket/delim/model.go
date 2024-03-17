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

import (
	"bufio"
	"io"
	"sync/atomic"

	libsiz "github.com/nabbar/golib/size"
)

type dlm struct {
	i *atomic.Value  // input io.ReadCloser
	d *atomic.Int32  // delimiter rune
	s *atomic.Uint64 // buffer libsiz.Size
	r *atomic.Value  // *bufio.Reader
}

func (o *dlm) SetDelim(delim rune) {
	o.d.Store(delim)
}

func (o *dlm) GetDelim() rune {
	return o.d.Load()
}

func (o *dlm) getDelimByte() byte {
	return byte(o.GetDelim())
}

func (o *dlm) SetBufferSize(b libsiz.Size) {
	o.s.Store(b.Uint64())
}

func (o *dlm) GetBufferSize() libsiz.Size {
	return libsiz.Size(o.s.Load())
}

func (o *dlm) SetInput(i io.ReadCloser) {
	if i == nil {
		i = &DiscardCloser{}
	}

	o.i.Store(i)
}

func (o *dlm) getInput() io.ReadCloser {
	if i := o.i.Load(); i == nil {
		return &DiscardCloser{}
	} else if v, k := i.(io.ReadCloser); !k {
		return &DiscardCloser{}
	} else {
		return v
	}
}

func (o *dlm) newReader() *bufio.Reader {
	if siz := o.GetBufferSize(); siz > 0 {
		return bufio.NewReaderSize(o.getInput(), siz.Int())
	} else {
		return bufio.NewReader(o.getInput())
	}
}

func (o *dlm) setReader(r *bufio.Reader) {
	if r == nil {
		r = o.newReader()
	}

	o.r.Store(r)
}

func (o *dlm) getReader() *bufio.Reader {
	if i := o.r.Load(); i == nil {
		r := o.newReader()
		o.setReader(r)
		return r
	} else if v, k := i.(*bufio.Reader); !k {
		r := o.newReader()
		o.setReader(r)
		return r
	} else {
		return v
	}
}
