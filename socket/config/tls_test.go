/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package config_test

import (
	libtls "github.com/nabbar/golib/certificates"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/nabbar/golib/socket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TLS Configuration Tests
//
// This file contains comprehensive tests for TLS functionality in socket configurations.
// It validates that TLS settings are properly validated, applied, and retrieved for both
// client and server configurations.
//
// Test coverage includes:
//   - TLS validation for different network protocols (TCP, TCP4, TCP6, UDP, Unix)
//   - Server-side TLS configuration with certificates
//   - Client-side TLS configuration with server name verification
//   - Default TLS configuration merging via DefaultTLS/GetTLS
//   - Error cases for invalid TLS configurations
//
// Performance optimization:
// The test suite pre-generates a server TLS configuration (cfgTLSSrv) in BeforeSuite
// to avoid the overhead of certificate generation in each test case.

var _ = Describe("TLS Configuration", func() {
	Context("Server TLS Configuration", func() {
		It("should validate server with TLS enabled on TCP", func() {
			// Use pre-generated TLS config from BeforeSuite
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8443",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			err := srv.Validate()
			expectNoValidationError(err)
		})

		It("should validate server with TLS enabled on TCP4", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP4,
				Address: "127.0.0.1:8443",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			err := srv.Validate()
			expectNoValidationError(err)
		})

		It("should validate server with TLS enabled on TCP6", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP6,
				Address: "[::1]:8443",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			err := srv.Validate()
			expectNoValidationError(err)
		})

		It("should reject TLS on UDP server", func() {
			srv := config.Server{
				Network: libptc.NetworkUDP,
				Address: ":9000",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			err := srv.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should reject TLS on Unix socket server", func() {
			skipIfWindows("Unix sockets not supported")

			srv := config.Server{
				Network: libptc.NetworkUnix,
				Address: "/tmp/test.sock",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			err := srv.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should reject TLS with nil config", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8443",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = libtls.Config{}

			err := srv.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should accept server without TLS", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}

			srv.TLS.Enabled = false

			err := srv.Validate()
			expectNoValidationError(err)
		})
	})

	Context("Client TLS Configuration", func() {
		It("should validate client with TLS enabled on TCP", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			err := cli.Validate()
			expectNoValidationError(err)
		})

		It("should validate client with TLS enabled on TCP4", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP4,
				Address: "127.0.0.1:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			err := cli.Validate()
			expectNoValidationError(err)
		})

		It("should validate client with TLS enabled on TCP6", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP6,
				Address: "[::1]:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			err := cli.Validate()
			expectNoValidationError(err)
		})

		It("should reject client TLS without server name", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "" // Missing server name

			err := cli.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should reject TLS on UDP client", func() {
			cli := config.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:9000",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			err := cli.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should reject TLS on Unix socket client", func() {
			skipIfWindows("Unix sockets not supported")

			cli := config.Client{
				Network: libptc.NetworkUnix,
				Address: "/tmp/test.sock",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			err := cli.Validate()
			expectValidationError(err, config.ErrInvalidTLSConfig)
		})

		It("should accept client without TLS", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}
			cli.TLS.Enabled = false

			err := cli.Validate()
			expectNoValidationError(err)
		})
	})

	Context("Server DefaultTLS and GetTLS", func() {
		It("should set and retrieve default TLS for server", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8443",
			}

			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			// Set default TLS configuration
			srv.DefaultTLS(testTLSConfigDefault())

			// Retrieve TLS configuration
			enabled, tlsCfg := srv.GetTLS()
			Expect(enabled).To(BeTrue())
			Expect(tlsCfg).ToNot(BeNil())

			// Verify we can create a stdlib TLS config
			stdTLS := tlsCfg.TLS("")
			Expect(stdTLS).ToNot(BeNil())
		})

		It("should return false when server TLS is disabled", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}
			srv.TLS.Enabled = false

			enabled, tlsCfg := srv.GetTLS()
			Expect(enabled).To(BeFalse())
			Expect(tlsCfg).To(BeNil())
		})

		It("should handle DefaultTLS with nil config", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8443",
			}
			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			// Should not panic with nil
			srv.DefaultTLS(nil)
			Succeed()
		})
	})

	Context("Client DefaultTLS and GetTLS", func() {
		It("should set and retrieve default TLS for client", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			// Set default TLS configuration
			cli.DefaultTLS(cfgTLSSrv.New())

			// Retrieve TLS configuration
			enabled, tlsCfg, serverName := cli.GetTLS()
			Expect(enabled).To(BeTrue())
			Expect(tlsCfg).ToNot(BeNil())
			Expect(serverName).To(Equal("localhost"))

			// Verify we can create a stdlib TLS config with server name
			stdTLS := tlsCfg.TLS(serverName)
			Expect(stdTLS).ToNot(BeNil())
			Expect(stdTLS.ServerName).To(Equal(serverName))
		})

		It("should return false when client TLS is disabled", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}
			cli.TLS.Enabled = false

			enabled, tlsCfg, serverName := cli.GetTLS()
			Expect(enabled).To(BeFalse())
			Expect(tlsCfg).To(BeNil())
			Expect(serverName).To(BeEmpty())
		})

		It("should handle DefaultTLS with nil config", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}

			// Should not panic with nil
			cli.DefaultTLS(nil)
			Succeed()
		})
	})

	Context("TLS Concurrent Access", func() {
		It("should handle concurrent GetTLS calls on server", func() {
			srv := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8443",
			}
			srv.TLS.Enabled = true
			srv.TLS.Config = cfgTLSSrv

			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func(s config.Server) {
					defer GinkgoRecover()
					enabled, tlsCfg := s.GetTLS()
					Expect(enabled).To(BeTrue())
					Expect(tlsCfg).ToNot(BeNil())
					done <- true
				}(srv)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent GetTLS calls on client", func() {
			cli := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8443",
			}

			cli.TLS.Enabled = true
			cli.TLS.Config = cfgTLSSrv
			cli.TLS.ServerName = "localhost"

			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func(c config.Client) {
					defer GinkgoRecover()
					enabled, tlsCfg, serverName := c.GetTLS()
					Expect(enabled).To(BeTrue())
					Expect(tlsCfg).ToNot(BeNil())
					Expect(serverName).To(Equal("localhost"))
					done <- true
				}(cli)
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})

		It("should handle concurrent DefaultTLS calls", func() {
			done := make(chan bool)
			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					srv := config.Server{
						Network: libptc.NetworkTCP,
						Address: ":8443",
					}
					srv.DefaultTLS(cfgTLSSrv.New())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})
})
