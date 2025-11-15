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

// Package curves provides elliptic curve configuration for TLS connections.
//
// This package defines the Curves type which represents elliptic curves used in ECDHE
// (Elliptic Curve Diffie-Hellman Ephemeral) cipher suites. It provides convenient
// parsing from strings and integers.
//
// Supported Curves:
//   - X25519: Modern, high-performance curve (preferred)
//   - P256 (secp256r1): NIST curve, widely supported
//   - P384 (secp384r1): NIST curve for higher security requirements
//   - P521 (secp521r1): NIST curve for maximum security
//
// Security Considerations:
//   - X25519 is preferred for its performance and security properties
//   - NIST curves (P256, P384, P521) are widely compatible but slower
//   - Curve selection affects ECDHE cipher suite performance
//
// Example:
//
//	curve := curves.Parse("X25519")
//	if curve != curves.Unknown {
//	    fmt.Println("Supported curve:", curve.String())
//	}
package curves

import (
	"crypto/tls"
	"math"
	"regexp"
	"strings"
)

var rx *regexp.Regexp

func init() {
	if r, e := regexp.Compile("[0-9]+"); e != nil {
		panic(e)
	} else {
		rx = r
	}
}

// Curves represents an elliptic curve identifier for TLS ECDHE cipher suites.
// It wraps the tls.CurveID values and provides parsing capabilities.
type Curves uint16

const (
	// Unknown represents an unsupported or unrecognized elliptic curve.
	Unknown Curves = iota

	// X25519 is a modern elliptic curve offering high performance and security.
	// Preferred for new deployments.
	X25519 = Curves(tls.X25519)

	// P256 (secp256r1) is a NIST curve providing good balance of security and performance.
	// Widely supported across different platforms and implementations.
	P256 = Curves(tls.CurveP256)

	// P384 (secp384r1) is a NIST curve for higher security requirements.
	// Slower than P256 but provides increased security margin.
	P384 = Curves(tls.CurveP384)

	// P521 (secp521r1) is a NIST curve for maximum security.
	// Slowest option but provides the highest security level.
	P521 = Curves(tls.CurveP521)
)

// List returns the list of all available curves.
//
// The list is ordered by the order the curves were added to the package.
// The list is not empty and contains all the curves in the package.
// The list is ordered by the order the curves were added to the package.
// Modifying the returned slice does not affect the original configuration.
func List() []Curves {
	return []Curves{
		X25519,
		P256,
		P384,
		P521,
	}
}

// ListString returns the list of all available curves as strings.
//
// The list is ordered by the order the curves were added to the package.
// The list is not empty and contains all the curves in the package.
// Modifying the returned slice does not affect the original configuration.
func ListString() []string {
	var res = make([]string, 0)
	for _, c := range List() {
		res = append(res, c.String())
	}
	return res
}

// Parse returns the curve corresponding to the given string.
//
// The Parse function is case-insensitive and accepts strings in the following formats:
// - "25519" for X25519
// - "256" for P256
// - "384" for P384
// - "521" for P521
//
// If the given string does not match any of the above formats, the Parse function returns Unknown.
//
// The Parse function is thread-safe.
// Multiple goroutines can call the Parse function at the same time without affecting the correctness of the TLS configuration.
func Parse(s string) Curves {
	s = strings.ToLower(s)
	s = rx.FindString(s)

	switch {
	case strings.EqualFold(s, "25519"):
		return X25519
	case strings.EqualFold(s, "256"):
		return P256
	case strings.EqualFold(s, "384"):
		return P384
	case strings.EqualFold(s, "521"):
		return P521
	default:
		return Unknown
	}
}

// ParseInt returns the curve corresponding to the given integer.
//
// The ParseInt function takes an integer as an argument and returns the corresponding curve.
// The integer should be in the range of [1, math.MaxUint16].
// If the given integer is not in the range, the ParseInt function adjusts the integer to the range.
// The ParseInt function is thread-safe.
// Multiple goroutines can call the ParseInt function at the same time without affecting the correctness of the TLS configuration.
//
// The ParseInt function returns Unknown if the given integer does not match any of the following curves:
// - tls.X25519 for X25519
// - tls.CurveP256 for P256
// - tls.CurveP384 for P384
// - tls.CurveP521 for P521
func ParseInt(d int) Curves {
	var r tls.CurveID
	if d > math.MaxUint16 {
		r = math.MaxUint16
	} else if d < 1 {
		r = 0
	} else {
		r = tls.CurveID(d)
	}

	switch r {
	case tls.X25519:
		return X25519
	case tls.CurveP256:
		return P256
	case tls.CurveP384:
		return P384
	case tls.CurveP521:
		return P521
	default:
		return Unknown
	}
}

// ParseBytes returns the curve corresponding to the given bytes.
//
// The ParseBytes function takes a []byte as an argument and returns the curve corresponding to the given string.
// The ParseBytes function is thread-safe.
// Multiple goroutines can call the ParseBytes function at the same time without affecting the correctness of the TLS configuration.
//
// The ParseBytes function returns Unknown if the given []byte does not match any of the following curves:
// - tls.X25519 for X25519
// - tls.CurveP256 for P256
// - tls.CurveP384 for P384
// - tls.CurveP521 for P521
func ParseBytes(p []byte) Curves {
	return Parse(string(p))
}

// Check returns true if the given integer is a valid curve, false otherwise.
//
// The Check function takes an integer as an argument and returns true if the integer corresponds to a valid curve, false otherwise.
// The Check function is thread-safe.
// Multiple goroutines can call the Check function at the same time without affecting the correctness of the TLS configuration.
//
// The Check function returns false if the given integer does not match any of the following curves:
// - tls.X25519 for X25519
// - tls.CurveP256 for P256
// - tls.CurveP384 for P384
// - tls.CurveP521 for P521
func Check(curves uint16) bool {
	if c := ParseInt(int(curves)); c == Unknown {
		return false
	}
	return true
}
