/*
 *  MIT License
 *
 *  Copyright (c) 2024 Salim Amine BOU ARAM & Nicolas JUHEL
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

package helper

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	arccmp "github.com/nabbar/golib/archive/compress"
)

const workBufSizeDeCompressWrite = 32 * 1024 // 32kB for buffer

type deCompressReader struct {
	src io.ReadCloser
}

func (o *deCompressReader) Read(p []byte) (n int, err error) {
	return o.src.Read(p)
}

func (o *deCompressReader) Write(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

func (o *deCompressReader) Close() error {
	return o.src.Close()
}

type bufNoEOF struct {
	m sync.Mutex
	b *bytes.Buffer
	c *atomic.Bool
}

func (o *bufNoEOF) Read(p []byte) (n int, err error) {
	if o.c.Load() && o.b.Len() < 1 {
		return 0, io.EOF
	}

	for o.b.Len() < 1 && !o.c.Load() {
		time.Sleep(100 * time.Microsecond)
	}

	n, _ = o.readBuff(p)

	if n < 1 && o.c.Load() {
		return 0, io.EOF
	} else {
		return n, nil
	}
}

func (o *bufNoEOF) Write(p []byte) (n int, err error) {
	if o.c.Load() {
		return 0, errors.New("closed buffer")
	}

	return o.writeBuff(p)
}

func (o *bufNoEOF) Close() error {
	o.c.Store(true)
	return nil
}

func (o *bufNoEOF) Len() int {
	o.m.Lock()
	defer o.m.Unlock()
	return o.b.Len()
}

func (o *bufNoEOF) readBuff(p []byte) (n int, err error) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.b.Len() < 1 {
		return 0, io.EOF
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(o.b.Bytes())

	n, err = buf.Read(p)
	o.b.Reset()

	var e error
	if buf.Len() > 0 {
		_, e = io.Copy(o.b, buf)
	}

	if err == nil && e != nil && e != io.EOF {
		return n, e
	}

	return n, err
}

func (o *bufNoEOF) writeBuff(p []byte) (n int, err error) {
	o.m.Lock()
	defer o.m.Unlock()
	return o.b.Write(p)
}

type deCompressWriter struct {
	alg arccmp.Algorithm
	wrt io.WriteCloser
	buf *bufNoEOF
	clo *atomic.Bool
	run *atomic.Bool
}

func (o *deCompressWriter) Read(p []byte) (n int, err error) {
	return 0, ErrInvalidSource
}

func (o *deCompressWriter) Write(p []byte) (n int, err error) {
	if o.clo.Load() {
		return 0, ErrClosedResource
	}

	n, err = o.buf.Write(p)
	if err != nil || o.run.Load() {
		return n, err
	}

	o.run.Store(true)
	if r, e := o.alg.Reader(o.buf); e != nil {
		return n, e
	} else {
		go func() {
			_, _ = io.Copy(o.wrt, r)
		}()
	}

	return n, err
}

func (o *deCompressWriter) Close() error {
	o.clo.Store(true)
	o.run.Store(false)
	_ = o.buf.Close()

	for o.buf.Len() > 0 {
		time.Sleep(time.Millisecond)
	}

	if err := o.wrt.Close(); err != nil {
		return err
	}

	return nil
}
