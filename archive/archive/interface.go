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
	"bufio"
	"io"

	arctps "github.com/nabbar/golib/archive/archive/types"
)

func Parse(s string) Algorithm {
	var alg = None
	if e := alg.UnmarshalText([]byte(s)); e != nil {
		return None
	} else {
		return alg
	}
}

// Detect will try to detect the algorithm of the input and returns the corresponding reader.
// This function will read the first 6 bytes of the input and try to detect the algorithm.
// If algorithm detection trigger a supported algorithm, the function will try to convert
// the input to appropriate reader for the archive.
// In any case, if en error occurs, the function will return an error.
// Otherwise, the function will return the algorithm, the reader and a nil error.
//
// If the input is a zip archive and the input is not a io.ReaderAt compatible, the function will return.
// If the input is a tar archive and the input is a strict io.ReadCloser, with no seek or read at compatible,
// the reader result could be use only for one time.
//
// This difference are based on how the algorithm work:
// - zip: will use dictionary / catalog to store position and metadata of each embedded file
// - tar: (TAR = Tape ARchive) will store each file continuously beginning with his metadata and following with his content.
//
// As that, a strict reader could be use only for tar archive and the reader result could be use only for one time.
// In this case, the best use if calling the walk function of the reader.
func Detect(r io.ReadCloser) (Algorithm, arctps.Reader, io.ReadCloser, error) {
	var (
		err error
		buf []byte
		bfr = &rdr{
			r: r,
			b: bufio.NewReader(r),
		}
	)

	if buf, err = bfr.Peek(265); err != nil {
		return None, nil, nil, err
	}

	switch {
	case Tar.DetectHeader(buf): // tar
		if t, e := Tar.Reader(bfr); e != nil {
			return None, nil, nil, e
		} else {
			return Tar, t, bfr, nil
		}

	case Zip.DetectHeader(buf): // zip
		bfr.b = nil // do not use buffer (using ReaderAt)
		if z, e := Zip.Reader(bfr); e != nil {
			return None, nil, nil, e
		} else {
			return Zip, z, bfr, nil
		}

	default:
		return None, nil, bfr, nil
	}
}
