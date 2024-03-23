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

package socket

import (
	"fmt"
	"io"
)

var (
	ErrInvalidInstance = fmt.Errorf("invalid instance")
	closedChanStruct   chan struct{}
)

func init() {
	closedChanStruct = make(chan struct{})
	close(closedChanStruct)
}

type FctWriter func(p []byte) (n int, err error)
type FctReader func(p []byte) (n int, err error)
type FctClose func() error
type FctCheck func() bool
type FctDone func() <-chan struct{}

type Reader interface {
	io.ReadCloser
	IsConnected() bool
	Done() <-chan struct{}
}

type Writer interface {
	io.WriteCloser
	IsConnected() bool
	Done() <-chan struct{}
}

type wrt struct {
	w FctWriter
	c FctClose
	d FctDone
	i FctCheck
}

func (o *wrt) Write(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInvalidInstance
	} else if o.w == nil {
		return 0, ErrInvalidInstance
	} else {
		return o.w(p)
	}
}

func (o *wrt) Close() error {
	if o == nil {
		return ErrInvalidInstance
	} else if o.w == nil {
		return ErrInvalidInstance
	} else {
		return o.c()
	}
}

func (o *wrt) IsConnected() bool {
	if o == nil {
		return false
	} else if o.i == nil {
		return false
	} else {
		return o.i()
	}
}

func (o *wrt) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	} else if o.d == nil {
		return closedChanStruct
	} else {
		return o.d()
	}
}

type rdr struct {
	r FctReader
	c FctClose
	d FctDone
	i FctCheck
}

func (o *rdr) Read(p []byte) (n int, err error) {
	if o == nil {
		return 0, ErrInvalidInstance
	} else if o.r == nil {
		return 0, ErrInvalidInstance
	} else {
		return o.r(p)
	}
}

func (o *rdr) Close() error {
	if o == nil {
		return ErrInvalidInstance
	} else if o.c == nil {
		return ErrInvalidInstance
	} else {
		return o.c()
	}
}

func (o *rdr) IsConnected() bool {
	if o == nil {
		return false
	} else if o.i == nil {
		return false
	} else {
		return o.i()
	}
}

func (o *rdr) Done() <-chan struct{} {
	if o == nil {
		return closedChanStruct
	} else if o.d == nil {
		return closedChanStruct
	} else {
		return o.d()
	}
}

func NewReader(fctRead FctReader, fctClose FctClose, fctCheck FctCheck, fctDone FctDone) Reader {
	return &rdr{
		r: fctRead,
		c: fctClose,
		d: fctDone,
		i: fctCheck,
	}
}

func NewWriter(fctWrite FctWriter, fctClose FctClose, fctCheck FctCheck, fctDone FctDone) Writer {
	return &wrt{
		w: fctWrite,
		c: fctClose,
		d: fctDone,
		i: fctCheck,
	}
}
