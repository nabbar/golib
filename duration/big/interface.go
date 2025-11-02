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

// Package big provides duration handling for very large time intervals beyond time.Duration limits.
//
// This package uses int64 seconds instead of int64 nanoseconds, allowing for durations
// up to ~292 billion years (compared to time.Duration's ~290 years limit).
//
// Features:
//   - Support for very large durations (>290 years)
//   - Days notation in parsing and formatting
//   - Multiple encoding support (JSON, YAML, TOML, CBOR, text)
//   - Viper configuration integration
//   - Compatible API with standard duration package
//   - Type conversions to/from time.Duration, int64, uint64, float64
//
// Maximum duration: ~106,751,991,167,300 days (~292 billion years)
//
// Trade-off: Second precision only (no nanosecond precision)
//
// Example usage:
//
//	import durbig "github.com/nabbar/golib/duration/big"
//
//	// Very large duration
//	d := durbig.Days(1000000)  // ~2740 years
//	fmt.Println(d.String())    // Output: 1000000d
//
//	// Parse large duration
//	large, _ := durbig.Parse("365000d")  // ~1000 years
//
//	// Convert to seconds
//	seconds := large.Int64()
//
//	// Use in JSON
//	type Config struct {
//	    MaxAge durbig.Duration `json:"max_age"`
//	}
package big

import (
	"errors"
	"math"
	"time"
)

const (
	Second Duration = 1
	Minute          = 60 * Second
	Hour            = 60 * Minute
	Day             = 24 * Hour
)

var (
	ErrOverFlow = errors.New("value overflow max int64")
)

// Max Value of Big Duration : 106,751,991,167,300 d 15 h 30 m 7 s

type Duration time.Duration

// Parse parses a string representing a duration and returns a Duration
// object. It will return an error if the string is invalid.
//
// The string must be in the format "XhYmZs" where X, Y, and Z are integers
// representing the number of hours, minutes, and seconds respectively.
// The letters "h", "m", and "s" are optional and can be omitted.
//
// Example:
//
// d, err := Parse("1h2m3s")
//
//	if err != nil {
//	    panic(err)
//	}
//
// fmt.Println(d.String()) // Output: 1h2m3s
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
//
// Example:
//
// d := Seconds(5)
// fmt.Println(d.String()) // Output: 5s
func Seconds(i int64) Duration {
	return Duration(i)
}

// Minutes returns a Duration representing i minutes.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
//
// Example:
//
// d := Minutes(5)
// fmt.Println(d.String()) // Output: 5m
func Minutes(i int64) Duration {
	return Duration(i) * Minute
}

// Hours returns a Duration representing i hours.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
//
// Example:
//
// d := Hours(5)
// fmt.Println(d.String()) // Output: 5h
func Hours(i int64) Duration {
	return Duration(i) * Hour
}

// Days returns a Duration representing i days.
//
// The returned Duration is a new Duration and does not modify the
// underlying time.Duration.
//
// The function panics if i is larger than math.MaxInt64 or smaller than -math.MaxInt64.
//
// Example:
//
// d := Days(7)
// fmt.Println(d.String()) // Output: 7d
func Days(i int64) Duration {
	return Duration(i) * Day
}

// ParseDuration returns a Duration representing d seconds.
//
// It does this by converting the time.Duration to a float64 and then passing
// it to ParseFloat64. This means that ParseDuration will round the input
// time.Duration to the nearest integer and then return a Duration representing
// that many seconds.
//
// If the input time.Duration is larger than math.MaxInt64 seconds, ParseDuration
// returns a Duration representing math.MaxInt64 seconds. If the input
// time.Duration is smaller than -math.MaxInt64 seconds, ParseDuration returns
// a Duration representing -math.MaxInt64 seconds.
//
// Example:
//
// d := time.Hour
// pd := libdur.ParseDuration(d)
// fmt.Println(pd) // Output: 1h0m0s
func ParseDuration(d time.Duration) Duration {
	return ParseFloat64(math.Floor(d.Seconds()))
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
