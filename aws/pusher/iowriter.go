/*
 *  MIT License
 *
 *  Copyright (c) 2024 Nicolas JUHEL
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

package pusher

import (
	"errors"
	"io"
)

func (o *psh) Write(p []byte) (n int, err error) {
	// checking incoming buffer is not empty
	if len(p) < 1 {
		return 0, nil
	}

	// writing into working part file
	if n, err = o.fileWrite(p); err != nil {
		return 0, err
	}

	// calculating md5 of part content (mandatory to upload part)
	if i, e := o.md5Write(p); e != nil {
		return 0, e
	} else if i != n {
		return n, io.ErrShortWrite
	}

	// calculating sha256 of full object (optional to upload part)
	if i, e := o.shaObjWrite(p); e != nil {
		return 0, e
	} else if i != n {
		return n, io.ErrShortWrite
	}

	// calculating sha256 of part content (optional to upload part)
	if i, e := o.shaPartWrite(p); e != nil {
		return 0, e
	} else if i != n {
		return n, io.ErrShortWrite
	}

	s := o.prtSize.Load()
	t := o.cfg.getPartSize().Int64()

	// checking if needed to push part
	if s >= t {
		// pushing part
		e := o.pushObject()

		if e != nil {
			return n, e
		}
	}

	// returning number of bytes written
	return n, nil
}

func (o *psh) Close() error {
	// closing uploaded part and aborting object
	return o.Abort()
}

func (o *psh) ReadFrom(r io.Reader) (n int64, err error) {
	var (
		i  int                                   // number of bytes read
		j  int                                   // number of bytes written
		er error                                 // read error
		ew error                                 // write error
		p  = make([]byte, o.cfg.getBufferSize()) // buffer
	)

	for {
		// checking context
		if o.ctx.Err() != nil {
			return n, o.ctx.Err()
		} else if !o.IsReady() {
			return n, ErrInvalidInstance
		}

		// reading from reader
		i, er = r.Read(p)

		// writing to object if possible
		if i > 0 {
			j, ew = o.Write(p[:i])
			n += int64(j)
		}

		// clearing buffer
		clear(p)

		// checking errors
		if er != nil && !errors.Is(er, io.EOF) {
			return n, er
		} else if ew != nil {
			return n, ew
		} else if er != nil {
			return n, nil
		}
	}
}
