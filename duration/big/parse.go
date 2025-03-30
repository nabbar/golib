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
	"errors"
	"fmt"
	"strings"
)

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed
var unitMap = map[string]uint64{
	"s": uint64(Second),
	"m": uint64(Minute),
	"h": uint64(Hour),
	"d": uint64(Day),
}

func parseString(s string) (Duration, error) {
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "'", "", -1)
	s = strings.Replace(s, " ", "", -1)

	// err: 99d55h44m33s123ms
	return parseDuration(s)
}

func (d *Duration) parseString(s string) error {
	if v, e := parseString(s); e != nil {
		return e
	} else {
		*d = v
		return nil
	}
}

func (d *Duration) unmarshall(val []byte) error {
	if tmp, err := ParseByte(val); err != nil {
		return err
	} else {
		*d = tmp
		return nil
	}
}

// parseDuration parses a duration string.
// code from time.ParseDuration
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "s", "m", "h", "d".
func parseDuration(s string) (Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]

		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}

	if s == "" {
		return 0, fmt.Errorf("time: invalid duration '%s'", orig)
	}

	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)

		if err != nil {
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false

		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}

		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		// Consume unit.
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

		if v > 1<<63/unit {
			// overflow
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}

		v *= unit

		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))

			if v > 1<<63 {
				// overflow
				return 0, fmt.Errorf("time: invalid duration '%s'", orig)
			}
		}
		d += v

		if d > 1<<63 {
			return 0, fmt.Errorf("time: invalid duration '%s'", orig)
		}
	}

	if neg {
		return -Duration(d), nil
	}

	if d > 1<<63-1 {
		return 0, fmt.Errorf("time: invalid duration '%s'", orig)
	}
	return Duration(d), nil
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, rem, errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, rem, errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
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
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}
