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

package unixgram_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"
)

var _ = Describe("Unix Datagram Server Boundary Tests", func() {
	var (
		ctx      context.Context
		cancel   context.CancelFunc
		sockPath string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		cleanupSocketFile(sockPath)
		time.Sleep(50 * time.Millisecond)
	})

	Describe("Configuration Boundaries", func() {
		It("should handle minimum valid GID (-1)", func() {
			cfg := createBasicConfig()
			cfg.GroupPerm = -1
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should handle zero GID", func() {
			cfg := createBasicConfig()
			cfg.GroupPerm = 0
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should handle maximum valid GID (32767)", func() {
			cfg := createBasicConfig()
			cfg.GroupPerm = 32767 // MaxGID
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should reject GID exceeding maximum (32768)", func() {
			cfg := createBasicConfig()
			cfg.GroupPerm = 32768 // MaxGID + 1
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})

		It("should handle minimum permissions (0000)", func() {
			cfg := createBasicConfig()
			cfg.PermFile = libprm.Perm(0000)
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should handle maximum permissions (0777)", func() {
			cfg := createBasicConfig()
			cfg.PermFile = libprm.Perm(0777)
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

	Describe("Datagram Size Boundaries", func() {
		It("should handle minimum datagram (1 byte)", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			err = sendUnixgramDatagram(sockPath, []byte{0xFF})
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))
		})

		It("should handle medium datagram (4KB)", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			data := make([]byte, 4096)
			for i := range data {
				data[i] = byte(i % 256)
			}

			err = sendUnixgramDatagram(sockPath, data)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))
		})

		It("should handle large datagram (16KB)", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			data := make([]byte, 16384)
			for i := range data {
				data[i] = byte(i % 256)
			}

			err = sendUnixgramDatagram(sockPath, data)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))
		})
	})

	Describe("Socket Path Boundaries", func() {
		It("should handle short path", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkUnixGram,
				Address:   "/tmp/s.sock",
				PermFile:  libprm.Perm(0600),
				GroupPerm: -1,
			}
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})

		It("should handle relative path", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkUnixGram,
				Address:   "./test.sock",
				PermFile:  libprm.Perm(0600),
				GroupPerm: -1,
			}
			sockPath = cfg.Address

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)

			Expect(err).ToNot(HaveOccurred())
			Expect(srv).ToNot(BeNil())
		})
	})

	Describe("Timing Boundaries", func() {
		It("should handle immediate context cancellation", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			// Cancel immediately
			immediateCtx, immediateCancel := context.WithCancel(testCtx)
			immediateCancel()

			err = srv.Listen(immediateCtx)
			Expect(err).To(HaveOccurred()) // Should get context cancelled error
		})

		It("should handle very short timeout", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			timeoutCtx, timeoutCancel := context.WithTimeout(testCtx, 1*time.Millisecond)
			defer timeoutCancel()

			err = srv.Listen(timeoutCtx)
			// May succeed or timeout depending on system speed
			_ = err
		})
	})

	Describe("State Transition Boundaries", func() {
		It("should handle rapid start-stop cycles", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 3; i++ {
				localCtx, localCancel := context.WithCancel(testCtx)

				startServer(srv, localCtx)
				Expect(srv.IsRunning()).To(BeTrue())

				stopServer(srv, localCancel)
				Eventually(func() bool {
					return !srv.IsRunning()
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())

				time.Sleep(100 * time.Millisecond)
			}
		})
	})
})
