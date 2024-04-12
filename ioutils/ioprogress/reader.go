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

type rdr struct {
	r io.ReadCloser

	cr *atomic.Int64
	fi *atomic.Value
	fe *atomic.Value
	fr *atomic.Value
}

func (r *rdr) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.inc(n)

	if errors.Is(err, io.EOF) {
		r.finish()
	}

	return n, err
}

func (r *rdr) Close() error {
	return r.r.Close()
}

func (r *rdr) RegisterFctIncrement(fct libfpg.FctIncrement) {
	if fct == nil {
		fct = func(size int64) {}
	}

	r.fi.Store(fct)
}

func (r *rdr) RegisterFctReset(fct libfpg.FctReset) {
	if fct == nil {
		fct = func(size, current int64) {}
	}

	r.fr.Store(fct)
}

func (r *rdr) RegisterFctEOF(fct libfpg.FctEOF) {
	if fct == nil {
		fct = func() {}
	}

	r.fe.Store(fct)
}

func (r *rdr) inc(n int) {
	if r == nil {
		return
	}

	r.cr.Add(int64(n))

	f := r.fi.Load()
	if f != nil {
		f.(libfpg.FctIncrement)(int64(n))
	}
}

func (r *rdr) finish() {
	if r == nil {
		return
	}

	f := r.fe.Load()
	if f != nil {
		f.(libfpg.FctEOF)()
	}
}

func (r *rdr) Reset(max int64) {
	if r == nil {
		return
	}

	f := r.fr.Load()
	f.(libfpg.FctReset)(max, r.cr.Load())
}
