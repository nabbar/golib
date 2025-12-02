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
	"io"
	"net"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/unixgram"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Datagram Client Edge Cases", func() {
	Describe("Read operations", func() {
		var (
			ctx        context.Context
			cancel     context.CancelFunc
			srv        scksrv.ServerUnixGram
			socketPath string
		)

		BeforeEach(func() {
			ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
			socketPath = getTestSocketPath()
			srv = createSimpleTestServer(ctx, socketPath)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			cleanupSocket(socketPath)
			if cancel != nil {
				cancel()
			}
		})

		It("should handle Read with closed socket", func() {
			cli := createClient(socketPath)
			connectClient(ctx, cli)

			// Close the socket
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Try to read after close
			buf := make([]byte, 100)
			n, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should handle Write with closed socket", func() {
			cli := createClient(socketPath)
			connectClient(ctx, cli)

			// Close the socket
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Try to write after close
			n, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})

		It("should trigger callbacks on Write", func() {
			cli := createClient(socketPath)

			writeCallbackCalled := new(atomic.Bool)
			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				if state == libsck.ConnectionWrite {
					writeCallbackCalled.Store(true)
				}
			})

			connectClient(ctx, cli)
			defer cli.Close()

			// Write data to trigger write callback
			_, err := cli.Write([]byte("test"))
			Expect(err).ToNot(HaveOccurred())

			// Callback should be triggered
			Eventually(func() bool {
				return writeCallbackCalled.Load()
			}, 500*time.Millisecond).Should(BeTrue())
		})
	})

	Describe("Once operation edge cases", func() {
		var (
			ctx        context.Context
			cancel     context.CancelFunc
			srv        scksrv.ServerUnixGram
			socketPath string
		)

		BeforeEach(func() {
			ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
			socketPath = getTestSocketPath()
			srv = createSimpleTestServer(ctx, socketPath)
		})

		AfterEach(func() {
			if srv != nil && srv.IsRunning() {
				_ = srv.Shutdown(ctx)
			}
			cleanupSocket(socketPath)
			if cancel != nil {
				cancel()
			}
		})

		It("should handle nil response callback", func() {
			cli := createClient(socketPath)

			err := cli.Once(ctx, nil, nil)
			// Function accepts nil callback (no-op behavior)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil reader with valid callback", func() {
			cli := createClient(socketPath)

			err := cli.Once(ctx, nil, func(r io.Reader) {
				// Callback provided but reader is nil
			})

			// Once should succeed even with nil reader (no data to send)
			// The function should connect, skip write, and call response callback
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Callback edge cases", func() {
		It("should handle nil error callback", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			cli.RegisterFuncError(nil)

			// Should not panic
			Expect(cli).ToNot(BeNil())
		})

		It("should handle nil info callback", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			cli.RegisterFuncInfo(nil)

			// Should not panic
			Expect(cli).ToNot(BeNil())
		})

		It("should handle multiple callback registrations", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			count := 0
			callback := func(errs ...error) {
				count++
			}

			// Register same callback multiple times
			cli.RegisterFuncError(callback)
			cli.RegisterFuncError(callback)
			cli.RegisterFuncError(callback)

			Expect(cli).ToNot(BeNil())
		})
	})

	Describe("Connection state edge cases", func() {
		It("should handle IsConnected after multiple calls", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			// Call multiple times before connect
			for i := 0; i < 10; i++ {
				Expect(cli.IsConnected()).To(BeFalse())
			}
		})

		It("should handle Close on unconnected client", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			err := cli.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should handle multiple Close calls", func() {
			ctx, cancel := context.WithTimeout(globalCtx, 5*time.Second)
			defer cancel()

			socketPath := getTestSocketPath()
			srv := createSimpleTestServer(ctx, socketPath)
			defer func() {
				if srv != nil && srv.IsRunning() {
					_ = srv.Shutdown(ctx)
				}
				cleanupSocket(socketPath)
			}()

			cli := createClient(socketPath)
			connectClient(ctx, cli)

			// First close should succeed
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Subsequent closes should fail
			err = cli.Close()
			Expect(err).To(HaveOccurred())

			err = cli.Close()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("TLS operations", func() {
		It("should handle SetTLS on unixgram client", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			// Unix sockets don't support TLS, should be no-op
			err := cli.SetTLS(true, nil, "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle SetTLS with various parameters", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			err := cli.SetTLS(false, nil, "test")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
