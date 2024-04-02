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
	"os"
	"path/filepath"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

type wrt struct {
	w io.WriteCloser
	z *zip.Writer
}

func (o *wrt) Close() error {
	if e := o.z.Flush(); e != nil {
		return e
	} else if e = o.z.Close(); e != nil {
		return e
	} else if e = o.w.Close(); e != nil {
		return e
	}

	return nil
}

func (o *wrt) Add(i fs.FileInfo, r io.ReadCloser, forcePath, notUse string) error {
	var (
		e error
		h *zip.FileHeader
		w io.Writer
	)

	if r == nil {
		return nil
	}

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if h, e = zip.FileInfoHeader(i); e != nil {
		return e
	} else if len(forcePath) > 0 {
		h.Name = forcePath
	}

	if w, e = o.z.CreateHeader(h); e != nil {
		return e
	} else if _, e = io.Copy(w, r); e != nil {
		return e
	}

	return nil
}

func (o *wrt) FromPath(source string, filter string, fct arctps.ReplaceName) error {
	if i, e := os.Stat(source); e == nil && !i.IsDir() {
		return o.addFiltering(source, filter, fct, i)
	}

	return filepath.Walk(source, func(path string, info fs.FileInfo, e error) error {
		if e != nil {
			return e
		}

		return o.addFiltering(path, filter, fct, info)
	})
}

func (o *wrt) addFiltering(source string, filter string, fct arctps.ReplaceName, info fs.FileInfo) error {
	var (
		ok  bool
		err error
		hdf *os.File
	)

	if len(filter) < 1 {
		filter = "*"
	}

	if fct == nil {
		fct = func(source string) string {
			return source
		}
	}

	if ok, err = filepath.Match(filter, source); err != nil {
		return err
	} else if !ok {
		return nil
	}

	if info == nil {
		return fs.ErrInvalid
	} else if info.IsDir() {
		return nil
	} else if info.Mode().IsRegular() {
		if hdf, err = os.Open(source); err != nil {
			return err
		} else {
			defer func() {
				_ = hdf.Close()
			}()
		}
	} else {
		return fs.ErrInvalid
	}

	return o.Add(info, hdf, fct(source), "")
}
