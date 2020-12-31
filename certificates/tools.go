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

package certificates

import (
	"crypto/tls"
	"crypto/x509"
	"runtime"
	"strings"
)

// Deprecated: use StringToCipherKey
func GetCipherKey(cipher string) uint16 {
	return StringToCipherKey(cipher)
}

// Deprecated: use StringToCurveID
func GetCurveID(curveRef string) tls.CurveID {
	return StringToCurveID(curveRef)
}

func SystemRootCA() *x509.CertPool {
	if runtime.GOOS == "windows" {
		return x509.NewCertPool()
	} else if c, e := x509.SystemCertPool(); e == nil {
		return c
	} else {
		return x509.NewCertPool()
	}
}

func StringToTlsVersion(tlsVersStr string) uint16 {
	tlsVersStr = strings.ToLower(tlsVersStr)
	tlsVersStr = strings.Replace(tlsVersStr, "TLS", "", -1)
	tlsVersStr = strings.TrimSpace(tlsVersStr)

	switch tlsVersStr {
	case "1", "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

func StringToClientAuth(auth string) tls.ClientAuthType {
	switch strings.ToLower(auth) {
	case "request":
		return tls.RequestClientCert
	case "require":
		return tls.RequireAnyClientCert
	case "verify":
		return tls.VerifyClientCertIfGiven
	case "strict":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}

func StringToCipherKey(cipher string) uint16 {
	cipher = strings.ToLower(cipher)

	RSA := strings.Contains(cipher, "rsa")
	DSA := strings.Contains(cipher, "ecdsa")
	GSM := strings.Contains(cipher, "gsm")
	AES128 := strings.Contains(cipher, "aes128") || strings.Contains(cipher, "aes_128") || strings.Contains(cipher, "aes-128")
	AES256 := strings.Contains(cipher, "aes256") || strings.Contains(cipher, "aes_256") || strings.Contains(cipher, "aes-256")
	CHACHA := strings.Contains(cipher, "chacha20") || strings.Contains(cipher, "chacha_20") || strings.Contains(cipher, "chacha-20")
	POLY := strings.Contains(cipher, "poly1305") || strings.Contains(cipher, "poly_1305") || strings.Contains(cipher, "poly-1305")

	if RSA && AES128 && !GSM {
		return tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
	} else if RSA && AES128 && GSM {
		return tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	} else if RSA && AES256 {
		return tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	} else if RSA && (CHACHA || POLY) {
		return tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
	} else if DSA && AES128 && !GSM {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
	} else if DSA && AES128 && GSM {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	} else if DSA && AES256 {
		return tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	} else if DSA && (CHACHA || POLY) {
		return tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305
	}

	return 0
}

func StringToCurveID(curveRef string) tls.CurveID {
	curveRef = strings.ToLower(curveRef)

	if strings.Contains(curveRef, "p256") {
		return tls.CurveP256
	} else if strings.Contains(curveRef, "p384") {
		return tls.CurveP384
	} else if strings.Contains(curveRef, "p521") {
		return tls.CurveP521
	} else if strings.Contains(curveRef, "x25519") {
		return tls.X25519
	}

	return 0
}
