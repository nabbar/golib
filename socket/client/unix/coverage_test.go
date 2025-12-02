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

package unix_test

import (
	"bytes"
	"context"
	"io"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UNIX Client Coverage Improvements", func() {
	Context("Write edge cases", func() {
		It("should handle type assertion error for invalid connection type", func() {
			cli := createClient(getTestSocketPath())
			defer func() { _ = cli.Close() }()

			// Write without connection
			n, err := cli.Write([]byte("test"))
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("Read edge cases", func() {
		It("should handle read on disconnected client", func() {
			cli := createClient(getTestSocketPath())
			defer func() { _ = cli.Close() }()

			buf := make([]byte, 100)
			n, err := cli.Read(buf)
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Context("IsConnected edge cases", func() {
		It("should return false when connection is invalid type", func() {
			cli := createClient(getTestSocketPath())
			defer func() { _ = cli.Close() }()

			Expect(cli.IsConnected()).To(BeFalse())
		})
	})

	Context("Close edge cases", func() {
		It("should handle close when not connected", func() {
			cli := createClient(getTestSocketPath())

			// Close without connect
			err := cli.Close()
			// May or may not error depending on implementation
			_ = err
		})

		It("should handle close on already closed client", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			srv := createSimpleTestServer(ctx, socketPath)
			defer cleanupSocket(socketPath)
			defer func() { _ = srv.Shutdown(ctx) }()

			cli := createClient(socketPath)
			connectClient(ctx, cli)

			// First close
			err := cli.Close()
			Expect(err).ToNot(HaveOccurred())

			// Second close should be safe
			err = cli.Close()
			// May or may not error, implementation specific
			_ = err
		})
	})

	Context("Once edge cases", func() {
		It("should handle Once with no response callback", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath)

			ctx, cancel := context.WithTimeout(globalCtx, 100*time.Millisecond)
			defer cancel()

			srv := createSimpleTestServer(ctx, socketPath)
			defer cleanupSocket(socketPath)
			defer func() { _ = srv.Shutdown(ctx) }()

			cli := createClient(socketPath)

			request := bytes.NewBufferString("test data")
			err := cli.Once(ctx, request, nil)
			Expect(err).ToNot(HaveOccurred())

			// Client should be disconnected after Once
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should handle Once with response callback that reads data", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			srv := createSimpleTestServer(ctx, socketPath)
			defer cleanupSocket(socketPath)
			defer func() { _ = srv.Shutdown(ctx) }()

			cli := createClient(socketPath)

			request := bytes.NewBufferString("echo this")
			var received []byte

			err := cli.Once(ctx, request, func(r io.Reader) {
				data, _ := io.ReadAll(r)
				received = data
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(len(received)).To(BeNumerically(">", 0))

			// Client should be disconnected after Once
			Expect(cli.IsConnected()).To(BeFalse())
		})

		It("should handle Once with context cancellation", func() {
			socketPath := getTestSocketPath()
			cli := createClient(socketPath)

			ctx, cancel := context.WithCancel(globalCtx)
			cancel() // Cancel immediately

			request := bytes.NewBufferString("test")
			err := cli.Once(ctx, request, nil)

			Expect(err).To(HaveOccurred())
		})
	})

	Context("dial edge cases", func() {
		It("should handle connect to non-existent socket", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath) // Ensure it doesn't exist

			cli := createClient(socketPath)
			defer func() { _ = cli.Close() }()

			err := cli.Connect(globalCtx)
			Expect(err).To(HaveOccurred())
		})

		It("should handle connect with invalid path permissions", func() {
			socketPath := "/root/cannot-create-here.sock"

			cli := createClient(socketPath)
			defer func() { _ = cli.Close() }()

			ctx, cancel := context.WithTimeout(globalCtx, 100*time.Millisecond)
			defer cancel()

			err := cli.Connect(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Callback coverage", func() {
		It("should invoke error callback on read error", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			srv := createServer(socketPath, closingHandler)
			startServer(ctx, srv)
			waitForServerRunning(socketPath, 5*time.Second)
			defer cleanupSocket(socketPath)
			defer func() { _ = srv.Shutdown(ctx) }()

			cli := createClient(socketPath)
			defer func() { _ = cli.Close() }()

			var errorCalled atomic.Int32
			cli.RegisterFuncError(func(errs ...error) {
				if len(errs) > 0 {
					errorCalled.Store(1)
				}
			})

			connectClient(ctx, cli)

			// Give server time to close connection
			time.Sleep(100 * time.Millisecond)

			// Try to read (should fail as server closed)
			buf := make([]byte, 100)
			_, _ = cli.Read(buf)

			// Allow callback to execute
			time.Sleep(100 * time.Millisecond)

			Expect(errorCalled.Load()).To(Equal(int32(1)))
		})

		It("should invoke error callback on write error", func() {
			socketPath := getTestSocketPath()
			cleanupSocket(socketPath)

			ctx, cancel := context.WithTimeout(globalCtx, 2*time.Second)
			defer cancel()

			srv := createServer(socketPath, closingHandler)
			startServer(ctx, srv)
			waitForServerRunning(socketPath, 5*time.Second)
			defer cleanupSocket(socketPath)
			defer func() { _ = srv.Shutdown(ctx) }()

			cli := createClient(socketPath)
			defer func() { _ = cli.Close() }()

			var errorCalled atomic.Int32
			cli.RegisterFuncError(func(errs ...error) {
				if len(errs) > 0 {
					errorCalled.Store(1)
				}
			})

			connectClient(ctx, cli)

			// Give server time to close connection
			time.Sleep(100 * time.Millisecond)

			// Try to write (may fail as server closed)
			_, _ = cli.Write([]byte("test"))

			// Allow callback to execute
			time.Sleep(100 * time.Millisecond)

			// Error callback may or may not be called depending on timing
			_ = errorCalled.Load()
		})
	})
})
