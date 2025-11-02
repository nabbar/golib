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

// Package cipher provides TLS cipher suite selection and management.
//
// This package defines the Cipher type which represents TLS cipher suites for both
// TLS 1.0-1.2 and TLS 1.3. It provides convenient parsing from strings and integers,
// as well as methods to check cipher suite validity.
//
// Supported Cipher Suites:
//   - TLS 1.0-1.2: RSA, ECDHE-RSA, ECDHE-ECDSA with AES-GCM and ChaCha20-Poly1305
//   - TLS 1.3: AES_128_GCM_SHA256, AES_256_GCM_SHA384, CHACHA20_POLY1305_SHA256
//
// Security Considerations:
//   - Only modern, secure cipher suites are supported
//   - Legacy cipher suites (RC4, 3DES, MD5) are not included
//   - Prefer ECDHE cipher suites for forward secrecy
//   - TLS 1.3 cipher suites provide improved security
//
// Example:
//
//	cipher := cipher.Parse("ECDHE-RSA-AES128-GCM-SHA256")
//	if cipher != cipher.Unknown {
//	    fmt.Println("Supported cipher:", cipher.String())
//	}
package cipher

import (
	"crypto/tls"
	"math"
	"slices"
	"strings"
)

// Cipher represents a TLS cipher suite identifier.
// It wraps the uint16 cipher suite values from crypto/tls and provides parsing capabilities.
type Cipher uint16

const (
	// Unknown represents an unsupported or unrecognized cipher suite.
	Unknown Cipher = Cipher(0)

	// TLS 1.0 - 1.2 cipher suites

	// TLS_RSA_WITH_AES_128_GCM_SHA256 uses RSA key exchange with AES-128-GCM.
	TLS_RSA_WITH_AES_128_GCM_SHA256 = Cipher(tls.TLS_RSA_WITH_AES_128_GCM_SHA256)

	// TLS_RSA_WITH_AES_256_GCM_SHA384 uses RSA key exchange with AES-256-GCM.
	TLS_RSA_WITH_AES_256_GCM_SHA384 = Cipher(tls.TLS_RSA_WITH_AES_256_GCM_SHA384)

	// TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 uses ECDHE key exchange with RSA signatures and AES-128-GCM.
	// Provides forward secrecy.
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 = Cipher(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)

	// TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 uses ECDHE key exchange with ECDSA signatures and AES-128-GCM.
	// Provides forward secrecy.
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 = Cipher(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)

	// TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 uses ECDHE key exchange with RSA signatures and AES-256-GCM.
	// Provides forward secrecy.
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384 = Cipher(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)

	// TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 uses ECDHE key exchange with ECDSA signatures and AES-256-GCM.
	// Provides forward secrecy.
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384 = Cipher(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)

	// TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256 uses ECDHE key exchange with RSA signatures and ChaCha20-Poly1305.
	// Provides forward secrecy. Optimized for mobile devices.
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256 = Cipher(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256)

	// TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 uses ECDHE key exchange with ECDSA signatures and ChaCha20-Poly1305.
	// Provides forward secrecy. Optimized for mobile devices.
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 = Cipher(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256)

	// TLS 1.3 cipher suites

	// TLS_AES_128_GCM_SHA256 is a TLS 1.3 cipher suite using AES-128-GCM.
	TLS_AES_128_GCM_SHA256 = Cipher(tls.TLS_AES_128_GCM_SHA256)

	// TLS_AES_256_GCM_SHA384 is a TLS 1.3 cipher suite using AES-256-GCM.
	TLS_AES_256_GCM_SHA384 = Cipher(tls.TLS_AES_256_GCM_SHA384)

	// TLS_CHACHA20_POLY1305_SHA256 is a TLS 1.3 cipher suite using ChaCha20-Poly1305.
	// Optimized for mobile devices.
	TLS_CHACHA20_POLY1305_SHA256 = Cipher(tls.TLS_CHACHA20_POLY1305_SHA256)
)
const (
	// TLS 1.0 - 1.2 cipher suites no sha for retro compt
	TLS_RSA_WITH_AES_128_GCM Cipher = iota + 1
	TLS_RSA_WITH_AES_256_GCM
	TLS_ECDHE_RSA_WITH_AES_128_GCM
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM
	TLS_ECDHE_RSA_WITH_AES_256_GCM
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM
	TLS_RSA_WITH_AES128_GCM Cipher = iota + 1
	TLS_RSA_WITH_AES256_GCM
	TLS_ECDHE_RSA_WITH_AES128_GCM
	TLS_ECDHE_ECDSA_WITH_AES128_GCM
	TLS_ECDHE_RSA_WITH_AES256_GCM
	TLS_ECDHE_ECDSA_WITH_AES256_GCM
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305

	// TLS 1.3 cipher suites retro compat
	TLS_AES_128_GCM
	TLS_AES_256_GCM
	TLS_AES128_GCM
	TLS_AES256_GCM
	TLS_CHACHA20_POLY1305
)

// List returns all the supported cipher suites.
//
// It includes both TLS 1.0 - 1.2 and TLS 1.3 cipher suites.
func List() []Cipher {
	return []Cipher{
		TLS_RSA_WITH_AES_128_GCM_SHA256,
		TLS_RSA_WITH_AES_256_GCM_SHA384,
		TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		TLS_AES_128_GCM_SHA256,
		TLS_AES_256_GCM_SHA384,
		TLS_CHACHA20_POLY1305_SHA256,
	}
}

// ListString returns a list of all supported cipher suites as strings.
//
// It includes both TLS 1.0 - 1.2 and TLS 1.3 cipher suites.
func ListString() []string {
	var res = make([]string, 0)
	for _, c := range List() {
		res = append(res, c.String())
	}
	return res
}

// Parse returns a Cipher from a given string.
//
// The string is cleaned up by removing any double quotes, single quotes, tls, periods, dashes, and whitespace.
// The cleaned up string is then split into parts separated by underscore.
// The parts are then matched against the codes of the available cipher suites.
//
// If a match is found, the corresponding corresponding Cipher is returned. If no match is found, Unknown is returned.
func Parse(s string) Cipher {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1)  // nolint
	s = strings.Replace(s, "'", "", -1)   // nolint
	s = strings.Replace(s, "tls", "", -1) // nolint
	s = strings.Replace(s, ".", "_", -1)  // nolint
	s = strings.Replace(s, "-", "_", -1)  // nolint
	s = strings.Replace(s, " ", "_", -1)  // nolint
	s = strings.TrimSpace(s)

	p := strings.Split(s, "_")

	switch {
	case containString(p, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256.Code()):
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384.Code()):
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256.Code()):
		return TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256.Code()):
		return TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_CHACHA20_POLY1305_SHA256.Code()):
		return TLS_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_RSA_WITH_AES_128_GCM_SHA256.Code()):
		return TLS_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_RSA_WITH_AES_256_GCM_SHA384.Code()):
		return TLS_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_AES_128_GCM_SHA256.Code()):
		return TLS_AES_128_GCM_SHA256
	case containString(p, TLS_AES_256_GCM_SHA384.Code()):
		return TLS_AES_256_GCM_SHA384
	// retro compat
	case containString(p, TLS_ECDHE_RSA_WITH_AES_128_GCM.Code()):
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES_128_GCM.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_RSA_WITH_AES_256_GCM.Code()):
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES_256_GCM.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305.Code()):
		return TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305.Code()):
		return TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_CHACHA20_POLY1305.Code()):
		return TLS_CHACHA20_POLY1305_SHA256
	case containString(p, TLS_RSA_WITH_AES_128_GCM.Code()):
		return TLS_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_RSA_WITH_AES_256_GCM.Code()):
		return TLS_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_AES_128_GCM.Code()):
		return TLS_AES_128_GCM_SHA256
	case containString(p, TLS_AES_256_GCM.Code()):
		return TLS_AES_256_GCM_SHA384

	case containString(p, TLS_ECDHE_RSA_WITH_AES128_GCM.Code()):
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES128_GCM.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_ECDHE_RSA_WITH_AES256_GCM.Code()):
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_ECDHE_ECDSA_WITH_AES256_GCM.Code()):
		return TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_RSA_WITH_AES128_GCM.Code()):
		return TLS_RSA_WITH_AES_128_GCM_SHA256
	case containString(p, TLS_RSA_WITH_AES256_GCM.Code()):
		return TLS_RSA_WITH_AES_256_GCM_SHA384
	case containString(p, TLS_AES128_GCM.Code()):
		return TLS_AES_128_GCM_SHA256
	case containString(p, TLS_AES256_GCM.Code()):
		return TLS_AES_256_GCM_SHA384
	// not found
	default:
		return Unknown
	}
}

// ParseInt takes an integer and returns a Cipher constant.
//
// If the integer is outside the range [1, math.MaxUint16], it is clamped to the nearest valid value.
// The function uses a switch statement to map the integer to a Cipher constant.
// If no matching Cipher constant is found, the function returns Unknown.
func ParseInt(d int) Cipher {
	var i uint16
	if d > math.MaxUint16 {
		i = math.MaxUint16
	} else if d < 1 {
		i = 0
	} else {
		i = uint16(d)
	}

	switch i {
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		return TLS_RSA_WITH_AES_128_GCM_SHA256
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		return TLS_RSA_WITH_AES_256_GCM_SHA384
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
	case tls.TLS_AES_128_GCM_SHA256:
		return TLS_AES_128_GCM_SHA256
	case tls.TLS_AES_256_GCM_SHA384:
		return TLS_AES_256_GCM_SHA384
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return TLS_CHACHA20_POLY1305_SHA256
	default:
		return Unknown
	}
}

// ParseBytes takes a byte slice and returns a Cipher constant.
//
// The byte slice is first converted to a string, and then passed to Parse.
// If no matching Cipher constant is found, the function returns Unknown.
func ParseBytes(p []byte) Cipher {
	return Parse(string(p))
}

// Check takes a Cipher constant and returns a boolean indicating
// whether the Cipher is valid or not.
//
// The function first calls ParseInt to convert the Cipher constant
// to a uint16 value. If the resulting value is Unknown, the
// function returns false. Otherwise, it returns true.
func Check(cipher uint16) bool {
	if c := ParseInt(int(cipher)); c == Unknown {
		return false
	}
	return true
}

func containString[S ~[]string](s S, v S) bool {
	keys := []string{
		"chacha20",
		"poly1305",
		"ecdhe",
		"rsa",
		"ecdsa",
		"aes",
		"128",
		"256",
		"sha256",
		"sha384",
		"gcm",
	}

	for _, k := range keys {
		if !keyContainString(s, v, k) {
			return false
		}
	}

	return true
}

func keyContainString[S ~[]string](s S, v S, k string) bool {
	if slices.Contains(s, k) && !slices.Contains(v, k) {
		return false
	} else if !slices.Contains(s, k) && slices.Contains(v, k) {
		return false
	}

	return true
}
