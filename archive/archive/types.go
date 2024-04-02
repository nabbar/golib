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

import "bytes"

type Algorithm uint8

const (
	None Algorithm = iota
	Tar
	Zip
)

func (a Algorithm) IsNone() bool {
	return a == None
}

func (a Algorithm) String() string {
	switch a {
	case Tar:
		return "tar"
	case Zip:
		return "zip"
	default:
		return "none"
	}
}

func (a Algorithm) Extension() string {
	switch a {
	case Tar:
		return ".tar"
	case Zip:
		return ".zip"
	default:
		return ""
	}
}

func (a Algorithm) DetectHeader(h []byte) bool {
	if len(h) < 263 {
		return false
	}

	switch a {
	case Tar:
		exp := append([]byte("ustar"), 0x00)
		val := h[257:263]
		return bytes.Equal(val, exp)
	case Zip:
		exp := []byte{0x50, 0x4b, 0x03, 0x04}
		return bytes.Equal(h[0:4], exp)
	default:
		return false
	}
}
