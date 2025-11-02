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
	_space       = " "
	FormatRound0 = "%.0f"
	FormatRound1 = "%.1f"
	FormatRound2 = "%.2f"
	FormatRound3 = "%.3f"
)

var (
	_maxFloat64 = uint64(math.Ceil(math.MaxFloat64))
	_maxFloat32 = uint64(math.Ceil(math.MaxFloat32))
)

func (s Size) String() string {
	u := s.Unit(0)

	if len(u) > 0 {
		return s.Format(FormatRound2) + _space + u
	} else {
		return s.Format(FormatRound2)
	}
}

func (s Size) Int64() int64 {
	if i := uint64(s); i > uint64(math.MaxInt64) {
		// overflow
		return math.MaxInt64
	} else {
		return int64(i)
	}
}

func (s Size) Int32() int32 {
	if i := uint64(s); i > uint64(math.MaxInt32) {
		// overflow
		return math.MaxInt32
	} else {
		return int32(i)
	}
}

func (s Size) Int() int {
	if i := uint64(s); i > uint64(math.MaxInt) {
		// overflow
		return math.MaxInt
	} else {
		return int(i)
	}
}

func (s Size) Uint64() uint64 {
	return uint64(s)
}

func (s Size) Uint32() uint32 {
	if i := uint64(s); i > uint64(math.MaxUint32) {
		// overflow
		return math.MaxUint32
	} else {
		return uint32(i)
	}
}

func (s Size) Uint() uint {
	if i := uint64(s); i > uint64(math.MaxUint) {
		// overflow
		return math.MaxUint
	} else {
		return uint(i)
	}
}

func (s Size) Float64() float64 {
	if i := uint64(s); i > _maxFloat64 {
		// overflow
		return math.MaxFloat64
	} else {
		return float64(i)
	}
}

func (s Size) Float32() float32 {
	if i := uint64(s); i > _maxFloat32 {
		// overflow
		return math.MaxFloat32
	} else {
		return float32(i)
	}
}

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

func (s Size) KiloBytes() uint64 {
	return s.floorByUnit(SizeKilo)
}

func (s Size) MegaBytes() uint64 {
	return s.floorByUnit(SizeMega)
}

func (s Size) GigaBytes() uint64 {
	return s.floorByUnit(SizeGiga)
}

func (s Size) TeraBytes() uint64 {
	return s.floorByUnit(SizeTera)
}

func (s Size) PetaBytes() uint64 {
	return s.floorByUnit(SizePeta)
}

func (s Size) ExaBytes() uint64 {
	return s.floorByUnit(SizeExa)
}
