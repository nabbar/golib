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

	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
)

func Create(archive libiot.FileProgress, content ...string) (bool, liberr.Error) {

	var (
		t *tar.Writer
		n int64

		err error
		lEr = ErrorTarCreateAddFile.Error(nil)
	)

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	t = tar.NewWriter(archive)

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

			// write header
			if e = t.WriteHeader(h); err != nil {
				return err
			}

			// if not a dir, write file content
			if !inf.IsDir() {
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

		return false, ErrorTarCreate.ErrorParent(fmt.Errorf("no file to add in archive"))
	} else if !lEr.HasParent() {
		lEr = nil
	}

	if err = t.Close(); err != nil {
		return false, ErrorTarCreate.ErrorParent(err)
	}

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, lEr
}

func CreateGzip(archive libiot.FileProgress, content ...string) (bool, liberr.Error) {

	var (
		z *gzip.Writer
		t *tar.Writer
		n int64

		err error
		lEr = ErrorTarCreateAddFile.Error(nil)
	)

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	z = gzip.NewWriter(archive)
	t = tar.NewWriter(z)

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

			// write header
			if e = t.WriteHeader(h); err != nil {
				return err
			}

			// if not a dir, write file content
			if !inf.IsDir() {
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

		return false, ErrorTarCreate.ErrorParent(fmt.Errorf("no file to add in archive"))
	} else if !lEr.HasParent() {
		lEr = nil
	}

	if err = t.Close(); err != nil {
		return false, ErrorTarCreate.ErrorParent(err)
	}

	if err = z.Close(); err != nil {
		return false, ErrorGzipCreate.ErrorParent(err)
	}

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, lEr
}
