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

// Round returns the result of rounding d toward zero to a multiple of unit.
// If unit <= 0, Round returns d unchanged.
//
// If the remainder of d divided by unit is less than half of unit, Round returns d.
// Otherwise, if d is negative, Round returns d - (d % unit) + unit.
// Otherwise, Round returns d + (unit - d % unit).
//
// If the result of the rounding overflows, Round returns either minDuration or maxDuration.
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
// Otherwise, it returns the duration d rounded to the nearest multiple of unit.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of unit,
// it rounds up to the next multiple of unit).
// For example, TruncateMinutes(ParseDuration("1.5m")) returns ParseDuration("2m").
func (d Duration) Truncate(unit Duration) Duration {
	if unit <= 0 {
		return d
	}
	return d - d%unit
}

// TruncateMinutes returns the result of rounding d toward zero to a multiple of a minute.
// If d is an exact multiple of a minute, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a minute.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a minute,
// it rounds up to the next multiple of a minute).
// For example, TruncateMinutes(ParseDuration("1.5m")) returns ParseDuration("2m").
func (d Duration) TruncateMinutes() Duration {
	return d.Truncate(Minute)
}

// TruncateHours returns the result of rounding d toward zero to a multiple of an hour.
// If d is an exact multiple of an hour, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of an hour.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of an hour,
// it rounds up to the next multiple of an hour).
// For example, TruncateHours(ParseDuration("1.5h")) returns ParseDuration("2h").
func (d Duration) TruncateHours() Duration {
	return d.Truncate(Hour)
}

// TruncateDays returns the result of rounding d toward zero to a multiple of a day.
// If d is an exact multiple of a day, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a day.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a day,
// it rounds up to the next multiple of a day).
// For example, TruncateDays(ParseDuration("1.5d")) returns ParseDuration("2d").
func (d Duration) TruncateDays() Duration {
	return d.Truncate(Day)
}

func lessThanHalf(x, y Duration) bool {
	return x.Uint64() < (y.Uint64() / 2) // same as x+x < y
}
