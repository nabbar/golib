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

package archive

import (
	"bufio"
	"io"
	"io/fs"
)

// struct to allow compatibility with io.ReadSeeker for all type of archive
type rdr struct {
	r io.ReadCloser
	b *bufio.Reader
}

func (o *rdr) Close() error {
	return o.r.Close()
}

func (o *rdr) Read(p []byte) (n int, err error) {
	if o.b != nil {
		return o.b.Read(p)
	}

	return o.r.Read(p)
}

func (o *rdr) Peek(n int) ([]byte, error) {
	if o.b != nil {
		return o.b.Peek(n)
	} else {
		b := bufio.NewReader(o.r)
		return b.Peek(n)
	}
}

func (o *rdr) ReadAt(p []byte, off int64) (n int, err error) {
	if r, k := o.r.(io.ReaderAt); k {
		return r.ReadAt(p, off)
	}

	if r, k := o.r.(io.ReadSeeker); k {
		if _, err = r.Seek(off, io.SeekStart); err != nil {
			return 0, err
		}

		if o.b != nil {
			o.b.Reset(r)
			return o.b.Read(p)
		} else {
			return o.r.Read(p)
		}
	}

	return 0, fs.ErrInvalid
}

func (o *rdr) Size() int64 {
	var s int64

	if r, k := o.r.(io.Seeker); k {
		if c, e := r.Seek(0, io.SeekCurrent); e != nil {
			return 0
		} else if _, e = r.Seek(c, io.SeekStart); e != nil {
			return 0
		} else if s, e = r.Seek(0, io.SeekEnd); e != nil {
			return 0
		} else if _, e = r.Seek(c, io.SeekStart); e != nil {
			return 0
		} else {
			return s
		}
	}

	if r, k := o.r.(io.ReaderAt); k {
		if _, e := r.ReadAt(nil, 0); e != nil {
			return 0
		} else if s, e = io.Copy(io.Discard, o.r); e != nil {
			return 0
		} else if _, e = r.ReadAt(nil, s); e != nil {
			return 0
		} else if o.b != nil {
			o.b.Reset(o.r)
			return s
		} else {
			return s
		}
	}

	return 0
}

func (o *rdr) Seek(offset int64, whence int) (int64, error) {
	if r, k := o.r.(io.Seeker); k {
		if n, e := r.Seek(offset, whence); e != nil {
			return n, e
		} else if o.b != nil {
			o.b.Reset(o.r)
			return n, nil
		} else {
			return n, nil
		}
	}

	if whence != io.SeekStart {
		return 0, fs.ErrInvalid
	}

	if _, e := o.ReadAt(make([]byte, 0), 0); e != nil {
		return 0, e
	} else if o.b != nil {
		o.b.Reset(o.r)
		return 0, nil
	} else {
		return 0, nil
	}
}

func (o *rdr) Reset() bool {
	if r, k := o.r.(io.Seeker); k {
		_, e := r.Seek(0, io.SeekStart)
		if e == nil {
			if o.b != nil {
				o.b.Reset(o.r)
			}
			return true
		}
	}

	if r, k := o.r.(io.ReaderAt); k {
		_, e := r.ReadAt(nil, 0)
		if e == nil {
			if o.b != nil {
				o.b.Reset(o.r)
			}
			return true
		}
	}

	return false
}
