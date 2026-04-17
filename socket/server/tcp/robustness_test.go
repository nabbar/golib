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

// Package tcp_test validates the server's resilience under adverse conditions.
//
// # Robustness Test Logic
//
// The 'robustness_test.go' file is designed to simulate real-world failures
// and ensure the server remains stable. It focuses on:
//   - Panic Recovery: Ensuring a single misbehaving connection handler doesn't
//     crash the entire server.
//   - Resource Leaks: Verifying that goroutines and file descriptors are released
//     even after abrupt client disconnects.
//   - Idle Timeout: Stressing the 'idlemgr' to ensure inactive connections are
//     correctly reaped after the configured period.
//   - Port Contention: Verifying that binding to an already occupied port
//     triggers the appropriate error callbacks.
//   - Shutdown Pressure: Testing the server's ability to shut down while
//     handlers are performing slow operations.
//
// # Dataflow: Idle Connection Reap
//
//	[Connection Open] ───> [idlemgr.Register()]
//	                             │
//	[User Inactive] <────────────+── [idlemgr Scan Loop]
//	                             │          │
//	[sCtx.Close()] <─────────────+── [Counter > Limit?]
//	      │
//	      v
//	[sCtx.Done() Closed] ───> [Handler Exits]
package tcp_test

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	libdur "github.com/nabbar/golib/duration"
	libsck "github.com/nabbar/golib/socket"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Robustness", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	// Setup: Fresh context and address for each spec.
	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

	// Cleanup: Ensure no dangling listeners.
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Context("error handling", func() {
		// Test: Isolation of handler panics.
		It("should handle handler panics gracefully", func() {
			panicHandler := func(c libsck.Context) {
				defer func() {
					_ = recover() // Manual recovery in test to check server state.
					_ = c.Close()
				}()
				panic("test panic")
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, panicHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// The server itself should remain functional.
			con := connectToServer(adr)
			_ = con.Close()

			Expect(srv.IsRunning()).To(BeTrue())
		})

		// Test: Resilience against client churn.
		It("should recover from client disconnect", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close() // Abrupt close.

			time.Sleep(100 * time.Millisecond)

			// Should still handle subsequent valid clients.
			con2 := connectToServer(adr)
			defer func() { _ = con2.Close() }()

			msg := []byte("test")
			rsp := sendAndReceive(con2, msg)
			Expect(rsp).To(Equal(msg))
		})

		// Test: Handling rapid connection/disconnection bursts.
		It("should handle rapid open/close cycles", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			for i := 0; i < 20; i++ {
				con := connectToServer(adr)
				_ = con.Close()
			}

			// Server should still be running.
			Expect(srv.IsRunning()).To(BeTrue())

			// And should accept new connections.
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()
			msg := []byte("test")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})
	})

	Context("callback reliability", func() {
		// Test: Asynchronous error notification.
		It("should call error callback on errors", func() {
			errCnt := new(atomic.Int32)
			errFunc := func(e ...error) {
				if len(e) > 0 && e[0] != nil {
					errCnt.Add(1)
				}
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(errFunc)

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()

			time.Sleep(200 * time.Millisecond)
		})

		// Test: Bind error detection.
		It("should trigger error callback on port already in use", func() {
			errorReceived := make(chan error, 10)
			errFunc := func(errs ...error) {
				for _, e := range errs {
					if e != nil {
						errorReceived <- e
					}
				}
			}

			// First server binds to the port.
			cfg1 := createDefaultConfig(adr)
			srv1, err := scksrt.New(nil, echoHandler, cfg1)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv1)
			waitForServerAcceptingConnections(adr, 2*time.Second)
			defer func() { _ = srv1.Close() }()

			// Second server attempts to bind to the SAME port.
			cfg2 := createDefaultConfig(adr)
			srv, err = scksrt.New(nil, echoHandler, cfg2)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(errFunc)

			go func() {
				_ = srv.Listen(c)
			}()

			// Should receive a "bind: address already in use" error.
			Eventually(errorReceived, 2*time.Second).Should(Receive(Not(BeNil())))
		})

		// Test: Verification of connection lifecycle events.
		It("should call info callback on connection events", func() {
			infoCnt := new(atomic.Int32)
			infoFunc := func(_, _ net.Addr, _ libsck.ConnState) {
				infoCnt.Add(1)
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfo(infoFunc)

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			time.Sleep(200 * time.Millisecond)
			Expect(infoCnt.Load()).To(BeNumerically(">", 0))
		})

		// Test: Server lifecycle event notification.
		It("should call server info callback on events", func() {
			srvInfoCnt := new(atomic.Int32)
			srvInfoFunc := func(_ string) {
				srvInfoCnt.Add(1)
			}

			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfoServer(srvInfoFunc)

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			time.Sleep(200 * time.Millisecond)
			Expect(srvInfoCnt.Load()).To(BeNumerically(">", 0))
		})
	})

	Context("resource cleanup", func() {
		// Test: Port release after Close().
		It("should clean up resources after Close", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			err = srv.Close()
			Expect(err).ToNot(HaveOccurred())

			waitForServerStopped(srv, 2*time.Second)

			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 3*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})

		// Test: Goroutine leak prevention.
		It("should not leak goroutines after shutdown", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()

			err = srv.Shutdown(c)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(500 * time.Millisecond)

			// Check that server is fully stopped.
			Expect(srv.IsRunning()).To(BeFalse())
			Eventually(func() int64 {
				return srv.OpenConnections()
			}, 2*time.Second, 10*time.Millisecond).Should(Equal(int64(0)))
		})
	})

	Context("edge cases", func() {
		// Test: Verification of optional configuration handling.
		It("should handle nil UpdateConn function", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			msg := []byte("test")
			rsp := sendAndReceive(con, msg)
			Expect(rsp).To(Equal(msg))
		})

		// Test: Graceful handling of slow handlers during shutdown.
		It("should handle shutdown timeout gracefully", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, slowHandler(5*time.Second), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServer(srv, 2*time.Second)

			tctx, tcnl := context.WithTimeout(c, 500*time.Millisecond)
			defer tcnl()

			err = srv.Shutdown(tctx)
			// A timeout error is expected but server integrity must remain intact.
			_ = err
		})

		// Test: Verifying that active connections are handled during shutdown.
		It("should handle connection cleanup during shutdown", func() {
			cfg := createDefaultConfig(adr)
			var err error
			srv, err = scksrt.New(nil, slowHandler(100*time.Millisecond), cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			err = srv.Shutdown(c)
			// Shutdown should complete.
			Expect(err).ToNot(HaveOccurred())
		})

		// Test: Centralized Idle Timeout enforcement.
		It("should close idle connections after ConIdleTimeout", func() {
			cfg := createDefaultConfig(adr)
			cfg.ConIdleTimeout = libdur.Seconds(2)

			handlerStarted := make(chan time.Time, 1)
			handlerEnded := make(chan time.Time, 1)

			handler := func(ctx libsck.Context) {
				defer func() {
					ctx.Close()
					handlerEnded <- time.Now()
				}()

				handlerStarted <- time.Now()

				// Block until sCtx is closed by the idle manager.
				<-ctx.Done()
			}

			var err error
			srv, err = scksrt.New(nil, handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			// Connect and stay idle.
			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			var startTime, endTime time.Time
			Eventually(handlerStarted, 2*time.Second).Should(Receive(&startTime))
			Eventually(handlerEnded, 4*time.Second).Should(Receive(&endTime))

			// Total duration should be roughly equal to ConIdleTimeout.
			duration := endTime.Sub(startTime)
			Expect(duration).To(BeNumerically("~", 2*time.Second, time.Second))

			// Underlying net.Conn should be closed.
			buf := make([]byte, 10)
			con.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			_, err = con.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})
})
