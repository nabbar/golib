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
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

// Code returns the unit code (e.g., "KB", "MB", "GB") for the given Size value.
//
// The unit parameter specifies the unit character to append (e.g., 'B', 'o').
// If unit is 0, the default unit (set by SetDefaultUnit) is used.
//
// This method returns the appropriate prefix based on the Size value:
//   - SizeUnit returns "B" (or the specified unit)
//   - SizeKilo returns "KB" (or "K" + unit)
//   - SizeMega returns "MB" (or "M" + unit)
//   - SizeGiga returns "GB" (or "G" + unit)
//   - SizeTera returns "TB" (or "T" + unit)
//   - SizePeta returns "PB" (or "P" + unit)
//   - SizeExa returns "EB" (or "E" + unit)
//
// Example:
//
//	fmt.Println(size.SizeKilo.Code('B'))  // Output: "KB"
//	fmt.Println(size.SizeMega.Code('o'))  // Output: "Mo"
//	fmt.Println(size.SizeGiga.Code(0))    // Output: "GB" (using default unit)
func (s Size) Code(unit rune) string {
	if unit == 0 {
		unit = defUnit
	}

	switch s {
	case SizeKilo:
		return "K" + string(unit)
	case SizeMega:
		return "M" + string(unit)
	case SizeGiga:
		return "G" + string(unit)
	case SizeTera:
		return "T" + string(unit)
	case SizePeta:
		return "P" + string(unit)
	case SizeExa:
		return "E" + string(unit)
	default:
		return "" + string(unit)
	}
}

// unmarshall is an internal method used by unmarshaling functions to parse
// a byte slice into a Size value.
//
// It uses the parseBytes function to handle the actual parsing logic.
func (s *Size) unmarshall(val []byte) error {
	if tmp, err := parseBytes(val); err != nil {
		return err
	} else {
		*s = tmp
		return nil
	}
}

// ViperDecoderHook returns a mapstructure decode hook function for use with Viper.
//
// This function allows Viper configuration files to automatically decode size values
// from various types (int, uint, float, string, []byte) into Size objects.
//
// The hook supports the following source types:
//   - All integer types (int, int8, int16, int32, int64)
//   - All unsigned integer types (uint, uint8, uint16, uint32, uint64)
//   - Float types (float32, float64)
//   - String (parsed using Parse)
//   - []byte (parsed using ParseByte)
//
// Example usage with Viper:
//
//	import (
//		"github.com/nabbar/golib/size"
//		"github.com/spf13/viper"
//		libmap "github.com/go-viper/mapstructure/v2"
//	)
//
//	type Config struct {
//		MaxFileSize size.Size `mapstructure:"max_file_size"`
//	}
//
//	v := viper.New()
//	v.SetConfigFile("config.yaml")
//
//	var cfg Config
//	err := v.Unmarshal(&cfg, viper.DecodeHook(
//		libmap.ComposeDecodeHookFunc(
//			size.ViperDecoderHook(),
//			// other hooks...
//		),
//	))
//
// With this hook, your config file can contain:
//
//	max_file_size: "10MB"   # String format
//	max_file_size: 10485760 # Integer format
//
// See also:
//   - github.com/nabbar/golib/viper for Viper configuration helpers
//   - github.com/nabbar/golib/config for complete configuration management
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
