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

package bz2

import (
	"compress/bzip2"
	"io"

	"github.com/nabbar/golib/errors"
)

func GetFile(src io.ReadSeeker, dst io.WriteSeeker) errors.Error {
	if _, e := src.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	}

	r := bzip2.NewReader(src)

	//nolint #nosec
	/* #nosec */
	if _, e := io.Copy(dst, r); e != nil {
		return ErrorIOCopy.Error(e)
	} else if _, e = dst.Seek(0, io.SeekStart); e != nil {
		return ErrorFileSeek.Error(e)
	} else {
		return nil
	}
}
