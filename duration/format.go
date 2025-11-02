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
	"fmt"
	"math"
	"time"
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
		s = fmt.Sprintf("%dd", n)
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

	if t > math.MaxInt64 {
		return math.MaxInt64
	}

	return int64(t)
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
