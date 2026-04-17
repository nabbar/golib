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

// Package tcp_test validates the security features of the TCP server.
//
// # TLS Test Logic
//
// The 'tls_test.go' file ensures that the server correctly handles encrypted 
// communication. It focuses on:
//   - TLS Handshake: Verifying that 'tls.NewListener' correctly wraps the TCP socket.
//   - Configuration: Validating that the server rejects incomplete TLS settings.
//   - Handover: Ensuring 'SetTLS(false)' correctly reverts the server to plain TCP mode.
//   - Cipher Suites: Verifying compatibility with the underlying certificates package.
//
// # Dataflow: Secure Connection Establishment
//
//	[tls.Dial] --(ClientHello)--> [net.Listener (TLS Wrapped)]
//	                                      │
//	[Handshake Completed] <──(ServerHello)──+── [srv.getTLS() lookup]
//	          │
//	          v
//	[Encrypted I/O] <──────────────> [sCtx.Read/Write]
//
// # Security Note
//
// In these tests, 'InsecureSkipVerify: true' is used because the certificates 
// are self-signed and generated dynamically in 'helper_test.go'.
package tcp_test

import (
	"context"
	"crypto/tls"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server TLS", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		ctx context.Context
		cnl context.CancelFunc
	)

	// Setup: Initialize address and context before each secure test.
	BeforeEach(func() {
		adr = getTestAddr()
		ctx, cnl = context.WithCancel(globalCtx)
	})

	// Cleanup: Stop the server and release resources.
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		// Allow some time for sockets to enter TIME_WAIT or be fully released.
		time.Sleep(100 * time.Millisecond)
	})

	Context("TLS configuration", func() {
		// Test: Standard secure server creation.
		It("should create server with TLS enabled", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		// Test: Verify the listener starts without error in TLS mode.
		It("should start TLS server successfully", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServer(srv, 2*time.Second)

			Expect(srv.IsRunning()).To(BeTrue())
		})

		// Test: Verify client connection and tracking.
		It("should accept TLS connections", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Create TLS client config with InsecureSkipVerify for testing self-signed certs.
			tlsCfg := &tls.Config{
				InsecureSkipVerify: true, // #nosec nolint
			}

			// Connect with TLS dialer.
			con, err := tls.Dial(libptc.NetworkTCP.Code(), adr, tlsCfg)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = con.Close() }()

			Expect(con).ToNot(BeNil())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		// Test: Verify data integrity over encrypted channel.
		It("should echo messages over TLS", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			tlsCfg := &tls.Config{
				InsecureSkipVerify: true, // #nosec nolint
			}

			con, err := tls.Dial(libptc.NetworkTCP.Code(), adr, tlsCfg)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = con.Close() }()

			msg := []byte("TLS test message")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})

		// Test: Verify dynamic reconfiguration of security settings.
		It("should disable TLS when SetTLS(false) is called", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Explicitly disable TLS.
			err = srv.SetTLS(false, nil)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Should now accept plain TCP connections instead of TLS.
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("plain TCP message")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})
	})

	Context("TLS errors", func() {
		// Test: Validation of TLS configuration inputs.
		It("should reject invalid TLS config with no certificates", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Enabling TLS with an empty config should fail validation.
			invalidTLS := libtls.New()
			err = srv.SetTLS(true, invalidTLS)
			Expect(err).To(HaveOccurred())
		})
	})
})
