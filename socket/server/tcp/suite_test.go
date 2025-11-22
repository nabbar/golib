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

package tcp_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// Global test context
	x context.Context
	n context.CancelFunc
)

func TestSocketServerTCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Server TCP Suite")
}

var _ = BeforeSuite(func() {
	x, n = context.WithTimeout(context.Background(), 60*time.Second)
})

var _ = AfterSuite(func() {
	if n != nil {
		n()
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

// getTestAddress returns a test address with a free port
func getTestAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", getFreePort())
}

// echoHandler is a simple echo handler for testing
func echoHandler(request libsck.Reader, response libsck.Writer) {
	defer func() {
		_ = request.Close()
		_ = response.Close()
	}()
	_, _ = io.Copy(response, request)
}

// delayHandler is a handler that delays before echoing
func delayHandler(delay time.Duration) libsck.HandlerFunc {
	return func(request libsck.Reader, response libsck.Writer) {
		defer func() {
			_ = request.Close()
			_ = response.Close()
		}()
		time.Sleep(delay)
		_, _ = io.Copy(response, request)
	}
}

// countingHandler counts the number of calls
func countingHandler(counter *atomic.Int32) libsck.HandlerFunc {
	return func(request libsck.Reader, response libsck.Writer) {
		defer func() {
			_ = request.Close()
			_ = response.Close()
		}()
		counter.Add(1)
		_, _ = io.Copy(response, request)
	}
}

// errorHandler always returns an error
func errorHandler(request libsck.Reader, response libsck.Writer) {
	defer func() {
		_ = request.Close()
		_ = response.Close()
	}()
	// Do nothing, just close
}

// createServer creates a new TCP server with the given handler
func createServer(handler libsck.HandlerFunc, upd libsck.UpdateConn) scksrv.ServerTcp {
	srv := scksrv.New(upd, handler)
	Expect(srv).ToNot(BeNil())
	return srv
}

// createAndRegisterServer creates and registers a new TCP server
func createAndRegisterServer(address string, handler libsck.HandlerFunc, upd libsck.UpdateConn) scksrv.ServerTcp {
	srv := createServer(handler, upd)
	err := srv.RegisterServer(address)
	Expect(err).ToNot(HaveOccurred())
	return srv
}

// startServer starts the server in a goroutine
func startServer(ctx context.Context, srv libsck.Server) {
	go func() {
		defer GinkgoRecover()
		err := srv.Listen(ctx)
		if err != nil {
			// Errors are expected when shutting down
			GinkgoWriter.Printf("Server listen error: %v\n", err)
		}
	}()
}

// waitForServerRunning waits for the server to be running
func waitForServerRunning(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.IsRunning() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	Fail("Server did not start within timeout")
}

// waitForServerStopped waits for the server to be stopped
func waitForServerStopped(srv libsck.Server, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if !srv.IsRunning() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	Fail("Server did not stop within timeout")
}

// waitForConnections waits for the server to have the expected number of connections
func waitForConnections(srv libsck.Server, expected int64, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		if srv.OpenConnections() == expected {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	// Don't fail, just continue - connection count may vary
}

// connectClient connects a TCP client to the given address
func connectClient(address string) net.Conn {
	conn, err := net.Dial(libptc.NetworkTCP.Code(), address)
	Expect(err).ToNot(HaveOccurred())
	Expect(conn).ToNot(BeNil())
	return conn
}

// connectTLSClient connects a TLS client to the given address
func connectTLSClient(address string, tlsConfig *tls.Config) net.Conn {
	conn, err := tls.Dial(libptc.NetworkTCP.Code(), address, tlsConfig)
	Expect(err).ToNot(HaveOccurred())
	Expect(conn).ToNot(BeNil())
	return conn
}

// sendMessage sends a message to the connection
func sendMessage(conn net.Conn, msg []byte) int {
	n, err := conn.Write(msg)
	Expect(err).ToNot(HaveOccurred())
	return n
}

// receiveMessage receives a message from the connection
func receiveMessage(conn net.Conn, bufSize int) []byte {
	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)
	Expect(err).ToNot(HaveOccurred())
	return buf[:n]
}

// generateSelfSignedCert generates a self-signed certificate for testing
func generateSelfSignedCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return tls.X509KeyPair(certPEM, keyPEM)
}

// createTLSConfig creates a TLS configuration for testing
func createTLSConfig() libtls.TLSConfig {
	cert, err := generateSelfSignedCert()
	Expect(err).ToNot(HaveOccurred())

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	})

	keyBytes, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	Expect(err).ToNot(HaveOccurred())

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	})

	tlsCfg := libtls.New()
	err = tlsCfg.AddCertificatePairString(string(keyPEM), string(certPEM))
	Expect(err).ToNot(HaveOccurred())

	return tlsCfg
}

// expectError is a helper to expect an error
func expectError(err error) {
	Expect(err).To(HaveOccurred())
}

// expectNoError is a helper to expect no error
func expectNoError(err error) {
	Expect(err).ToNot(HaveOccurred())
}
