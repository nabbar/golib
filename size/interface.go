/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package bytes

type Size uint64

const (
	SizeNul  Size = 0
	SizeUnit Size = 1
	SizeKilo Size = 1 << 10
	SizeMega Size = 1 << 20
	SizeGiga Size = 1 << 30
	SizeTera Size = 1 << 40
	SizePeta Size = 1 << 50
	SizeExa  Size = 1 << 60
)

var defUnit = 'B'

func SetDefaultUnit(unit rune) {
	if unit == 0 {
		defUnit = 'B'
	} else if s := string(unit); len(s) < 1 {
		defUnit = 'B'
	} else {
		defUnit = unit
	}
}

func GetSize(s string) (sizeBytes Size, success bool) {
	if z, e := parseString(s); e != nil {
		return SizeNul, false
	} else {
		return z, true
	}
}

func SizeFromInt64(val int64) Size {
	v := uint64(val)
	return Size(v)
}

func Parse(s string) (Size, error) {
	return parseString(s)
}

func ParseSize(s string) (Size, error) {
	return parseString(s)
}

func ParseByteAsSize(p []byte) (Size, error) {
	return parseBytes(p)
}
