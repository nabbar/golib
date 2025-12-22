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

// helper_test.go provides shared test utilities and helper functions.
// Includes server configuration creation, connection helpers, test address
// management, and common handler implementations used across all test files.
package tcp_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync/atomic"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	tlscrt "github.com/nabbar/golib/certificates/certs"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	tcp "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
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

// getFreePort returns a free TCP port for testing
func getFreePort() int {
	adr, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), "localhost:0")
	Expect(err).ToNot(HaveOccurred())

	lis, err := net.ListenTCP(libptc.NetworkTCP.Code(), adr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lis.Close()
	}()

	return lis.Addr().(*net.TCPAddr).Port
}

// getTestAddr returns a test address with a free port
func getTestAddr() string {
	return fmt.Sprintf("localhost:%d", getFreePort())
}

// echoHandler is a simple echo handler for testing
func echoHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

	buf := make([]byte, 1024)
	for {
		n, err := c.Read(buf)
		if err != nil {
			return
		}

		if n > 0 {
			_, err = c.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}
}

// counterHandler counts connections
func counterHandler(cnt *atomic.Int32) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		cnt.Add(1)

		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				_, err = c.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}
}

// slowHandler adds delay before echoing
func slowHandler(dly time.Duration) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()

		time.Sleep(dly)
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				_, err = c.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}
}

// closeHandler immediately closes connections
func closeHandler(c libsck.Context) {
	_ = c.Close()
}

// writeOnlyHandler writes but doesn't read
func writeOnlyHandler(msg string) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		_, _ = c.Write([]byte(msg))
	}
}

// readOnlyHandler reads but doesn't write
func readOnlyHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

	buf := make([]byte, 1024)
	for {
		_, err := c.Read(buf)
		if err != nil {
			return
		}
	}
}

// createDefaultConfig creates a default server configuration
func createDefaultConfig(addr string) sckcfg.Server {
	return sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: addr,
	}
}

// createTLSConfig creates a TLS-enabled server configuration
func createTLSConfig(addr string) sckcfg.Server {
	cfg := createDefaultConfig(addr)
	cfg.TLS.Enabled = true
	cfg.TLS.Config = srvTLSCfg
	return cfg
}

// waitForServer waits for the server to be running
func waitForServer(srv tcp.ServerTcp, timeout time.Duration) {
	Eventually(func() bool {
		return srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForServerStopped waits for the server to stop
func waitForServerStopped(srv tcp.ServerTcp, timeout time.Duration) {
	Eventually(func() bool {
		return !srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForConnections waits for specific connection count
func waitForConnections(srv tcp.ServerTcp, exp int64, timeout time.Duration) {
	Eventually(func() int64 {
		return srv.OpenConnections()
	}, timeout, 10*time.Millisecond).Should(Equal(exp))
}

// waitForGone waits for server to be fully gone
func waitForGone(srv tcp.ServerTcp, timeout time.Duration) {
	Eventually(func() bool {
		return srv.IsGone()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// connectToServer establishes a TCP connection to the server
func connectToServer(addr string) net.Conn {
	con, err := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 2*time.Second)
	Expect(err).ToNot(HaveOccurred())
	Expect(con).ToNot(BeNil())
	return con
}

// sendAndReceive sends data and receives response
func sendAndReceive(con net.Conn, data []byte) []byte {
	n, err := con.Write(data)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	buf := make([]byte, len(data))
	n, err = io.ReadFull(con, buf)
	Expect(err).ToNot(HaveOccurred())
	Expect(n).To(Equal(len(data)))

	return buf
}

// startServerInBackground starts a server in a goroutine
func startServerInBackground(c context.Context, srv tcp.ServerTcp) {
	go func() {
		_ = srv.Listen(c)
	}()
}

// waitForServerAcceptingConnections waits for server to accept connections
func waitForServerAcceptingConnections(addr string, timeout time.Duration) {
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	tck := time.NewTicker(50 * time.Millisecond)
	defer tck.Stop()

	for {
		select {
		case <-tmr.C:
			Fail(fmt.Sprintf("Timeout waiting for server to accept connections at %s after %v", addr, timeout))
			return
		case <-tck.C:
			if c, e := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}
