/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package perm

import (
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

// ViperDecoderHook returns a DecodeHookFuncType function for Viper configuration decoding.
//
// This hook enables automatic conversion of string values to Perm types when
// unmarshaling configuration files with Viper. It checks if the source data is
// a string and the target type is Perm, then parses the string into a Perm value.
//
// Usage with Viper:
//
//	import (
//	    "github.com/spf13/viper"
//	    "github.com/nabbar/golib/file/perm"
//	)
//
//	type Config struct {
//	    FilePermission perm.Perm `mapstructure:"file_perm"`
//	}
//
//	v := viper.New()
//	v.SetConfigFile("config.yaml")
//
//	cfg := Config{}
//	opts := viper.DecoderConfigOption(func(c *mapstructure.DecoderConfig) {
//	    c.DecodeHook = perm.ViperDecoderHook()
//	})
//	v.Unmarshal(&cfg, opts)
//
// The hook supports all permission formats: octal strings ("0644"),
// symbolic notation ("rwxr-xr-x"), and quoted strings.
func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Perm(0)
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
		return parseString(t)
	}
}
