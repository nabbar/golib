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

package protocol

import (
	"fmt"
	"math"
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

func (n NetworkProtocol) IsTCP() bool {
	switch n {
	case NetworkTCP, NetworkTCP4, NetworkTCP6:
		return true
	default:
		return false
	}
}

func (n NetworkProtocol) IsUDP() bool {
	switch n {
	case NetworkUDP, NetworkUDP4, NetworkUDP6:
		return true
	default:
		return false
	}
}

func (n NetworkProtocol) IsUnixLike() bool {
	switch n {
	case NetworkUnix, NetworkUnixGram:
		return true
	default:
		return false
	}
}

func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = NetworkProtocol(0)
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
					if c := ParseInt64(int64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Int8 {
			if i, k := data.(int8); k {
				f = func() error {
					if c := ParseInt64(int64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Int16 {
			if i, k := data.(int16); k {
				f = func() error {
					if c := ParseInt64(int64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Int32 {
			if i, k := data.(int32); k {
				f = func() error {
					if c := ParseInt64(int64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Int64 {
			if i, k := data.(int64); k {
				f = func() error {
					if c := ParseInt64(i); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint {
			if i, k := data.(uint); k {
				f = func() error {
					if uint64(i) > uint64(math.MaxUint16) {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else if c := ParseUint64(uint64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint8 {
			if i, k := data.(uint8); k {
				f = func() error {
					if c := ParseUint64(uint64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint16 {
			if i, k := data.(uint16); k {
				f = func() error {
					if c := ParseUint64(uint64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint32 {
			if i, k := data.(uint32); k {
				f = func() error {
					if c := ParseUint64(uint64(i)); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
					return nil
				}
			}
		} else if from.Kind() == reflect.Uint64 {
			if i, k := data.(uint64); k {
				f = func() error {
					if i > uint64(math.MaxUint16) {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else if c := ParseUint64(i); c == NetworkEmpty {
						return fmt.Errorf("network protocol: invalid value '%d'", i)
					} else {
						z = c
					}
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
