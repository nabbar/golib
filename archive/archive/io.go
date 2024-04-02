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

package archive

import (
	"errors"
	"io"

	arctar "github.com/nabbar/golib/archive/archive/tar"
	arctps "github.com/nabbar/golib/archive/archive/types"
	arczip "github.com/nabbar/golib/archive/archive/zip"
)

var (
	ErrInvalidAlgorithm = errors.New("invalid algorithm")
)

func (a Algorithm) Reader(r io.ReadCloser) (arctps.Reader, error) {
	switch a {
	case Tar:
		return arctar.NewReader(r)
	case Zip:
		return arczip.NewReader(r)
	default:
		return nil, ErrInvalidAlgorithm
	}
}

func (a Algorithm) Writer(w io.WriteCloser) (arctps.Writer, error) {
	switch a {
	case Tar:
		return arctar.NewWriter(w)
	case Zip:
		return arczip.NewWriter(w)
	default:
		return nil, ErrInvalidAlgorithm
	}
}
