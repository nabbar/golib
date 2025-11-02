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
 *
 */

package status

import (
	"math"
	"strings"
)

// Status represents the health status of a monitor.
// It is an enumeration type with three possible values: KO, Warn, and OK.
// The underlying type is uint8, making it efficient for storage and comparison.
type Status uint8

const (
	// KO represents a failed or critical status (value: 0).
	KO Status = iota
	// Warn represents a warning or degraded status (value: 1).
	Warn
	// OK represents a healthy or successful status (value: 2).
	OK
)

// Parse converts a string to a Status value.
// It is case-insensitive and handles various formats:
//   - Removes quotes (single and double)
//   - Trims whitespace
//   - Removes "status" prefix if present
//   - Matches "OK", "Warn", or "KO" (case-insensitive)
//
// Returns KO for any unrecognized string.
//
// Examples:
//
//	Parse("OK") -> OK
//	Parse(" warn ") -> Warn
//	Parse("'ko'") -> KO
//	Parse("unknown") -> KO
func Parse(s string) Status {
	s = strings.ToLower(strings.TrimSpace(s)) // nolint
	s = strings.Replace(s, "\"", "", -1)      // nolint
	s = strings.Replace(s, "'", "", -1)       // nolint
	s = strings.Replace(s, " ", "", -1)       // nolint
	s = strings.Replace(s, "status", "", -1)  // nolint

	switch {
	case strings.EqualFold(s, OK.String()):
		return OK
	case strings.EqualFold(s, Warn.String()):
		return Warn
	default:
		return KO
	}
}

// ParseByte converts a byte slice to a Status value.
// It delegates to Parse after converting bytes to string.
func ParseByte(p []byte) Status {
	return Parse(string(p))
}

// ParseUint converts an unsigned integer to a Status value.
// Valid values: 0 (KO), 1 (Warn), 2 (OK).
// Returns KO for any other value.
func ParseUint(i uint) Status {
	return ParseUint64(uint64(i))
}

// ParseUint8 converts a uint8 to a Status value.
// Valid values: 0 (KO), 1 (Warn), 2 (OK).
// Returns KO for any other value.
func ParseUint8(i uint8) Status {
	s := Status(i)
	switch s {
	case OK, Warn:
		return s
	default:
		return KO
	}
}

// ParseUint64 converts a uint64 to a Status value.
// Values greater than math.MaxUint8 return KO.
// Valid values: 0 (KO), 1 (Warn), 2 (OK).
// Returns KO for any other value.
func ParseUint64(i uint64) Status {
	if i > uint64(math.MaxUint8) {
		return KO
	} else {
		return ParseUint8(uint8(i))
	}
}

// ParseInt converts an int to a Status value.
// Negative values and values greater than math.MaxUint8 return KO.
// Valid values: 0 (KO), 1 (Warn), 2 (OK).
// Returns KO for any other value.
func ParseInt(i int) Status {
	return ParseInt64(int64(i))
}

// ParseInt64 converts an int64 to a Status value.
// Negative values and values greater than math.MaxUint8 return KO.
// Valid values: 0 (KO), 1 (Warn), 2 (OK).
// Returns KO for any other value.
func ParseInt64(i int64) Status {
	if i > int64(math.MaxUint8) {
		return KO
	} else if i <= 0 {
		return ParseUint8(0)
	} else {
		return ParseUint8(uint8(i))
	}
}

// ParseFloat64 converts a float64 to a Status value.
// The value is floored before conversion.
// Negative values and values greater than math.MaxUint8 return KO.
// Valid values: 0.x (KO), 1.x (Warn), 2.x (OK).
// Returns KO for any other value.
func ParseFloat64(i float64) Status {
	if p := math.Floor(i); p > float64(math.MaxUint8) {
		return KO
	} else if p <= 0 {
		return ParseUint8(0)
	} else {
		return ParseUint8(uint8(p))
	}
}

// NewFromString returns a Status based on the given string.
// It is case-insensitive and supports the following values:
// - "ok" for OK status
// - "warn" for Warn status
// Any other value will return KO status.
// Deprecated: see Parse
func NewFromString(sts string) Status {
	return Parse(sts)
}

// NewFromInt returns a Status based on the given int64 value.
// If the value is greater than the maximum value of a uint8, it will return KO status.
// Otherwise, it will return the corresponding Status value.
// If the value doesn't match any valid Status value, it will return KO status.
// Deprecated: see ParseInt64
func NewFromInt(i int64) Status {
	return ParseInt64(i)
}
