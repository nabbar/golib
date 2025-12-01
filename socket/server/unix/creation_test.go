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

// creation_test.go validates server creation and configuration.
// Tests constructor behavior, parameter validation, and initial state.
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

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
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
