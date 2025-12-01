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

// context_test.go verifies the context interface implementation (sCtx).
// Tests include context methods (Deadline, Done, Err, Value), connection state
// queries, host information retrieval, and I/O operations through the context.
package tcp_test

import (
	"context"
	"io"
	"net"
	"time"

	libsck "github.com/nabbar/golib/socket"
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

	BeforeEach(func() {
		adr = getTestAddr()
		c, cnl = context.WithCancel(globalCtx)
	})

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

				// Keep connection alive briefly
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

			// Cancel context to trigger Done
			cancel()

			Eventually(doneCalled, 3*time.Second).Should(Receive(BeTrue()))
			_ = con.Close()
		}, SpecTimeout(5*time.Second))

		It("should provide Err from parent context", func() {
			cancelCtx, cancel := context.WithCancel(c)

			errReceived := make(chan error, 1)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				// Check Err immediately
				if ctxIf, ok := interface{}(connCtx).(interface{ Err() error }); ok {
					// Wait for context cancellation
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
			cancel() // Cancel after connection is established

			// The error could be either context.Canceled or io.ErrClosedPipe
			// depending on timing, both are valid
			Eventually(errReceived, 2*time.Second).Should(Receive(Not(BeNil())))
			_ = con.Close()
		})

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
		It("should report IsConnected correctly", func() {
			connStateChanges := make(chan bool, 10)
			handler := func(connCtx libsck.Context) {
				defer connCtx.Close()

				if stateIf, ok := interface{}(connCtx).(interface{ IsConnected() bool }); ok {
					// Initially connected
					connStateChanges <- stateIf.IsConnected()

					// Still connected
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

			// Should receive true twice (connected state)
			Eventually(connStateChanges, 2*time.Second).Should(Receive(BeTrue()))
			Eventually(connStateChanges, 2*time.Second).Should(Receive(BeTrue()))
		})

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

		It("should handle Close correctly", func() {
			closeCalled := make(chan bool, 1)
			handler := func(connCtx libsck.Context) {
				defer func() {
					connCtx.Close()
					closeCalled <- true
				}()

				// Simulate some work
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
})
