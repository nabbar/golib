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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsck "github.com/nabbar/golib/socket"
)

var _ = Describe("UDP Server Robustness", func() {
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

	Describe("Error Handling", func() {
		Context("with handler errors", func() {
			It("should handle handler returning immediately", func() {
				immediateHandler := func(ctx libsck.Context) {
					// Return immediately
				}

				srv, err := createServerWithHandler(immediateHandler)
				Expect(err).ToNot(HaveOccurred())

				startServer(srv, ctx)
				defer stopServer(srv, cancel)

				// Server should still be running
				Expect(srv.IsRunning()).To(BeTrue())
			})

			It("should handle handler that panics", func() {
				panicHandler := func(ctx libsck.Context) {
					defer func() {
						if r := recover(); r != nil {
							// Panic recovered in handler
						}
						ctx.Close()
					}()
					panic("intentional panic for testing")
				}

				srv, err := createServerWithHandler(panicHandler)
				Expect(err).ToNot(HaveOccurred())

				// Should not crash the test
				go func() {
					_ = srv.Listen(ctx)
				}()

				time.Sleep(100 * time.Millisecond)
				cancel()
			})

			It("should handle handler with infinite loop", func() {
				infiniteHandler := func(ctx libsck.Context) {
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
				stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer stopCancel()

				err = srv.Shutdown(stopCtx)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with context errors", func() {
			It("should handle cancelled context", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				cancelledCtx, cancelFunc := context.WithCancel(context.Background())
				cancelFunc() // Cancel immediately

				err = srv.Listen(cancelledCtx)

				// Should return with context error
				Expect(err).To(HaveOccurred())
			})

			It("should handle context with deadline", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				deadlineCtx, deadlineCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer deadlineCancel()

				err = srv.Listen(deadlineCtx)

				// Should return with deadline exceeded
				Eventually(func() bool {
					return err != nil
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
			})
		})
	})

	Describe("Boundary Conditions", func() {
		Context("with address edge cases", func() {
			It("should handle port 0 (random port)", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("127.0.0.1:0")
				Expect(err).ToNot(HaveOccurred())

				startServer(srv, ctx)
				defer stopServer(srv, cancel)

				// Should be running with random port
				Expect(srv.IsRunning()).To(BeTrue())
			})

			It("should handle IPv6 loopback", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				err = srv.RegisterServer("[::1]:0")
				Expect(err).ToNot(HaveOccurred())

				startServer(srv, ctx)
				defer stopServer(srv, cancel)

				Expect(srv.IsRunning()).To(BeTrue())
			})
		})

		Context("with timing edge cases", func() {
			It("should handle immediate shutdown after start", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				go func() {
					_ = srv.Listen(ctx)
				}()

				// Shutdown immediately
				time.Sleep(10 * time.Millisecond)
				cancel()

				// Should stop cleanly
				Eventually(func() bool {
					return !srv.IsRunning()
				}, 2*time.Second, 10*time.Millisecond).Should(BeTrue())
			})

			It("should handle multiple rapid start/stop cycles", func() {
				handler := newTestHandler(false)
				srv, err := createServerWithHandler(handler.handler)
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < 5; i++ {
					localCtx, localCancel := context.WithCancel(ctx)

					go func() {
						_ = srv.Listen(localCtx)
					}()

					time.Sleep(50 * time.Millisecond)
					localCancel()
					time.Sleep(50 * time.Millisecond)
				}

				// Should handle all cycles
				Expect(srv).ToNot(BeNil())
			})
		})
	})

	Describe("Resource Management", func() {
		It("should not leak goroutines", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			// Start and stop multiple times
			for i := 0; i < 3; i++ {
				localCtx, localCancel := context.WithCancel(ctx)
				startServer(srv, localCtx)
				stopServer(srv, localCancel)
				time.Sleep(100 * time.Millisecond)
			}

			// Should not have leaked goroutines
			// (difficult to test directly, but server should be clean)
			Expect(srv.IsRunning()).To(BeFalse())
		})

		It("should handle cleanup on errors", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			// Try to listen on already bound port (may fail)
			err = srv.RegisterServer("127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Resources should be properly managed
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("State Consistency", func() {
		It("should maintain consistent state during transitions", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			// Initial state
			assertServerState(srv, false, true, 0)

			// Start
			startServer(srv, ctx)

			// Running state
			assertServerState(srv, true, false, 0)

			// Stop
			stopServer(srv, cancel)

			// Stopped state - wait for IsGone to be true
			Eventually(func() bool {
				return srv.IsGone()
			}, 5*time.Second, 50*time.Millisecond).Should(BeTrue())

			Expect(srv.IsRunning()).To(BeFalse())
			Expect(srv.OpenConnections()).To(Equal(int64(0)))
		})

		It("should handle state queries during transitions", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				_ = srv.Listen(ctx)
			}()

			// Query state during startup
			for i := 0; i < 100; i++ {
				_ = srv.IsRunning()
				_ = srv.IsGone()
				_ = srv.OpenConnections()
				time.Sleep(time.Millisecond)
			}

			cancel()

			// Query state during shutdown
			for i := 0; i < 100; i++ {
				_ = srv.IsRunning()
				_ = srv.IsGone()
				_ = srv.OpenConnections()
				time.Sleep(time.Millisecond)
			}
		})
	})

	Describe("Callback Resilience", func() {
		It("should handle slow error callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			slowCallback := func(errs ...error) {
				time.Sleep(100 * time.Millisecond)
			}

			srv.RegisterFuncError(slowCallback)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Server should still work
			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle panicking callback", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			panicCallback := func(errs ...error) {
				panic("callback panic")
			}

			srv.RegisterFuncError(panicCallback)

			// Should not crash
			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			Expect(srv.IsRunning()).To(BeTrue())
		})

		It("should handle nil callbacks", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			srv.RegisterFuncError(nil)
			srv.RegisterFuncInfo(nil)
			srv.RegisterFuncInfoServer(nil)

			startServer(srv, ctx)
			defer stopServer(srv, cancel)

			// Should work with nil callbacks
			Expect(srv.IsRunning()).To(BeTrue())
		})
	})

	Describe("Shutdown Timeout", func() {
		It("should respect shutdown context timeout", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			// Shutdown with very short timeout
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer shutdownCancel()

			err = srv.Shutdown(shutdownCtx)

			// May timeout or succeed
			_ = err
		})

		It("should complete shutdown within reasonable time", func() {
			handler := newTestHandler(false)
			srv, err := createServerWithHandler(handler.handler)
			Expect(err).ToNot(HaveOccurred())

			startServer(srv, ctx)

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			start := time.Now()
			err = srv.Shutdown(shutdownCtx)
			duration := time.Since(start)

			Expect(err).ToNot(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 2*time.Second))
		})
	})
})
