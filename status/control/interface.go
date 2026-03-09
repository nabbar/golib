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

// Package control provides the validation modes that govern how component health
// affects the overall application status. These modes are used to define flexible
// and robust health check policies.
//
// The package defines a set of standard control modes (Ignore, Should, Must,
// AnyOf, Quorum) and provides utilities for parsing and serializing them.
package control

import (
	"math"
	"strings"
)

// Mode represents a control mode for mandatory component validation. It determines
// the strictness of the health check for a component or a group of components.
// It is implemented as a `uint8` for efficiency.
type Mode uint8

const (
	// Ignore indicates that no validation is required for the component.
	// This is the default mode (zero value). Components in this mode are
	// monitored, but their status (even if KO) does not affect the overall
	// application status.
	Ignore Mode = iota

	// Should indicates that the component is important but not critical. If the
	// component is unhealthy (KO or WARN), it will generate a warning but will
	// not cause a critical failure (KO) of the overall application. This is
	// useful for optional features or degraded modes.
	Should

	// Must indicates that the component is critical and must be healthy. If the
	// component is unhealthy (KO), the overall application status will be marked
	// as failed (KO). If it is WARN, the overall status will be WARN.
	Must

	// AnyOf is used for redundant groups of components (e.g., a cluster of
	// read-only databases). It requires at least one component in the group to
	// be healthy (OK or WARN). If all components are KO, the group is KO.
	AnyOf

	// Quorum is used for distributed systems requiring consensus. It requires a
	// majority (>50%) of the components in the group to be healthy (OK or WARN).
	// If 50% or fewer are healthy, the group is considered KO.
	Quorum
)

// Parse converts a string to a `Mode`. The parsing is case-insensitive.
// If the string does not match any known mode, `Ignore` is returned as the default.
//
// Supported values:
//   - "ignore" -> Ignore
//   - "should" -> Should
//   - "must"   -> Must
//   - "anyof"  -> AnyOf
//   - "quorum" -> Quorum
func Parse(s string) Mode {
	switch {
	case strings.EqualFold(Should.Code(), s):
		return Should
	case strings.EqualFold(Must.Code(), s):
		return Must
	case strings.EqualFold(AnyOf.Code(), s):
		return AnyOf
	case strings.EqualFold(Quorum.Code(), s):
		return Quorum
	}

	return Ignore
}

// ParseBytes is a convenience wrapper for `Parse` that accepts a byte slice.
// It converts the byte slice to a string and calls `Parse`.
func ParseBytes(p []byte) Mode {
	return Parse(string(p))
}

// ParseUint64 converts a `uint64` to a `Mode`. This is useful when reading mode
// values from numeric configurations or databases. If the value is out of the
// valid range for `Mode`, `Ignore` is returned.
//
// Mapping:
//   - 0 -> Ignore
//   - 1 -> Should
//   - 2 -> Must
//   - 3 -> AnyOf
//   - 4 -> Quorum
func ParseUint64(p uint64) Mode {
	var m Mode
	if p > uint64(math.MaxUint8) {
		m = Mode(math.MaxUint8)
	} else {
		m = Mode(p)
	}

	switch m {
	case Should, Must, AnyOf, Quorum:
		return m
	default:
		return Ignore
	}
}

// ParseInt64 converts an `int64` to a `Mode`. Negative values are treated as 0
// (`Ignore`). This is useful for signed numeric configurations.
//
// Mapping:
//   - < 0 -> Ignore
//   - 0 -> Ignore
//   - 1 -> Should
//   - 2 -> Must
//   - 3 -> AnyOf
//   - 4 -> Quorum
func ParseInt64(p int64) Mode {
	if p < 0 {
		return ParseUint64(0)
	} else {
		return ParseUint64(uint64(p))
	}
}
