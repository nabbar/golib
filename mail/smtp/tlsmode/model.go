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

package tlsmode

import (
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

// ViperDecoderHook returns a mapstructure decode hook function for TLSMode.
//
// This hook enables automatic decoding of TLSMode values when using Viper
// for configuration management. It supports decoding from various primitive
// types including all integer types, float types, strings, and byte slices.
//
// Supported input types:
//   - int, int8, int16, int32, int64: decoded with ParseInt64
//   - uint, uint8, uint16, uint32, uint64: decoded with ParseUint64
//   - float32, float64: decoded with ParseFloat64
//   - string: decoded with Parse
//   - []byte: decoded with ParseBytes
//
// The hook returns the original data unchanged if:
//   - The target type is not TLSMode
//   - The source type is not supported
//
// Example usage with Viper:
//
//	import (
//	    "github.com/spf13/viper"
//	    "github.com/nabbar/golib/mail/smtp/tlsmode"
//	)
//
//	type Config struct {
//	    TLS tlsmode.TLSMode `mapstructure:"tls"`
//	}
//
//	v := viper.New()
//	v.SetConfigType("yaml")
//	// ... configure viper ...
//
//	var cfg Config
//	err := v.Unmarshal(&cfg, viper.DecodeHook(
//	    tlsmode.ViperDecoderHook(),
//	))
//
// See github.com/go-viper/mapstructure/v2 for mapstructure details.
func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = TLSMode(0)
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
					z = Parse(s)
					return nil
				}
			}
		} else if from.Kind() == reflect.Slice {
			if p, k := data.([]byte); k {
				f = func() error {
					z = ParseBytes(p)
					return nil
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
