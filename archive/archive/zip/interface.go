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

type readerSize interface {
	Size() int64
}

type readerSeek interface {
	io.Seeker
}

type readerAt interface {
	io.ReadCloser
	io.ReaderAt
}

func NewReader(r io.ReadCloser) (arctps.Reader, error) {
	if s, k := r.(readerSize); !k {
		return nil, fs.ErrInvalid
	} else if ra, ok := r.(readerAt); !ok {
		return nil, fs.ErrInvalid
	} else if rs, o := r.(io.Seeker); !o {
		return nil, fs.ErrInvalid
	} else if siz := s.Size(); siz <= 0 {
		return nil, fs.ErrInvalid
	} else if _, e := rs.Seek(0, io.SeekStart); e != nil {
		return nil, e
	} else if z, err := zip.NewReader(ra, siz); err != nil {
		return nil, err
	} else {
		return &rdr{
			r: r,
			z: z,
		}, nil
	}
}

func NewWriter(w io.WriteCloser) (arctps.Writer, error) {
	return &wrt{
		w: w,
		z: zip.NewWriter(w),
	}, nil
}
