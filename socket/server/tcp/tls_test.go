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

// tls_test.go validates TLS/SSL functionality of the TCP server.
// Tests include TLS handshake, encrypted communication, certificate validation,
// TLS configuration management, and secure connection lifecycle.
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

	BeforeEach(func() {
		adr = getTestAddr()
		ctx, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Context("TLS configuration", func() {
		It("should create server with TLS enabled", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should start TLS server successfully", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServer(srv, 2*time.Second)

			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should accept TLS connections", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Create TLS client config with InsecureSkipVerify for testing
			tlsCfg := &tls.Config{
				InsecureSkipVerify: true, // #nosec nolint
			}

			// Connect with TLS
			con, err := tls.Dial(libptc.NetworkTCP.Code(), adr, tlsCfg)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = con.Close() }()

			Expect(con).ToNot(BeNil())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		It("should echo messages over TLS", func() {
			cfg := createTLSConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Create TLS client config with InsecureSkipVerify for testing
			tlsCfg := &tls.Config{
				InsecureSkipVerify: true, // #nosec nolint
			}

			// Connect with TLS
			con, err := tls.Dial(libptc.NetworkTCP.Code(), adr, tlsCfg)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = con.Close() }()

			msg := []byte("TLS test message")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})

		It("should disable TLS when SetTLS(false) is called", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = srv.SetTLS(false, nil)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(ctx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Should accept plain TCP connections
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("plain TCP message")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})
	})

	Context("TLS errors", func() {
		It("should reject invalid TLS config with no certificates", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Try to enable TLS with invalid config
			invalidTLS := libtls.New()
			err = srv.SetTLS(true, invalidTLS)
			Expect(err).To(HaveOccurred())
		})
	})
})
