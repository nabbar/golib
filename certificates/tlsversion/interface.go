/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package tlsversion provides TLS version management and parsing.
//
// This package defines the Version type which represents TLS protocol versions.
// It provides convenient parsing from strings and integers, as well as methods
// to set minimum and maximum supported TLS versions.
//
// Supported Versions:
//   - TLS 1.0: Legacy, not recommended (deprecated)
//   - TLS 1.1: Legacy, not recommended (deprecated)
//   - TLS 1.2: Widely supported, secure
//   - TLS 1.3: Modern, most secure (preferred)
//
// Security Recommendations:
//   - Use TLS 1.2 as minimum version for compatibility
//   - Use TLS 1.3 as maximum version for best security
//   - Avoid TLS 1.0 and 1.1 (deprecated and insecure)
//   - TLS 1.3 provides improved performance and security
//
// Example:
//
//	minVer := tlsversion.Parse("1.2")
//	maxVer := tlsversion.Parse("1.3")
//	config.SetVersionMin(minVer)
//	config.SetVersionMax(maxVer)
package tlsversion

import (
	"crypto/tls"
	"strings"
)

// Version represents a TLS protocol version.
// It wraps the int version values from crypto/tls and provides parsing capabilities.
type Version int

const (
	// VersionUnknown represents an unsupported or unrecognized TLS version.
	VersionUnknown Version = iota

	// VersionTLS10 represents TLS 1.0 (deprecated, not recommended).
	// Should only be used for legacy compatibility.
	VersionTLS10 = Version(tls.VersionTLS10)

	// VersionTLS11 represents TLS 1.1 (deprecated, not recommended).
	// Should only be used for legacy compatibility.
	VersionTLS11 = Version(tls.VersionTLS11)

	// VersionTLS12 represents TLS 1.2 (secure, widely supported).
	// Recommended as minimum version for most deployments.
	VersionTLS12 = Version(tls.VersionTLS12)

	// VersionTLS13 represents TLS 1.3 (modern, most secure).
	// Recommended for new deployments, provides improved performance and security.
	VersionTLS13 = Version(tls.VersionTLS13)
)

// List returns a slice of all known TLS versions, in descending order of version.
//
// The returned slice contains all known TLS versions, from the highest to the lowest.
// The returned slice is not a copy of the known TLS versions.
// The returned slice is a reference to the known TLS versions.
func List() []Version {
	return []Version{
		VersionTLS13,
		VersionTLS12,
		VersionTLS11,
		VersionTLS10,
	}
}

// ListHigh returns a slice of the highest TLS versions, in descending order of version.
//
// The returned slice contains the highest TLS versions, from the highest to the second highest.
// The returned slice is not a copy of the highest TLS versions.
// The returned slice is a reference to the highest TLS versions.
// The returned slice contains at most two TLS versions.
func ListHigh() []Version {
	return []Version{
		VersionTLS13,
		VersionTLS12,
	}
}

// Parse returns the TLS version corresponding as a Version.
//
// The function takes a string that represents a TLS version.
// The string is case-insensitive and can contain any of the following characters:
//   - " (double quote)
//   - ' (single quote)
//   - tls (the string "tls" regardless of case)
//   - ssl (the string "ssl" regardless of case)
//   - . (period)
//   - - (hyphen)
//   - _ (underscore)
//   - (space)
//
// The function returns the TLS version that matches the string, or VersionUnknown if no match is found.
//
// The returned value is a reference to a known TLS version.
// The returned value is not a copy of a known TLS version.
// The returned value is thread-safe.
// Multiple goroutines can call the Parse function at the same time without affecting the correctness of the TLS configuration.
func Parse(s string) Version {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1)  // nolint
	s = strings.Replace(s, "'", "", -1)   // nolint
	s = strings.Replace(s, "tls", "", -1) // nolint
	s = strings.Replace(s, "ssl", "", -1) // nolint
	s = strings.Replace(s, ".", "", -1)   // nolint
	s = strings.Replace(s, "-", "", -1)   // nolint
	s = strings.Replace(s, "_", "", -1)   // nolint
	s = strings.Replace(s, " ", "", -1)   // nolint
	s = strings.TrimSpace(s)

	switch {
	case strings.EqualFold(s, "1"):
		return VersionTLS10
	case strings.EqualFold(s, "10"):
		return VersionTLS10
	case strings.EqualFold(s, "11"):
		return VersionTLS11
	case strings.EqualFold(s, "12"):
		return VersionTLS12
	case strings.EqualFold(s, "13"):
		return VersionTLS13
	default:
		return VersionUnknown
	}
}

// ParseInt returns the TLS version corresponding as a Version, given an integer d.
//
// The function takes an integer d that represents a TLS version.
// The function returns the TLS version that matches the integer, or VersionUnknown if no match is found.
//
// The returned value is a reference to a known TLS version.
// The returned value is not a copy of a known TLS version.
// The returned value is thread-safe.
// Multiple goroutines can call the ParseInt function at the same time without affecting the correctness of the TLS configuration.
func ParseInt(d int) Version {
	switch d {
	case tls.VersionTLS10:
		return VersionTLS10
	case tls.VersionTLS11:
		return VersionTLS11
	case tls.VersionTLS12:
		return VersionTLS12
	case tls.VersionTLS13:
		return VersionTLS13
	default:
		return VersionUnknown
	}
}

// ParseBytes returns the TLS version corresponding as a Version, given a byte slice p.
//
// The function takes a byte slice p that represents a TLS version.
// The function returns the TLS version that matches the byte slice, or VersionUnknown if no match is found.
//
// The returned value is a reference to a known TLS version.
// The returned value is not a copy of a known TLS version.
// The returned value is thread-safe.
// Multiple goroutines can call the ParseBytes function at the same time without affecting the correctness of the TLS configuration.
func ParseBytes(p []byte) Version {
	return Parse(string(p))
}
