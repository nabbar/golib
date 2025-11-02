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
		return fmt.Sprintf(uni, "P")
	case SizeExa:
		return fmt.Sprintf(uni, "E")
	}

	return fmt.Sprintf(uni, "")
}

func (s Size) isMax(size Size) bool {
	val := math.Abs(size.Float64())
	uni := math.Abs(s.Float64())
	return val > uni
}

func (s Size) sizeByUnit(unit Size) float64 {
	if uint64(s/unit) > _maxFloat64 {
		return math.MaxFloat64
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

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Size(0)
			f func() error
		)

		// Check if the target type matches the expected one
		if to != reflect.TypeOf(z) {
			return data, nil
		}

		// Check if the data type matches the expected one
		if from.Kind() == reflect.Int {
			if i, k := data.(int); k {
				f = func() error {
					z = ParseInt64(int64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Int8 {
			if i, k := data.(int8); k {
				f = func() error {
					z = ParseInt64(int64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Int16 {
			if i, k := data.(int16); k {
				f = func() error {
					z = ParseInt64(int64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Int32 {
			if i, k := data.(int32); k {
				f = func() error {
					z = ParseInt64(int64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Int64 {
			if i, k := data.(int64); k {
				f = func() error {
					z = ParseInt64(i)
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint {
			if i, k := data.(uint); k {
				f = func() error {
					z = ParseUint64(uint64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint8 {
			if i, k := data.(uint8); k {
				f = func() error {
					z = ParseUint64(uint64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint16 {
			if i, k := data.(uint16); k {
				f = func() error {
					z = ParseUint64(uint64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint32 {
			if i, k := data.(uint32); k {
				f = func() error {
					z = ParseUint64(uint64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint64 {
			if i, k := data.(uint64); k {
				f = func() error {
					z = ParseUint64(i)
					return nil
				}
			}
		} else if from.Kind() == reflect.Float32 {
			if i, k := data.(float32); k {
				f = func() error {
					z = ParseFloat64(float64(i))
					return nil
				}
			}
		} else if from.Kind() == reflect.Float64 {
			if i, k := data.(float64); k {
				f = func() error {
					z = ParseFloat64(i)
					return nil
				}
			}
		} else if from.Kind() == reflect.String {
			if s, k := data.(string); k {
				f = func() error {
					return z.unmarshall([]byte(s))
				}
			}
		} else if from.Kind() == reflect.Slice {
			if p, k := data.([]byte); k {
				f = func() error {
					return z.unmarshall(p)
				}
			}
		}

		if f == nil {
			return data, nil
		} else if err := f(); err != nil {
			return nil, err
		} else {
			return z, nil
		}
	}
}
