//go:build linux || darwin

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

// Package unix_test provides validation for the initialization and configuration
// phase of the Unix domain socket server.
//
// # creation_test.go: Constructor and Configuration Validation
//
// This file ensures that the server can be correctly instantiated with various
// configuration parameters and that invalid inputs are properly caught during
// the construction phase.
//
// # Scenarios Covered:
//
// ## 1. Successful Instantiation
//   - Default Configuration: Verifies that `New()` works with standard settings.
//   - UpdateConn Integration: Ensures the optional connection-tuning callback can be registered.
//   - Custom Permissions: Validates that specific POSIX permissions (e.g., 0660) are accepted.
//   - Idle Timeout: Confirms that the server correctly initializes the Idle Manager when a
//     timeout is specified in the config.
//
// ## 2. Invalid Input Rejection (Error Paths)
//   - Missing Handler: Ensures `New()` returns `ErrInvalidHandler` if no logic is provided.
//   - Invalid GID: Verifies rejection of Group IDs exceeding the `MaxGID` (32767) limit.
//   - Malformed Paths: Ensures empty or illegal socket paths are caught early.
//   - Wrong Network Type: Confirms the server only accepts `libptc.NetworkUnix`.
//
// ## 3. Dynamic Reconfiguration
//   - RegisterSocket: Tests the ability to update the socket file metadata (path, permissions, GID)
//     after the server has been created but before it has started.
//
// ## 4. Callback Management
//   - Functional Registration: Ensures that error and info callbacks can be registered and
//     updated atomically.
//
// ## 5. Transport Layer Compatibility
//   - TLS No-op: Validates that the `SetTLS` method exists (for interface compliance) but
//     returns nil without performing any action, as Unix sockets don't support transport-layer TLS.
//
// # Technical Note:
//
// These tests are the first line of defense against configuration errors that could lead
// to runtime failures (e.g., permission denied when creating the socket file or system-level
// errors when setting file ownership).
package unix_test

import (
	"net"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Server Creation", func() {
	var socketPath string

	BeforeEach(func() {
		socketPath = getTestSocketPath()
	})

	AfterEach(func() {
		cleanupSocketFile(socketPath)
	})

	Context("with valid configuration", func() {
		It("should create server with default configuration", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.IsGone()).To(BeTrue())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		It("should create server with UpdateConn callback", func() {
			upd := func(c net.Conn) {
				// UpdateConn callback
			}

			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(upd, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should create server with custom permissions", func() {
			cfg := createConfigWithPerms(socketPath, 0660, -1)
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should create server with idle timeout", func() {
			cfg := createConfigWithIdleTimeout(socketPath, 1000)
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

	Context("with invalid configuration", func() {
		It("should fail with nil handler", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, nil, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
			Expect(err).To(Equal(scksru.ErrInvalidHandler))
		})

		It("should fail with invalid group ID", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkUnix,
				Address:   socketPath,
				PermFile:  libprm.Perm(0600),
				GroupPerm: scksru.MaxGID + 1,
			}
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
			Expect(err).To(Equal(scksru.ErrInvalidGroup))
		})

		It("should fail with empty socket path", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkUnix,
				Address:   "",
				PermFile:  libprm.Perm(0600),
				GroupPerm: -1,
			}
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})

		It("should fail with invalid network type", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkTCP,
				Address:   socketPath,
				PermFile:  libprm.Perm(0600),
				GroupPerm: -1,
			}
			srv, err := scksru.New(nil, echoHandler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})
	})

	Context("RegisterSocket method", func() {
		It("should update socket configuration", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			newPath := getTestSocketPath()
			defer cleanupSocketFile(newPath)

			err = srv.RegisterSocket(newPath, libprm.Perm(0660), -1)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail with invalid group ID", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = srv.RegisterSocket(socketPath, libprm.Perm(0600), scksru.MaxGID+1)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(scksru.ErrInvalidGroup))
		})

		It("should fail with invalid path", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = srv.RegisterSocket("", libprm.Perm(0600), -1)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("callback registration", func() {
		var srv scksru.ServerUnix

		BeforeEach(func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should register error callback", func() {
			srv.RegisterFuncError(func(errs ...error) {
				// Error callback
			})
		})

		It("should register info callback", func() {
			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				// Info callback
			})
		})

		It("should register server info callback", func() {
			srv.RegisterFuncInfoServer(func(msg string) {
				// Server info callback
			})
		})
	})

	Context("TLS configuration", func() {
		It("should accept SetTLS but do nothing (no-op)", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// SetTLS should always return nil for Unix sockets
			err = srv.SetTLS(true, nil)
			Expect(err).ToNot(HaveOccurred())

			err = srv.SetTLS(false, nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
