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

// creation_test.go tests server initialization and configuration validation.
// Verifies proper server instance creation, configuration parameter handling,
// and error conditions during the initialization phase.
package tcp_test

import (
	"net"

	libdur "github.com/nabbar/golib/duration"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Creation", func() {
	Context("with valid configuration", func() {
		It("should create server with minimal configuration", func() {
			cfg := createDefaultConfig(getTestAddr())
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.IsGone()).To(BeTrue())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		It("should create server with TLS configuration", func() {
			cfg := createTLSConfig(getTestAddr())
			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should create server with custom update function", func() {
			upd := func(c net.Conn) {
				// Custom connection configuration
				_ = c
			}

			cfg := createDefaultConfig(getTestAddr())
			srv, err := scksrt.New(upd, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should create server with idle timeout configuration", func() {
			cfg := createDefaultConfig(getTestAddr())
			cfg.ConIdleTimeout = libdur.Seconds(30)

			srv, err := scksrt.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

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
