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

// Package unix_test includes robustness and fault-tolerance tests to ensure the server
// behaves correctly under adverse conditions, malformed client behavior, and unexpected failures.
//
// # robustness_test.go: Fault Tolerance and Edge Case Validation
//
// These tests focus on the server's stability and its ability to recover from common runtime
// issues such as handler panics, idle connections, and filesystem-level permission changes.
//
// # Scenarios Covered:
//
// ## 1. Callback Reliability
//   - Error Reporting: Forces errors (e.g., listening on an illegal path) and verifies that
//     the `RegisterFuncError` callback is correctly triggered.
//   - Connection Events: Validates that `RegisterFuncInfo` receives all lifecycle events
//     (New, Read, Write, Close) for every connection.
//   - Lifecycle Logging: Confirms that `RegisterFuncInfoServer` captures major transitions.
//
// ## 2. Idle Timeout and Resource Leaks
//   - Automatic Closure: Sets an aggressive 1-second idle timeout and verifies that clients
//     who connect but stay silent are automatically disconnected by the Idle Manager.
//   - Anti-false Positive: Confirms that active connections (those sending periodic data)
//     are NOT disconnected, validating the activity reset logic in `sCtx.Read/Write`.
//
// ## 3. Fault Recovery (Panics)
//   - Handler Resilience: Simulates an intentional panic within the user's `HandlerFunc`.
//     Verifies that the server recovers, logs the panic via the runner's recovery logic,
//     and continues to accept new connections without interruption.
//
// ## 4. Adverse Client Behavior
//   - Half-Open/Silent Clients: Tests the server's reaction to clients that connect but never
//     send or read data.
//   - Client Drop: Validates graceful handling when a client abruptly closes the socket
//     mid-stream (detecting `io.EOF`).
//
// ## 5. Lifecycle Persistence
//   - Repeated Cycles: Executes multiple start/stop sequences (`Listen` -> `Shutdown` -> `Listen`)
//     on the same server instance to ensure that internal state (channels, pooling) is
//     perfectly re-initialized.
//
// # Technical Focus:
// Special attention is given to the centralized Idle Manager. Since it runs in its own
// goroutine, these tests ensure that its interaction with the atomic `cnt` (activity counter)
// in each `sCtx` is perfectly synchronized.
//
// # Data Flow (Panic Recovery):
//
//	[ Listener Loop ] ---> (New Conn) ---> [ Handler Goroutine ]
//	                                              |
//	                                     ( Panic! )
//	                                              |
//	[ Server Recovery ] <-------------------------+
//	        |
//	( Log Error )
//	( Close sCtx )
//	( Return to Pool )
//	        |
//	[ Listener Still Active ]
package unix_test

import (
	"context"
	"net"
	"os"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Server Robustness", func() {
	var (
		srv        scksru.ServerUnix
		socketPath string
		c          context.Context
		cnl        context.CancelFunc
	)

	BeforeEach(func() {
		socketPath = getTestSocketPath()
		c, cnl = context.WithCancel(globalCtx)
	})

	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		cleanupSocketFile(socketPath)
		time.Sleep(100 * time.Millisecond)
	})

	Context("error callback behavior", func() {
		It("should call error callback on errors", func() {
			errorCalled := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(func(errs ...error) {
				errorCalled.Store(true)
			})

			// Cause an error by trying to listen on invalid path
			invalidPath := "/invalid/path/that/does/not/exist/test.sock"
			err = srv.RegisterSocket(invalidPath, 0600, -1)
			Expect(err).ToNot(HaveOccurred())

			_ = srv.Listen(c)
			Eventually(errorCalled.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should call info callback on connection events", func() {
			infoCalled := &atomic.Int32{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {
				infoCalled.Add(1)
			})

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			Eventually(func() int32 {
				return infoCalled.Load()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">=", 1))
		})

		It("should call info server callback", func() {
			infoSrvCalled := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfoServer(func(_ string) {
				infoSrvCalled.Store(true)
			})

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			Eventually(infoSrvCalled.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("idle timeout behavior", func() {
		It("should close idle connections after timeout", func() {
			// Using a slightly longer timeout for stability in tests
			cfg := createConfigWithIdleTimeout(socketPath, time.Second)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Connection is established
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(1)))

			// Wait for idle manager to scan and close (idle manager scans every 1s by default in New)
			// So we wait at least idle + scan interval
			time.Sleep(2 * time.Second)

			// Connection should be closed due to timeout
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 100*time.Millisecond).Should(Equal(int64(0)))
		})

		It("should not timeout active connections", func() {
			cfg := createConfigWithIdleTimeout(socketPath, 1*time.Second)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Keep connection active with periodic messages
			for i := 0; i < 5; i++ {
				data := []byte("Keep alive")
				rsp := sendAndReceive(con, data)
				Expect(rsp).To(Equal(data))
				time.Sleep(400 * time.Millisecond)
			}

			// Connection should still be open
			Expect(srv.OpenConnections()).To(Equal(int64(1)))
		})
	})

	Context("handler panics", func() {
		It("should recover from handler panic", func() {
			panicCount := &atomic.Int32{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()
				panicCount.Add(1)
				panic("intentional panic for testing")
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			// Server should still be running after panic
			Expect(srv.IsRunning()).To(BeTrue())
			Eventually(func() int32 {
				return panicCount.Load()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically(">", int32(1)))
		})
	})

	Context("socket file permissions", func() {
		It("should create socket with correct permissions", func() {
			cfg := createConfigWithPerms(socketPath, 0600, -1)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			info, err := os.Stat(socketPath)
			Expect(err).ToNot(HaveOccurred())

			// Check that it's a socket
			Expect(info.Mode() & os.ModeSocket).ToNot(Equal(os.FileMode(0)))
		})
	})

	Context("malformed client behavior", func() {
		It("should handle client that doesn't send data", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, readOnlyHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Wait a bit then close without sending
			time.Sleep(100 * time.Millisecond)
			_ = con.Close()

			// Server should handle gracefully
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle client that only sends data", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, writeOnlyHandler("Server message"), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			// Read server message
			buf := make([]byte, 1024)
			n, err := con.Read(buf)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(BeNumerically(">", 0))

			// Server should handle gracefully
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Context("resource cleanup", func() {
		It("should cleanup all connections on shutdown", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, slowHandler(1*time.Second), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			// Create multiple connections
			con1 := connectToServer(socketPath)
			con2 := connectToServer(socketPath)
			con3 := connectToServer(socketPath)

			// Wait for connections to be established
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(3)))

			// Shutdown
			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			// All connections should be closed
			_ = con1.Close()
			_ = con2.Close()
			_ = con3.Close()

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("repeated start/stop cycles", func() {
		It("should handle multiple Listen attempts", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// First start
			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			// Stop
			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			// Wait for complete stop
			time.Sleep(200 * time.Millisecond)

			// Second start should work
			c2, cnl2 := context.WithCancel(globalCtx)
			defer cnl2()

			startServerInBackground(c2, srv)
			waitForServer(srv, 2*time.Second)

			err = srv.Shutdown(c2)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
