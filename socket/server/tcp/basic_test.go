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

// Package tcp_test validates the core operational features of the TCP server.
//
// # Basic Operations Test Logic
//
// The 'basic_test.go' file covers the fundamental lifecycle of a running server:
//   - Listening: Verifying the server binds correctly to the OS port.
//   - Acceptance: Ensuring clients can connect and exchange data (Echo pattern).
//   - State Monitoring: Validating OpenConnections, IsRunning, and IsGone flags.
//   - Graceful Shutdown: Ensuring the server drains connections correctly.
//
// # Dataflow: Echo Test Scenario
//
//	[Client] --(Dial)--> [Server Accept] --(Spawn Handler)
//	   │                                       │
//	   +-----(Send: "Hello")----------------> [Handler Read]
//	                                           │
//	   +<----(Recv: "Hello")----------------- [Handler Write]
//	   │                                       │
//	[Close] <-------------------------------- [Defer ctx.Close()]
//
// # Performance Considerations for Tests
//
// We use Eventually() with small polling intervals to account for the 
// asynchronous nature of goroutine spawning and connection establishment 
// in a high-performance environment.
package tcp_test

import (
	"context"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Basic Operations", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	// Setup: Create a fresh server instance before each test spec.
	BeforeEach(func() {
		adr = getTestAddr()
		cfg := createDefaultConfig(adr)
		var err error
		srv, err = scksrt.New(nil, echoHandler, cfg)
		Expect(err).ToNot(HaveOccurred())

		// Create a test-specific cancellation context.
		c, cnl = context.WithCancel(globalCtx)
		// Small sleep to ensure OS resources are settled.
		time.Sleep(5 * time.Millisecond)
	})

	// Cleanup: Ensure the server is closed to prevent port leaks.
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(50 * time.Millisecond)
	})

	Context("starting and stopping", func() {
		// Test: Verify the server correctly reports its running state.
		It("should start server successfully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			Expect(srv.IsRunning()).To(BeTrue())
			Expect(srv.IsGone()).To(BeFalse())

			nw, ls, tls := srv.Listener()
			Expect(nw).To(Equal(libptc.NetworkTCP))
			Expect(ls).To(Equal(adr))
			Expect(tls).To(BeFalse())
		})

		// Test: Verify connection acceptance and tracking.
		It("should accept connections when running", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Expect(con).ToNot(BeNil())
			// OpenConnections uses an atomic counter.
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))
		})

		// Test: Functional I/O verification.
		It("should echo messages correctly", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("Hello, World!")
			rsp := sendAndReceive(con, msg)

			Expect(rsp).To(Equal(msg))
		})

		// Test: Lifecycle management via Context.
		It("should stop server with context cancellation", func() {
			tctx, tcnl := context.WithCancel(c)
			startServerInBackground(tctx, srv)
			waitForServer(srv, 2*time.Second)

			tcnl() // Cancel the context passed to Listen()

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})
	})

	Context("connection management", func() {
		// Test: Multi-client tracking.
		It("should track multiple connections", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con1 := connectToServer(adr)
			defer func() { _ = con1.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			con2 := connectToServer(adr)
			defer func() { _ = con2.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(2)))

			con3 := connectToServer(adr)
			defer func() { _ = con3.Close() }()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(3)))
		})

		// Test: Atomic decrement on close.
		It("should decrement count when connection closes", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		// Test: Stress testing connection churn.
		It("should handle rapid connection open/close cycles", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			for i := 0; i < 10; i++ {
				con := connectToServer(adr)
				_ = con.Close()
			}

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("graceful shutdown", func() {
		// Test: Draining phase logic.
		It("should shutdown gracefully with Shutdown()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			_ = con.Close()

			// Shutdown() will block until OpenConnections() == 0.
			err := srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(srv.IsRunning()).To(BeFalse())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		// Test: Forced closure.
		It("should close immediately with Close()", func() {
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			err := srv.Close()
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 2*time.Second)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		// Test: State transitions verification.
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
		// Test: Boundary conditions for context timeouts.
		It("should handle context cancellation during Listen", func() {
			tctx, tcnl := context.WithTimeout(c, 500*time.Microsecond)
			defer tcnl()

			err := srv.Listen(tctx)
			// Context expiration is expected in this setup.
			Expect(err).To(HaveOccurred())
			Expect(srv.IsRunning()).To(BeFalse())
		})

		// Test: Idempotency check.
		It("should handle double close gracefully", func() {
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			err1 := srv.Close()
			Expect(err1).ToNot(HaveOccurred())

			// Second close should not panic or hang.
			err2 := srv.Close()
			_ = err2
		})
	})
})
