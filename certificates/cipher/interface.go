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

package curves

import (
	"crypto/tls"
	"math"
	"slices"
	"strings"
)

type Cipher uint16

const (
	Unknown Cipher = Cipher(0)

	// TLS 1.0 - 1.2 cipher suites.
	TLS_RSA_WITH_AES_128_GCM_SHA256               = Cipher(tls.TLS_RSA_WITH_AES_128_GCM_SHA256)
	TLS_RSA_WITH_AES_256_GCM_SHA384               = Cipher(tls.TLS_RSA_WITH_AES_256_GCM_SHA384)
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256         = Cipher(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256       = Cipher(tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
	TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384         = Cipher(tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
	TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384       = Cipher(tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256   = Cipher(tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256)
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 = Cipher(tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256)

	// TLS 1.3 cipher suites.
	TLS_AES_128_GCM_SHA256       = Cipher(tls.TLS_AES_128_GCM_SHA256)
	TLS_AES_256_GCM_SHA384       = Cipher(tls.TLS_AES_256_GCM_SHA384)
	TLS_CHACHA20_POLY1305_SHA256 = Cipher(tls.TLS_CHACHA20_POLY1305_SHA256)
)

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

func ListString() []string {
	var res = make([]string, 0)
	for _, c := range List() {
		res = append(res, c.String())
	}
	return res
}

func Parse(s string) Cipher {
	s = strings.ToLower(s)
	s = strings.Replace(s, "\"", "", -1)
	s = strings.Replace(s, "'", "", -1)
	s = strings.Replace(s, "tls", "", -1)
	s = strings.Replace(s, ".", "_", -1)
	s = strings.Replace(s, "-", "_", -1)
	s = strings.Replace(s, " ", "_", -1)
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
	default:
		return Unknown
	}
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

func ParseInt(d int) Cipher {
	if d > math.MaxUint16 {
		d = math.MaxUint16
	} else if d < 1 {
		d = 0
	}

	switch uint16(d) {
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

func Check(cipher uint16) bool {
	if c := ParseInt(int(cipher)); c == Unknown {
		return false
	}
	return true
}

func parseBytes(p []byte) Cipher {
	return Parse(string(p))
}
