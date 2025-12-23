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

package httpserver_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"

	. "github.com/onsi/gomega"
)

var (
	// TLS configuration for server tests (initialized in BeforeSuite)
	srvTLSCfg libtls.Config
	// TLS certificate and key as PEM strings
	genTLSCrt string
	genTLSKey string
)

// initTLSConfigs initializes TLS configurations for testing
func initTLSConfigs() {
	var err error
	genTLSCrt, genTLSKey, err = genCertPair()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(genTLSCrt)).To(BeNumerically(">", 0))
	Expect(len(genTLSKey)).To(BeNumerically(">", 0))

	var crt tlscrt.Cert
	crt, err = tlscrt.ParsePair(genTLSKey, genTLSCrt)
	Expect(err).ToNot(HaveOccurred())
	Expect(crt).ToNot(BeNil())

	srvTLSCfg = libtls.Config{
		CurveList:  tlscrv.List(),
		CipherList: tlscpr.List(),
		Certs:      []tlscrt.Certif{crt.Model()},
		VersionMin: tlsvrs.VersionTLS12,
		VersionMax: tlsvrs.VersionTLS13,
	}
}

// genCertPair generates a self-signed certificate pair for testing
func genCertPair() (pub string, key string, err error) {
	var (
		tpl x509.Certificate
		ser *big.Int
		prv *ecdsa.PrivateKey
		crt []byte
		cbu *bytes.Buffer
		kyd []byte
		kbu *bytes.Buffer
	)

	prv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	ser, err = rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", err
	}

	tpl = x509.Certificate{
		SerialNumber: ser,
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

	crt, err = x509.CreateCertificate(rand.Reader, &tpl, &tpl, &prv.PublicKey, prv)
	if err != nil {
		return "", "", err
	}

	cbu = bytes.NewBufferString("")
	if err = pem.Encode(cbu, &pem.Block{Type: "CERTIFICATE", Bytes: crt}); err != nil {
		return "", "", err
	}

	kyd, err = x509.MarshalECPrivateKey(prv)
	if err != nil {
		return "", "", err
	}

	kbu = bytes.NewBufferString("")
	if err = pem.Encode(kbu, &pem.Block{Type: "EC PRIVATE KEY", Bytes: kyd}); err != nil {
		return "", "", err
	}

	return cbu.String(), kbu.String(), nil
}

// GetFreePort returns a free TCP port for testing
func GetFreePort() int {
	adr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	lis, err := net.ListenTCP("tcp", adr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lis.Close()
	}()

	return lis.Addr().(*net.TCPAddr).Port
}

// Mock HTTP handler for testing
type mockHandler struct {
	called bool
	status int
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	if m.status == 0 {
		m.status = http.StatusOK
	}
	w.WriteHeader(m.status)
	_, _ = w.Write([]byte("mock response"))
}
