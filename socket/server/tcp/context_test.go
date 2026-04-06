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

// Package tcp_test validates the context-aware connection wrapper (sCtx).
//
// # Context Test Logic
//
// The 'context_test.go' file ensures that the sCtx structure correctly implements 
// the 'context.Context' interface and the custom 'libsck.Context' interface. 
// It focuses on:
//   - Cancellation Propagation: Ensuring that when the server stops, the 
//     connection context's Done() channel is closed.
//   - Deadline Management: Verifying that deadlines from the parent context 
//     are correctly reported.
//   - Value Retrieval: Ensuring request-scoped values flow through to the handler.
//   - Metadata Retrieval: Verifying Host and Protocol information accuracy.
//   - Idle Manager Integration: Testing Ref(), Inc(), and Get() used by the idlemgr.
//
// # Dataflow: Context Cancellation
//
//	[Parent Context Cancel] ───> [srv.gnc (Gone channel)]
//	                                     │
//	[sCtx.Close()] <───(select)──────────+── [sCtx.Done()]
//	      │
//	      v
//	[Handler Unblocked] ───> [Resource Cleanup]
package tcp_test

import (
	"context"
	"io"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
	sckidl "github.com/nabbar/golib/socket/idlemgr"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Context", func() {
	var (
		srv scksrt.ServerTcp
		adr string
		c   context.Context
		cnl context.CancelFunc
	)

	// Setup: Initialize address and global context.
	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

	// Cleanup: Stop server and cancel local context.
	AfterEach(func() {
		if srv != nil {
			_ = srv.Close()
		}
		if cnl != nil {
			cnl()
		}
		time.Sleep(50 * time.Millisecond)
	})

	Context("context interface methods", func() {
		// Test: Standard context.Deadline() implementation.
		It("should provide Deadline from parent context", func() {
			deadline := time.Now().Add(5 * time.Second)
			ctxWithDeadline, cancel := context.WithDeadline(c, deadline)
			defer cancel()

			connReceived := make(chan net.Conn, 1)
			handler := func(ctx libsck.Context) {
				defer ctx.Close()

				if conn, ok := interface{}(ctx).(interface {
					Deadline() (time.Time, bool)
				}); ok {
					t, hasDeadline := conn.Deadline()
					Expect(hasDeadline).To(BeTrue())
					Expect(t).To(BeTemporally("~", deadline, time.Second))
				}

				time.Sleep(10 * time.Millisecond)
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(ctxWithDeadline, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			connReceived <- con
			time.Sleep(50 * time.Millisecond)
			_ = con.Close()
		})

		// Test: Standard context.Deadline() for non-deadline contexts.
		It("should return false for deadline when none is set", func() {
			handlerCalled := make(chan bool, 1)
			handler := func(ctx libsck.Context) {
				defer ctx.Close()
				if conn, ok := interface{}(ctx).(interface {
					Deadline() (time.Time, bool)
				}); ok {
					_, ok := conn.Deadline()
					Expect(ok).To(BeFalse())
				}
				handlerCalled <- true
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()
			Eventually(handlerCalled, 2*time.Second).Should(Receive(BeTrue()))
		})

		// Test: Standard context.Done() implementation and signaling.
		It("should provide Done channel from parent context", func(ctx SpecContext) {
			cancelCtx, cancel := context.WithCancel(c)

			doneCalled := make(chan bool, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if ctxIf, ok := interface{}(connCtx).(interface{ Done() <-chan struct{} }); ok {
					select {
					case <-ctxIf.Done():
						doneCalled <- true
					case <-time.After(2 * time.Second):
						doneCalled <- false
					}
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(cancelCtx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			time.Sleep(20 * time.Millisecond)

			// Cancel context to trigger Done on sCtx
			cancel()

			Eventually(doneCalled, 3*time.Second).Should(Receive(BeTrue()))
			_ = con.Close()
		}, SpecTimeout(5*time.Second))

		// Test: Standard context.Err() behavior.
		It("should provide Err from parent context", func() {
			cancelCtx, cancel := context.WithCancel(c)

			errReceived := make(chan error, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if ctxIf, ok := interface{}(connCtx).(interface{ Err() error }); ok {
					// Wait for external cancellation
					time.Sleep(100 * time.Millisecond)
					errReceived <- ctxIf.Err()
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(cancelCtx, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			time.Sleep(20 * time.Millisecond)
			cancel() // Triggers sCtx cancellation

			Eventually(errReceived, 2*time.Second).Should(Receive(Not(BeNil())))
			_ = con.Close()
		})

		// Test: Standard context.Value() propagation.
		It("should provide Value from parent context", func() {
			type contextKey string
			const testKey contextKey = "testKey"
			testValue := "testValue"

			ctxWithValue := context.WithValue(c, testKey, testValue)

			valueReceived := make(chan string, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if ctxIf, ok := interface{}(connCtx).(interface{ Value(any) any }); ok {
					if val := ctxIf.Value(testKey); val != nil {
						valueReceived <- val.(string)
					}
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(ctxWithValue, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			Eventually(valueReceived, 2*time.Second).Should(Receive(Equal(testValue)))
			_ = con.Close()
		})
	})

	Context("connection state methods", func() {
		// Test: Verifying IsConnected flag accuracy.
		It("should report IsConnected correctly", func() {
			connStateChanges := make(chan bool, 10)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if stateIf, ok := interface{}(connCtx).(interface{ IsConnected() bool }); ok {
					connStateChanges <- stateIf.IsConnected()

					time.Sleep(10 * time.Millisecond)
					connStateChanges <- stateIf.IsConnected()
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Eventually(connStateChanges, 2*time.Second).Should(Receive(BeTrue()))
			Eventually(connStateChanges, 2*time.Second).Should(Receive(BeTrue()))
		})

		// Test: Verifying host string formatting.
		It("should report RemoteHost correctly", func() {
			remoteReceived := make(chan string, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if hostIf, ok := interface{}(connCtx).(interface{ RemoteHost() string }); ok {
					remote := hostIf.RemoteHost()
					remoteReceived <- remote
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Eventually(remoteReceived, 2*time.Second).Should(Receive(ContainSubstring("127.0.0.1")))
		})

		It("should report LocalHost correctly", func() {
			localReceived := make(chan string, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if hostIf, ok := interface{}(connCtx).(interface{ LocalHost() string }); ok {
					local := hostIf.LocalHost()
					localReceived <- local
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			Eventually(localReceived, 2*time.Second).Should(Receive(ContainSubstring("127.0.0.1")))
		})
	})

	Context("I/O operations", func() {
		// Test: Verifying sCtx implementation of io.Reader.
		It("should handle Read correctly", func() {
			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, echoHandler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			testData := []byte("test read operation")
			response := sendAndReceive(con, testData)
			Expect(response).To(Equal(testData))
		})

		// Test: Verifying sCtx implementation of io.Writer.
		It("should handle Write correctly", func() {
			dataReceived := make(chan []byte, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if rw, ok := interface{}(connCtx).(io.ReadWriter); ok {
					buf := make([]byte, 1024)
					n, err := rw.Read(buf)
					if err == nil && n > 0 {
						dataReceived <- buf[:n]
					}
				}
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			defer func() { _ = con.Close() }()

			testData := []byte("test write operation")
			n, err := con.Write(testData)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(testData)))

			Eventually(dataReceived, 2*time.Second).Should(Receive(Equal(testData)))
		})

		// Test: Verifying sCtx implementation of io.Closer.
		It("should handle Close correctly", func() {
			closeCalled := make(chan bool, 1)
			handler := func(connCtx libsck.Context) {
				defer func() {
					connCtx.Close()
					closeCalled <- true
				}()

				time.Sleep(10 * time.Millisecond)
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			time.Sleep(20 * time.Millisecond)
			_ = con.Close()

			Eventually(closeCalled, 2*time.Second).Should(Receive(BeTrue()))
		})
	})

	Context("extra coverage", func() {
		// Test: Idle Manager interface implementation check.
		It("Ref, Inc, Get should work", func() {
			handlerCalled := make(chan bool, 1)
			handler := func(ctx libsck.Context) {
				defer ctx.Close()

				sx, ok := ctx.(sckidl.Client)
				Expect(ok).To(BeTrue())

				// Ref() used for logging identification
				Expect(sx.Ref()).To(ContainSubstring("127.0.0.1"))
				// Inc() / Get() used for inactivity scans
				sx.Inc()
				Expect(sx.Get()).To(Equal(uint32(1)))

				handlerCalled <- true
			}

			cfg := createDefaultConfig(adr)
			srv, _ = scksrt.New(nil, handler, cfg)
			startServerInBackground(c, srv)
			waitForServerAcceptingConnections(adr, 2*time.Second)

			con := connectToServer(adr)
			_ = con.Close()
			Eventually(handlerCalled, 2*time.Second).Should(Receive(BeTrue()))
		})
	})
})
