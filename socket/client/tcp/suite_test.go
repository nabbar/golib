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
	certca "github.com/nabbar/golib/certificates/ca"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckclt "github.com/nabbar/golib/socket/client/tcp"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	// Global test context
	globalCtx      context.Context
	globalCnl      context.CancelFunc
	srvTLS, cliTLS = createTLSConfig()
)

func TestSocketClientTCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Client TCP Suite")
}

var _ = BeforeSuite(func() {
	globalCtx, globalCnl = context.WithCancel(context.Background())
})

var _ = AfterSuite(func() {
	if globalCnl != nil {
		globalCnl()
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
	return fmt.Sprintf("localhost:%d", getFreePort())
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

// silentHandler reads but doesn't write back
func silentHandler(request libsck.Reader, response libsck.Writer) {
	defer func() {
		_ = request.Close()
		_ = response.Close()
	}()
	buf := make([]byte, 1024)
	for {
		_, err := request.Read(buf)
		if err != nil {
			return
		}
	}
}

// closingHandler immediately closes the connection
func closingHandler(request libsck.Reader, response libsck.Writer) {
	_ = request.Close()
	_ = response.Close()
}

// createServer creates a new TCP server with the given handler
func createServer(handler libsck.HandlerFunc) scksrt.ServerTcp {
	srv := scksrt.New(nil, handler)
	Expect(srv).ToNot(BeNil())
	return srv
}

// createAndRegisterServer creates and registers a new TCP server
func createAndRegisterServer(address string, handler libsck.HandlerFunc) scksrt.ServerTcp {
	srv := createServer(handler)
	err := srv.RegisterServer(address)
	Expect(err).ToNot(HaveOccurred())
	return srv
}

// startServer starts a server in a goroutine
func startServer(ctx context.Context, srv scksrt.ServerTcp) {
	go func() {
		_ = srv.Listen(ctx)
	}()
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

// createClient creates a new TCP client
func createClient(address string) sckclt.ClientTCP {
	cli, err := sckclt.New(address)
	Expect(err).ToNot(HaveOccurred())
	Expect(cli).ToNot(BeNil())
	return cli
}

// connectClient connects a client to the server
func connectClient(ctx context.Context, cli sckclt.ClientTCP) {
	err := cli.Connect(ctx)
	Expect(err).ToNot(HaveOccurred())
}

// waitForClientConnected waits for the client to be connected
func waitForClientConnected(cli sckclt.ClientTCP, timeout time.Duration) {
	Eventually(func() bool {
		return cli.IsConnected()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// sendAndReceive sends data and receives response
func sendAndReceive(cli sckclt.ClientTCP, data []byte) []byte {
	n, err := cli.Write(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	response := make([]byte, len(data))
	n, err = io.ReadFull(cli, response)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	return response
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

// createTLSServer creates a TLS-enabled server
func createTLSServer(address string, handler libsck.HandlerFunc) scksrt.ServerTcp {
	srv := createAndRegisterServer(address, handler)
	err := srv.SetTLS(true, srvTLS)
	Expect(err).ToNot(HaveOccurred())

	return srv
}

// createTLSClient creates a TLS-enabled client
func createTLSClient(address string) sckclt.ClientTCP {
	cli := createClient(address)
	err := cli.SetTLS(true, cliTLS, "")
	Expect(err).ToNot(HaveOccurred())

	return cli
}

// createSimpleTestServer creates and starts a simple echo server for testing
func createSimpleTestServer(ctx context.Context, address string) scksrt.ServerTcp {
	srv := createAndRegisterServer(address, echoHandler)
	startServer(ctx, srv)
	waitForServerRunning(address, 2*time.Second)
	return srv
}

// waitForConnections waits for the server to have the expected number of connections
func waitForConnections(srv scksrt.ServerTcp, expected int64, timeout time.Duration) {
	Eventually(func() int64 {
		return srv.OpenConnections()
	}, timeout, 10*time.Millisecond).Should(Equal(expected))
}
