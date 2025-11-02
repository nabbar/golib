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

package network

import (
	"fmt"
	"strings"

	libsiz "github.com/nabbar/golib/size"
)

// Bytes represents a byte size using the github.com/nabbar/golib/size package
// This provides consistent byte size handling with binary units (KB=1024, MB=1024², etc.)
type Bytes libsiz.Size

// String returns the numeric string representation of the byte value
func (n Bytes) String() string {
	return fmt.Sprintf("%d", n)
}

// FormatUnitFloat formats the bytes with a floating point precision and appropriate unit suffix (KB, MB, GB, etc.)
// If precision < 1, it delegates to FormatUnitInt()
// The format includes padding to ensure consistent alignment
func (n Bytes) FormatUnitFloat(precision int) string {
	if precision < 1 {
		return n.FormatUnitInt()
	}

	s := libsiz.Size(n)
	format := fmt.Sprintf("%%.%df", precision)

	// Determine the appropriate unit and value
	var valueStr string
	var unitStr string

	switch {
	case s >= libsiz.SizeExa:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizeExa))
		unitStr = "EB"
	case s >= libsiz.SizePeta:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizePeta))
		unitStr = "PB"
	case s >= libsiz.SizeTera:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizeTera))
		unitStr = "TB"
	case s >= libsiz.SizeGiga:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizeGiga))
		unitStr = "GB"
	case s >= libsiz.SizeMega:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizeMega))
		unitStr = "MB"
	case s >= libsiz.SizeKilo:
		valueStr = fmt.Sprintf(format, float64(s)/float64(libsiz.SizeKilo))
		unitStr = "KB"
	default:
		valueStr = fmt.Sprintf(format, float64(s))
		unitStr = " "
	}

	// Add padding for alignment (4 characters for number part before decimal)
	parts := strings.SplitN(valueStr, ".", 2)
	if len(parts) > 0 && len(parts[0]) < _MaxSizeOfPad_ {
		padding := strings.Repeat(" ", _MaxSizeOfPad_-len(parts[0]))
		return padding + valueStr + " " + unitStr
	}

	return valueStr + " " + unitStr
}

// FormatUnitInt formats the bytes as an integer with appropriate unit suffix (KB, MB, GB, etc.)
// The format includes padding to ensure 4-character alignment for the numeric part
func (n Bytes) FormatUnitInt() string {
	s := libsiz.Size(n)

	// Determine the appropriate unit
	var valueStr string
	var unitStr string

	switch {
	case s >= libsiz.SizeExa:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizeExa))
		unitStr = "EB"
	case s >= libsiz.SizePeta:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizePeta))
		unitStr = "PB"
	case s >= libsiz.SizeTera:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizeTera))
		unitStr = "TB"
	case s >= libsiz.SizeGiga:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizeGiga))
		unitStr = "GB"
	case s >= libsiz.SizeMega:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizeMega))
		unitStr = "MB"
	case s >= libsiz.SizeKilo:
		valueStr = fmt.Sprintf("%.0f", float64(s)/float64(libsiz.SizeKilo))
		unitStr = "KB"
	default:
		valueStr = fmt.Sprintf("%d", s)
		unitStr = " "
	}

	// Add padding for alignment (4 characters for number part)
	if len(valueStr) < _MaxSizeOfPad_ {
		padding := strings.Repeat(" ", _MaxSizeOfPad_-len(valueStr))
		return padding + valueStr + " " + unitStr
	}

	return valueStr + " " + unitStr
}

// AsNumber converts Bytes to Number type for decimal unit operations
func (n Bytes) AsNumber() Number {
	return Number(n)
}

// AsUint64 returns the byte value as uint64
func (n Bytes) AsUint64() uint64 {
	return uint64(n)
}

// AsFloat64 returns the byte value as float64
func (n Bytes) AsFloat64() float64 {
	return float64(n)
}
