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

package bytes

import (
	"errors"
	"fmt"
	"strings"
)

// Based on ParseDuration from time package

var unitMap = map[string]uint64{
	"B":  uint64(SizeUnit),
	"b":  uint64(SizeUnit),
	"k":  uint64(SizeKilo),
	"K":  uint64(SizeKilo),
	"kb": uint64(SizeKilo),
	"Kb": uint64(SizeKilo),
	"kB": uint64(SizeKilo),
	"KB": uint64(SizeKilo),
	"m":  uint64(SizeMega),
	"M":  uint64(SizeMega),
	"mb": uint64(SizeMega),
	"Mb": uint64(SizeMega),
	"mB": uint64(SizeMega),
	"MB": uint64(SizeMega),
	"g":  uint64(SizeGiga),
	"G":  uint64(SizeGiga),
	"gb": uint64(SizeGiga),
	"Gb": uint64(SizeGiga),
	"gB": uint64(SizeGiga),
	"GB": uint64(SizeGiga),
	"t":  uint64(SizeTera),
	"T":  uint64(SizeTera),
	"tb": uint64(SizeTera),
	"Tb": uint64(SizeTera),
	"tB": uint64(SizeTera),
	"TB": uint64(SizeTera),
	"p":  uint64(SizePeta),
	"P":  uint64(SizePeta),
	"pb": uint64(SizePeta),
	"Pb": uint64(SizePeta),
	"pB": uint64(SizePeta),
	"PB": uint64(SizePeta),
	// no e/E to prevent mismatching with notation 1.23E45/1.23e+45/1.23E-45
	"eb": uint64(SizeExa),
	"Eb": uint64(SizeExa),
	"eB": uint64(SizeExa),
	"EB": uint64(SizeExa),
}

func parseBytes(p []byte) (Size, error) {
	return parseString(string(p))
}

func parseString(s string) (Size, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+

	var (
		orig = s
		neg  bool

		d          uint64
		errInvalid = fmt.Errorf("size: invalid size '%s'", orig)
		errUnit    = fmt.Errorf("size: missing unit '%s'", orig)
		errUnkUnit = fmt.Errorf("size: unknown unit '%s'", orig)
	)

	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	s = strings.TrimSpace(s)

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "" {
		return 0, errInvalid
	}

	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errInvalid
		}

		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errInvalid
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
			return 0, errInvalid
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
			return 0, errUnit
		}

		u := strings.TrimSpace(s[:i])
		s = s[i:]
		unit, ok := unitMap[u]

		if !ok {
			return 0, errUnkUnit
		}

		if v > 1<<63/unit {
			// overflow
			return 0, errInvalid
		}

		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))

			if v > 1<<63 {
				// overflow
				return 0, errInvalid
			}
		}
		d += v

		if d > 1<<63 {
			return 0, errInvalid
		}
	}

	if neg {
		return -Size(d), nil
	}

	if d > 1<<63-1 {
		return 0, errInvalid
	}

	return Size(d), nil

}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x uint64, rem string, err error) {
	var errLeadingInt = errors.New("size: bad [0-9]*") // never printed

	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, "", errLeadingInt
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
