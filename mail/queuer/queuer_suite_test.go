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
 */

package queuer_test

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
	"sync/atomic"
	"testing"
	"time"

	saslsv "github.com/emersion/go-sasl"
	smtpsv "github.com/emersion/go-smtp"
	libtls "github.com/nabbar/golib/certificates"
	certca "github.com/nabbar/golib/certificates/ca"
	libsmtp "github.com/nabbar/golib/mail/smtp"
	smtpcfg "github.com/nabbar/golib/mail/smtp/config"
	smtptp "github.com/nabbar/golib/mail/smtp/tlsmode"
	libptc "github.com/nabbar/golib/network/protocol"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// Global context for all tests
	testCtx    context.Context
	testCancel context.CancelFunc

	// Test SMTP server
	testSMTPHost = "localhost"
	testSMTPUser = "testuser"
	testSMTPPass = "testpass"

	srvTLS, cliTLS = createTLSConfig()
)

// TestQueuer is the entry point for the Ginkgo test suite
func TestQueuer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mail/Queuer Package Suite")
}

var _ = BeforeSuite(func() {
	testCtx, testCancel = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	if testCancel != nil {
		testCancel()
	}
})

// Helper functions

// getFreePort returns a free TCP port
func getFreePort() int {
	addr, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	lstn, err := net.ListenTCP(libptc.NetworkTCP.Code(), addr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.Addr().(*net.TCPAddr).Port
}

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
	requireAuth bool
	messages    []testMessage
	msgCount    atomic.Int32
	mu          sync.Mutex
}

func (b *testBackend) NewSession(_ *smtpsv.Conn) (smtpsv.Session, error) {
	return &testSession{backend: b}, nil
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
		s.AllowInsecureAuth = true
	}

	return s
}

// startTestSMTPServer starts a test SMTP server and returns server, host, port
func startTestSMTPServer(backend *testBackend, useTLS bool) (*smtpsv.Server, string, int, error) {
	srv := getNewServer(backend, useTLS)

	if useTLS {
		go func() {
			_ = srv.ListenAndServeTLS()
		}()
	} else {
		go func() {
			_ = srv.ListenAndServe()
		}()
	}

	// Wait for server to be ready
	waitForServerRunning(srv.Addr, 5*time.Second)

	// Extract host and port
	if i := strings.Split(srv.Addr, ":"); len(i) != 2 {
		return nil, "", 0, fmt.Errorf("invalid server address: %s", srv.Addr)
	} else if p, e := strconv.Atoi(i[1]); e != nil {
		return nil, "", 0, fmt.Errorf("invalid server address: %s", srv.Addr)
	} else {
		return srv, i[0], p, nil
	}
}

// waitForServerRunning waits for the server to be running
func waitForServerRunning(address string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(testCtx, timeout)
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

func (s *testSession) Auth(mech string) (saslsv.Server, error) {
	return saslsv.NewPlainServer(func(identity, username, password string) error {
		if username != testSMTPUser || password != testSMTPPass {
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
	s.backend.mu.Lock()
	s.backend.messages = append(s.backend.messages, testMessage{
		From: s.from,
		To:   s.to,
		Data: data,
	})
	s.backend.mu.Unlock()
	s.backend.msgCount.Add(1)
	return nil
}

func (s *testSession) Reset() {
	s.from = ""
	s.to = nil
}

func (s *testSession) Logout() error {
	return nil
}

// newTestConfig creates a test SMTP config
func newTestConfig(host string, port int, tlsMode smtptp.TLSMode) smtpcfg.Config {
	dsn := fmt.Sprintf("tcp(%s:%d)/%s", host, port, tlsMode.String())
	model := smtpcfg.ConfigModel{DSN: dsn}
	cfg, err := model.Config()
	Expect(err).ToNot(HaveOccurred())
	return cfg
}

// newTestSMTPClient creates a real test SMTP client
func newTestSMTPClient(host string, port int) libsmtp.SMTP {
	cfg := newTestConfig(host, port, smtptp.TLSNone)
	cli, err := libsmtp.New(cfg, cliTLS.TlsConfig(""))
	Expect(err).ToNot(HaveOccurred())
	Expect(cli).ToNot(BeNil())
	return cli
}

// simpleWriterTo implements io.WriterTo for simple message sending
type simpleWriterTo struct {
	content string
}

func (s *simpleWriterTo) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(s.content))
	return int64(n), err
}

func newSimpleMessage(content string) io.WriterTo {
	return &simpleWriterTo{content: content}
}
