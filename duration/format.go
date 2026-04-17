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
	"math"
	"strconv"
	"time"

	libpid "github.com/nabbar/golib/pidcontroller"
)

// Time returns a time.Duration representation of the duration.
// It is a simple wrapper around the conversion of the underlying int64
// value to a time.Duration.
//
// Time is useful when working with the time package, as it allows for
// easy conversion between the duration package and the time package.
//
// Example:
//
// d := libdur.ParseDuration("1h30m")
// td := d.Time()
// fmt.Println(td) // Output: 1h30m0s
func (d Duration) Time() time.Duration {
	return time.Duration(d)
}

// String returns a string representation of the duration.
// The string is in the format "NdNhNmNs" where N is a number.
// The days are omitted if n is 0 or negative. The hours, minutes, and seconds
// are omitted if they are 0.
//
// Example:
//
// d := libdur.ParseDuration("1d2h3m4s")
// fmt.Println(d.String()) // Output: 1d2h3m4s
func (d Duration) String() string {
	var (
		s string
		n = d.Days()
		i = d.Time()
	)

	if n > 0 {
		i = i - (time.Duration(n) * 24 * time.Hour)
		s = strconv.FormatInt(n, 10) + "d"
	}

	if n < 1 || i > 0 {
		s = s + i.String()
	}

	return s
}

// Days returns the number of days in the duration.
// The number of days is calculated by dividing the total number of hours
// by 24 and rounding down to the nearest integer.
// If the total number of hours is greater than the maximum value of int64,
// the maximum value of int64 is returned.
func (d Duration) Days() int64 {
	t := math.Floor(d.Time().Hours() / 24)
	return libpid.Float64ToInt64(t)
}

// Hours returns the number of hours in the duration.
// It calculates the total hours from the underlying time.Duration,
// rounds down to the nearest integer, and converts it to int64.
// This provides the total duration expressed in full hours.
func (d Duration) Hours() int64 {
	t := math.Floor(d.Time().Hours())
	return libpid.Float64ToInt64(t)
}

// Minutes returns the number of minutes in the duration.
// It calculates the total minutes from the underlying time.Duration,
// rounds down to the nearest integer, and converts it to int64.
// This provides the total duration expressed in full minutes.
func (d Duration) Minutes() int64 {
	t := math.Floor(d.Time().Minutes())
	return libpid.Float64ToInt64(t)
}

// Seconds returns the number of seconds in the duration.
// It calculates the total seconds from the underlying time.Duration,
// rounds down to the nearest integer, and converts it to int64.
// This provides the total duration expressed in full seconds.
func (d Duration) Seconds() int64 {
	t := math.Floor(d.Time().Seconds())
	return libpid.Float64ToInt64(t)
}

// Milliseconds returns the duration as an integer millisecond count.
// It delegates to the underlying time.Duration.Milliseconds() method.
func (d Duration) Milliseconds() int64 {
	return d.Time().Milliseconds()
}

// Microseconds returns the duration as an integer microsecond count.
// It delegates to the underlying time.Duration.Microseconds() method.
func (d Duration) Microseconds() int64 {
	return d.Time().Microseconds()
}

// Nanoseconds returns the duration as an integer nanosecond count.
// It delegates to the underlying time.Duration.Nanoseconds() method.
func (d Duration) Nanoseconds() int64 {
	return d.Time().Nanoseconds()
}

// Float64 returns the underlying int64 value of the duration as a float64.
//
// This can be useful when working with libraries or functions that expect
// a float64 value, as it allows for easy conversion between the duration
// package and the required type.
//
// Example:
//
// d := libdur.ParseDuration("1h30m")
// f := d.Float64()
// fmt.Println(f) // Output: 5400.0
func (d Duration) Float64() float64 {
	return float64(d)
}

// Uint64 returns the duration as an unsigned 64-bit integer.
// If the duration is negative, it returns the absolute value cast to uint64.
// Otherwise, it returns the duration cast to uint64.
func (d Duration) Uint64() uint64 {
	if t := d.Time(); t < 0 {
		return uint64(-t)
	} else {
		return uint64(t)
	}
}

// Uint32 returns the duration as an unsigned 32-bit integer.
// If the duration is negative, it returns the absolute value cast to uint32.
// Otherwise, it returns the duration cast to uint32.
func (d Duration) Uint32() uint32 {
	if t := d.Time(); t > math.MaxUint32 || t < -math.MaxUint32 {
		return math.MaxUint32
	} else if t > 0 {
		return uint32(t)
	} else if t < 0 {
		return uint32(-t)
	}
	return 0
}

// Int64 returns the duration as a signed 64-bit integer.
// It simply casts the underlying Duration (which is int64) to int64.
func (d Duration) Int64() int64 {
	return int64(d)
}

// Duration returns the time.Duration value.
// It is equivalent to d.Time().
func (d Duration) Duration() time.Duration {
	return d.Time()
}
