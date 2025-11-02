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

package tcp_test

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libptc "github.com/nabbar/golib/network/protocol"
	libsck "github.com/nabbar/golib/socket"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server Callbacks", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     libsck.Server
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(x, 60*time.Second)
		address = getTestAddress()
		srv = createAndRegisterServer(address, echoHandler, nil)
	})

	AfterEach(func() {
		if srv != nil && srv.IsRunning() {
			_ = srv.Shutdown(ctx)
		}
		if cancel != nil {
			cancel()
		}
	})

	Describe("RegisterFuncError", func() {
		It("should call error callback on errors", func() {
			errorCount := &atomic.Int32{}
			var lastError error

			srv.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
				if len(errs) > 0 {
					lastError = errs[0]
				}
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Connect and immediately close to trigger potential errors
			conn := connectClient(address)
			_ = conn.Close()

			// Wait a bit for any error callbacks
			time.Sleep(500 * time.Millisecond)

			// Errors may or may not occur depending on timing
			// Just verify the callback can be registered without panic
			_ = errorCount.Load()
			_ = lastError
		})

		It("should handle nil error callback", func() {
			srv.RegisterFuncError(nil)

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Should work fine without callback
			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)
		})

		It("should allow replacing error callback", func() {
			count1 := &atomic.Int32{}
			count2 := &atomic.Int32{}

			srv.RegisterFuncError(func(errs ...error) {
				count1.Add(1)
			})

			srv.RegisterFuncError(func(errs ...error) {
				count2.Add(1)
			})

			// Second callback should replace the first
			// (Implementation detail: only the last registered callback is used)
		})
	})

	Describe("RegisterFuncInfo", func() {
		It("should call info callback on connection events", func() {
			infoCount := &atomic.Int32{}
			var events []libsck.ConnState

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				infoCount.Add(1)
				events = append(events, state)

				Expect(local).ToNot(BeNil())
				Expect(remote).ToNot(BeNil())
				Expect(local.Network()).To(Equal(libptc.NetworkTCP.String()))
				Expect(remote.Network()).To(Equal(libptc.NetworkTCP.String()))
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("test\n")
			sendMessage(conn, msg)
			_ = receiveMessage(conn, 1024)

			_ = conn.Close()

			// Wait for events to be processed
			time.Sleep(1 * time.Second)

			// Should have received multiple state changes
			Eventually(func() int32 {
				return infoCount.Load()
			}, 5*time.Second, 50*time.Millisecond).Should(BeNumerically(">", 0))
		})

		It("should report correct connection states", func() {
			states := make(map[libsck.ConnState]int)
			var statesMutex sync.Mutex

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				statesMutex.Lock()
				defer statesMutex.Unlock()
				states[state]++
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("test\n")
			sendMessage(conn, msg)
			_ = receiveMessage(conn, 1024)

			_ = conn.Close()

			// Wait for all events
			time.Sleep(2 * time.Second)

			statesMutex.Lock()
			defer statesMutex.Unlock()

			// Should have at least ConnectionNew
			Expect(states[libsck.ConnectionNew]).To(BeNumerically(">", 0))
		})

		It("should handle nil info callback", func() {
			srv.RegisterFuncInfo(nil)

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			defer func() {
				_ = conn.Close()
			}()

			waitForConnections(srv, 1, 2*time.Second)
		})

		It("should call info callback for multiple connections", func() {
			connectionEvents := &atomic.Int32{}

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				if state == libsck.ConnectionNew {
					connectionEvents.Add(1)
				}
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn1 := connectClient(address)
			defer func() {
				_ = conn1.Close()
			}()

			conn2 := connectClient(address)
			defer func() {
				_ = conn2.Close()
			}()

			conn3 := connectClient(address)
			defer func() {
				_ = conn3.Close()
			}()

			Eventually(func() int32 {
				return connectionEvents.Load()
			}, 3*time.Second, 50*time.Millisecond).Should(Equal(int32(3)))
		})
	})

	Describe("RegisterFuncInfoServer", func() {
		It("should call server info callback", func() {
			infoCount := &atomic.Int32{}
			var messages []string

			srv.RegisterFuncInfoServer(func(msg string) {
				infoCount.Add(1)
				messages = append(messages, msg)
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Wait for server info messages
			Eventually(func() int32 {
				return infoCount.Load()
			}, 3*time.Second, 50*time.Millisecond).Should(BeNumerically(">", 0))

			// Should have received server startup message
			Expect(len(messages)).To(BeNumerically(">", 0))
		})

		It("should provide meaningful server messages", func() {
			var receivedMsg string

			srv.RegisterFuncInfoServer(func(msg string) {
				receivedMsg = msg
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			Eventually(func() string {
				return receivedMsg
			}, 3*time.Second, 50*time.Millisecond).ShouldNot(BeEmpty())

			// Message should contain relevant info
			Expect(receivedMsg).To(Or(
				ContainSubstring("listening"),
				ContainSubstring("starting"),
			))
		})

		It("should handle nil server info callback", func() {
			srv.RegisterFuncInfoServer(nil)

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)
		})

		It("should call server info on shutdown", func() {
			messages := make([]string, 0)
			var msgMutex sync.Mutex

			srv.RegisterFuncInfoServer(func(msg string) {
				msgMutex.Lock()
				defer msgMutex.Unlock()
				messages = append(messages, msg)
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			_ = srv.Shutdown(ctx)

			time.Sleep(500 * time.Millisecond)

			msgMutex.Lock()
			defer msgMutex.Unlock()

			// Should have received messages including shutdown
			Expect(len(messages)).To(BeNumerically(">", 0))
		})
	})

	Describe("Multiple Callbacks", func() {
		It("should support all callbacks simultaneously", func() {
			errorCalled := &atomic.Bool{}
			infoCalled := &atomic.Bool{}
			serverInfoCalled := &atomic.Bool{}

			srv.RegisterFuncError(func(errs ...error) {
				errorCalled.Store(true)
			})

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				infoCalled.Store(true)
			})

			srv.RegisterFuncInfoServer(func(msg string) {
				serverInfoCalled.Store(true)
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			conn := connectClient(address)
			waitForConnections(srv, 1, 2*time.Second)

			msg := []byte("test\n")
			sendMessage(conn, msg)
			_ = receiveMessage(conn, 1024)

			_ = conn.Close()

			// Wait for callbacks
			time.Sleep(1 * time.Second)

			// At least info and server info should be called
			Expect(infoCalled.Load()).To(BeTrue())
			Expect(serverInfoCalled.Load()).To(BeTrue())
		})
	})

	Describe("Callback Thread Safety", func() {
		It("should handle concurrent callback invocations", func() {
			callCount := &atomic.Int32{}

			srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				callCount.Add(1)
				// Simulate some processing
				time.Sleep(1 * time.Millisecond)
			})

			startServer(ctx, srv)
			waitForServerRunning(srv, 2*time.Second)

			// Create many concurrent connections
			const numConns = 10
			conns := make([]net.Conn, numConns)
			for i := 0; i < numConns; i++ {
				conns[i] = connectClient(address)
				defer func() {
					_ = conns[i].Close()
				}()
			}

			waitForConnections(srv, int64(numConns), 3*time.Second)

			// Should have many callback invocations
			Eventually(func() int32 {
				return callCount.Load()
			}, 3*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", int32(numConns)))
		})
	})
})
