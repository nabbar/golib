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

// Package control provides control modes for mandatory component validation.
//
// This package defines different validation modes that can be applied to
// mandatory components in a status monitoring system. These modes determine
// how strictly components must be validated.
//
// # Available Modes
//
// The package defines five control modes:
//
//   - Ignore: No validation required (default)
//   - Should: Component should be present but is not mandatory
//   - Must: Component must be present and healthy
//   - AnyOf: At least one of the components must be healthy
//   - Quorum: A majority of components must be healthy
//
// # Usage Example
//
//	import "github.com/nabbar/golib/status/control"
//
//	// Parse a mode from string
//	mode := control.Parse("must")
//	fmt.Println(mode.String()) // Output: Must
//
//	// Use in configuration
//	if mode == control.Must {
//	    // Enforce strict validation
//	}
//
// # Serialization
//
// The Mode type supports multiple serialization formats:
//   - JSON: Marshals to/from string representation
//   - YAML: Marshals to/from string representation
//   - TOML: Marshals to/from string representation
//   - Text: Marshals to/from string representation
//   - CBOR: Marshals to/from string representation
//
// All parsing is case-insensitive for convenience.
//
// # Integration
//
// This package is designed to work with:
//   - github.com/nabbar/golib/status/mandatory: For managing mandatory components
//   - github.com/nabbar/golib/status: For overall status management
//
// See also: github.com/nabbar/golib/status/mandatory for usage with mandatory components.
package control

import (
	"math"
	"strings"
)

// Mode represents a control mode for mandatory component validation.
//
// Mode determines how strictly components must be validated in a status
// monitoring system. It is implemented as a uint8 for efficient storage
// and comparison.
//
// The zero value (Ignore) means no validation is required.
type Mode uint8

const (
	// Ignore indicates no validation is required for the component.
	// This is the default mode and allows components to be absent or unhealthy
	// without affecting the overall status.
	Ignore Mode = iota

	// Should indicates the component should be present but is not mandatory.
	// If the component is absent or unhealthy, it may generate a warning
	// but will not cause a failure.
	Should

	// Must indicates the component must be present and healthy.
	// If the component is absent or unhealthy, the overall status will fail.
	Must

	// AnyOf indicates at least one of the components in the group must be healthy.
	// This mode is useful for redundant components where any one can satisfy
	// the requirement.
	AnyOf

	// Quorum indicates a majority of components in the group must be healthy.
	// This mode is useful for distributed systems where a quorum is required
	// for proper operation.
	Quorum
)

// Parse converts a string to a Mode.
//
// The parsing is case-insensitive and supports the following values:
//   - "should" -> Should
//   - "must" -> Must
//   - "anyof" -> AnyOf
//   - "quorum" -> Quorum
//   - any other value -> Ignore
//
// Example:
//
//	mode := control.Parse("MUST")
//	fmt.Println(mode) // Output: Must
//
//	mode = control.Parse("invalid")
//	fmt.Println(mode) // Output: (empty string for Ignore)
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

// ParseBytes converts a byte slice to a Mode.
//
// This is a convenience wrapper around Parse that converts the byte slice
// to a string before parsing. The parsing is case-insensitive.
//
// Example:
//
//	mode := control.ParseBytes([]byte("must"))
//	fmt.Println(mode) // Output: Must
func ParseBytes(p []byte) Mode {
	return Parse(string(p))
}

// ParseUint64 converts a uint64 to a Mode.
//
// This function is useful when reading Mode values from numeric configuration
// or database fields. Values are mapped as follows:
//   - 0 -> Ignore
//   - 1 -> Should
//   - 2 -> Must
//   - 3 -> AnyOf
//   - 4 -> Quorum
//   - any other value -> Ignore
//
// Values larger than math.MaxUint8 are clamped to MaxUint8 before conversion.
//
// Example:
//
//	mode := control.ParseUint64(2)
//	fmt.Println(mode) // Output: Must
//
//	mode = control.ParseUint64(999)
//	fmt.Println(mode) // Output: (empty string for Ignore)
func ParseUint64(p uint64) Mode {
	var m Mode
	if p > uint64(math.MaxUint8) {
		m = Mode(math.MaxUint8)
	} else {
		m = Mode(p)
	}

	switch m {
	case Should:
		return Should
	case Must:
		return Must
	case AnyOf:
		return AnyOf
	case Quorum:
		return Quorum
	default:
		return Ignore
	}
}

// ParseInt64 converts an int64 to a Mode.
//
// This function is useful when reading Mode values from signed numeric
// configuration or database fields. Negative values are treated as 0 (Ignore).
//
// Values are mapped as follows:
//   - 0 or negative -> Ignore
//   - 1 -> Should
//   - 2 -> Must
//   - 3 -> AnyOf
//   - 4 -> Quorum
//   - any other value -> Ignore
//
// Example:
//
//	mode := control.ParseInt64(2)
//	fmt.Println(mode) // Output: Must
//
//	mode = control.ParseInt64(-1)
//	fmt.Println(mode) // Output: (empty string for Ignore)
func ParseInt64(p int64) Mode {
	if p < 0 {
		return ParseUint64(0)
	} else {
		return ParseUint64(uint64(p))
	}
}
