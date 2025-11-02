/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package testhelpers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// TempCertPair represents a temporary certificate pair for testing
type TempCertPair struct {
	CertFile string
	KeyFile  string
	TempDir  string
}

// GenerateTempCert generates a temporary self-signed certificate for testing
func GenerateTempCert() (*TempCertPair, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "httpserver-test-certs-*")
	if err != nil {
		return nil, err
	}

	// Generate private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
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
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	// Write certificate file
	certFile := filepath.Join(tempDir, "cert.pem")

	rt, err := os.OpenRoot(tempDir)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	defer func() {
		_ = rt.Close()
	}()

	certOut, err := rt.Create(filepath.Base(certFile))
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	defer func() {
		_ = certOut.Close()
	}()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	// Write private key file
	keyFile := filepath.Join(tempDir, "key.pem")

	keyOut, err := rt.Create(filepath.Base(keyFile))

	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	defer func() {
		_ = keyOut.Close()
	}()

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, err
	}

	return &TempCertPair{
		CertFile: certFile,
		KeyFile:  keyFile,
		TempDir:  tempDir,
	}, nil
}

// Cleanup removes the temporary certificate files and directory
func (t *TempCertPair) Cleanup() error {
	if t.TempDir != "" {
		return os.RemoveAll(t.TempDir)
	}
	return nil
}
