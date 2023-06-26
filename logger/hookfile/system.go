/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package hookfile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	libiot "github.com/nabbar/golib/ioutils"
)

const sizeBuffer = 32 * 1024

func (o *hkf) writeBuffer(buf *bytes.Buffer) error {
	var (
		e error
		h *os.File
		p = o.getFilepath()
		m = o.getFileMode()
		n = o.getPathMode()
		f = o.getFlags()
	)

	if o.getCreatePath() {
		if e = libiot.PathCheckCreate(true, p, m, n); e != nil {
			return e
		}
	}

	defer func() {
		if h != nil {
			_ = h.Close()
		}
	}()

	if h, e = os.OpenFile(p, f, m); e != nil {
		return e
	} else if _, e = h.Seek(0, io.SeekEnd); e != nil {
		return e
	} else if _, e = h.Write(buf.Bytes()); e != nil {
		return e
	} else if e = h.Close(); e != nil {
		h = nil
		buf.Reset()
		return e
	} else {
		h = nil
	}

	buf.Reset()
	return nil
}

func (o *hkf) freeBuffer(buf *bytes.Buffer, size int) *bytes.Buffer {
	var a = bytes.NewBuffer(buf.Bytes()[size:])
	a.Grow(size)
	return a
}

func (o *hkf) Run(ctx context.Context) {
	var (
		b = bytes.NewBuffer(make([]byte, sizeBuffer))
		t = time.NewTicker(time.Second)
		e error
	)

	defer func() {
		//flush buffer before exit function
		if b.Len() > 0 {
			if e = o.writeBuffer(b); e != nil {
				fmt.Println(e.Error())
			}
			b.Reset()
		}
	}()

	o.prepareChan()
	fmt.Printf("starting hook for log file '%s'\n", o.getFilepath())

	for {
		select {
		case <-ctx.Done():
			return

		case <-o.Done():
			return

		case <-t.C:
			if b.Len() < 1 {
				continue
			} else if e = o.writeBuffer(b); e != nil {
				fmt.Println(e.Error())
			}

		case p := <-o.Data():
			// prevent buffer overflow
			if b.Len()+len(p) >= sizeBuffer {
				b = o.freeBuffer(b, len(p))
				b.Write(p)
			} else {
				_, _ = b.Write(p)
			}
		}
	}
}
