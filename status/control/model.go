/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package control

import (
	"math"
	"reflect"
	"strings"

	libmap "github.com/go-viper/mapstructure/v2"
)

// String returns the PascalCase string representation of the `Mode`.
// The `Ignore` mode returns an empty string, which is its defined string value.
//
// This method implements the `fmt.Stringer` interface, allowing `Mode` values
// to be printed directly and clearly in logs and other outputs.
//
// Example:
//
//	fmt.Println(control.Must) // Output: Must
func (c Mode) String() string {
	switch c {
	case Should:
		return "Should"
	case Must:
		return "Must"
	case AnyOf:
		return "AnyOf"
	case Quorum:
		return "Quorum"
	}

	return "Ignore"
}

// Code returns the lowercase string representation of the `Mode`.
// This is primarily used for case-insensitive comparisons and for storing the
// mode in configuration files in a consistent format.
//
// Example:
//
//	fmt.Println(control.Must.Code()) // Output: must
func (c Mode) Code() string {
	return strings.ToLower(c.String())
}

// Uint8 returns the `Mode` value as a `uint8`.
// This is the underlying type of `Mode` and allows for easy conversion
// when interfacing with libraries that expect standard integer types.
func (c Mode) Uint8() uint8 {
	return uint8(c)
}

// Uint16 returns the `Mode` value as a `uint16`.
// Useful for compatibility with APIs expecting 16-bit unsigned integers.
func (c Mode) Uint16() uint16 {
	return uint16(c)
}

// Uint32 returns the `Mode` value as a `uint32`.
// Useful for compatibility with APIs expecting 32-bit unsigned integers.
func (c Mode) Uint32() uint32 {
	return uint32(c)
}

// Uint64 returns the `Mode` value as a `uint64`.
// Useful for compatibility with APIs expecting 64-bit unsigned integers.
func (c Mode) Uint64() uint64 {
	return uint64(c)
}

// Int8 returns the `Mode` value as an `int8`.
// Currently, this implementation returns 0 (Ignore) for safety or compatibility reasons.
// Developers should check if this specific behavior meets their needs.
func (c Mode) Int8() int8 {
	if i := c.Uint8(); i > 0 && i < math.MaxInt8 {
		return int8(i)
	}

	return 0
}

// Int16 returns the `Mode` value as an `int16`.
// Currently, this implementation returns 0 (Ignore).
func (c Mode) Int16() int16 {
	return int16(c)
}

// Int32 returns the `Mode` value as an `int32`.
// Currently, this implementation returns 0 (Ignore).
func (c Mode) Int32() int32 {
	return int32(c)
}

// Int64 returns the `Mode` value as an `int64`.
// Unlike smaller signed integer conversions in this package, this method returns
// the actual underlying value of the `Mode` cast to `int64`.
func (c Mode) Int64() int64 {
	return int64(c)
}

// ViperDecoderHook returns a `mapstructure.DecodeHookFunc` that can be used with
// Viper to automatically decode string values from configuration files (e.g., YAML,
// JSON, TOML) into `Mode` types during the unmarshaling process.
//
// This hook is essential for a seamless configuration experience, as it allows
// developers to use human-readable strings (like "must" or "Should") in their
// config files, which are then automatically converted to the correct `Mode` type.
// The decoding is case-insensitive.
//
// Example usage with Viper:
//
//	import (
//		"github.com/spf13/viper"
//		"github.com/mitchellh/mapstructure"
//		"github.com/nabbar/golib/status/control"
//	)
//
//	type MyConfig struct {
//	    ValidationMode control.Mode `mapstructure:"validation_mode"`
//	}
//
//	v := viper.New()
//	// ... load config from file or other source ...
//
//	var cfg MyConfig
//	// The hook is passed to Viper's Unmarshal function.
//	err := v.Unmarshal(&cfg, viper.DecodeHook(control.ViperDecoderHook()))
//
// In a YAML file:
//
//	validation_mode: "Must"
//
// The string "Must" will be correctly decoded into `control.Must`.
func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Mode(0)
			t string
			k bool
		)

		// This hook is only interested in converting strings to the control.Mode type.
		// If the target type is not control.Mode, we pass the data through unchanged.
		if to != reflect.TypeOf(z) {
			return data, nil
		}

		// The source data must be a string to be parsed.
		if from.Kind() != reflect.String {
			return data, nil
		} else if t, k = data.(string); !k {
			return data, nil
		}

		// Parse the string data into a Mode type.
		if e := z.unmarshall([]byte(t)); e != nil {
			return nil, e
		} else {
			return z, nil
		}
	}
}
