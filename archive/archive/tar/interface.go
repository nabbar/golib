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

	arctps "github.com/nabbar/golib/archive/archive/types"
)

// NewReader will create a new reader from the provided io.ReadCloser.
// It returns the reader and a nil error if the creation succeed.
// The reader is a io.ReadCloser compatible with the tar archive algorithm.
// It is the caller responsibility to close the provided io.ReadCloser to release resources.
func NewReader(r io.ReadCloser) (arctps.Reader, error) {
	return &rdr{
		r: r,
		z: tar.NewReader(r),
	}, nil
}

// NewWriter will create a new writer from the provided io.WriteCloser.
// It returns the writer and a nil error if the creation succeed.
// The writer is a io.WriteCloser compatible with the tar archive algorithm.
// It is the caller responsibility to close the provided io.WriteCloser to release resources.
func NewWriter(w io.WriteCloser) (arctps.Writer, error) {
	return &wrt{
		w: w,
		z: tar.NewWriter(w),
	}, nil
}
