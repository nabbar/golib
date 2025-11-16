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

import (
	"fmt"
	"math"
)

const (
	// _space is the separator used between the numeric value and unit in formatted output.
	_space = " "

	// FormatRound0 formats size values with no decimal places (e.g., "1 MB").
	FormatRound0 = "%.0f"

	// FormatRound1 formats size values with 1 decimal place (e.g., "1.5 MB").
	FormatRound1 = "%.1f"

	// FormatRound2 formats size values with 2 decimal places (e.g., "1.50 MB").
	// This is the default format used by the String() method.
	FormatRound2 = "%.2f"

	// FormatRound3 formats size values with 3 decimal places (e.g., "1.500 MB").
	FormatRound3 = "%.3f"
)

var (
	// _maxFloat64 is the maximum uint64 value that can be represented as a float64.
	_maxFloat64 = uint64(math.Ceil(math.MaxFloat64))

	// _maxFloat32 is the maximum uint64 value that can be represented as a float32.
	_maxFloat32 = uint64(math.Ceil(math.MaxFloat32))
)

// String returns a human-readable string representation of the Size.
//
// The method automatically selects the most appropriate unit (B, KB, MB, GB, TB, PB, EB)
// and formats the value with 2 decimal places (using FormatRound2).
//
// Example:
//
//	size := size.ParseUint64(1572864)
//	fmt.Println(size.String()) // Output: "1.50 MB"
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.String()) // Output: "1.00 KB"
func (s Size) String() string {
	u := s.Unit(0)

	if len(u) > 0 {
		return s.Format(FormatRound2) + _space + u
	} else {
		return s.Format(FormatRound2)
	}
}

// Int64 converts the Size to an int64 value representing bytes.
//
// If the Size value is greater than math.MaxInt64, the method returns
// math.MaxInt64 to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Int64()) // Output: 1024
func (s Size) Int64() int64 {
	if i := uint64(s); i > uint64(math.MaxInt64) {
		// overflow
		return math.MaxInt64
	} else {
		return int64(i)
	}
}

// Int32 converts the Size to an int32 value representing bytes.
//
// If the Size value is greater than math.MaxInt32, the method returns
// math.MaxInt32 to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Int32()) // Output: 1024
func (s Size) Int32() int32 {
	if i := uint64(s); i > uint64(math.MaxInt32) {
		// overflow
		return math.MaxInt32
	} else {
		return int32(i)
	}
}

// Int converts the Size to an int value representing bytes.
//
// If the Size value is greater than math.MaxInt, the method returns
// math.MaxInt to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Int()) // Output: 1024
func (s Size) Int() int {
	if i := uint64(s); i > uint64(math.MaxInt) {
		// overflow
		return math.MaxInt
	} else {
		return int(i)
	}
}

// Uint64 returns the Size value as a uint64 representing bytes.
//
// This is the most direct way to get the raw byte count.
//
// Example:
//
//	size := size.ParseUint64(1048576)
//	fmt.Println(size.Uint64()) // Output: 1048576
func (s Size) Uint64() uint64 {
	return uint64(s)
}

// Uint32 converts the Size to a uint32 value representing bytes.
//
// If the Size value is greater than math.MaxUint32, the method returns
// math.MaxUint32 to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Uint32()) // Output: 1024
func (s Size) Uint32() uint32 {
	if i := uint64(s); i > uint64(math.MaxUint32) {
		// overflow
		return math.MaxUint32
	} else {
		return uint32(i)
	}
}

// Uint converts the Size to a uint value representing bytes.
//
// If the Size value is greater than math.MaxUint, the method returns
// math.MaxUint to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Uint()) // Output: 1024
func (s Size) Uint() uint {
	if i := uint64(s); i > uint64(math.MaxUint) {
		// overflow
		return math.MaxUint
	} else {
		return uint(i)
	}
}

// Float64 converts the Size to a float64 value representing bytes.
//
// If the Size value is greater than what can be represented as a float64,
// the method returns math.MaxFloat64 to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1048576)
//	fmt.Println(size.Float64()) // Output: 1048576.0
func (s Size) Float64() float64 {
	if i := uint64(s); i > _maxFloat64 {
		// overflow
		return math.MaxFloat64
	} else {
		return float64(i)
	}
}

// Float32 converts the Size to a float32 value representing bytes.
//
// If the Size value is greater than what can be represented as a float32,
// the method returns math.MaxFloat32 to prevent overflow.
//
// Example:
//
//	size := size.ParseUint64(1024)
//	fmt.Println(size.Float32()) // Output: 1024.0
func (s Size) Float32() float32 {
	if i := uint64(s); i > _maxFloat32 {
		// overflow
		return math.MaxFloat32
	} else {
		return float32(i)
	}
}

// Format returns a formatted string representation of the Size using the specified format.
//
// The format string should be a printf-style float format (e.g., "%.2f").
// The method automatically selects the most appropriate unit and formats
// the value accordingly.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	fmt.Println(size.Format(FormatRound0)) // Output: "2"
//	fmt.Println(size.Format(FormatRound1)) // Output: "1.5"
//	fmt.Println(size.Format(FormatRound2)) // Output: "1.50"
//
// Note: This method returns only the numeric part. Use String() to include the unit.
func (s Size) Format(format string) string {
	switch {
	case SizeExa.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizeExa))
	case SizePeta.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizePeta))
	case SizeTera.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizeTera))
	case SizeGiga.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizeGiga))
	case SizeMega.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizeMega))
	case SizeKilo.isMax(s):
		return fmt.Sprintf(format, s.sizeByUnit(SizeKilo))
	default:
		return fmt.Sprintf(format, s.sizeByUnit(SizeUnit))
	}
}

// Unit returns the unit string for the Size value.
//
// The method automatically selects the most appropriate unit (B, KB, MB, GB, TB, PB, EB)
// based on the Size value. The unit parameter specifies the unit character to use
// (e.g., 'B' for Byte, 'o' for octet). If unit is 0, the default unit is used.
//
// Example:
//
//	size := size.ParseUint64(1048576) // 1 MB
//	fmt.Println(size.Unit('B'))        // Output: "MB"
//	fmt.Println(size.Unit('o'))        // Output: "Mo"
//
// See also: Code() for getting the unit code for predefined Size constants.
func (s Size) Unit(unit rune) string {
	switch {
	case SizeExa.isMax(s):
		return SizeExa.Code(unit)
	case SizePeta.isMax(s):
		return SizePeta.Code(unit)
	case SizeTera.isMax(s):
		return SizeTera.Code(unit)
	case SizeGiga.isMax(s):
		return SizeGiga.Code(unit)
	case SizeMega.isMax(s):
		return SizeMega.Code(unit)
	case SizeKilo.isMax(s):
		return SizeKilo.Code(unit)
	default:
		return SizeUnit.Code(unit)
	}
}

// KiloBytes returns the size in kilobytes (KB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1536) // 1.5 KB
//	fmt.Println(size.KiloBytes())   // Output: 1
//
//	size := size.ParseUint64(2048) // 2 KB
//	fmt.Println(size.KiloBytes())   // Output: 2
func (s Size) KiloBytes() uint64 {
	return s.floorByUnit(SizeKilo)
}

// MegaBytes returns the size in megabytes (MB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1572864) // 1.5 MB
//	fmt.Println(size.MegaBytes())      // Output: 1
//
//	size := size.ParseUint64(2097152) // 2 MB
//	fmt.Println(size.MegaBytes())      // Output: 2
func (s Size) MegaBytes() uint64 {
	return s.floorByUnit(SizeMega)
}

// GigaBytes returns the size in gigabytes (GB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1610612736) // 1.5 GB
//	fmt.Println(size.GigaBytes())         // Output: 1
//
//	size := size.ParseUint64(2147483648) // 2 GB
//	fmt.Println(size.GigaBytes())         // Output: 2
func (s Size) GigaBytes() uint64 {
	return s.floorByUnit(SizeGiga)
}

// TeraBytes returns the size in terabytes (TB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1649267441664) // 1.5 TB
//	fmt.Println(size.TeraBytes())            // Output: 1
func (s Size) TeraBytes() uint64 {
	return s.floorByUnit(SizeTera)
}

// PetaBytes returns the size in petabytes (PB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1688849860263936) // 1.5 PB
//	fmt.Println(size.PetaBytes())               // Output: 1
func (s Size) PetaBytes() uint64 {
	return s.floorByUnit(SizePeta)
}

// ExaBytes returns the size in exabytes (EB), floored to the nearest whole number.
//
// Example:
//
//	size := size.ParseUint64(1729382256910270464) // 1.5 EB
//	fmt.Println(size.ExaBytes())                   // Output: 1
func (s Size) ExaBytes() uint64 {
	return s.floorByUnit(SizeExa)
}
