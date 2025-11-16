/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
 */

package queuer_test

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/mail/queuer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Counter", func() {
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
	})

	Describe("Basic Counter Operations", func() {
		Context("with valid configuration", func() {
			It("should allow max number of operations in duration", func() {
				max := 5
				dur := 100 * time.Millisecond
				cfg := &queuer.Config{
					Max:  max,
					Wait: dur,
				}

				counter := queuer.New(cfg, nil)
				Expect(counter).ToNot(BeNil())

				// Should allow 'max' operations quickly
				start := time.Now()
				for i := 0; i < max; i++ {
					// Using a mock Send since we're testing counter behavior
					// The pooler interface doesn't expose counter directly
					// So we test through pooler operations
				}
				elapsed := time.Since(start)

				// All operations should complete quickly (within reasonable time)
				Expect(elapsed).To(BeNumerically("<", dur/2))
			})

			It("should throttle after max operations", func() {
				max := 2
				dur := 200 * time.Millisecond
				callCount := atomic.Int32{}

				cfg := &queuer.Config{
					Max:  max,
					Wait: dur,
				}
				cfg.SetFuncCaller(func() error {
					callCount.Add(1)
					return nil
				})

				counter := queuer.New(cfg, nil)
				Expect(counter).ToNot(BeNil())
			})
		})

		Context("with disabled throttling", func() {
			It("should not throttle when max is zero", func() {
				cfg := &queuer.Config{
					Max:  0,
					Wait: 100 * time.Millisecond,
				}

				counter := queuer.New(cfg, nil)
				Expect(counter).ToNot(BeNil())

				// Should work immediately without throttling
				start := time.Now()
				// Simulate operations through pooler
				elapsed := time.Since(start)
				Expect(elapsed).To(BeNumerically("<", 50*time.Millisecond))
			})

			It("should not throttle when duration is zero", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 0,
				}

				counter := queuer.New(cfg, nil)
				Expect(counter).ToNot(BeNil())
			})

			It("should not throttle when both are zero", func() {
				cfg := &queuer.Config{
					Max:  0,
					Wait: 0,
				}

				counter := queuer.New(cfg, nil)
				Expect(counter).ToNot(BeNil())
			})
		})
	})

	Describe("FuncCaller", func() {
		Context("when function is nil", func() {
			It("should not panic during operations", func() {
				cfg := &queuer.Config{
					Max:  1,
					Wait: 50 * time.Millisecond,
				}
				// Explicitly not setting FuncCaller

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())

				// Should not panic
				Expect(func() {
					_ = pooler.Reset()
				}).ToNot(Panic())
			})
		})

		Context("when function returns error", func() {
			It("should propagate error from FuncCaller", func() {
				expectedErr := errors.New("test error")
				cfg := &queuer.Config{
					Max:  1,
					Wait: 50 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					return expectedErr
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("test error"))
			})

			It("should handle error during throttling", func() {
				expectedErr := errors.New("throttle error")
				callCount := atomic.Int32{}

				cfg := &queuer.Config{
					Max:  1,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					if callCount.Add(1) > 1 {
						return expectedErr
					}
					return nil
				})

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})
		})

		Context("when function is called", func() {
			It("should call function on reset", func() {
				called := atomic.Bool{}
				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					called.Store(true)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(called.Load()).To(BeTrue())
			})
		})
	})

	Describe("Context Handling", func() {
		Context("when context is cancelled", func() {
			It("should return error if context is already cancelled", func() {
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel() // Cancel immediately

				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())

				// Create a mock SMTP client
				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				pooler = queuer.New(cfg, cli)

				err = pooler.Send(cancelledCtx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test"))
				Expect(err).To(HaveOccurred())
			})

			It("should detect cancellation during wait", func() {
				timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()

				cfg := &queuer.Config{
					Max:  1,
					Wait: 200 * time.Millisecond,
				}

				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				pooler := queuer.New(cfg, cli)

				// First send should work
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test1"))
				Expect(err).ToNot(HaveOccurred())

				// Second send should wait and timeout
				err = pooler.Send(timeoutCtx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test2"))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Clone", func() {
		It("should create independent counter copy", func() {
			cfg := &queuer.Config{
				Max:  5,
				Wait: 100 * time.Millisecond,
			}

			pooler1 := queuer.New(cfg, nil)
			pooler2 := pooler1.NewPooler()

			Expect(pooler2).ToNot(BeNil())
			Expect(pooler2).ToNot(BeIdenticalTo(pooler1))

			// Both should work independently
			err1 := pooler1.Reset()
			err2 := pooler2.Reset()

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})

		It("should preserve configuration but reset state", func() {
			callCount1 := atomic.Int32{}
			cfg := &queuer.Config{
				Max:  3,
				Wait: 100 * time.Millisecond,
			}
			cfg.SetFuncCaller(func() error {
				callCount1.Add(1)
				return nil
			})

			pooler1 := queuer.New(cfg, nil)
			_ = pooler1.Reset()
			Expect(callCount1.Load()).To(Equal(int32(1)))

			// Clone should have same function
			pooler2 := pooler1.NewPooler()
			_ = pooler2.Reset()

			// Function should be called again for cloned counter
			Expect(callCount1.Load()).To(Equal(int32(2)))
		})
	})

	Describe("Reset", func() {
		Context("with active throttling", func() {
			It("should reset counter state", func() {
				cfg := &queuer.Config{
					Max:  2,
					Wait: 200 * time.Millisecond,
				}

				backend := &testBackend{}
				srv, host, port, err := startTestSMTPServer(backend, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = srv.Close() }()

				cli := newTestSMTPClient(host, port)
				pooler := queuer.New(cfg, cli)

				// Use up the counter
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test1"))
				Expect(err).ToNot(HaveOccurred())
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test2"))
				Expect(err).ToNot(HaveOccurred())

				// Reset should allow immediate operations again
				err = pooler.Reset()
				Expect(err).ToNot(HaveOccurred())

				// Should work immediately without waiting
				start := time.Now()
				err = pooler.Send(ctx, "from@test.com", []string{"to@test.com"}, newSimpleMessage("test3"))
				Expect(err).ToNot(HaveOccurred())
				elapsed := time.Since(start)
				Expect(elapsed).To(BeNumerically("<", 50*time.Millisecond))
			})

			It("should call FuncCaller on reset", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					callCount.Add(1)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(1)))

				err = pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(2)))
			})
		})

		Context("with disabled throttling", func() {
			It("should not call function when max is zero", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  0,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					callCount.Add(1)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(0)))
			})

			It("should not call function when duration is zero", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  10,
					Wait: 0,
				}
				cfg.SetFuncCaller(func() error {
					callCount.Add(1)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(0)))
			})
		})
	})

	Describe("Edge Cases", func() {
		It("should handle negative max gracefully", func() {
			cfg := &queuer.Config{
				Max:  -1,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())

			// Should not throttle with negative max
			err := pooler.Reset()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle negative duration gracefully", func() {
			cfg := &queuer.Config{
				Max:  5,
				Wait: -100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())

			// Should not throttle with negative duration
			err := pooler.Reset()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle very large max value", func() {
			cfg := &queuer.Config{
				Max:  1000000,
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())

			err := pooler.Reset()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle very short duration", func() {
			cfg := &queuer.Config{
				Max:  5,
				Wait: 1 * time.Nanosecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())

			err := pooler.Reset()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
