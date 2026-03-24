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

// Round returns the result of rounding d to the nearest multiple of m.
// The rounding behavior for halfway values is to round away from zero.
// If the result exceeds the maximum (or minimum)
// value that can be stored in a Duration, Round returns the maximum (or minimum) duration.
// If m <= 0, Round returns d unchanged.
//
// Example:
//
//	d := Seconds(10)
//	r := d.Round(Seconds(3))
//	fmt.Println(r) // Output: 9s
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

// Truncate returns the result of rounding d toward zero to a multiple of unit.
// If unit <= 0, Truncate returns d unchanged.
//
// Example:
//
//	d := Seconds(10)
//	t := d.Truncate(Seconds(3))
//	fmt.Println(t) // Output: 9s
func (d Duration) Truncate(unit Duration) Duration {
	if unit <= 0 {
		return d
	}
	return d - d%unit
}

// TruncateMinutes returns the result of rounding d toward zero to a multiple of a minute.
// It is a shorthand for d.Truncate(Minute).
//
// Example:
//
//	d := Seconds(90)
//	t := d.TruncateMinutes()
//	fmt.Println(t) // Output: 1m
func (d Duration) TruncateMinutes() Duration {
	return d.Truncate(Minute)
}

// TruncateHours returns the result of rounding d toward zero to a multiple of an hour.
// It is a shorthand for d.Truncate(Hour).
//
// Example:
//
//	d := Minutes(90)
//	t := d.TruncateHours()
//	fmt.Println(t) // Output: 1h
func (d Duration) TruncateHours() Duration {
	return d.Truncate(Hour)
}

// TruncateDays returns the result of rounding d toward zero to a multiple of a day.
// It is a shorthand for d.Truncate(Day).
//
// Example:
//
//	d := Hours(30)
//	t := d.TruncateDays()
//	fmt.Println(t) // Output: 1d
func (d Duration) TruncateDays() Duration {
	return d.Truncate(Day)
}

func lessThanHalf(x, y Duration) bool {
	return x.Uint64() < (y.Uint64() / 2) // same as x+x < y, but prevent potential overflow
}
