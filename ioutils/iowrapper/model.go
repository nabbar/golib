/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package iowrapper

import (
	"io"

	libatm "github.com/nabbar/golib/atomic"
)

type iow struct {
	i any
	r libatm.Value[FuncRead]
	w libatm.Value[FuncWrite]
	s libatm.Value[FuncSeek]
	c libatm.Value[FuncClose]
}

func (o *iow) SetRead(read FuncRead) {
	if read == nil {
		read = o.fakeRead
	}

	o.r.Store(read)
}

func (o *iow) SetWrite(write FuncWrite) {
	if write == nil {
		write = o.fakeWrite
	}

	o.w.Store(write)
}

func (o *iow) SetSeek(seek FuncSeek) {
	if seek == nil {
		seek = o.fakeSeek
	}

	o.s.Store(seek)
}

func (o *iow) SetClose(close FuncClose) {
	if close == nil {
		close = o.fakeClose
	}

	o.c.Store(close)
}

func (o *iow) Read(p []byte) (n int, err error) {
	var r []byte

	if o.r != nil {
		r = o.r.Load()(p)
	} else {
		r = o.fakeRead(p)
	}

	if r == nil {
		return 0, io.ErrUnexpectedEOF
	}

	copy(p, r)
	return len(r), nil
}

func (o *iow) Write(p []byte) (n int, err error) {
	var r []byte

	if o.w != nil {
		r = o.w.Load()(p)
	} else {
		r = o.fakeWrite(p)
	}

	if r == nil {
		return 0, io.ErrUnexpectedEOF
	}

	copy(p, r)
	return len(r), nil
}

func (o *iow) Seek(offset int64, whence int) (int64, error) {
	if o.s != nil {
		return o.s.Load()(offset, whence)
	} else {
		return o.fakeSeek(offset, whence)
	}
}

func (o *iow) Close() error {
	if o.c != nil {
		return o.c.Load()()
	} else {
		return o.fakeClose()
	}
}

func (o *iow) fakeRead(p []byte) []byte {
	if r, k := o.i.(io.Reader); k {
		n, err := r.Read(p)
		if err != nil {
			return p[:n]
		}
		return p[:n]
	} else {
		return nil
	}
}

func (o *iow) fakeWrite(p []byte) []byte {
	if r, k := o.i.(io.Writer); k {
		n, err := r.Write(p)
		if err != nil {
			return p[:n]
		}
		return p[:n]
	} else {
		return nil
	}
}

func (o *iow) fakeSeek(offset int64, whence int) (int64, error) {
	if r, k := o.i.(io.Seeker); k {
		return r.Seek(offset, whence)
	} else {
		return 0, io.ErrUnexpectedEOF
	}
}

func (o *iow) fakeClose() error {
	if r, k := o.i.(io.Closer); k {
		return r.Close()
	} else {
		return nil
	}
}
