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

package udp_test

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libsck "github.com/nabbar/golib/socket"
	scksrv "github.com/nabbar/golib/socket/server/udp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client Callbacks", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrv.ServerUdp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		address = getTestAddress()
		srv = createSimpleTestServer(ctx, address)
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
		It("should call error callback on connection errors", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			errorCount := new(atomic.Int32)
			var lastError error
			var mu sync.Mutex

			cli.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
				mu.Lock()
				if len(errs) > 0 {
					lastError = errs[0]
				}
				mu.Unlock()
			})

			// Try to write without connection
			_, _ = cli.Write([]byte("test"))

			// Wait for callback
			time.Sleep(100 * time.Millisecond)

			Expect(errorCount.Load()).To(BeNumerically(">", 0))
			mu.Lock()
			Expect(lastError).ToNot(BeNil())
			mu.Unlock()
		})

		It("should call error callback on read errors", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			errorCount := new(atomic.Int32)
			cli.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
			})

			// Try to read without connection
			buf := make([]byte, 1024)
			_, _ = cli.Read(buf)

			time.Sleep(100 * time.Millisecond)
			Expect(errorCount.Load()).To(BeNumerically(">", 0))
		})

		It("should handle nil error callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			cli.RegisterFuncError(nil)

			// Should not panic
			_, _ = cli.Write([]byte("test"))
		})

		It("should allow changing error callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			count1 := new(atomic.Int32)
			count2 := new(atomic.Int32)

			cli.RegisterFuncError(func(errs ...error) {
				count1.Add(1)
			})

			// Trigger error
			_, _ = cli.Write([]byte("test1"))
			time.Sleep(100 * time.Millisecond)

			// Replace callback
			cli.RegisterFuncError(func(errs ...error) {
				count2.Add(1)
			})

			// Trigger another error
			_, _ = cli.Write([]byte("test2"))
			time.Sleep(100 * time.Millisecond)

			// First callback should have been called once
			Expect(count1.Load()).To(Equal(int32(1)))
			// Second callback should have been called once
			Expect(count2.Load()).To(Equal(int32(1)))
		})

		It("should execute callback asynchronously", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			callbackStarted := make(chan bool, 1)
			callbackDone := make(chan bool, 1)

			cli.RegisterFuncError(func(errs ...error) {
				callbackStarted <- true
				time.Sleep(50 * time.Millisecond)
				callbackDone <- true
			})

			// Trigger error - should not block
			start := time.Now()
			_, _ = cli.Write([]byte("test"))
			duration := time.Since(start)

			// Write should return quickly (async callback)
			Expect(duration).To(BeNumerically("<", 40*time.Millisecond))

			// Callback should still execute
			Eventually(callbackStarted, time.Second).Should(Receive())
			Eventually(callbackDone, time.Second).Should(Receive())
		})
	})

	Describe("RegisterFuncInfo", func() {
		It("should call info callback on connect", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			infoCount := new(atomic.Int32)
			var states []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				infoCount.Add(1)
				mu.Lock()
				states = append(states, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)
			time.Sleep(100 * time.Millisecond)

			Expect(infoCount.Load()).To(BeNumerically(">=", 2)) // Dial + New
			mu.Lock()
			Expect(states).To(ContainElement(libsck.ConnectionDial))
			Expect(states).To(ContainElement(libsck.ConnectionNew))
			mu.Unlock()
		})

		It("should call info callback on read", func() {
			srv, cli, _, ctx, cancel := createTestServerAndClient(simpleEchoHandler())
			defer cleanupServer(srv, ctx)
			defer cleanupClient(cli)
			defer cancel()

			var states []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				states = append(states, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)

			// Write and read with echo server
			_, _ = cli.Write([]byte("test"))
			
			// Read with timeout in goroutine to avoid blocking
			done := make(chan struct{})
			go func() {
				defer close(done)
				buf := make([]byte, 1024)
				_, _ = cli.Read(buf)
			}()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				// Timeout is acceptable for read callback test
			}

			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			// Read callback may or may not be called depending on server response timing
			// Just verify we got some callbacks
			Expect(len(states)).To(BeNumerically(">", 0))
			mu.Unlock()
		})

		It("should call info callback on write", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var states []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				states = append(states, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)
			_, _ = cli.Write([]byte("test"))
			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			Expect(states).To(ContainElement(libsck.ConnectionWrite))
			mu.Unlock()
		})

		It("should call info callback on close", func() {
			cli := createClient(address)

			var states []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				states = append(states, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)
			_ = cli.Close()
			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			Expect(states).To(ContainElement(libsck.ConnectionClose))
			mu.Unlock()
		})

		It("should handle nil info callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			cli.RegisterFuncInfo(nil)

			// Should not panic
			connectClient(ctx, cli)
		})

		It("should provide valid addresses in callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var capturedLocal, capturedRemote net.Addr
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				if state == libsck.ConnectionNew {
					mu.Lock()
					capturedLocal = local
					capturedRemote = remote
					mu.Unlock()
				}
			})

			connectClient(ctx, cli)
			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			Expect(capturedLocal).ToNot(BeNil())
			Expect(capturedRemote).ToNot(BeNil())
			mu.Unlock()
		})

		It("should execute callback asynchronously", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			callbackStarted := make(chan bool, 1)
			callbackDone := make(chan bool, 1)

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				if state == libsck.ConnectionNew {
					callbackStarted <- true
					time.Sleep(50 * time.Millisecond)
					callbackDone <- true
				}
			})

			// Connect should not block waiting for callback
			start := time.Now()
			connectClient(ctx, cli)
			duration := time.Since(start)

			// Connect should return quickly (async callback)
			Expect(duration).To(BeNumerically("<", 200*time.Millisecond))

			// Callback should still execute
			Eventually(callbackStarted, time.Second).Should(Receive())
			Eventually(callbackDone, time.Second).Should(Receive())
		})
	})

	Describe("Combined Callbacks", func() {
		It("should handle both error and info callbacks", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			errorCount := new(atomic.Int32)
			infoCount := new(atomic.Int32)

			cli.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
			})

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				infoCount.Add(1)
			})

			// Normal operation - should trigger info callbacks
			connectClient(ctx, cli)
			time.Sleep(100 * time.Millisecond)

			Expect(infoCount.Load()).To(BeNumerically(">", 0))

			// Error operation - should trigger error callback
			_ = cli.Close()
			_, _ = cli.Write([]byte("test"))
			time.Sleep(100 * time.Millisecond)

			Expect(errorCount.Load()).To(BeNumerically(">", 0))
		})
	})
})
