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

package tcp_test

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	libsck "github.com/nabbar/golib/socket"
	scksrt "github.com/nabbar/golib/socket/server/tcp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Client Callbacks", func() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		srv     scksrt.ServerTcp
		address string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
		address = getTestAddress()
		srv = createSimpleTestServer(ctx, address)
	})

	AfterEach(func() {
		if srv != nil {
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

			var (
				errorCount = new(atomic.Int32)
				lastError  = libatm.NewValue[error]()
			)

			cli.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)

				for _, e := range errs {
					if e != nil {
						lastError.Store(e)
						break
					}
				}
			})

			// Try to write without connection
			msg := []byte("test")
			_, err := cli.Write(msg)
			Expect(err).To(HaveOccurred())
			Expect(cli.Close()).To(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			Expect(errorCount.Load()).To(BeNumerically(">", 0))
			Expect(lastError.Load()).To(HaveOccurred())
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
			_, e := cli.Read(buf)
			Expect(e).To(HaveOccurred())
			Expect(cli.Close()).To(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			Expect(errorCount.Load()).To(BeNumerically(">", int32(0)))
		})

		It("should handle nil error callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			cli.RegisterFuncError(nil)

			// Should not panic
			_, e := cli.Write([]byte("test"))
			Expect(e).To(HaveOccurred())
		})

		It("should allow changing error callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			count1 := new(atomic.Int32)
			cli.RegisterFuncError(func(errs ...error) {
				count1.Add(1)
			})

			// Trigger error
			_, e := cli.Write([]byte("test1"))
			Expect(e).To(HaveOccurred())
			time.Sleep(50 * time.Millisecond)
			firstCount := count1.Load()

			// Change callback
			count2 := new(atomic.Int32)
			cli.RegisterFuncError(func(errs ...error) {
				count2.Add(1)
			})

			// Trigger another error
			_, e = cli.Write([]byte("test2"))
			Expect(e).To(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			// First callback should not increase
			Expect(count1.Load()).To(Equal(firstCount))
			// Second callback should be called
			Expect(count2.Load()).To(BeNumerically(">", 0))
		})

		It("should call error callback with multiple errors", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var errors []error
			var mu sync.Mutex

			cli.RegisterFuncError(func(errs ...error) {
				mu.Lock()
				errors = append(errors, errs...)
				mu.Unlock()
			})

			// Generate multiple errors
			for i := 0; i < 5; i++ {
				_, e := cli.Write([]byte("test"))
				Expect(e).To(HaveOccurred())
			}

			time.Sleep(200 * time.Millisecond)

			mu.Lock()
			errorCount := len(errors)
			mu.Unlock()

			Expect(errorCount).To(BeNumerically(">", 0))
		})
	})

	Describe("RegisterFuncInfo", func() {
		It("should call info callback on connection events", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			eventCount := new(atomic.Int32)
			var events []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				eventCount.Add(1)
				mu.Lock()
				events = append(events, state)
				mu.Unlock()
			})

			err := cli.Connect(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for callbacks
			time.Sleep(200 * time.Millisecond)

			Expect(eventCount.Load()).To(BeNumerically(">", 0))

			mu.Lock()
			hasDialEvent := false
			hasNewEvent := false
			for _, event := range events {
				if event == libsck.ConnectionDial {
					hasDialEvent = true
				}
				if event == libsck.ConnectionNew {
					hasNewEvent = true
				}
			}
			mu.Unlock()

			Expect(hasDialEvent || hasNewEvent).To(BeTrue())
		})

		It("should call info callback on read events", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var events []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				events = append(events, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)

			msg := []byte("Test message\n")
			_ = sendAndReceive(cli, msg)

			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			hasReadEvent := false
			for _, event := range events {
				if event == libsck.ConnectionRead {
					hasReadEvent = true
					break
				}
			}
			mu.Unlock()

			Expect(hasReadEvent).To(BeTrue())
		})

		It("should call info callback on write events", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var events []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				events = append(events, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)

			msg := []byte("Test write\n")
			_, _ = cli.Write(msg)

			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			hasWriteEvent := false
			for _, event := range events {
				if event == libsck.ConnectionWrite {
					hasWriteEvent = true
					break
				}
			}
			mu.Unlock()

			Expect(hasWriteEvent).To(BeTrue())
		})

		It("should call info callback on close events", func() {
			cli := createClient(address)

			var events []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				events = append(events, state)
				mu.Unlock()
			})

			connectClient(ctx, cli)
			_ = cli.Close()

			time.Sleep(200 * time.Millisecond)

			mu.Lock()
			hasCloseEvent := false
			for _, event := range events {
				if event == libsck.ConnectionClose {
					hasCloseEvent = true
					break
				}
			}
			mu.Unlock()

			Expect(hasCloseEvent).To(BeTrue())
		})

		It("should handle nil info callback", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			cli.RegisterFuncInfo(nil)

			// Should not panic
			connectClient(ctx, cli)
			_, _ = cli.Write([]byte("test\n"))
		})

		It("should provide valid addresses in callbacks", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			var localAddr, remoteAddr net.Addr
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				if state == libsck.ConnectionNew {
					mu.Lock()
					localAddr = local
					remoteAddr = remote
					mu.Unlock()
				}
			})

			connectClient(ctx, cli)
			time.Sleep(200 * time.Millisecond)

			mu.Lock()
			Expect(localAddr).ToNot(BeNil())
			Expect(remoteAddr).ToNot(BeNil())
			mu.Unlock()
		})

		It("should track connection lifecycle", func() {
			cli := createClient(address)

			var lifecycle []libsck.ConnState
			var mu sync.Mutex

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				mu.Lock()
				lifecycle = append(lifecycle, state)
				mu.Unlock()
			})

			// Full lifecycle
			connectClient(ctx, cli)
			_, _ = cli.Write([]byte("test\n"))
			buf := make([]byte, 1024)
			_, _ = cli.Read(buf)
			_ = cli.Close()

			time.Sleep(300 * time.Millisecond)

			mu.Lock()
			lifecycleLen := len(lifecycle)
			mu.Unlock()

			Expect(lifecycleLen).To(BeNumerically(">=", 3))
		})
	})

	Describe("Callback Concurrency", func() {
		It("should handle concurrent callback invocations", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			errorCount := new(atomic.Int32)
			infoCount := new(atomic.Int32)

			cli.RegisterFuncError(func(errs ...error) {
				errorCount.Add(1)
				time.Sleep(1 * time.Millisecond) // Simulate processing
			})

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				infoCount.Add(1)
				time.Sleep(1 * time.Millisecond) // Simulate processing
			})

			connectClient(ctx, cli)

			// Generate multiple events concurrently
			done := make(chan bool)
			go func() {
				defer close(done)
				for i := 0; i < 10; i++ {
					msg := []byte("test\n")
					_, _ = cli.Write(msg)
					_, _ = cli.Read(make([]byte, 1024))
				}
			}()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				Fail("Timeout waiting for concurrent callbacks")
			}

			time.Sleep(300 * time.Millisecond)

			// Should have received multiple callbacks
			Expect(infoCount.Load()).To(BeNumerically(">", 0))
		})

		It("should not block on slow callbacks", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			slowCallbackCalled := new(atomic.Bool)

			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				slowCallbackCalled.Store(true)
				time.Sleep(500 * time.Millisecond) // Slow callback
			})

			start := time.Now()
			connectClient(ctx, cli)

			msg := []byte("test\n")
			response := sendAndReceive(cli, msg)
			elapsed := time.Since(start)

			// Communication should not be blocked by slow callback
			Expect(response).To(Equal(msg))
			Expect(elapsed).To(BeNumerically("<", 5*time.Second))

			// Wait for slow callback to complete
			time.Sleep(600 * time.Millisecond)
			Expect(slowCallbackCalled.Load()).To(BeTrue())
		})
	})

	Describe("Callback Edge Cases", func() {
		It("should handle callback that panics", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			cli.RegisterFuncError(func(errs ...error) {
				panic("callback panic")
			})

			// Should not crash the client
			connectClient(ctx, cli)
			msg := []byte("test\n")
			response := sendAndReceive(cli, msg)
			Expect(response).To(Equal(msg))
		})

		It("should handle callback registration after connection", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			connectClient(ctx, cli)

			// Register callback after connection
			callbackCalled := new(atomic.Bool)
			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				callbackCalled.Store(true)
			})

			// Should still receive callbacks for subsequent events
			msg := []byte("test\n")
			_ = sendAndReceive(cli, msg)

			time.Sleep(200 * time.Millisecond)
			Expect(callbackCalled.Load()).To(BeTrue())
		})

		It("should handle multiple callback registrations", func() {
			cli := createClient(address)
			defer func() {
				_ = cli.Close()
			}()

			count1 := new(atomic.Int32)
			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				count1.Add(1)
			})

			count2 := new(atomic.Int32)
			cli.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
				count2.Add(1)
			})

			connectClient(ctx, cli)
			time.Sleep(200 * time.Millisecond)

			// Only the last registered callback should be called
			Expect(count1.Load()).To(Equal(int32(0)))
			Expect(count2.Load()).To(BeNumerically(">", 0))
		})
	})
})
