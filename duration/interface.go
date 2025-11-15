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

// Package duration provides an extended duration type with days support and multiple encoding formats.
//
// This package wraps time.Duration and extends it with:
//   - Days notation in parsing and formatting (e.g., "5d23h15m13s")
//   - Multiple encoding support (JSON, YAML, TOML, CBOR, text)
//   - Viper configuration integration
//   - Arithmetic operations and helper functions
//   - Truncation and rounding to various time units
//   - PID controller-based range generation
//
// The package is limited to time.Duration's range (±290 years).
// For larger durations, use the big sub-package.
//
// Example usage:
//
//	import "github.com/nabbar/golib/duration"
//
//	// Parse duration with days
//	d, _ := duration.Parse("5d23h15m13s")
//	fmt.Println(d.String())  // Output: 5d23h15m13s
//
//	// Create durations
//	timeout := duration.Days(2) + duration.Hours(3)
//
//	// Convert to time.Duration
//	std := timeout.Time()
//
//	// Use in JSON
//	type Config struct {
//	    Timeout duration.Duration `json:"timeout"`
//	}
package duration

import (
	"math"
	"time"
)

type Duration time.Duration

// Parse parses a string representing a duration and returns a Duration
// object. It will return an error if the string is invalid.
//
// The string must be in the format "XhYmZs" where X, Y, and Z are integers
// representing the number of hours, minutes, and seconds respectively.
// The letters "h", "m", and "s" are optional and can be omitted.
//
// For example, "2h" represents 2 hours, "3m" represents 3 minutes,
// and "4s" represents 4 seconds.
//
// The function is case insensitive.
func Parse(s string) (Duration, error) {
	return parseString(s)
}

// ParseByte parses a byte array representing a duration and returns a Duration
// object. It will return an error if the byte array is invalid.
//
// The byte array must be in the format "XhYmZs" where X, Y, and Z are integers
// representing the number of hours, minutes, and seconds respectively.
// The letters "h", "m", and "s" are optional and can be omitted.
//
// For example, "2h" represents 2 hours, "3m" represents 3 minutes,
// and "4s" represents 4 seconds.
//
// The function is case insensitive.
func ParseByte(p []byte) (Duration, error) {
	return parseString(string(p))
}

// Seconds returns a Duration representing i seconds.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
func Seconds(i int64) Duration {
	return Duration(time.Duration(i) * time.Second)
}

// Minutes returns a Duration representing i minutes.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
func Minutes(i int64) Duration {
	return Duration(time.Duration(i) * time.Minute)
}

// Hours returns a Duration representing i hours.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
func Hours(i int64) Duration {
	return Duration(time.Duration(i) * time.Hour)
}

// Days returns a Duration representing i days.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
//
// The duration is calculated by multiplying i by 24 hours (1 day).
func Days(i int64) Duration {
	return Duration(time.Duration(i) * time.Hour * 24)
}

// ParseDuration returns a Duration representing d time.Duration.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function is a no-op and simply returns the input time.Duration as a
// Duration. It can be used to convert a time.Duration to a Duration
// without modifying the underlying time.Duration.
//
// Example:
//
//	d := 5*time.Hour
//	dd := ParseDuration(d)
//	fmt.Println(dd) // 5h0m0s
func ParseDuration(d time.Duration) Duration {
	return Duration(d)
}

// ParseFloat64 returns a Duration representing f seconds.
//
// If f is larger than math.MaxInt64, ParseFloat64 returns a Duration
// representing math.MaxInt64 seconds. If f is smaller than -math.MaxInt64,
// ParseFloat64 returns a Duration representing -math.MaxInt64 seconds.
//
// Otherwise, ParseFloat64 returns a Duration representing the closest
// integer to f seconds. The returned Duration is a new Duration and
// does not modify the underlying float64.
func ParseFloat64(f float64) Duration {
	const (
		mx float64 = math.MaxInt64
		mi         = -mx
	)

	if f > mx {
		return Duration(math.MaxInt64)
	} else if f < mi {
		return Duration(-math.MaxInt64)
	} else {
		return Duration(math.Round(f))
	}
}

func ParseUint32(i uint32) Duration {
	if uint64(i) > uint64(math.MaxInt64) {
		return Duration(math.MaxInt64)
	} else {
		return Duration(i)
	}
}
