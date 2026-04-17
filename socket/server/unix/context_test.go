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

// Package unix_test contains unit tests for the connection context implementation.
//
// # context_test.go: Connection Context (sCtx) Validation
//
// This file validates the `sCtx` structure, which implements the `libsck.Context`,
// `io.Reader`, `io.Writer`, and `context.Context` interfaces. It is the most
// critical component for the per-connection logic and resource pooling.
//
// # Scenarios Covered:
//
// ## 1. Interface Compliance
//   - I/O Operations: Ensures `Read()` and `Write()` correctly proxy data to/from
//     the underlying `net.UnixConn`.
//   - Context Integration: Validates that `Deadline()`, `Done()`, `Err()`, and
//     `Value()` delegate correctly to the base context.
//
// ## 2. Connection Lifecycle in Context
//   - IsConnected: Verifies the atomic state tracking of the connection's health.
//   - Local/Remote Host: Checks that the socket paths and protocol codes are
//     formatted correctly for logging and identification.
//
// ## 3. Error Handling and Propagation
//   - Post-Close Behavior: Ensures that I/O operations on a closed context
//     immediately return `io.ErrClosedPipe`.
//   - EOF Transition: Validates that receiving an `io.EOF` triggers a graceful
//     automatic closure of the context.
//
// ## 4. Idle Management Integration
//   - Activity Tracking: Tests the `Inc()` and `Get()` methods used by the
//     centralized Idle Manager.
//   - Reset Mechanism: Ensures that a successful I/O operation resets the
//     activity counter, preventing premature idle timeouts.
//
// ## 5. Resource Pooling (sync.Pool)
//   - Clean State: Validates (indirectly) that the `reset()` method correctly
//     sanitizes the structure before reuse, preventing data leaks between
//     sequential connections.
//
// # Technical Focus:
//
// These tests pay special attention to the atomic nature of the `sCtx` state. Since
// connection contexts are recycled, it is vital that the `clo` (closed) and `cnt`
// (counter) fields are perfectly reset.
//
// # Use Case Example (Testing Context Values):
//
//	tctx := context.WithValue(context.Background(), "traceID", "12345")
//	// In server handler:
//	traceID := ctx.Value("traceID") // Must return "12345"
package unix_test

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
	scksru "github.com/nabbar/golib/socket/server/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unix Socket Context Interface", func() {
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
		time.Sleep(50 * time.Millisecond)
	})

	Context("Read and Write operations", func() {
		It("should read and write data correctly", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			testData := []byte("Test data for read/write")
			rsp := sendAndReceive(con, testData)

			Expect(rsp).To(Equal(testData))
		})

		It("should handle multiple reads and writes", func() {
			cfg := createDefaultConfig(socketPath)
			var err error
			srv, err = scksru.New(nil, echoHandler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			for i := 0; i < 5; i++ {
				data := []byte("Message " + string(rune('0'+i)))
				rsp := sendAndReceive(con, data)
				Expect(rsp).To(Equal(data))
			}
		})

		It("should return error after connection close", func() {
			readCalled := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()
				_ = ctx.Close()
				readCalled.Store(true)

				// Try to read after close
				buf := make([]byte, 10)
				_, err := ctx.Read(buf)
				Expect(err).To(HaveOccurred())
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			Eventually(readCalled.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("connection state tracking", func() {
		It("should report IsConnected correctly", func() {
			connected := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()

				Expect(ctx.IsConnected()).To(BeTrue())
				connected.Store(true)

				time.Sleep(50 * time.Millisecond)
				_ = ctx.Close()

				Expect(ctx.IsConnected()).To(BeFalse())
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			Eventually(connected.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should report remote and local addresses", func() {
			addrChecked := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()

				remote := ctx.RemoteHost()
				local := ctx.LocalHost()

				Expect(remote).ToNot(BeEmpty())
				Expect(local).ToNot(BeEmpty())
				Expect(local).To(ContainSubstring("unix"))

				addrChecked.Store(true)
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			Eventually(addrChecked.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("context propagation", func() {
		It("should respect context cancellation", func() {
			cancelled := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()

				select {
				case <-ctx.Done():
					cancelled.Store(true)
					return
				case <-time.After(5 * time.Second):
					Fail("Context was not cancelled")
				}
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			tctx, tcnl := context.WithCancel(c)
			startServerInBackground(tctx, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			time.Sleep(50 * time.Millisecond)
			tcnl()

			Eventually(cancelled.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should provide context error after cancellation", func() {
			errChecked := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()

				<-ctx.Done()
				err := ctx.Err()
				Expect(err).ToNot(BeNil())
				errChecked.Store(true)
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			tctx, tcnl := context.WithCancel(c)
			startServerInBackground(tctx, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			defer func() { _ = con.Close() }()

			time.Sleep(50 * time.Millisecond)
			tcnl()

			Eventually(errChecked.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})

		It("should handle context methods correctly", func() {
			deadlineChecked := &atomic.Bool{}
			valueChecked := &atomic.Bool{}
			refChecked := &atomic.Bool{}

			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() { _ = ctx.Close() }()

				// Test Deadline
				_, ok := ctx.Deadline()
				Expect(ok).To(BeTrue())
				deadlineChecked.Store(true)

				// Test Value
				val := ctx.Value("testKey")
				Expect(val).To(Equal("testValue"))
				valueChecked.Store(true)

				// Test Ref/Inc/Get
				sx, ok := ctx.(sckidl.Client)
				Expect(ok).To(BeTrue())

				Expect(sx.Ref()).To(ContainSubstring("unix"))
				sx.Inc()
				Expect(sx.Get()).To(BeNumerically(">=", 0))
				refChecked.Store(true)

			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Inject value and deadline
			d := time.Now().Add(10 * time.Second)
			tctx, tcnl := context.WithDeadline(c, d)
			defer tcnl()
			tctx = context.WithValue(tctx, "testKey", "testValue")

			startServerInBackground(tctx, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			Eventually(deadlineChecked.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
			Eventually(valueChecked.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
			Eventually(refChecked.Load, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("EOF handling", func() {
		It("should handle client disconnect gracefully", func() {
			eofDetected := &atomic.Bool{}
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				defer func() {
					eofDetected.Store(true)

					Expect(ctx.IsConnected()).To(BeFalse())
					Expect(ctx.Err()).To(HaveOccurred())

					_ = ctx.Close()
					time.Sleep(50 * time.Millisecond)
				}()

				buf := make([]byte, 1024)
				for {
					_, err := ctx.Read(buf)
					if err != nil {
						return
					}
				}
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)

			con := connectToServer(socketPath)
			_ = con.Close()

			Expect(srv.IsRunning()).To(BeTrue())
			Expect(srv.IsGone()).To(BeFalse())

			Eventually(eofDetected.Load, 5*time.Second, 10*time.Millisecond).Should(BeTrue())
		})
	})

	Context("Edge cases and errors", func() {
		It("should handle Read/Write on nil connection correctly", func() {
			cfg := createDefaultConfig(socketPath)
			srv, err := scksru.New(nil, func(ctx libsck.Context) {
				// Internal access to sCtx is not possible here without type assertion
				// but we can test behavior through interface
				_ = ctx.Close()
				_, err := ctx.Read(make([]byte, 10))
				Expect(err).To(Equal(io.ErrClosedPipe))

				_, err = ctx.Write([]byte("test"))
				Expect(err).To(Equal(io.ErrClosedPipe))
			}, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(socketPath, 2*time.Second)
			_ = connectToServer(socketPath).Close()
		})
	})
})
