/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

// Package tlsmode provides TLS connection mode types and parsing for SMTP connections.
//
// This package defines three TLS connection modes for SMTP:
//   - TLSNone: Plain SMTP connection without encryption
//   - TLSStartTLS: SMTP connection with opportunistic STARTTLS upgrade
//   - TLSStrictTLS: Direct TLS connection (SMTPS)
//
// The package supports parsing from various formats (string, numeric, bytes) and
// marshaling/unmarshaling to/from JSON, YAML, TOML, CBOR, and binary formats.
//
// Example:
//
//	mode := tlsmode.Parse("starttls")
//	fmt.Println(mode.String()) // Output: starttls
//
//	data, _ := json.Marshal(mode)
//	fmt.Println(string(data)) // Output: "starttls"
package tlsmode

import (
	"math"
	"strings"
)

// TLSMode represents the TLS connection mode for SMTP.
//
// This type is used to specify how TLS should be established for an SMTP connection.
// It can be marshaled to and unmarshaled from various formats including JSON, YAML,
// TOML, CBOR, and binary formats.
//
// The type is represented as a uint8 internally, allowing efficient storage and comparison.
type TLSMode uint8

// TLS connection modes for SMTP.
const (
	// TLSNone indicates a plain SMTP connection without encryption.
	// This mode sends all data in clear text and should only be used
	// for testing or on trusted networks.
	// Standard port: 25
	TLSNone TLSMode = iota

	// TLSStartTLS indicates an SMTP connection that starts unencrypted
	// and upgrades to TLS using the STARTTLS command.
	// This is also known as opportunistic TLS or explicit TLS.
	// Standard port: 587 (submission port)
	TLSStartTLS

	// TLSStrictTLS indicates a direct TLS connection (SMTPS).
	// The connection is encrypted from the start using implicit TLS.
	// Also known as SMTPS or implicit TLS.
	// Standard port: 465
	TLSStrictTLS
)

// Parse parses a string into a TLSMode.
//
// This function is case-insensitive and strips common formatting characters
// (whitespace, quotes, underscores, hyphens) before parsing.
//
// Recognized values:
//   - "starttls", "start-tls", "start_tls", "STARTTLS" → TLSStartTLS
//   - "tls", "TLS" → TLSStrictTLS
//   - "" (empty string), "none", or any other value → TLSNone
//
// Example:
//
//	mode := Parse("starttls")        // TLSStartTLS
//	mode = Parse("  TLS  ")          // TLSStrictTLS
//	mode = Parse("start-tls")        // TLSStartTLS
//	mode = Parse("invalid")          // TLSNone
func Parse(s string) TLSMode {
	// Normalize the string by removing common formatting characters
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\r", "", -1) // nolint
	s = strings.Replace(s, "\n", "", -1) // nolint
	s = strings.Replace(s, "\"", "", -1) // nolint
	s = strings.Replace(s, "'", "", -1)  // nolint
	s = strings.Replace(s, " ", "", -1)  // nolint
	s = strings.Replace(s, "_", "", -1)  // nolint
	s = strings.Replace(s, "-", "", -1)  // nolint

	// Check for exact matches first before stripping "tls"
	switch {
	case strings.EqualFold(s, "starttls"):
		return TLSStartTLS
	case strings.EqualFold(s, "tls"):
		return TLSStrictTLS
	default:
		return TLSNone
	}
}

// ParseBytes parses a byte slice into a TLSMode.
//
// This is a convenience function that converts the byte slice to a string
// and calls Parse. See Parse for supported values.
//
// Example:
//
//	mode := ParseBytes([]byte("starttls")) // TLSStartTLS
func ParseBytes(p []byte) TLSMode {
	return Parse(string(p))
}

// ParseUint64 parses an unsigned integer into a TLSMode.
//
// Valid values:
//   - 0 → TLSNone
//   - 1 → TLSStartTLS
//   - 2 → TLSStrictTLS
//   - Any value > 255 or invalid → TLSNone
//
// Example:
//
//	mode := ParseUint64(1) // TLSStartTLS
//	mode = ParseUint64(2)  // TLSStrictTLS
func ParseUint64(i uint64) TLSMode {
	var p TLSMode
	if i > uint64(math.MaxUint8) {
		return TLSNone
	} else {
		p = TLSMode(i)
	}

	switch p {
	case TLSStrictTLS:
		return TLSStrictTLS
	case TLSStartTLS:
		return TLSStartTLS
	default:
		return TLSNone
	}
}

// ParseInt64 parses a signed integer into a TLSMode.
//
// Negative values return TLSNone. Otherwise, the value is converted to
// uint64 and parsed with ParseUint64.
//
// Valid values:
//   - 0 → TLSNone
//   - 1 → TLSStartTLS
//   - 2 → TLSStrictTLS
//   - Negative or > 255 → TLSNone
//
// Example:
//
//	mode := ParseInt64(1)  // TLSStartTLS
//	mode = ParseInt64(-1)  // TLSNone
func ParseInt64(i int64) TLSMode {
	if i < 0 {
		return TLSNone
	} else {
		return ParseUint64(uint64(i))
	}
}

// ParseFloat64 parses a floating-point number into a TLSMode.
//
// The float is floored to an integer before parsing. Values greater than
// 255 or with a floored value greater than 255 return TLSNone.
//
// Valid values:
//   - 0.0 to 0.999 → TLSNone
//   - 1.0 to 1.999 → TLSStartTLS
//   - 2.0 to 2.999 → TLSStrictTLS
//   - > 255.0 → TLSNone
//
// Example:
//
//	mode := ParseFloat64(1.5) // TLSStartTLS (floored to 1)
//	mode = ParseFloat64(2.9)  // TLSStrictTLS (floored to 2)
func ParseFloat64(f float64) TLSMode {
	if f > math.MaxUint8 {
		return TLSNone
	} else if p := math.Floor(f); p > math.MaxUint8 {
		return TLSNone
	} else {
		return ParseInt64(int64(p))
	}
}

// TLSModeFromString parses a string into a TLSMode.
//
// Deprecated: Use Parse instead. This function is maintained for backward compatibility.
func TLSModeFromString(str string) TLSMode {
	return Parse(str)
}

// TLSModeFromInt parses an integer into a TLSMode.
//
// Deprecated: Use ParseInt64 instead. This function is maintained for backward compatibility.
func TLSModeFromInt(i int64) TLSMode {
	return ParseInt64(i)
}
