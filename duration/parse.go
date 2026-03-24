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
	"errors"
	"fmt"
	"strings"
	"time"
)

// errLeadingInt is a private error used to signal an issue with parsing a leading integer from a string.
// It is not exposed to the user but helps in internal error handling.
var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// unitMap maps the string representation of time units to their corresponding duration in nanoseconds (as uint64).
// It supports standard units like "ns", "us", "ms", "s", "m", "h", and the custom "d" for days.
// It also includes Unicode symbols for microseconds.
var unitMap = map[string]uint64{
	"ns": uint64(time.Nanosecond),
	"us": uint64(time.Microsecond),
	"µs": uint64(time.Microsecond), // U+00B5 = micro symbol
	"μs": uint64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": uint64(time.Millisecond),
	"s":  uint64(time.Second),
	"m":  uint64(time.Minute),
	"h":  uint64(time.Hour),
	"d":  uint64(24 * time.Hour),
}

// parseString is an internal helper that cleans and parses a duration string.
// It removes quotes and spaces before passing the string to the main parsing logic.
func parseString(s string) (Duration, error) {
	s = strings.Replace(s, "\"", "", -1) // nolint
	s = strings.Replace(s, "'", "", -1)  // nolint
	s = strings.Replace(s, " ", "", -1)  // nolint

	// err: 99d55h44m33s123ms
	return parseDuration(s)
}

// parseString is a method on the Duration pointer that allows a Duration object to be updated by parsing a string.
// It's primarily used for unmarshalling tasks where the duration object already exists.
func (d *Duration) parseString(s string) error {
	if v, e := parseString(s); e != nil {
		return e
	} else {
		*d = v
		return nil
	}
}

// unmarshall is a helper method for unmarshalling a duration from a byte slice.
// It wraps the ParseByte function to update the value of the Duration pointer.
func (d *Duration) unmarshall(val []byte) error {
	if tmp, err := ParseByte(val); err != nil {
		return err
	} else {
		*d = tmp
		return nil
	}
}

// parseDuration is the core parsing logic, adapted from the standard library's time.ParseDuration.
// It parses a duration string, which is a sequence of decimal numbers with optional fractions and unit suffixes.
// It supports "d" for days in addition to standard units ("ns", "us", "ms", "s", "m", "h").
// The string can be signed (e.g., "-1.5h").
func parseDuration(s string) (Duration, error) {
	// The overall structure is a signed sequence of segments, like:
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume optional sign prefix.
	if s != "" {
		c := s[0]

		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: "0" is a zero duration.
	if s == "0" {
		return 0, nil
	}

	if s == "" {
		return 0, fmt.Errorf("time: invalid duration '%s'", orig)
	}

	for s != "" {
		var (
			v, f  uint64      // integers before and after the decimal point
			scale float64 = 1 // scale factor for the fractional part
		)

		var err error

		// The next character must be a digit or a period.
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') { // nolint
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		// Consume the integer part of the number.
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		pre := pl != len(s) // check if we consumed any digits

		// Consume the fractional part of the number.
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}

		if !pre && !post {
			// No digits were found (e.g., ".s" or "-.s").
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		// Consume the unit suffix.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}

		if i == 0 {
			return 0, fmt.Errorf("time: missing unit in duration '%s'", orig)
		}

		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]

		if !ok {
			return 0, fmt.Errorf("time: unknown unit '%s' in duration '%s'", u, orig)
		}

		// Check for overflow when scaling the integer part by the unit.
		if v > 1<<63/unit {
			return 0, fmt.Errorf("time: invalid duration '%s' (overflow)", orig)
		}

		v *= unit

		// Add the fractional part, scaled appropriately.
		if f > 0 {
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				return 0, fmt.Errorf("time: invalid duration '%s' (overflow)", orig)
			}
		}

		d += v

		if d > 1<<63 {
			return 0, fmt.Errorf("time: invalid duration '%s' (overflow)", orig)
		}
	}

	if neg {
		return -Duration(d), nil
	}

	if d > 1<<63-1 {
		return 0, fmt.Errorf("time: invalid duration '%s' (overflow)", orig)
	}

	return Duration(d), nil
}

// leadingInt consumes a leading integer from a byte slice or string.
// It returns the parsed integer, the remaining part of the string, and an error on overflow.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]

		if c < '0' || c > '9' {
			break
		}

		// Check for overflow before multiplication.
		if x > 1<<63/10 {
			return 0, rem, errLeadingInt
		}

		x = x*10 + uint64(c) - '0'

		// Check for overflow after addition.
		if x > 1<<63 {
			return 0, rem, errLeadingInt
		}
	}

	return x, s[i:], nil
}

// leadingFraction consumes a leading fractional part of a number from a string.
// It returns the parsed fraction as an integer, its scale, and the remainder of the string.
// It doesn't return an error on overflow but simply stops accumulating precision to avoid panics.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false

	for ; i < len(s); i++ {
		c := s[i]

		if c < '0' || c > '9' {
			break
		}

		if overflow {
			continue
		}

		// Check for potential overflow before multiplication.
		if x > (1<<63-1)/10 {
			overflow = true
			continue
		}

		y := x*10 + uint64(c) - '0'

		// Check for overflow after addition.
		if y > 1<<63 {
			overflow = true
			continue
		}

		x = y
		scale *= 10
	}

	return x, scale, s[i:]
}
