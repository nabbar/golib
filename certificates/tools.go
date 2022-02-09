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

// Deprecated: use StringToCipherKey.
func GetCipherKey(cipher string) uint16 {
	return StringToCipherKey(cipher)
}

// Deprecated: use StringToCurveID.
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
	cipher = strings.ReplaceAll(cipher, " ", "")
	cipher = strings.ReplaceAll(cipher, "_", "")
	cipher = strings.ReplaceAll(cipher, "-", "")

	rsa := strings.Contains(cipher, "rsa")
	dhe := strings.Contains(cipher, "ecdhe")
	dsa := strings.Contains(cipher, "ecdsa")
	gcm := strings.Contains(cipher, "gcm")
	aes1 := strings.Contains(cipher, "aes128")
	aes2 := strings.Contains(cipher, "aes256")
	cha := strings.Contains(cipher, "chacha20")
	poly := strings.Contains(cipher, "poly1305")

	switch {
	// TLS v1.0 - v1.2
	case !dhe && rsa && aes1 && !gcm:
		return tls.TLS_RSA_WITH_AES_128_CBC_SHA256
	case !dhe && rsa && aes1 && gcm:
		return tls.TLS_RSA_WITH_AES_128_GCM_SHA256
	case !dhe && rsa && aes2 && gcm:
		return tls.TLS_RSA_WITH_AES_256_GCM_SHA384
	case dhe && dsa && aes1 && !gcm:
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
	case dhe && rsa && aes1 && !gcm:
		return tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
	case dhe && rsa && aes1 && gcm:
		return tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
	case dhe && dsa && aes1 && gcm:
		return tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
	case dhe && rsa && aes2 && gcm:
		return tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
	case dhe && dsa && aes2 && gcm:
		return tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
	case rsa && (cha || poly):
		return tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
	case dsa && (cha || poly):
		return tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256

	// TLS v1.3
	case aes1 && gcm:
		return tls.TLS_AES_128_GCM_SHA256
	case aes2 && gcm:
		return tls.TLS_AES_256_GCM_SHA384
	case cha || poly:
		return tls.TLS_CHACHA20_POLY1305_SHA256

	}

	return 0
}

func StringToCurveID(curveRef string) tls.CurveID {
	curveRef = strings.ToLower(curveRef)

	if strings.Contains(curveRef, "p2") {
		return tls.CurveP256
	} else if strings.Contains(curveRef, "p3") {
		return tls.CurveP384
	} else if strings.Contains(curveRef, "p5") {
		return tls.CurveP521
	} else if strings.Contains(curveRef, "x25") {
		return tls.X25519
	}

	return 0
}
