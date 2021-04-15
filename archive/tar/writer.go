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

package tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	liberr "github.com/nabbar/golib/errors"
)

func Create(archive io.WriteSeeker, stripPath string, content ...string) (bool, liberr.Error) {

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	if ok, err := createTar(archive, stripPath, content...); err != nil || !ok {
		return ok, err
	}

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, nil
}

func CreateGzip(archive io.WriteSeeker, stripPath string, content ...string) (bool, liberr.Error) {

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	z := gzip.NewWriter(archive)

	if ok, err := createTar(z, stripPath, content...); err != nil || !ok {
		return ok, err
	}

	if err := z.Close(); err != nil {
		return false, ErrorGzipCreate.ErrorParent(err)
	}

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, nil
}

func createTar(w io.Writer, stripPath string, content ...string) (bool, liberr.Error) {
	var (
		t *tar.Writer
		n int64

		err error
		lEr = ErrorTarCreateAddFile.Error(nil)
	)

	stripPath = strings.TrimLeft(stripPath, "/")
	t = tar.NewWriter(w)

	for i := 0; i < len(content); i++ {
		if content[i] == "" {
			continue
		}

		err = filepath.Walk(content[i], func(file string, inf os.FileInfo, err error) error {
			var (
				e error
				h *tar.Header
				f *os.File
			)

			// generate tar header
			h, e = tar.FileInfoHeader(inf, file)
			if e != nil {
				return e
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			h.Name = filepath.ToSlash(file)

			if stripPath != "" {
				h.Name = filepath.Clean(strings.Replace(strings.TrimLeft(h.Name, "/"), stripPath, "", 1))
			}
			h.Name = strings.TrimLeft(h.Name, "/")

			if h.Name == "" || h.Name == "." {
				return nil
			}

			// write header
			if e = t.WriteHeader(h); e != nil {
				return e
			}

			// if not a dir, write file content
			if !inf.IsDir() {
				//nolint #gosec
				/* #nosec */
				f, e = os.Open(file)

				if e != nil {
					return e
				}

				if _, e = io.Copy(t, f); e != nil {
					return e
				}
			}

			n++
			return nil
		})

		if err != nil {
			lEr.AddParent(err)
			continue
		}
	}

	if n < 1 {
		if lEr.HasParent() {
			return false, lEr
		}

		//nolint #goerr113
		return false, ErrorTarCreate.ErrorParent(fmt.Errorf("no file to add in archive"))
	} else if !lEr.HasParent() {
		lEr = nil
	}

	if err = t.Close(); err != nil {
		return false, ErrorTarCreate.ErrorParent(err)
	}

	return true, lEr
}
