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
	"bufio"
	"io"
)

func Parse(s string) Algorithm {
	var alg = None
	if e := alg.UnmarshalText([]byte(s)); e != nil {
		return None
	} else {
		return alg
	}
}

func Detect(r io.Reader) (Algorithm, io.ReadCloser, error) {
	var (
		err error
		alg Algorithm
		rdr io.ReadCloser
	)

	if alg, rdr, err = DetectOnly(r); err != nil {
		return None, nil, err
	} else if rdr, err = alg.Reader(rdr); err != nil {
		return None, nil, err
	} else {
		return alg, rdr, nil
	}
}

func DetectOnly(r io.Reader) (Algorithm, io.ReadCloser, error) {
	var (
		err error
		alg Algorithm
		bfr = bufio.NewReader(r)
		buf []byte
	)

	if buf, err = bfr.Peek(6); err != nil {
		return None, nil, err
	}

	// Vérifier le type de compression
	switch {
	case Gzip.DetectHeader(buf): // gzip
		alg = Gzip
	case Bzip2.DetectHeader(buf): // bzip2
		alg = Bzip2
	case LZ4.DetectHeader(buf): // lz4
		alg = LZ4
	case XZ.DetectHeader(buf): // xz
		alg = XZ
	default:
		alg = None
	}

	return alg, io.NopCloser(bfr), err
}
