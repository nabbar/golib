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

package size

import "math"

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

// Parse parses a size string into a Size value.
//
// The size string is of the form "<number><unit>", where "<number>" is a
// decimal number and "<unit>" is one of the following:
//   - "B" for byte
//   - "K" for kilobyte
//   - "M" for megabyte
//   - "G" for gigabyte
//   - "T" for terabyte
//   - "P" for petabyte
//   - "E" for exabyte
//
// Examples:
//   - "1B" for 1 byte
//   - "2K" for 2 kilobytes
//   - "3M" for 3 megabytes
//   - "4G" for 4 gigabytes
//   - "5T" for 5 terabytes
//   - "6P" for 6 petabytes
//   - "7E" for 7 exabytes
//
// The function returns an error if the size string is invalid.
func Parse(s string) (Size, error) {
	return parseString(s)
}

// ParseByte parses a byte slice into a Size value.
//
// The function is a simple wrapper around ParseBytes.
//
// Examples:
//   - ParseByte([]byte("1B")) for 1 byte
//   - ParseByte([]byte("2K")) for 2 kilobytes
//   - ParseByte([]byte("3M")) for 3 megabytes
//   - ParseByte([]byte("4G")) for 4 gigabytes
//   - ParseByte([]byte("5T")) for 5 terabytes
//   - ParseByte([]byte("6P")) for 6 petabytes
//   - ParseByte([]byte("7E")) for 7 exabytes
//
// The function returns an error if the size string is invalid.
func ParseByte(p []byte) (Size, error) {
	return parseBytes(p)
}

func ParseUint64(s uint64) Size {
	return Size(s)
}

// ParseInt64 converts an int64 value into a Size value.
// The function will always return a positive Size value.
//
// Examples:
//   - ParseInt64(-1)) for -1 byte but will return 1 byte
//   - ParseInt64(1)) for 1 byte
//   - ParseInt64(-1024)) for -1024 bytes but will return 1024 byte
//   - ParseInt64(1024)) for 1024 bytes
func ParseInt64(s int64) Size {
	if s < 0 {
		return Size(uint64(-s))
	} else {
		return Size(uint64(s))
	}
}

// ParseFloat64 converts a float64 value into a Size value.
// The function will always return a positive Size value.
//
// Examples:
//   - ParseFloat64(-1.0)) for -1.0 bytes but will return 1 byte
//   - ParseFloat64(1.0)) for 1.0 bytes
//   - ParseFloat64(-1024.0)) for -1024.0 bytes but will return 1024 bytes
//   - ParseFloat64(1024.0)) for 1024.0 bytes
func ParseFloat64(s float64) Size {
	s = math.Floor(s)

	if s > math.MaxUint64 {
		return Size(uint64(math.MaxUint64))
	} else if -s > math.MaxUint64 {
		return Size(uint64(math.MaxUint64))
	} else if s > 0 {
		return Size(uint64(s))
	} else {
		return Size(uint64(-s))
	}
}

// GetSize parses a size string and returns a Size value and a boolean indicating success.
// Deprecated: see Parse
func GetSize(s string) (sizeBytes Size, success bool) {
	if z, e := parseString(s); e != nil {
		return SizeNul, false
	} else {
		return z, true
	}
}

// SizeFromInt64 converts an int64 value into a Size value.
// Deprecated: see ParseInt64
func SizeFromInt64(val int64) Size {
	return ParseInt64(val)
}

// SizeFromFloat64 converts a float64 value into a Size value.
// Deprecated: see ParseFloat64
func SizeFromFloat64(val float64) Size {
	return ParseFloat64(val)
}

// ParseSize parses a size string into a Size value.
// Deprecated: see Parse
func ParseSize(s string) (Size, error) {
	return Parse(s)
}

// ParseByteAsSize parses a byte slice into a Size value.
// Deprecated: see ParseByte
func ParseByteAsSize(p []byte) (Size, error) {
	return ParseByte(p)
}
