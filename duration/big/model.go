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
 */

package big

import (
	"reflect"

	libmap "github.com/mitchellh/mapstructure"
)

const (
	minDuration Duration = -1 << 63
	maxDuration Duration = 1<<63 - 1
)

// IsDays checks if the duration is at least one day.
//
// It returns true if the duration represents 24 hours or more,
// false otherwise.
//
// Example:
//
//	d := Days(1)
//	if d.IsDays() {
//	    fmt.Println("At least one day")
//	}
func (d Duration) IsDays() bool {
	return d >= Days(1)
}

// IsHours checks if the duration is at least one hour.
//
// It returns true if the duration represents 60 minutes or more,
// false otherwise.
//
// Example:
//
//	d := Hours(1)
//	if d.IsHours() {
//	    fmt.Println("At least one hour")
//	}
func (d Duration) IsHours() bool {
	return d >= Hours(1)
}

// IsMinutes checks if the duration is at least one minute.
//
// It returns true if the duration represents 60 seconds or more,
// false otherwise.
//
// Example:
//
//	d := Minutes(1)
//	if d.IsMinutes() {
//	    fmt.Println("At least one minute")
//	}
func (d Duration) IsMinutes() bool {
	return d >= Minutes(1)
}

// IsSeconds checks if the duration is at least one second.
//
// It returns true if the duration represents 1 second or more,
// false otherwise.
//
// Example:
//
//	d := Seconds(1)
//	if d.IsSeconds() {
//	    fmt.Println("At least one second")
//	}
func (d Duration) IsSeconds() bool {
	return d >= Seconds(1)
}

// ViperDecoderHook is a libmap.DecodeHookFuncType that is used to decode strings into Duration values.
// It takes a reflect.Type, a reflect.Type, and an interface{} as parameters, and returns an interface{} and an error.
// If the data type is not a string, it returns the data as is and a nil error.
// If the target type is not a Duration, it returns the data as is and a nil error.
// Otherwise, it formats/decodes/parses the data and returns the new value.
func ViperDecoderHook() libmap.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		var (
			z = Duration(0)
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
