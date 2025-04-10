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

package big

import (
	"errors"
	"math"
	"time"
)

const mx float64 = math.MaxInt64

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

func Parse(s string) (Duration, error) {
	return parseString(s)
}

func ParseByte(p []byte) (Duration, error) {
	return parseString(string(p))
}

func Seconds(i int64) Duration {
	return Duration(i)
}

func Minutes(i int64) Duration {
	return Duration(i) * Minute
}

func Hours(i int64) Duration {
	return Duration(i) * Hour
}

func Days(i int64) Duration {
	return Duration(i) * Day
}

func ParseDuration(d time.Duration) Duration {
	return ParseFloat64(math.Floor(d.Seconds()))
}

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
