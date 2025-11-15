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

// Package auth provides client authentication mode types and parsing for TLS connections.
//
// This package defines the ClientAuth type which wraps tls.ClientAuthType and provides
// convenient parsing from strings and other formats. It supports all standard TLS client
// authentication modes.
//
// Client Authentication Modes:
//   - NoClientCert: Server will not request client certificates
//   - RequestClientCert: Server requests but doesn't require client certificates
//   - RequireAnyClientCert: Server requires client certificate but doesn't verify it
//   - VerifyClientCertIfGiven: Server verifies client certificate if provided
//   - RequireAndVerifyClientCert: Server requires and verifies client certificates
//
// Example:
//
//	auth := auth.Parse("require")
//	if auth == auth.RequireAndVerifyClientCert {
//	    // Configure strict client authentication
//	}
package auth

import (
	"crypto/tls"
	"strings"
)

const (
	strict  = "strict"
	require = "require"
	verify  = "verify"
	request = "request"
	none    = "none"
)

// ClientAuth represents the client authentication policy for TLS connections.
// It wraps tls.ClientAuthType and provides parsing and formatting capabilities.
type ClientAuth tls.ClientAuthType

const (
	// NoClientCert indicates that no client certificate is requested or required.
	NoClientCert = ClientAuth(tls.NoClientCert)

	// RequestClientCert indicates that a client certificate is requested but not required.
	RequestClientCert = ClientAuth(tls.RequestClientCert)

	// RequireAnyClientCert indicates that at least one client certificate is required,
	// but it doesn't need to be valid or verified.
	RequireAnyClientCert = ClientAuth(tls.RequireAnyClientCert)

	// VerifyClientCertIfGiven indicates that if a client certificate is provided,
	// it must be valid and verified.
	VerifyClientCertIfGiven = ClientAuth(tls.VerifyClientCertIfGiven)

	// RequireAndVerifyClientCert indicates that a valid client certificate is required
	// and must be verified against the client CA pool.
	RequireAndVerifyClientCert = ClientAuth(tls.RequireAndVerifyClientCert)
)

// List returns all possible ClientAuth values.
//
// The returned slice is in an arbitrary order.
//
// The returned value is not a copy of a known ClientAuth.
// The returned value is thread-safe.
// Multiple goroutines can call the List function at the same time without affecting the correctness of the TLS configuration.
func List() []ClientAuth {
	return []ClientAuth{
		NoClientCert,
		RequestClientCert,
		RequireAnyClientCert,
		VerifyClientCertIfGiven,
		RequireAndVerifyClientCert,
	}
}

// Parse returns the ClientAuth corresponding as a ClientAuth, given a string s.
//
// The function takes a string s that represents a tls.ClientAuthType.
// The function returns the ClientAuth that matches the string, or NoClientCert if no match is found.
//
// The returned value is a reference to a known ClientAuth.
// The returned value is not a copy of a known ClientAuth.
// The returned value is thread-safe.
// Multiple goroutines can call the Parse function at the same time without affecting the correctness of the TLS configuration.
//
// The string s can contain one of the following keywords:
// - "strict"
// - "require"
// - "verify"
// - "request"
// - "none"
//
// The function will return the corresponding ClientAuth value if the string s contains the keyword.
// If the string s does not contain any of the keywords, the function will return NoClientCert.
func Parse(s string) ClientAuth {
	s = cleanString(s)

	switch {
	case strings.Contains(s, strict) || (strings.Contains(s, require) && strings.Contains(s, verify)):
		return RequireAndVerifyClientCert
	case strings.Contains(s, verify):
		return VerifyClientCertIfGiven
	case strings.Contains(s, require) && !strings.Contains(s, verify):
		return RequireAnyClientCert
	case strings.Contains(s, request):
		return RequestClientCert
	default:
		return NoClientCert
	}
}

// ParseInt returns the ClientAuth corresponding as a ClientAuth, given an integer d.
//
// The function takes an integer d that represents a tls.ClientAuthType.
// The function returns the ClientAuth that matches the integer, or NoClientCert if no match is found.
//
// The returned value is a reference to a known ClientAuth.
// The returned value is not a copy of a known ClientAuth.
// The returned value is thread-safe.
// Multiple goroutines can call the ParseInt function at the same time without affecting the correctness of the TLS configuration.
func ParseInt(d int) ClientAuth {
	switch tls.ClientAuthType(d) {
	case tls.RequireAndVerifyClientCert:
		return RequireAndVerifyClientCert
	case tls.VerifyClientCertIfGiven:
		return VerifyClientCertIfGiven
	case tls.RequireAnyClientCert:
		return RequireAnyClientCert
	case tls.RequestClientCert:
		return RequestClientCert
	default:
		return NoClientCert
	}
}

// ParseBytes returns the ClientAuth corresponding as a ClientAuth, given a byte slice p.
//
// The function takes a byte slice p that represents a ClientAuth.
// The function returns the ClientAuth that matches the byte slice, or NoClientCert if no match is found.
//
// The returned value is a reference to a known ClientAuth.
// The returned value is not a copy of a known ClientAuth.
// The returned value is thread-safe.
// Multiple goroutines can call the ParseBytes function at the same time without affecting the correctness of the TLS configuration.
func ParseBytes(p []byte) ClientAuth {
	return Parse(string(p))
}

func cleanString(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1) // nolint
	s = strings.Replace(s, "'", "", -1)  // nolint
	return strings.TrimSpace(s)
}
