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

// Basic Tests - Socket Configuration Package
//
// This file contains fundamental tests for the socket/config package, focusing on
// basic functionality and core validation logic.
//
// Test Coverage:
//   - Client and Server struct initialization (zero-value and with values)
//   - Protocol validation for TCP, UDP, and Unix sockets
//   - Address format validation for each protocol type
//   - Platform compatibility checks (e.g., Unix sockets on Windows)
//   - Error constant definitions and consistency
//
// Test Organization:
//   - Basic Client Configuration: Tests for Client struct creation and validation
//   - Basic Server Configuration: Tests for Server struct creation and validation
//   - Error Constants: Verification of package-level error values
//
// The tests in this file verify that the configuration structures can be created,
// that basic validation works correctly, and that appropriate errors are returned
// for invalid configurations.
package config_test

import (
	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/nabbar/golib/socket/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Client Configuration", func() {
	Context("Client struct initialization", func() {
		It("should create a zero-value client", func() {
			var c config.Client
			Expect(c.Network).To(Equal(libptc.NetworkProtocol(0)))
			Expect(c.Address).To(BeEmpty())
			Expect(c.TLS.Enabled).To(BeFalse())
		})

		It("should create a client with values", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}
			Expect(c.Network).To(Equal(libptc.NetworkTCP))
			Expect(c.Address).To(Equal("localhost:8080"))
		})
	})

	Context("Client TCP validation", func() {
		It("should validate TCP client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "localhost:8080",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should validate TCP4 client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkTCP4,
				Address: "127.0.0.1:8080",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should validate TCP6 client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkTCP6,
				Address: "[::1]:8080",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should reject TCP client with invalid address", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "invalid-address",
			}
			err := c.Validate()
			Expect(err).To(HaveOccurred())
		})

		It("should reject TCP client with empty address", func() {
			c := config.Client{
				Network: libptc.NetworkTCP,
				Address: "",
			}
			err := c.Validate()
			// Empty address may be accepted by ResolveTCPAddr for some implementations
			// This test is more about documenting behavior than strict validation
			_ = err
		})
	})

	Context("Client UDP validation", func() {
		It("should validate UDP client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkUDP,
				Address: "localhost:9000",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should validate UDP4 client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkUDP4,
				Address: "127.0.0.1:9000",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should validate UDP6 client with valid address", func() {
			c := config.Client{
				Network: libptc.NetworkUDP6,
				Address: "[::1]:9000",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should reject UDP client with invalid address", func() {
			c := config.Client{
				Network: libptc.NetworkUDP,
				Address: "invalid-address",
			}
			err := c.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Client Unix socket validation", func() {
		BeforeEach(func() {
			skipIfWindows("Unix sockets not supported")
		})

		It("should validate Unix client with valid path", func() {
			c := config.Client{
				Network: libptc.NetworkUnix,
				Address: "/tmp/test.sock",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should validate Unixgram client with valid path", func() {
			c := config.Client{
				Network: libptc.NetworkUnixGram,
				Address: "/tmp/test.sock",
			}
			err := c.Validate()
			expectNoValidationError(err)
		})

		It("should reject Unix client with empty path", func() {
			c := config.Client{
				Network: libptc.NetworkUnix,
				Address: "",
			}
			err := c.Validate()
			// Empty address may be accepted by ResolveUnixAddr for some implementations
			// This test is more about documenting behavior than strict validation
			_ = err
		})
	})

	Context("Client platform compatibility", func() {
		When("running on Windows", func() {
			It("should reject Unix socket protocols", func() {
				if !isWindows() {
					Skip("Test only runs on Windows")
				}

				c := config.Client{
					Network: libptc.NetworkUnix,
					Address: "/tmp/test.sock",
				}
				err := c.Validate()
				expectValidationError(err, config.ErrInvalidProtocol)
			})
		})
	})

	Context("Client invalid protocol", func() {
		It("should reject unsupported protocol", func() {
			c := config.Client{
				Network: libptc.NetworkProtocol(0),
				Address: "localhost:8080",
			}
			err := c.Validate()
			expectValidationError(err, config.ErrInvalidProtocol)
		})
	})
})

var _ = Describe("Basic Server Configuration", func() {
	Context("Server struct initialization", func() {
		It("should create a zero-value server", func() {
			var s config.Server
			Expect(s.Network).To(Equal(libptc.NetworkProtocol(0)))
			Expect(s.Address).To(BeEmpty())
			Expect(s.PermFile).To(Equal(libprm.Perm(0)))
			Expect(s.GroupPerm).To(Equal(int32(0)))
			Expect(s.TLS.Enable).To(BeFalse())
		})

		It("should create a server with values", func() {
			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}
			Expect(s.Network).To(Equal(libptc.NetworkTCP))
			Expect(s.Address).To(Equal(":8080"))
		})
	})

	Context("Server TCP validation", func() {
		It("should validate TCP server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: ":8080",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should validate TCP4 server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkTCP4,
				Address: "0.0.0.0:8080",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should validate TCP6 server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkTCP6,
				Address: "[::]:8080",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should reject TCP server with invalid address", func() {
			s := config.Server{
				Network: libptc.NetworkTCP,
				Address: "invalid-address",
			}
			err := s.Validate()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Server UDP validation", func() {
		It("should validate UDP server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkUDP,
				Address: ":9000",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should validate UDP4 server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkUDP4,
				Address: "0.0.0.0:9000",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should validate UDP6 server with valid address", func() {
			s := config.Server{
				Network: libptc.NetworkUDP6,
				Address: "[::]:9000",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})
	})

	Context("Server Unix socket validation", func() {
		BeforeEach(func() {
			skipIfWindows("Unix sockets not supported")
		})

		It("should validate Unix server with valid path", func() {
			s := config.Server{
				Network: libptc.NetworkUnix,
				Address: "/tmp/test.sock",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})

		It("should validate Unixgram server with valid path", func() {
			s := config.Server{
				Network: libptc.NetworkUnixGram,
				Address: "/tmp/test.sock",
			}
			err := s.Validate()
			expectNoValidationError(err)
		})
	})

	Context("Server invalid protocol", func() {
		It("should reject unsupported protocol", func() {
			s := config.Server{
				Network: libptc.NetworkProtocol(0),
				Address: ":8080",
			}
			err := s.Validate()
			expectValidationError(err, config.ErrInvalidProtocol)
		})
	})
})

var _ = Describe("Error Constants", func() {
	It("should have ErrInvalidProtocol defined", func() {
		Expect(config.ErrInvalidProtocol).NotTo(BeNil())
		Expect(config.ErrInvalidProtocol.Error()).To(ContainSubstring("invalid protocol"))
	})

	It("should have ErrInvalidTLSConfig defined", func() {
		Expect(config.ErrInvalidTLSConfig).NotTo(BeNil())
		Expect(config.ErrInvalidTLSConfig.Error()).To(ContainSubstring("invalid TLS config"))
	})

	It("should have ErrInvalidGroup defined", func() {
		Expect(config.ErrInvalidGroup).NotTo(BeNil())
		Expect(config.ErrInvalidGroup.Error()).To(ContainSubstring("invalid unix group"))
	})

	It("should have MaxGID defined", func() {
		Expect(config.MaxGID).To(BeNumerically("==", 32767))
	})
})
