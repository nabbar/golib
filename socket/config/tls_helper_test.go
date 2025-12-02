/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package config_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"

	. "github.com/onsi/gomega"
)

var (
	// cfgTLSSrv is the pre-generated Server TLS configuration initialized in BeforeSuite.
	// This avoids regenerating certificates inline during each test, significantly improving test performance.
	// The configuration includes a self-signed certificate valid for localhost with TLS 1.2-1.3 support.
	cfgTLSSrv libtls.Config

	genTLSKey  string
	genTLSCert string
)

// genPairPEM generates a temporary self-signed certificate pair for testing purposes.
//
// This function creates an ECDSA P-256 key pair and a self-signed X.509 certificate
// valid for localhost. The certificate is suitable for testing TLS configurations
// but should never be used in production.
//
// The generated certificate has the following characteristics:
//   - Algorithm: ECDSA with P-256 curve (provides good security and performance)
//   - Validity: 24 hours from generation time
//   - Usage: Key encipherment and digital signature
//   - Extended usage: Server authentication
//   - Subject: CN=localhost, O=Test Organization
//   - DNS names: localhost, 127.0.0.1
//
// Returns:
//   - pub: PEM-encoded certificate string
//   - key: PEM-encoded private key string (EC PRIVATE KEY format)
//   - err: error if certificate generation fails
//
// Example usage:
//
//	certPEM, keyPEM, err := genPairPEM()
//	if err != nil {
//	    // handle error
//	}
//	// Use certPEM and keyPEM for TLS configuration
func genPairPEM() (pub string, key string, err error) {
	var (
		tpl x509.Certificate

		serNbr  *big.Int
		privKey *ecdsa.PrivateKey

		crtDER []byte
		crtBuf *bytes.Buffer

		keyDER []byte
		keyBuf *bytes.Buffer
	)

	// Generate private key
	privKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	// Create certificate tpl
	serNbr, err = rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", err
	}

	tpl = x509.Certificate{
		SerialNumber: serNbr,
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "127.0.0.1"},
	}

	// Create self-signed certificate
	crtDER, err = x509.CreateCertificate(rand.Reader, &tpl, &tpl, &privKey.PublicKey, privKey)
	if err != nil {
		return "", "", err
	}

	// Write certificate file
	crtBuf = bytes.NewBufferString("")
	if err = pem.Encode(crtBuf, &pem.Block{Type: "CERTIFICATE", Bytes: crtDER}); err != nil {
		return "", "", err
	}

	// Write private key file
	keyDER, err = x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", "", err
	}

	keyBuf = bytes.NewBufferString("")
	if err = pem.Encode(keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return "", "", err
	}

	return crtBuf.String(), keyBuf.String(), nil
}

// testTLSConfigDefault returns a default TLS configuration for testing without certificates.
//
// This configuration includes:
//   - All supported elliptic curves (P-256, P-384, P-521, X25519)
//   - All supported cipher suites from the certificates package
//   - TLS version range: 1.2 to 1.3
//
// This is useful for client configurations that don't require client certificates
// or for base configurations that will be extended with certificates later.
//
// Returns:
//   - A certificates.Config with default TLS parameters
func testTLSConfigDefault() libtls.TLSConfig {
	c := &libtls.Config{
		CurveList:  tlscrv.List(),
		CipherList: tlscpr.List(),
		VersionMin: tlsvrs.VersionTLS12,
		VersionMax: tlsvrs.VersionTLS13,
	}

	return c.New()
}

// testTLSConfigClient returns a minimal TLS configuration suitable for client connections.
//
// This configuration includes only the TLS version range (1.2 to 1.3) and relies on
// default values for curves and cipher suites. This is appropriate for client-side
// configurations where the server dictates the negotiated parameters.
//
// The configuration does not include client certificates, making it suitable for
// server authentication only (one-way TLS).
//
// Returns:
//   - A certificates.Config with minimal client TLS parameters
func testTLSConfigClient() libtls.Config {
	return libtls.Config{
		VersionMin: tlsvrs.VersionTLS12,
		VersionMax: tlsvrs.VersionTLS13,
	}
}

// testTLSConfigServer returns a complete TLS configuration for server testing.
//
// This function generates a fresh self-signed certificate and returns a fully
// configured TLS setup suitable for server-side socket testing. The configuration
// includes:
//   - A self-signed ECDSA certificate valid for localhost
//   - All supported elliptic curves
//   - All supported cipher suites
//   - TLS version range: 1.2 to 1.3
//
// This function is designed to be called once in BeforeSuite and the result cached
// in cfgTLSSrv to avoid the performance overhead of generating certificates for
// each test case.
//
// The function uses Gomega assertions and will cause test failure if certificate
// generation fails.
//
// Returns:
//   - A complete certificates.Config ready for server socket testing
//
// Performance note:
//
//	Generating certificates is expensive (typically 10-50ms). Calling this function
//	in BeforeSuite instead of inline in tests can reduce total test time by 50% or more.
func testTLSConfigServer() libtls.Config {
	var (
		err error
		cfg tlscrt.Cert
	)

	// Generate a fresh certificate pair
	genTLSCert, genTLSKey, err = genPairPEM()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(genTLSCert)).To(BeNumerically(">", 0))
	Expect(len(genTLSKey)).To(BeNumerically(">", 0))

	// Parse the certificate pair into a usable format
	cfg, err = tlscrt.ParsePair(genTLSKey, genTLSCert)
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	// Return a complete TLS configuration
	return libtls.Config{
		CurveList:  tlscrv.List(),
		CipherList: tlscpr.List(),
		Certs:      []tlscrt.Certif{cfg.Model()},
		VersionMin: tlsvrs.VersionTLS12,
		VersionMax: tlsvrs.VersionTLS13,
	}
}
