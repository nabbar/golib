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
	"math"
)

const (
	errPtnNeg = "size: negative size"
	errPtnInv = "size: invalid"
	errPtnOvf = "(overflow)"
	errPtnUnt = "size: missing unit"
	errPtnUkn = "size: unknown unit"
)

var (
	ErrLeadingInt = errors.New("size: bad [0-9]*")
)

// The parsing logic is based on ParseDuration from the standard library's time package.
// It supports flexible parsing of size strings with various unit representations.

// parseBytes converts a byte slice to a Size by delegating to parseString.
//
// This is an internal helper function used by ParseByte and unmarshaling methods.
func parseBytes(p []byte) (Size, error) {
	return parseString(string(p))
}

func haveSpace(s string, l int) bool {
	for i := 0; i < l; i++ {
		if s[i] == '"' || s[i] == '\'' || s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' {
			return true
		}
	}
	return false
}

func cleanSpace(s string, l int) string {
	p := make([]byte, 0, l)
	for i := 0; i < l; i++ {
		if s[i] == '"' || s[i] == '\'' || s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' {
			continue
		} else {
			p = append(p, s[i])
		}
	}

	return string(p)
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
	// The overall structure is a signed sequence of segments, like:
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64

	if haveSpace(s, len(s)) {
		s = cleanSpace(s, len(s))
	}

	// Consume optional sign prefix.
	if s != "" {
		c := s[0]

		if c == '-' {
			return 0, errNegative(orig)
		}

		if c == '+' {
			s = s[1:]
		}
	}

	// Special case: "0" is a zero duration.
	if s == "0" {
		return 0, nil
	}

	if s == "" {
		return 0, errInvalid(orig)
	}

	for s != "" {
		var (
			v, f  uint64      // integers before and after the decimal point
			scale float64 = 1 // scale factor for the fractional part
		)

		var err error

		// The next character must be a digit or a period.
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') { // nolint
			return 0, errInvalid(orig)
		}

		// Consume the integer part of the number.
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errInvalid(orig)
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
			return 0, errInvalid(orig)
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
			return 0, errUnit(orig)
		}

		u := s[:i]
		s = s[i:]
		//unit, ok := unitMap[u]
		unit, ok := getUnit(u)

		if !ok {
			return 0, errUnkUnit(orig)
		}

		// Check for overflow when scaling the integer part by the unit.
		if float64(v) > float64(math.MaxUint64)/float64(unit) {
			return 0, errOverflow(orig)
		}

		v *= unit

		// Add the fractional part, scaled appropriately.
		if f > 0 {
			frac := uint64(float64(f) * (float64(unit) / scale))

			if v > math.MaxUint64-frac {
				return 0, errOverflow(orig)
			}

			v += frac
		}

		if v > math.MaxUint64-d {
			return 0, errOverflow(orig)
		}

		d += v
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
	i := 0
	for ; i < len(s); i++ {
		c := s[i]

		if c < '0' || c > '9' {
			break
		}

		// Check for overflow before multiplication.
		if x > math.MaxUint64/10 {
			return 0, rem, ErrLeadingInt
		}

		x = x*10 + uint64(c) - '0'
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
		if x > math.MaxUint64/10 {
			overflow = true
			continue
		}

		x = x*10 + uint64(c) - '0'
		scale *= 10
	}

	return x, scale, s[i:]
}

// getUnit return the byte multipliers defined by his letter.
//
// The map supports multiple variations of each unit to provide flexibility:
//   - Single letter units (case-insensitive): b/B, k/K, m/M, g/G, t/T, p/P
//   - Two-letter units with various capitalizations: kb/Kb/kB/KB, mb/Mb/mB/MB, etc.
//
// Note: Single 'e' and 'E' are not supported to prevent conflicts with scientific
// notation (e.g., 1.23E45). Use 'eb', 'Eb', 'eB', or 'EB' for exabytes.
func getUnit(s string) (uint64, bool) {
	if len(s) == 1 {
		switch s {
		case "B":
			return uint64(SizeUnit), true
		case "b":
			return uint64(SizeUnit), true
		case "k":
			return uint64(SizeKilo), true
		case "K":
			return uint64(SizeKilo), true
		case "m":
			return uint64(SizeMega), true
		case "M":
			return uint64(SizeMega), true
		case "g":
			return uint64(SizeGiga), true
		case "G":
			return uint64(SizeGiga), true
		case "t":
			return uint64(SizeTera), true
		case "T":
			return uint64(SizeTera), true
		case "p":
			return uint64(SizePeta), true
		case "P":
			return uint64(SizePeta), true
		// no e/E to prevent scientific writing like 1E23KB
		default:
			return 0, false
		}
	}

	switch s {
	case "kb":
		return uint64(SizeKilo), true
	case "Kb":
		return uint64(SizeKilo), true
	case "kB":
		return uint64(SizeKilo), true
	case "KB":
		return uint64(SizeKilo), true
	case "mb":
		return uint64(SizeMega), true
	case "Mb":
		return uint64(SizeMega), true
	case "mB":
		return uint64(SizeMega), true
	case "MB":
		return uint64(SizeMega), true
	case "gb":
		return uint64(SizeGiga), true
	case "Gb":
		return uint64(SizeGiga), true
	case "gB":
		return uint64(SizeGiga), true
	case "GB":
		return uint64(SizeGiga), true
	case "tb":
		return uint64(SizeTera), true
	case "Tb":
		return uint64(SizeTera), true
	case "tB":
		return uint64(SizeTera), true
	case "TB":
		return uint64(SizeTera), true
	case "pb":
		return uint64(SizePeta), true
	case "Pb":
		return uint64(SizePeta), true
	case "pB":
		return uint64(SizePeta), true
	case "PB":
		return uint64(SizePeta), true
	case "eb":
		return uint64(SizeExa), true
	case "Eb":
		return uint64(SizeExa), true
	case "eB":
		return uint64(SizeExa), true
	case "EB":
		return uint64(SizeExa), true
	default:
		return 0, false
	}
}

func errNegative(s string) error {
	return errors.New(errPtnNeg + " '" + s + "'")
}

func errInvalid(s string) error {
	return errors.New(errPtnInv + " '" + s + "'")
}

func errOverflow(s string) error {
	return errors.New(errPtnInv + " '" + s + "' " + errPtnOvf)
}

func errUnit(s string) error {
	return errors.New(errPtnUnt + " '" + s + "'")
}

func errUnkUnit(s string) error {
	return errors.New(errPtnUkn + " '" + s + "'")
}
