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

package gzip

import (
	gz "compress/gzip"
	"io"

	libfpg "github.com/nabbar/golib/file/progress"

	"github.com/nabbar/golib/errors"
)

func GetFile(src io.ReadSeeker, dst io.WriteSeeker) errors.Error {
	var siz = getGunZipSize(src)

	if d, k := dst.(libfpg.Progress); k && siz > 0 {
		d.Reset(siz)
	}

	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else if siz > 0 {
	}

	r, e := gz.NewReader(src)
	if e != nil {
		return ErrorGZReader.Error(e)
	}

	defer func() {
		_ = r.Close()
	}()

	//nolint #nosec
	/* #nosec */
	if _, e = io.Copy(dst, r); e != nil {
		return ErrorIOCopy.Error(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else {
		return nil
	}
}

func getGunZipSize(src io.ReadSeeker) int64 {
	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return 0
	}

	r, e := gz.NewReader(src)
	if e != nil {
		return 0
	}

	defer func() {
		_ = r.Close()
	}()

	if s, k := src.(libfpg.Progress); k {
		s.RegisterFctReset(func(size, current int64) {

		})
		s.RegisterFctIncrement(func(size int64) {

		})
		s.RegisterFctEOF(func() {

		})
	}

	var n int64
	n, e = io.Copy(io.Discard, r)

	if e != nil {
		return 0
	}

	return n
}
