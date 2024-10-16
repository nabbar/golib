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

import "bytes"

type Algorithm uint8

const (
	None Algorithm = iota
	Bzip2
	Gzip
	LZ4
	XZ
)

func List() []Algorithm {
	return []Algorithm{
		None,
		Bzip2,
		Gzip,
		LZ4,
		XZ,
	}
}

func ListString() []string {
	var (
		lst = List()
		res = make([]string, len(lst))
	)
	for i := range lst {
		res[i] = lst[i].String()
	}
	return res
}

func (a Algorithm) IsNone() bool {
	return a == None
}

func (a Algorithm) String() string {
	switch a {
	case Gzip:
		return "gzip"
	case Bzip2:
		return "bzip2"
	case LZ4:
		return "lz4"
	case XZ:
		return "xz"
	default:
		return "none"
	}
}

func (a Algorithm) Extension() string {
	switch a {
	case Gzip:
		return ".gz"
	case Bzip2:
		return ".bz2"
	case LZ4:
		return ".lz4"
	case XZ:
		return ".xz"
	default:
		return ""
	}
}

func (a Algorithm) DetectHeader(h []byte) bool {
	if len(h) < 6 {
		return false
	}

	switch a {
	case Gzip:
		exp := []byte{31, 139}
		return bytes.Equal(h[0:2], exp)
	case Bzip2:
		exp := []byte{'B', 'Z', 'h'}
		return bytes.Equal(h[0:3], exp) && h[3] >= '0' && h[3] <= '9'
	case LZ4:
		exp := []byte{0x04, 0x22, 0x4D, 0x18}
		return bytes.Equal(h[0:4], exp)
	case XZ:
		exp := []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}
		alt := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		return bytes.Equal(h[0:6], exp) || bytes.Equal(h[0:6], alt)
	default:
		return false
	}
}
