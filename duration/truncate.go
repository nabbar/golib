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
	"time"
)

// TruncateMicroseconds returns the result of rounding d toward zero to a multiple of a microsecond.
// If d is an exact multiple of a microsecond, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a microsecond.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a microsecond,
// it rounds up to the next multiple of a microsecond).
// For example, TruncateMicroseconds(ParseDuration("1.5µs")) returns ParseDuration("2µs").
func (d Duration) TruncateMicroseconds() Duration {
	return Duration(time.Duration(d.Time().Microseconds()) * time.Microsecond)
}

// TruncateMilliseconds returns the result of rounding d toward zero to a multiple of a millisecond.
// If d is an exact multiple of a millisecond, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a millisecond.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a millisecond,
// it rounds up to the next multiple of a millisecond.
// For example, TruncateMilliseconds(ParseDuration("1.5ms")) returns ParseDuration("2ms").
func (d Duration) TruncateMilliseconds() Duration {
	return Duration(time.Duration(d.Time().Milliseconds()) * time.Millisecond)
}

// TruncateSeconds returns the result of rounding d toward zero to a multiple of a second.
// If d is an exact multiple of a second, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a second.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a second,
// it rounds up to the next multiple of a second.
// For example, TruncateSeconds(ParseDuration("1.5s")) returns ParseDuration("2s").
func (d Duration) TruncateSeconds() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Seconds())) * time.Second)
}

// TruncateMinutes returns the result of rounding d toward zero to a multiple of a minute.
// If d is an exact multiple of a minute, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a minute.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a minute,
// it rounds up to the next multiple of a minute.
// For example, TruncateMinutes(ParseDuration("1.5m")) returns ParseDuration("2m").
func (d Duration) TruncateMinutes() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Minutes())) * time.Minute)
}

// TruncateHours returns the result of rounding d toward zero to a multiple of an hour.
// If d is an exact multiple of an hour, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of an hour.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of an hour,
// it rounds up to the next multiple of an hour.
// For example, TruncateHours(ParseDuration("1.5h")) returns ParseDuration("2h").
func (d Duration) TruncateHours() Duration {
	return Duration(time.Duration(math.Floor(d.Time().Hours())) * time.Hour)
}

// TruncateDays returns the result of rounding d toward zero to a multiple of a day.
// If d is an exact multiple of a day, it returns d unchanged.
// Otherwise, it returns the duration d rounded to the nearest multiple of a day.
// The rounding mode is to round half even up (i.e. if d is halfway between two multiples of a day,
// it rounds up to the next multiple of a day.
// For example, TruncateDays(ParseDuration("1.5d")) returns ParseDuration("2d").
func (d Duration) TruncateDays() Duration {
	return Duration(time.Duration(d.Days()) * 24 * time.Hour)
}
