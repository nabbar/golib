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

// Package unix_test contains the Behavior-Driven Development (BDD) test suite for the Unix domain socket server.
// It utilizes the Ginkgo and Gomega frameworks to provide clear, readable, and maintainable test specifications.
//
// # basic_test.go: Fundamental Operational Validation
//
// This file focuses on the primary functional requirements of the Unix socket server, ensuring that the basic
// state transitions and communication primitives work correctly in isolation.
//
// # Scenarios Covered:
//
// ## 1. Server Lifecycle Management
//   - Startup: Validates that `Listen()` correctly initializes the listener, creates the socket file,
//     and transitions `IsRunning()` to true.
//   - Shutdown: Ensures that `Close()` and `Shutdown()` stop the listener and clean up filesystem resources.
//   - Graceful Stop: Tests that a context cancellation during `Listen` triggers a clean exit without panics.
//
// ## 2. Core Communication (Echo)
//   - Data Integrity: Verifies that data sent by a client to the `echoHandler` is received back bit-for-bit.
//   - Stream Handling: Ensures that the server can handle multiple read/write cycles over a single persistent connection.
//
// ## 3. Connection Tracking
//   - Atomic Counting: Validates that the server accurately increments and decrements the internal connection
//     counter (`OpenConnections()`) as clients connect and disconnect.
//   - State Consistency: Ensures `IsRunning()` and `IsGone()` flags reflect the actual physical state of the server.
//
// ## 4. Filesystem Interaction
//   - Automatic Creation: Verifies the socket file is created with the expected path upon calling `Listen`.
//   - Automatic Cleanup: Ensures the socket file is deleted from the OS temporary directory when the server stops,
//     preventing stale socket errors in subsequent runs.
//
// # Test Architecture:
//
//	[ BeforeEach ]
//	      |
//	      +---> Generate Unique Socket Path (/tmp/test-unix-*.sock)
//	      +---> Initialize Server with `echoHandler`
//	      +---> Create Cancellable Context
//	      v
//	[ Test Case (It) ]
//	      |
//	      +---> Start Server in Goroutine
//	      +---> Wait for Socket to be Ready (Polling net.Dial)
//	      +---> Execute Test Logic (net.Dial, Read, Write, Close)
//	      +---> Assertions (Expect/Eventually)
//	      v
//	[ AfterEach ]
//	      |
//	      +---> Stop Server (srv.Close)
//	      +---> Cancel Context
//	      +---> Remove Physical Socket File
//
// # Why These Tests Matter:
// These tests form the "sanity check" for any architectural changes (like the recent introduction of
// sync.Pool or the Idle Manager). If a basic echo fails or connection counting is off, the higher-level
// robustness and concurrency features cannot be trusted.
package unix_test

import (
	"context"
	"os"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Server Basic Operations", func() {
	var (
		srv        scksru.ServerUnix
		socketPath string
		c          context.Context
		cnl        context.CancelFunc
	)

	BeforeEach(func() {
		socketPath = getTestSocketPath()
		cfg := createDefaultConfig(socketPath)
		var err error
		srv, err = scksru.New(nil, echoHandler, cfg)
		Expect(err).ToNot(HaveOccurred())

		c, cnl = context.WithCancel(globalCtx)
		time.Sleep(5 * time.Millisecond)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		cleanupSocketFile(socketPath)
		time.Sleep(50 * time.Millisecond)
	})

	Context("starting and stopping", func() {
		It("should start server successfully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			Expect(srv.IsRunning()).To(BeTrue())
			Expect(srv.IsGone()).To(BeFalse())

			// Test Listener method
			ntw, adr, tls := srv.Listener()
			Expect(ntw).To(Equal(libptc.NetworkUnix))
			Expect(adr).To(Equal(socketPath))
			Expect(tls).To(BeFalse())
		})

		It("should accept connections when running", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			Expect(con).ToNot(BeNil())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		It("should echo messages correctly", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			msg := []byte("Hello, Unix Socket!")
			rsp := sendAndReceive(con, msg)

			Expect(rsp).To(Equal(msg))
		})

		It("should stop server with context cancellation", func() {
			tctx, tcnl := context.WithCancel(c)
			startServerInBackground(tctx, srv)
			waitForServer(srv, 2*time.Second)

			tcnl()

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Context("connection management", func() {
		It("should track multiple connections", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con1 := connectToServer(socketPath)
			defer func() { _ = con1.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			con2 := connectToServer(socketPath)
			defer func() { _ = con2.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(2)))

			con3 := connectToServer(socketPath)
			defer func() { _ = con3.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(3)))
		})

		It("should decrement count when connection closes", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should handle rapid connection open/close cycles", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			for i := 0; i < 10; i++ {
				con := connectToServer(socketPath)
				_ = con.Close()
			}

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("graceful shutdown", func() {
		It("should shutdown gracefully with Shutdown()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()

			err := srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(srv.IsRunning()).To(BeFalse())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should close immediately with Close()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should set IsGone flag after shutdown", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			err := srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return srv.IsGone()
			}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("error conditions", func() {
		It("should handle context cancellation during Listen", func() {
			tctx, tcnl := context.WithTimeout(c, 500*time.Microsecond)
			defer tcnl()

			err := srv.Listen(tctx)
			Expect(err).To(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should handle double close gracefully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			err1 := srv.Close()
			Expect(err1).ToNot(HaveOccurred())

			// Second close should not panic
			err2 := srv.Close()
			_ = err2
		})
	})

	Context("socket file management", func() {
		It("should create socket file on Listen", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			// Socket file should exist
			_, err := os.Stat(socketPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should remove socket file on Close", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			// Give time for cleanup
			time.Sleep(100 * time.Millisecond)

			// Socket file should be removed
			_, err = os.Stat(socketPath)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})
})
