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

package tar

import (
	"archive/tar"
	"io"
	"io/fs"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

type reset interface {
	Reset() bool
}

type rdr struct {
	r io.ReadCloser
	z *tar.Reader
}

func (o *rdr) Reset() bool {
	if r, k := o.r.(reset); k {
		return r.Reset()
	}

	return false
}

func (o *rdr) Close() error {
	return o.r.Close()
}

func (o *rdr) List() ([]string, error) {
	var (
		e error
		h *tar.Header
		l = make([]string, 0)
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil {
			l = append(l, h.Name)
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return l, nil
}

func (o *rdr) Info(s string) (fs.FileInfo, error) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return h.FileInfo(), nil
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return nil, fs.ErrNotExist

}

func (o *rdr) Get(s string) (io.ReadCloser, error) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return io.NopCloser(o.z), nil
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return nil, fs.ErrNotExist
}

func (o *rdr) Has(s string) bool {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()
		if h != nil && h.Name == s {
			return true
		} else if h != nil {
			_, _ = io.Copy(io.Discard, o.z)
		}
	}

	return false
}

func (o *rdr) Walk(fct arctps.FuncExtract) {
	var (
		e error
		h *tar.Header
	)

	if o.Reset() {
		o.z = tar.NewReader(o.r)
	}

	for e == nil {
		h, e = o.z.Next()

		if h == nil || e != nil {
			continue
		}

		if !fct(h.FileInfo(), io.NopCloser(o.z), h.Name, h.Linkname) {
			return
		}

		// prevent file cursor not at EOF of current file
		_, _ = io.Copy(io.Discard, o.z)
	}
}
