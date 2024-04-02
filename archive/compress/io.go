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

package compress

import (
	"compress/bzip2"
	"compress/gzip"
	"io"

	bz2 "github.com/dsnet/compress/bzip2"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz"
)

func (a Algorithm) Reader(r io.Reader) (io.ReadCloser, error) {
	switch a {
	case Bzip2:
		return io.NopCloser(bzip2.NewReader(r)), nil
	case Gzip:
		return gzip.NewReader(r)
	case LZ4:
		return io.NopCloser(lz4.NewReader(r)), nil
	case XZ:
		c, e := xz.NewReader(r)
		return io.NopCloser(c), e
	default:
		return io.NopCloser(r), nil
	}
}

func (a Algorithm) Writer(w io.WriteCloser) (io.WriteCloser, error) {
	switch a {
	case Bzip2:
		return bz2.NewWriter(w, nil)
	case Gzip:
		return gzip.NewWriter(w), nil
	case LZ4:
		return lz4.NewWriter(w), nil
	case XZ:
		return xz.NewWriter(w)
	default:
		return w, nil
	}
}
