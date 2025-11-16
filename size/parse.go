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

package size

import (
	"errors"
	"fmt"
	"strings"
)

// The parsing logic is based on ParseDuration from the standard library's time package.
// It supports flexible parsing of size strings with various unit representations.

// unitMap defines the mapping of unit strings to their byte multipliers.
//
// The map supports multiple variations of each unit to provide flexibility:
//   - Single letter units (case-insensitive): b/B, k/K, m/M, g/G, t/T, p/P
//   - Two-letter units with various capitalizations: kb/Kb/kB/KB, mb/Mb/mB/MB, etc.
//
// Note: Single 'e' and 'E' are not supported to prevent conflicts with scientific
// notation (e.g., 1.23E45). Use 'eb', 'Eb', 'eB', or 'EB' for exabytes.
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

// parseBytes converts a byte slice to a Size by delegating to parseString.
//
// This is an internal helper function used by ParseByte and unmarshaling methods.
func parseBytes(p []byte) (Size, error) {
	return parseString(string(p))
}

// parseString parses a size string into a Size value.
//
// The function implements a flexible parser that accepts size strings in the format:
//
//	[-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
//
// The parser supports:
//   - Optional sign prefix: +/- (negative values return an error)
//   - Integer and decimal numbers: 10, 10.5, .5
//   - Multiple size components: "1GB500MB" (parsed as 1GB + 500MB)
//   - Various unit representations (see unitMap)
//   - Quoted strings: "10MB", '10MB'
//   - Whitespace trimming
//
// Examples of valid inputs:
//   - "10MB"
//   - "1.5GB"
//   - "100"
//   - "1GB500MB"
//   - " 10 MB " (whitespace trimmed)
//
// This is an internal function. Use Parse or ParseByte for public API.
func parseString(s string) (Size, error) {
	// Expected format: [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+

	var (
		orig = s
		neg  bool

		d           uint64
		errNegative = fmt.Errorf("size: negative size '%s'", orig)
		errInvalid  = fmt.Errorf("size: invalid size '%s'", orig)
		errUnit     = fmt.Errorf("size: missing unit '%s'", orig)
		errUnkUnit  = fmt.Errorf("size: unknown unit '%s'", orig)
	)

	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.Trim(s, "'")
	s = strings.TrimSpace(s)

	// Consume optional sign prefix
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Validate that we have content to parse
	if s == "" {
		return 0, errInvalid
	} else if neg {
		return 0, errNegative
	}

	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// Validate the next character is a digit or decimal point
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') { // nolint
			return 0, errInvalid
		}

		// Parse the integer part of the number
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errInvalid
		}
		// Track if we consumed any digits before a decimal point
		pre := pl != len(s)

		// Parse the fractional part (if present)
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

		// Extract and validate the unit string
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

		// Check for overflow before multiplication
		if v > 1<<63/unit {
			return 0, errInvalid
		}

		// Multiply by unit and add fractional part
		v *= unit
		if f > 0 {
			// Use float64 for fractional calculations to maintain precision
			v += uint64(float64(f) * (float64(unit) / scale))

			// Check for overflow after adding fractional part
			if v > 1<<63 {
				return 0, errInvalid
			}
		}
		// Accumulate the parsed value
		d += v

		// Check for overflow in total
		if d > 1<<63 {
			return 0, errInvalid
		}
	}

	return Size(d), nil

}

// leadingInt parses and consumes the leading digits from a string.
//
// This is an internal helper function that extracts the integer part of a number.
// It returns the parsed value, the remaining string, and an error if overflow occurs.
//
// Parameters:
//   - s: The input string
//
// Returns:
//   - x: The parsed uint64 value
//   - rem: The remaining string after consuming digits
//   - err: An error if the value overflows uint64
func leadingInt(s string) (x uint64, rem string, err error) {
	// Internal error (never exposed to user)
	var errLeadingInt = errors.New("size: bad [0-9]*")

	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		// Stop when we hit a non-digit character
		if c < '0' || c > '9' {
			break
		}
		// Check for overflow before multiplying by 10
		if x > 1<<63/10 {
			return 0, "", errLeadingInt
		}
		// Accumulate the digit
		x = x*10 + uint64(c) - '0'

		// Check for overflow after adding digit
		if x > 1<<63 {
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction parses and consumes the fractional digits from a string.
//
// This is an internal helper function that extracts the decimal part of a number.
// Unlike leadingInt, this function does not return an error on overflow; instead,
// it stops accumulating precision once overflow would occur. This is acceptable
// for fractional parts since we only need finite precision.
//
// Parameters:
//   - s: The input string (after the decimal point)
//
// Returns:
//   - x: The parsed digits as a uint64 (without decimal point)
//   - scale: The divisor to convert x back to a fraction (e.g., 100 for two digits)
//   - rem: The remaining string after consuming digits
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	// Track overflow state
	overflow := false

	for ; i < len(s); i++ {
		c := s[i]

		// Stop when we hit a non-digit character
		if c < '0' || c > '9' {
			break
		}
		// Once overflow occurs, just skip remaining digits
		if overflow {
			continue
		}

		// Check for potential overflow
		if x > (1<<63-1)/10 {
			overflow = true
			continue
		}
		// Try to accumulate the digit
		y := x*10 + uint64(c) - '0'

		// Check if accumulation caused overflow
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}
