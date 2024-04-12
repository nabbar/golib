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
	"errors"
	"io"
	"sync/atomic"

	libfpg "github.com/nabbar/golib/file/progress"
)

type wrt struct {
	w io.WriteCloser

	cr *atomic.Int64
	fi *atomic.Value
	fe *atomic.Value
	fr *atomic.Value
}

func (w *wrt) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.inc(n)

	if errors.Is(err, io.EOF) {
		w.finish()
	}

	return n, err
}

func (w *wrt) Close() error {
	return w.w.Close()
}

func (w *wrt) RegisterFctIncrement(fct libfpg.FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	w.fi.Store(fct)
}

func (w *wrt) RegisterFctReset(fct libfpg.FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	w.fr.Store(fct)
}

func (w *wrt) RegisterFctEOF(fct libfpg.FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	w.fe.Store(fct)
}

func (w *wrt) inc(n int) {
	if w == nil {
		return
	}

	w.cr.Add(int64(n))

	f := w.fi.Load()
	if f != nil {
		f.(libfpg.FctIncrement)(int64(n))
	}
}

func (w *wrt) finish() {
	if w == nil {
		return
	}

	f := w.fe.Load()
	if f != nil {
		f.(libfpg.FctEOF)()
	}
}

func (w *wrt) Reset(max int64) {
	if w == nil {
		return
	}

	f := w.fr.Load()
	f.(libfpg.FctReset)(max, w.cr.Load())
}
