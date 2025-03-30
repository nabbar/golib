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

import (
	"fmt"
	"math"
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

func (s Size) Code(unit rune) string {
	var uni string

	if unit == 0 {
		uni = "%s" + string(defUnit)
	} else {
		uni = "%s" + string(unit)
	}

	switch s {
	case SizeUnit:
		return fmt.Sprintf(uni, "")
	case SizeKilo:
		return fmt.Sprintf(uni, "K")
	case SizeMega:
		return fmt.Sprintf(uni, "M")
	case SizeGiga:
		return fmt.Sprintf(uni, "G")
	case SizeTera:
		return fmt.Sprintf(uni, "T")
	case SizePeta:
		return fmt.Sprintf(uni, "E")
	}

	return fmt.Sprintf(uni, "")
}

func (s Size) isMax(size Size) bool {
	val := math.Abs(size.Float64())
	uni := math.Abs(s.Float64())
	return val >= (10 * uni)
}

func (s Size) sizeByUnit(unit Size) float64 {
	if s > 0 && uint64(s/unit) > _maxFloat64 {
		// overflow
		return math.MaxFloat64
	} else if s < 0 && uint64(-s/unit) > _maxFloat64 {
		// overflow
		return -math.MaxFloat64
	} else {
		return float64(s / unit)
	}
}

func (s Size) floorByUnit(unit Size) uint64 {
	return uint64(math.Floor(s.sizeByUnit(unit)))
}

func (s *Size) unmarshall(val []byte) error {
	if tmp, err := parseBytes(val); err != nil {
		return err
	} else {
		*s = tmp
		return nil
	}
}

func (s *Size) Mul(v float64) {
	v = s.Float64() * v
	if v > math.MaxUint64 {
		*s = math.MaxUint64
	} else {
		*s = Size(v)
	}
}

func (s *Size) Div(v float64) {
	v = math.Ceil(s.Float64() / v)
	if v > math.MaxUint64 {
		*s = math.MaxUint64
	} else {
		*s = Size(v)
	}
}

func (s *Size) Add(v uint64) {
	v = s.Uint64() + v
	if v > math.MaxUint64 {
		*s = math.MaxUint64
	} else {
		*s = Size(v)
	}
}

func (s *Size) Sub(v uint64) {
	v = s.Uint64() - v
	if v > math.MaxUint64 {
		*s = math.MaxUint64
	} else {
		*s = Size(v)
	}
}

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Size(0)
			t string
			k bool
		)

		// Check if the data type matches the expected one
		if from.Kind() != reflect.String {
			return data, nil
		} else if t, k = data.(string); !k {
			return data, nil
		}

		// Check if the target type matches the expected one
		if to != reflect.TypeOf(z) {
			return data, nil
		}

		// Format/decode/parse the data and return the new value
		if e := z.unmarshall([]byte(t)); e != nil {
			return nil, e
		} else {
			return z, nil
		}
	}
}
