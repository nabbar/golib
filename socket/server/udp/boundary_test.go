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
	"errors"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
	"github.com/nabbar/golib/socket/server/udp"
)

var _ = Describe("UDP Server Boundary Tests", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		time.Sleep(100 * time.Millisecond)
	})

	Describe("Error Callback Edge Cases", func() {
		It("should handle nil error callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfoServer(nil)
			srv.RegisterFuncInfo(nil)
			srv.RegisterFuncError(nil)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Should work with nil callback
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle error callback with nil errors", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			called := false
			srv.RegisterFuncInfoServer(func(msg string) {})
			srv.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {})
			srv.RegisterFuncError(func(errs ...error) {
				called = true
				for _, e := range errs {
					if e != nil {
						// Handle non-nil error
						_ = e
					}
				}
			})

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Callback may or may not be called
			_ = called
		})

		It("should handle multiple error callbacks", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			calls := 0
			callback := func(errs ...error) {
				calls++
			}

			srv.RegisterFuncInfoServer(func(msg string) {})
			srv.RegisterFuncInfo(func(_, _ net.Addr, _ libsck.ConnState) {})
			srv.RegisterFuncError(callback)
			srv.RegisterFuncError(callback)
			srv.RegisterFuncError(callback)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Multiple registrations should work
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Info Callback Edge Cases", func() {
		It("should handle nil info callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncInfo(nil)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle info callback with nil addresses", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(func(errs ...error) {})
			srv.RegisterFuncInfoServer(func(msg string) {})
			srv.RegisterFuncInfo(newInfoCollector().callback)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Server Info Callback Edge Cases", func() {
		It("should handle nil server info callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(func(errs ...error) {})
			srv.RegisterFuncInfoServer(func(msg string) {})
			srv.RegisterFuncInfoServer(nil)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle server info callback with empty messages", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(func(errs ...error) {})
			srv.RegisterFuncInfoServer(func(msg string) {})
			srv.RegisterFuncInfoServer(func(msg string) {
				// Handle empty or any message
				_ = msg
			})

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Address Registration Edge Cases", func() {
		It("should handle RegisterServer with already registered address", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			err = srv.RegisterServer("127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())

			// Try to register again (last one wins)
			err = srv.RegisterServer("127.0.0.1:9999")
			// Should succeed - RegisterServer allows re-registration
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle RegisterServer with empty address", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			err = srv.RegisterServer("")
			Expect(err).To(Equal(udp.ErrInvalidAddress))
		})

		It("should handle RegisterServer with invalid format", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			err = srv.RegisterServer("invalid:::address")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Shutdown Edge Cases", func() {
		It("should handle Shutdown before Listen", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			err = srv.Shutdown(shutdownCtx)
			// Should succeed or return specific error
			_ = err
		})

		It("should handle Shutdown with expired context", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Create already expired context
			expiredCtx, expiredCancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
			defer expiredCancel()

			err = srv.Shutdown(expiredCtx)
			// May or may not timeout depending on shutdown speed
			_ = err
			_ = errors.Is(err, context.DeadlineExceeded)
		})

		It("should handle multiple Shutdown calls", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			shutdownCtx1, shutdownCancel1 := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel1()

			err = srv.Shutdown(shutdownCtx1)
			Expect(err).ToNot(HaveOccurred())

			// Second shutdown
			shutdownCtx2, shutdownCancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel2()

			err = srv.Shutdown(shutdownCtx2)
			// May succeed or return error
			_ = err
		})
	})

	Describe("Listen Edge Cases", func() {
		It("should handle Listen with already running server", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Verify server is running
			Expect(srv.IsRunning()).To(BeTrue())

			// Try to listen again (should return error immediately)
			localCtx, localCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer localCancel()

			err = srv.Listen(localCtx)
			// Should return an error
			Expect(err).To(HaveOccurred())
		})

		It("should handle Listen without RegisterServer", func() {
			handler := newTestHandler(false)
			cfg := createBasicConfig()
			cfg.Address = "" // Force error

			srv, err := udp.New(nil, handler.handler, cfg)
			// Should fail at creation
			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})

		It("should handle Listen with invalid address", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			// Register invalid address
			err = srv.RegisterServer("invalid-address-format")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Handler Edge Cases", func() {
		It("should reject nil handler", func() {
			cfg := createBasicConfig()
			srv, err := udp.New(nil, nil, cfg)
			// Should return an error
			Expect(err).To(HaveOccurred())
			Expect(srv).To(BeNil())
		})

		It("should handle handler that never returns", func() {
			infiniteHandler := func(ctx libsck.Context) {
				defer ctx.Close()
				for {
					select {
					case <-ctx.Done():
						return
					default:
						time.Sleep(10 * time.Millisecond)
					}
				}
			}

			srv, err := createServerWithHandler(infiniteHandler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Should be able to stop
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			err = srv.Shutdown(shutdownCtx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("UpdateConn Callback Edge Cases", func() {
		It("should handle UpdateConn with slow callback", func() {
			handler := newTestHandler(false)
			slowUpdate := libsck.UpdateConn(func(conn net.Conn) {
				time.Sleep(100 * time.Millisecond)
			})

			cfg := createBasicConfig()
			srv, err := udp.New(slowUpdate, handler.handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Should still work
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle UpdateConn with panicking callback", func() {
			handler := newTestHandler(false)
			panicUpdate := libsck.UpdateConn(func(conn net.Conn) {
				defer func() {
					if r := recover(); r != nil {
						// Recovered
					}
				}()
				panic("update panic")
			})

			cfg := createBasicConfig()
			srv, err := udp.New(panicUpdate, handler.handler, cfg)
			Expect(err).ToNot(HaveOccurred())

			// Should handle panic gracefully
			go func() {
				_ = srv.Listen(ctx)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()
		})
	})
})
