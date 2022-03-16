/*
 *  MIT License
 *
 *  Copyright (c) 2022 Nicolas JUHEL
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
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	liberr "github.com/nabbar/golib/errors"
)

func Create(archive io.WriteSeeker, stripPath string, comment string, content ...string) (bool, liberr.Error) {

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	var (
		z = zip.NewWriter(archive)
	)

	if z == nil {
		return false, ErrorZipOpen.Error(nil)
	} else {
		z.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
			return flate.NewWriter(out, flate.BestCompression)
		})
	}

	if comment != "" {
		if e := z.SetComment(comment); e != nil {
			return false, ErrorZipComment.ErrorParent(e)
		}
	}

	if ok, err := addFileToZip(z, stripPath, content...); err != nil || !ok {
		return ok, err
	}

	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, nil
}

func addFileToZip(z *zip.Writer, stripPath string, content ...string) (bool, liberr.Error) {
	var (
		n   int64
		err error
		w   io.Writer
		lEr = ErrorZipAddFile.Error(nil)
	)

	stripPath = strings.TrimLeft(stripPath, "/")

	for i := 0; i < len(content); i++ {
		if content[i] == "" {
			continue
		}

		err = filepath.Walk(content[i], func(file string, inf os.FileInfo, err error) error {
			var (
				e error
				h *zip.FileHeader
				f *os.File
			)

			// generate tar header
			h, e = zip.FileInfoHeader(inf)
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
			if w, e = z.CreateHeader(h); e != nil {
				return e
			} else if !inf.IsDir() {
				// if not a dir, write file content

				//nolint #gosec
				/* #nosec */
				f, e = os.Open(file)

				if e != nil {
					return e
				}

				if _, e = io.Copy(w, f); e != nil {
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

		return false, ErrorZipCreate.ErrorParent(fmt.Errorf("no file to add in archive"))
	} else if !lEr.HasParent() {
		lEr = nil
	}

	if err = z.Close(); err != nil {
		return false, ErrorZipCreate.ErrorParent(err)
	}

	return true, lEr
}
