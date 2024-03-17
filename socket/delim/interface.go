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
	"io"
	"sync/atomic"

	libsiz "github.com/nabbar/golib/size"
)

type BufferDelim interface {
	io.ReadCloser
	io.WriterTo

	SetDelim(d rune)
	GetDelim() rune

	SetBufferSize(b libsiz.Size)
	GetBufferSize() libsiz.Size

	SetInput(i io.ReadCloser)
	Reader() io.ReadCloser
	Copy(w io.Writer) (n int64, err error)
	ReadBytes() ([]byte, error)
}

func New(r io.ReadCloser, delim rune, sizeBufferRead libsiz.Size) BufferDelim {
	d := &dlm{
		i: new(atomic.Value),
		d: new(atomic.Int32),
		s: new(atomic.Uint64),
		r: new(atomic.Value),
	}

	d.SetDelim(delim)
	d.SetBufferSize(sizeBufferRead)
	d.SetInput(r)

	return d
}
