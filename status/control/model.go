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
	"reflect"
	"strings"

	libmap "github.com/go-viper/mapstructure/v2"
)

// String returns the string representation of the Mode.
//
// The string representation uses PascalCase for readability:
//   - Should -> "Should"
//   - Must -> "Must"
//   - AnyOf -> "AnyOf"
//   - Quorum -> "Quorum"
//   - Ignore -> "" (empty string)
//
// This method is used by the fmt package when printing Mode values.
//
// Example:
//
//	mode := control.Must
//	fmt.Println(mode.String()) // Output: Must
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

	return ""
}

// Code returns the lowercase code representation of the Mode.
//
// This is a convenience method that returns the lowercase version of String().
// It's useful for case-insensitive comparisons and configuration files.
//
// Returns:
//   - Should -> "should"
//   - Must -> "must"
//   - AnyOf -> "anyof"
//   - Quorum -> "quorum"
//   - Ignore -> "" (empty string)
//
// Example:
//
//	mode := control.Must
//	fmt.Println(mode.Code()) // Output: must
func (c Mode) Code() string {
	return strings.ToLower(c.String())
}

// ViperDecoderHook returns a mapstructure decode hook for Viper configuration.
//
// This hook allows Mode values to be automatically decoded from configuration
// files (YAML, JSON, TOML, etc.) when using Viper. The hook converts string
// values to Mode types during configuration unmarshaling.
//
// The decoding is case-insensitive and supports all Mode string representations.
//
// Usage with Viper:
//
//	import (
//	    "github.com/spf13/viper"
//	    "github.com/nabbar/golib/status/control"
//	)
//
//	v := viper.New()
//	v.SetConfigType("yaml")
//
//	// Register the decode hook
//	decoderConfig := &mapstructure.DecoderConfig{
//	    DecodeHook: control.ViperDecoderHook(),
//	    Result:     &config,
//	}
//
// Configuration file example (YAML):
//
//	mandatory:
//	  mode: must
//	  keys:
//	    - database
//	    - cache
//
// The "must" string will be automatically decoded to control.Must.
//
// See also: github.com/spf13/viper for Viper configuration management.
func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Mode(0)
			t string
			k bool
		)

		// Check if the target type matches the expected one
		if to != reflect.TypeOf(z) {
			return data, nil
		}

		// Check if the data type matches the expected one
		if from.Kind() != reflect.String {
			return data, nil
		} else if t, k = data.(string); !k {
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
