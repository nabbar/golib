/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package smtp_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	saslsv "github.com/emersion/go-sasl"
	smtpsv "github.com/emersion/go-smtp"
	libtls "github.com/nabbar/golib/certificates"
	certca "github.com/nabbar/golib/certificates/ca"

	. "github.com/onsi/ginkgo/v2"
)

var (
	testSMTPHost     = "localhost"
	testSMTPPort     int
	testSMTPUser     = "testuser"
	testSMTPPassword = "testpass"

	srvTLS, cliTLS = createTLSConfig()
)

// createTLSConfig creates a TLS configuration for testing
func createTLSConfig() (serverConfig, clientConfig libtls.TLSConfig) {
	certPEM, keyPEM := generateSelfSignedCert()

	// Server config
	serverConfig = libtls.New()
	err := serverConfig.AddCertificatePairString(string(keyPEM), string(certPEM))
	if err != nil {
		panic(err)
	}

	// Client config with server cert as CA
	ca, err := certca.Parse(string(certPEM))
	if err != nil {
		panic(err)
	}

	clientConfig = libtls.New()
	if !clientConfig.AddRootCA(ca) {
		panic("failed to add root CA")
	}

	return
}

// generateSelfSignedCert generates a self-signed certificate for testing
func generateSelfSignedCert() (certPEM, keyPEM []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Co"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return
}

// testBackend implements smtpsv.Backend
type testBackend struct {
	sync.Mutex
	requireAuth bool
	failAuth    bool
	messages    []testMessage
}

func (b *testBackend) NewSession(_ *smtpsv.Conn) (smtpsv.Session, error) {
	return &testSession{backend: b}, nil
}

func getServerHostPort(s *smtpsv.Server) (string, int, error) {
	if i := strings.Split(s.Addr, ":"); len(i) != 2 {
		return "", 0, fmt.Errorf("invalid server address: %s", s.Addr)
	} else if p, e := strconv.Atoi(i[1]); e != nil {
		return "", 0, fmt.Errorf("invalid server address: %s", s.Addr)
	} else {
		return i[0], p, nil
	}
}

func getNewServer(backend *testBackend, useTLS bool) *smtpsv.Server {
	s := smtpsv.NewServer(backend)
	s.Addr = fmt.Sprintf("localhost:%d", getFreePort())
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50

	if useTLS {
		s.TLSConfig = srvTLS.TlsConfig("")
		s.AllowInsecureAuth = false
	} else {
		// Allow PLAIN auth over non-TLS for testing
		s.AllowInsecureAuth = true
	}

	return s
}

// startTestSMTPServer starts a test SMTP server
func startTestSMTPServer(backend *testBackend, useTLS bool) (*smtpsv.Server, error) {
	var srv = getNewServer(backend, useTLS)

	if useTLS {
		go func() {
			// Ignore error - it's normal when server.Close() is called
			_ = srv.ListenAndServeTLS()
		}()
	} else {
		go func() {
			// Ignore error - it's normal when server.Close() is called
			_ = srv.ListenAndServe()
		}()
	}

	// Wait for server to be ready
	waitForServerRunning(srv.Addr, 5*time.Second)

	return srv, nil
}

// waitForServerRunning waits for the server to be running by attempting to connect
func waitForServerRunning(address string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(globalCtx, timeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			Fail(fmt.Sprintf("Timeout waiting for server to start at %s after %v", address, timeout))
			return
		case <-ticker.C:
			if c, e := net.DialTimeout("tcp", address, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}

type testMessage struct {
	From string
	To   []string
	Data []byte
}

type testSession struct {
	backend       *testBackend
	from          string
	to            []string
	authenticated bool
}

func (s *testSession) AuthMechanisms() []string {
	return []string{saslsv.Plain}
}

// Auth is the handler for supported authenticators.
func (s *testSession) Auth(mech string) (saslsv.Server, error) {
	return saslsv.NewPlainServer(func(identity, username, password string) error {
		if username != testSMTPUser || password != testSMTPPassword {
			return fmt.Errorf("invalid credentials")
		}
		s.authenticated = true
		return nil
	}), nil
}

func (s *testSession) Mail(from string, _ *smtpsv.MailOptions) error {
	if s.backend.requireAuth && !s.authenticated {
		return fmt.Errorf("authentication required")
	}
	if strings.Contains(from, "\n") || strings.Contains(from, "\r") {
		return fmt.Errorf("invalid from address")
	}
	s.from = from
	return nil
}

func (s *testSession) Rcpt(to string, _ *smtpsv.RcptOptions) error {
	if strings.Contains(to, "\n") || strings.Contains(to, "\r") {
		return fmt.Errorf("invalid to address")
	}
	s.to = append(s.to, to)
	return nil
}

func (s *testSession) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	s.backend.Lock()
	defer s.backend.Unlock()
	s.backend.messages = append(s.backend.messages, testMessage{
		From: s.from,
		To:   s.to,
		Data: data,
	})
	return nil
}

func (s *testSession) Reset() {
	s.from = ""
	s.to = nil
	// Don't reset authenticated - it persists for the connection
}

func (s *testSession) Logout() error {
	return nil
}
