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

// Package tcp_test provides functional and unit testing utilities for the TCP server.
//
// # Helper Test Utilities
//
// The 'helper_test.go' file is the backbone of the test suite. It provides:
//   - TLS Certificate Generation: Functions to create self-signed certificates on-the-fly.
//   - Network Helpers: Utilities to find free ports and manage test addresses.
//   - Mock Handlers: A variety of server handlers (Echo, Counter, Slow, WriteOnly, etc.).
//   - Waiters & Synchronizers: Functions to ensure the server is ready before running tests.
//
// # Design Pattern: Memory Pooling in Tests
//
// Even in tests, we use a 'sync.Pool' (bufPool) for I/O buffers. This reflects the 
// production server's behavior and ensures benchmarks are measuring the server's 
// performance rather than the test's memory allocations.
//
// # Dataflow: Self-Signed Certificate Generation
//
//	[ecdsa.GenerateKey] ───> [Private Key]
//	                             │
//	[x509.Certificate] ───+──────v
//	 (Template)           │  [x509.CreateCertificate] ───> [Raw DER]
//	                      │                                   │
//	[PEM Encoder] <───────+───────────────────────────────────+
//	      │
//	      v
//	[srvTLSCfg (PEM string)]
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
	"sync"
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

const (
	// bufferSize defines the standard buffer size for test handlers (256KB).
	bufferSize = 256 * 1024

	// Network-related constants
	addrLocalhost = "localhost"
	addrFreePort  = "localhost:0"
	addrIPLocal   = "127.0.0.1"

	// Certificate-related constants
	certOrg     = "Test Organization"
	pemCertType = "CERTIFICATE"
	pemKeyType  = "EC PRIVATE KEY"
	emptyString = ""

	// Common error messages for failed assertions
	errGenCert      = "failed to gen cert pair: %v"
	errParsePair    = "failed to parse pair: %v"
	errResolveAddr  = "failed to resolve addr: %v"
	errListen       = "failed to listen: %v"
	errConnect      = "failed to connect to server: %v"
	errConnectTLS   = "failed to connect to TLS server: %v"
	errWriteData    = "failed to write data: %v"
	errReadData     = "failed to read data: %v"
	errCreateServer = "failed to create server: %v"
	errStartServer  = "server failed to start"
	errWaitAccept   = "Timeout waiting for server to accept connections at %s after %v"
	errWriteLen     = "wrote %d bytes, expected %d"
	errReadLen      = "read %d bytes, expected %d"

	msgEchoLatency = "benchmark message for echo latency"
	msgConcurrent  = "concurrent client message"
)

var (
	// srvTLSCfg is the global TLS configuration used for secure server tests.
	srvTLSCfg libtls.Config
	// genTLSCrt stores the generated PEM certificate.
	genTLSCrt string
	// genTLSKey stores the generated PEM private key.
	genTLSKey string

	// bufPool is a sync.Pool for managing test-side I/O buffers.
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, bufferSize)
		},
	}
)

// initTLSConfigs initializes the global srvTLSCfg by generating a fresh 
// self-signed certificate. This is typically called in BeforeSuite.
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

// genCertPair is a helper that creates an ECDSA P-256 private key and 
// a corresponding self-signed X.509 certificate for 'localhost'.
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

	// Generate ECDSA key pair
	prv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return emptyString, emptyString, err
	}

	// Generate a unique serial number
	ser, err = rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return emptyString, emptyString, err
	}

	// Prepare certificate template
	tpl = x509.Certificate{
		SerialNumber: ser,
		Subject: pkix.Name{
			Organization: []string{certOrg},
			CommonName:   addrLocalhost,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{addrLocalhost, addrIPLocal},
	}

	// Create self-signed certificate
	crt, err = x509.CreateCertificate(rand.Reader, &tpl, &tpl, &prv.PublicKey, prv)
	if err != nil {
		return emptyString, emptyString, err
	}

	// PEM Encode Certificate
	cbu = bytes.NewBufferString(emptyString)
	if err = pem.Encode(cbu, &pem.Block{Type: pemCertType, Bytes: crt}); err != nil {
		return emptyString, emptyString, err
	}

	// PEM Encode Private Key
	kyd, err = x509.MarshalECPrivateKey(prv)
	if err != nil {
		return emptyString, emptyString, err
	}

	kbu = bytes.NewBufferString(emptyString)
	if err = pem.Encode(kbu, &pem.Block{Type: pemKeyType, Bytes: kyd}); err != nil {
		return emptyString, emptyString, err
	}

	return cbu.String(), kbu.String(), nil
}

// getFreePort attempts to bind a listener to port 0 to let the OS 
// assign an available TCP port. Returns the assigned port.
func getFreePort() int {
	adr, err := net.ResolveTCPAddr(libptc.NetworkTCP.Code(), addrFreePort)
	Expect(err).ToNot(HaveOccurred())

	lis, err := net.ListenTCP(libptc.NetworkTCP.Code(), adr)
	Expect(err).ToNot(HaveOccurred())

	defer func() {
		_ = lis.Close()
	}()

	return lis.Addr().(*net.TCPAddr).Port
}

// getTestAddr returns a full loopback address (localhost:port) with 
// a guaranteed free port.
func getTestAddr() string {
	return fmt.Sprintf("%s:%d", addrLocalhost, getFreePort())
}

// echoHandler is a high-performance mock handler that reads from the 
// socket and writes back the exact same data. It uses the global bufPool.
func echoHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

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

// echoHandlerBench is an alias to echoHandler specifically for benchmarks.
func echoHandlerBench(c libsck.Context) {
	echoHandler(c)
}

// counterHandler is a mock handler that increments an atomic counter 
// for every connection it processes.
func counterHandler(cnt *atomic.Int32) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		cnt.Add(1)

		buf := bufPool.Get().([]byte)
		defer bufPool.Put(buf)

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

// slowHandler simulates a connection with high latency by sleeping 
// before starting the echo loop.
func slowHandler(dly time.Duration) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()

		time.Sleep(dly)
		buf := bufPool.Get().([]byte)
		defer bufPool.Put(buf)

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

// closeHandler is a minimal handler that drops the connection immediately.
func closeHandler(c libsck.Context) {
	_ = c.Close()
}

// writeOnlyHandler writes a static message to the client and then closes 
// the connection without reading any input.
func writeOnlyHandler(msg string) libsck.HandlerFunc {
	return func(c libsck.Context) {
		defer func() {
			_ = c.Close()
		}()
		_, _ = c.Write([]byte(msg))
	}
}

// readOnlyHandler consumes all data from the client until EOF but 
// never writes anything back.
func readOnlyHandler(c libsck.Context) {
	defer func() {
		_ = c.Close()
	}()

	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)

	for {
		_, err := c.Read(buf)
		if err != nil {
			return
		}
	}
}

// createDefaultConfig creates a Server structure with TCP protocol 
// and the specified address.
func createDefaultConfig(addr string) sckcfg.Server {
	return sckcfg.Server{
		Network: libptc.NetworkTCP,
		Address: addr,
	}
}

// createTLSConfig creates a Server structure with TLS enabled and 
// uses the global srvTLSCfg generated in initTLSConfigs.
func createTLSConfig(addr string) sckcfg.Server {
	cfg := createDefaultConfig(addr)
	cfg.TLS.Enabled = true
	cfg.TLS.Config = srvTLSCfg
	return cfg
}

// waitForServer uses Gomega's Eventually to poll the server's IsRunning flag.
func waitForServer(srv tcp.ServerTcp, timeout time.Duration) {
	Eventually(func() bool {
		return srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// waitForServerStopped waits for the server's listener loop to exit.
func waitForServerStopped(srv tcp.ServerTcp, timeout time.Duration) {
	Eventually(func() bool {
		return !srv.IsRunning()
	}, timeout, 10*time.Millisecond).Should(BeTrue())
}

// connectToServer dials the server with a 2-second timeout and 
// returns the net.Conn or fails the test.
func connectToServer(addr string) net.Conn {
	con, err := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 2*time.Second)
	Expect(err).ToNot(HaveOccurred())
	Expect(con).ToNot(BeNil())
	return con
}

// sendAndReceive is a blocking helper that writes 'data' to the connection 
// and waits until 'len(data)' bytes are read back.
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

// startServerInBackground triggers srv.Listen() in a separate goroutine.
func startServerInBackground(c context.Context, srv tcp.ServerTcp) {
	go func() {
		_ = srv.Listen(c)
	}()
}

// waitForServerAcceptingConnections repeatedly dials the address until 
// a connection is established or the timeout expires.
func waitForServerAcceptingConnections(addr string, timeout time.Duration) {
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	tck := time.NewTicker(50 * time.Millisecond)
	defer tck.Stop()

	for {
		select {
		case <-tmr.C:
			Fail(fmt.Sprintf(errWaitAccept, addr, timeout))
			return
		case <-tck.C:
			if c, e := net.DialTimeout(libptc.NetworkTCP.Code(), addr, 100*time.Millisecond); e == nil {
				_ = c.Close()
				return
			}
		}
	}
}
