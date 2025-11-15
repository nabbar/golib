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
	"os"
	"path/filepath"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

type wrt struct {
	w io.WriteCloser
	z *tar.Writer
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

// Add adds a file to the tar archive.
//
// It takes in the file information, the file reader, and the target path if the new file is a link.
// It returns an error if any operation fails.
func (o *wrt) Add(i fs.FileInfo, r io.ReadCloser, forcePath, target string) error {
	var (
		e error
		h *tar.Header
	)

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if h, e = tar.FileInfoHeader(i, target); e != nil {
		return e
	} else if len(target) > 0 {
		h.Linkname = target
	}

	if len(forcePath) > 0 {
		h.Name = forcePath
	}

	if e = o.z.WriteHeader(h); e != nil {
		return e
	}

	if r != nil {
		if _, e = io.Copy(o.z, r); e != nil {
			return e
		}
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
		ok     bool
		err    error
		rpt    *os.Root
		hdf    *os.File
		target string
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
	} else if info.Mode()&os.ModeSymlink != 0 { // SymLink
		if target, err = os.Readlink(source); err != nil {
			return err
		}
	} else if info.Mode()&os.ModeDevice != 0 { // HardLink
		if target, err = os.Readlink(source); err != nil {
			return err
		}
	} else if info.Mode().IsRegular() {
		rpt, err = os.OpenRoot(filepath.Dir(source))

		defer func() {
			if rpt != nil {
				_ = rpt.Close()
			}
		}()

		if err != nil {
			return err
		}

		hdf, err = rpt.Open(filepath.Base(source))

		defer func() {
			_ = hdf.Close()
		}()

		if err != nil {
			return err
		}
	} else {
		return fs.ErrInvalid
	}

	return o.Add(info, hdf, fct(source), target)
}
