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

package duration

import (
	"reflect"

	libmap "github.com/go-viper/mapstructure/v2"
)

// IsDays checks if the duration is greater than or equal to one day.
// It returns true if the duration is at least 24 hours.
func (d Duration) IsDays() bool {
	return d >= Days(1)
}

// IsHours checks if the duration is greater than or equal to one hour.
// It returns true if the duration is at least 60 minutes.
func (d Duration) IsHours() bool {
	return d >= Hours(1)
}

// IsMinutes checks if the duration is greater than or equal to one minute.
// It returns true if the duration is at least 60 seconds.
func (d Duration) IsMinutes() bool {
	return d >= Minutes(1)
}

// IsSeconds checks if the duration is greater than or equal to one second.
// It returns true if the duration is at least 1000 milliseconds.
func (d Duration) IsSeconds() bool {
	return d >= Seconds(1)
}

// IsMilliseconds checks if the duration is greater than or equal to one millisecond.
// It returns true if the duration is at least 1000 microseconds.
func (d Duration) IsMilliseconds() bool {
	return d >= Milliseconds(1)
}

// IsMicroseconds checks if the duration is greater than or equal to one microsecond.
// It returns true if the duration is at least 1000 nanoseconds.
func (d Duration) IsMicroseconds() bool {
	return d >= Microseconds(1)
}

// IsNanoseconds checks if the duration is greater than or equal to one nanosecond.
// This is effectively checking if the duration is non-zero (assuming positive duration),
// as 1ns is the smallest unit.
func (d Duration) IsNanoseconds() bool {
	return d >= Nanoseconds(1)
}

// ViperDecoderHook is a libmap.DecodeHookFuncType that is used to decode strings into libdur.Duration values.
// It takes a reflect.Type, a reflect.Type, and an interface{} as parameters, and returns an interface{} and an error.
// If the data type is not a string, it returns the data as is and a nil error.
// If the target type is not a libdur.Duration, it returns the data as is and a nil error.
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
