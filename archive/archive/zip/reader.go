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

package zip

import (
	"archive/zip"
	"io"
	"io/fs"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

type rdr struct {
	r io.ReadCloser
	z *zip.Reader
}

func (o *rdr) Close() error {
	return o.r.Close()
}

func (o *rdr) List() ([]string, error) {
	var res = make([]string, 0, len(o.z.File))

	for _, f := range o.z.File {
		res = append(res, f.Name)
	}

	return res, nil
}

func (o *rdr) Info(s string) (fs.FileInfo, error) {
	for _, f := range o.z.File {
		if f.Name == s {
			return f.FileInfo(), nil
		}
	}

	return nil, fs.ErrNotExist
}

func (o *rdr) Get(s string) (io.ReadCloser, error) {
	for _, f := range o.z.File {
		if f.Name == s {
			return f.Open()
		}
	}

	return nil, fs.ErrNotExist
}

func (o *rdr) Has(s string) bool {
	for _, f := range o.z.File {
		if f.Name == s {
			return true
		}
	}

	return false
}

func (o *rdr) Walk(fct arctps.FuncExtract) {
	for _, f := range o.z.File {
		r, _ := f.Open()
		if !fct(f.FileInfo(), r, f.Name, "") {
			return
		}
	}
}
