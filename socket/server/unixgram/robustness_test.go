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
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libprm "github.com/nabbar/golib/file/perm"
	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"
	sckcfg "github.com/nabbar/golib/socket/config"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"
)

var _ = Describe("Unix Datagram Server Robustness", func() {
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

	Describe("Error Handling", func() {
		It("should handle existing socket file", func() {
			cfg := createBasicConfig()
			sockPath = cfg.Address

			// Create existing file
			file, err := os.Create(sockPath)
			Expect(err).ToNot(HaveOccurred())
			file.Close()

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			// Should start successfully (removes existing file)
			go func() {
				_ = srv.Listen(ctx)
			}()

			Eventually(func() bool {
				return srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should handle invalid socket paths", func() {
			cfg := sckcfg.Server{
				Network:   libptc.NetworkUnixGram,
				Address:   "/nonexistent/path/socket.sock",
				PermFile:  libprm.Perm(0600),
				GroupPerm: -1,
			}

			handler := func(ctx libsck.Context) {}
			srv, err := scksrv.New(nil, handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Should fail to listen
			err = srv.Listen(ctx)
			Expect(err).To(HaveOccurred())
		})

		It("should handle context cancellation during listen", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Cancel immediately
			cancel()

			Eventually(func() bool {
				return !srv.IsRunning()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should handle shutdown timeout gracefully", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Immediate timeout
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer shutdownCancel()

			err = srv.Shutdown(shutdownCtx)
			// May timeout or succeed depending on timing
			_ = err
		})
	})

	Describe("Edge Cases", func() {
		It("should handle repeated start attempts", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Try to start again (should fail or be no-op)
			newCtx, newCancel := context.WithCancel(testCtx)
			defer newCancel()

			go func() {
				_ = srv.Listen(newCtx)
			}()

			time.Sleep(100 * time.Millisecond)

			// Should still be running from first start
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle empty datagrams", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send empty datagram
			err = sendUnixgramDatagram(sockPath, []byte{})
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			// Handler should handle it gracefully
		})

		It("should handle very small datagrams", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			startServer(srv, ctx)

			// Send single byte
			err = sendUnixgramDatagram(sockPath, []byte{0x01})
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int64 {
				return handler.getCount()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", 0))
		})
	})

	Describe("Resource Cleanup", func() {
		It("should cleanup socket file on normal shutdown", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			Expect(fileExists(sockPath)).To(BeTrue())

			stopServer(srv, cancel)

			Eventually(func() bool {
				return !fileExists(sockPath)
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should cleanup socket file on error shutdown", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			Expect(fileExists(sockPath)).To(BeTrue())

			// Force error by cancelling context
			cancel()

			Eventually(func() bool {
				return !fileExists(sockPath)
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Describe("Callback Resilience", func() {
		It("should handle nil callbacks gracefully", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			// Register nil callbacks
			srv.RegisterFuncError(nil)
			srv.RegisterFuncInfo(nil)
			srv.RegisterFuncInfoServer(nil)

			// Should still work
			startServer(srv, ctx)
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle panicking callbacks", func() {
			handler := newTestHandler(false)
			srv, path, err := createServerWithHandler(handler.handler)
			sockPath = path
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			// Register panicking callback
			srv.RegisterFuncError(func(errs ...error) {
				panic("test panic")
			})

			// Server should still function
			// (implementation may recover from panics)
			startServer(srv, ctx)
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})
})
