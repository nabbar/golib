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
	"fmt"
	"math"
	"time"
)

func (d Duration) Time() (time.Duration, error) {
	mxt := float64(math.MaxInt64) / float64(time.Second)
	if d.Float64() > mxt {
		return 0, fmt.Errorf("overflow max time.Duration value")
	}

	return time.Duration(d) * time.Second, nil
}

// String returns a string representing the duration in the form "135d72h3m5s".
// Leading zero units are omitted. The zero duration formats as 0s.
func (d Duration) String() string {
	var s string

	if d < 0 {
		s = "-"
	} else if d == 0 {
		return "0s"
	}

	// Days
	r, p := stringUnit(int64(d.Abs()), Day.Int64(), "d")
	s += p

	// Hours
	r, p = stringUnit(r, Hour.Int64(), "h")
	s += p

	// Minutes
	r, p = stringUnit(r, Minute.Int64(), "m")
	s += p

	// Seconds
	if r > 0 {
		s += fmt.Sprintf("%ds", r)
	}

	return s
}

func stringUnit(val, div int64, unit string) (rest int64, str string) {
	if val == val%div {
		// same value so no unit in value, so skip
		return val, ""
	}

	n := val % div
	v := (val - n) / div

	if v > 0 {
		return n, fmt.Sprintf("%d%s", v, unit)
	} else {
		return val, ""
	}
}

func (d Duration) Int64() int64 {
	return int64(d)
}

func (d Duration) Uint64() uint64 {
	if d < 0 {
		return uint64(0)
	}

	return uint64(d)
}

func (d Duration) Float64() float64 {
	return float64(d)
}
