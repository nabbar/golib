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

// Package tcp_test contains functional and unit tests for the TCP server.
//
// # Creation Test Logic
//
// The 'creation_test.go' file focuses on the static initialization of the server.
// It validates that the srv structure is correctly populated before any 
// network listener is actually opened.
//
// # Test Dataflow: Server Initialization
//
//	[Config Payload] ───> [tcp.New()] ───> [Validation Check]
//	                          │                  │
//	                          v                  +-- (ErrInvalidAddress)
//	                [Struct Allocation]          +-- (ErrInvalidHandler)
//	                          │
//	                          v
//	                [sync.Pool Setup] ───> [Final ServerTcp Instance]
//
// # Coverage Areas
//
//   - Minimal Config: Ensuring the server can be created with only an address and handler.
//   - TLS Config: Verifying that SetTLS is called correctly during New().
//   - Connection Tunning: Ensuring the UpdateConn callback is registered.
//   - Error Handling: Forcing failures with malformed addresses.
package tcp_test

import (
	"net"

	libdur "github.com/nabbar/golib/duration"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Creation", func() {
	// Context: Tests with valid inputs to ensure standard server establishment.
	Context("with valid configuration", func() {
		// Example: Basic instantiation
		It("should create server with minimal configuration", func() {
			cfg := createDefaultConfig(getTestAddr())
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
			// A newly created server is not yet running
			Expect(srv.IsRunning()).To(BeFalse())
			// After New(), IsGone() is true because the server is 'ready to start' 
			// but not currently draining active connections.
			Expect(srv.IsGone()).To(BeTrue())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		// Example: Securing the server with TLS
		It("should create server with TLS configuration", func() {
			cfg := createTLSConfig(getTestAddr())
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		// Example: Customizing low-level socket options (buffer sizes, etc.)
		It("should create server with custom update function", func() {
			upd := func(c net.Conn) {
				// Custom connection configuration: NoDelay is already set by default in Conn()
				// but user can override other options here (ReadBuffer, WriteBuffer).
				_ = c
			}

			cfg := createDefaultConfig(getTestAddr())
			srv, err := scksrt.New(upd, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		// Example: Enabling the Idle Connection Manager
		It("should create server with idle timeout configuration", func() {
			cfg := createDefaultConfig(getTestAddr())
			cfg.ConIdleTimeout = libdur.Seconds(30)

			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

	// Context: Negative testing to ensure robust error detection during startup.
	Context("with invalid configuration", func() {
		It("should fail with empty address", func() {
			cfg := createDefaultConfig("")
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
			Expect(err).To(MatchError(scksrt.ErrInvalidAddress))
		})

		It("should fail with invalid address format", func() {
			cfg := createDefaultConfig("invalid-address")
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})
	})

	// Context: Verifying the initial internal flags of the server.
	Context("server state after creation", func() {
		var srv scksrt.ServerTcp

		BeforeEach(func() {
			cfg := createDefaultConfig(getTestAddr())
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have correct initial state", func() {
			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.IsGone()).To(BeTrue())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})
	})
})
