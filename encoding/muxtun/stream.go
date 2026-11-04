/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package muxtun

import (
	"io"
	"sync"
)

type stream struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func (o *stream) Read(p []byte) (n int, err error) {
	return o.r.Read(p)
}

func (o *stream) Write(p []byte) (n int, err error) {
	return o.w.Write(p)
}

type strList struct {
	m sync.Mutex
	l map[uint8]*stream
}

func (o *strList) Read(k uint8, p []byte) (n int, err error) {
	o.m.Lock()
	s := o.l[k]
	o.m.Unlock()

	if s == nil {
		return 0, io.EOF
	}

	return s.Read(p)
}

func (o *strList) Write(k uint8, p []byte) (n int, err error) {
	o.m.Lock()
	s := o.l[k]
	o.m.Unlock()

	if s == nil {
		return 0, io.EOF
	}

	return s.Write(p)
}

func (o *strList) new(k uint8) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.l == nil {
		o.l = make(map[uint8]*stream)
	}

	if s, ok := o.l[k]; ok {
		_ = s.w.Close()
		_ = s.r.Close()
	}

	o.l[k] = new(stream)
	o.l[k].r, o.l[k].w = io.Pipe()
}

func (o *strList) del(k uint8) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.l == nil {
		return
	}

	if s, ok := o.l[k]; ok {
		_ = s.w.Close()
		_ = s.r.Close()
		delete(o.l, k)
	}
}

func (o *strList) Close() error {
	o.m.Lock()
	defer o.m.Unlock()

	for k := range o.l {
		_ = o.l[k].w.Close()
		_ = o.l[k].r.Close()
	}

	o.l = nil
	return nil
}

type vStream struct {
	m sync.Mutex
	v map[uint8]bool
}

func (o *vStream) open(v uint8) bool {
	o.m.Lock()
	defer o.m.Unlock()

	if o.v == nil {
		return false
	}

	if s, k := o.v[v]; k {
		return s
	}

	return false
}

func (o *vStream) add(v uint8) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.v == nil {
		o.v = make(map[uint8]bool)
	}

	if _, k := o.v[v]; !k {
		o.v[v] = false
	}
}

func (o *vStream) state(v uint8, f bool) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.v == nil {
		o.v = make(map[uint8]bool)
	}

	if _, k := o.v[v]; k {
		o.v[v] = f
	}
}

func (o *vStream) delete(v uint8) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.v == nil {
		o.v = make(map[uint8]bool)
	}

	delete(o.v, v)
}
