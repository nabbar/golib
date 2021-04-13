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

package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	liberr "github.com/nabbar/golib/errors"
)

func Create(archive io.WriteSeeker, content ...string) (bool, liberr.Error) {
	var (
		w *gzip.Writer
		f *os.File

		err error
	)

	if len(content) != 1 {
		//nolint #goerr113
		return false, ErrorParamsMismatching.ErrorParent(fmt.Errorf("content path must be limited to strictly one contents"))
	}

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	if _, err = os.Stat(content[0]); err != nil {
		return false, ErrorParamsEmpty.ErrorParent(err)
	}

	w = gzip.NewWriter(archive)

	defer func() {
		if w != nil {
			_ = w.Close()
		}
	}()

	if f, err = os.Open(content[0]); err != nil {
		return false, ErrorFileOpen.ErrorParent(err)
	}

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	if _, err = io.Copy(w, f); err != nil {
		return false, ErrorIOCopy.ErrorParent(err)
	}

	if err = w.Close(); err != nil {
		return false, ErrorGZCreate.ErrorParent(err)
	}

	if _, err = archive.Seek(0, io.SeekStart); err != nil {
		return false, ErrorFileSeek.ErrorParent(err)
	}

	return true, nil
}
