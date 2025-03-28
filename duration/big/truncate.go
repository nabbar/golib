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

// Round returns the result of rounding d to the nearest multiple of unit.
// The rounding behavior for halfway values is to round away from zero.
// If the result exceeds the maximum (or minimum)
// value that can be stored in a [Duration],
// Round returns the maximum (or minimum) duration.
// If unit <= 0, Round returns d unchanged.
// code from time.Duration
func (d Duration) Round(unit Duration) Duration {
	if unit <= 0 {
		return d
	}

	r := d % unit

	if d < 0 {
		r = -r

		if lessThanHalf(r, unit) {
			return d + r
		}

		if d1 := d - unit + r; d1 < d {
			return d1
		}

		return minDuration // overflow
	}

	if lessThanHalf(r, unit) {
		return d - r
	}

	if d1 := d + unit - r; d1 > d {
		return d1
	}

	return maxDuration // overflow
}

// Truncate returns the result of rounding d toward zero to a multiple of m.
// If unit <= 0, Truncate returns d unchanged.
// code from time.Duration
func (d Duration) Truncate(unit Duration) Duration {
	if unit <= 0 {
		return d
	}
	return d - d%unit
}

// TruncateMinutes returns the result of rounding d toward zero to a multiple of Minute.
// If unit <= 0, TruncateMinutes returns d unchanged.
func (d Duration) TruncateMinutes() Duration {
	return d.Truncate(Minute)
}

// TruncateHours returns the result of rounding d toward zero to a multiple of Hour.
// If unit <= 0, TruncateHours returns d unchanged.
func (d Duration) TruncateHours() Duration {
	return d.Truncate(Hour)
}

// TruncateDays returns the result of rounding d toward zero to a multiple of Day.
// If unit <= 0, TruncateDays returns d unchanged.
func (d Duration) TruncateDays() Duration {
	return d.Truncate(Day)
}

// lessThanHalf reports whether x+x < y but avoids overflow,
// assuming x and y are both positive (Duration is signed).
// code from time.Duration
func lessThanHalf(x, y Duration) bool {
	return uint64(x)+uint64(x) < uint64(y)
}
