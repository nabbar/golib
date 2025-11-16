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
	"errors"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/mail/queuer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Configuration Fields", func() {
		Context("with default values", func() {
			It("should allow zero max", func() {
				cfg := &queuer.Config{
					Max:  0,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should allow zero wait duration", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 0,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should allow both zero", func() {
				cfg := &queuer.Config{
					Max:  0,
					Wait: 0,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})
		})

		Context("with positive values", func() {
			It("should accept valid configuration", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should accept large max value", func() {
				cfg := &queuer.Config{
					Max:  1000000,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should accept large duration", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 1 * time.Hour,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should accept small duration", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 1 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})
		})

		Context("with negative values", func() {
			It("should handle negative max", func() {
				cfg := &queuer.Config{
					Max:  -5,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should handle negative duration", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: -100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})
		})
	})

	Describe("FuncCaller", func() {
		Context("setting function", func() {
			It("should allow setting nil function", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				Expect(func() {
					cfg.SetFuncCaller(nil)
				}).ToNot(Panic())

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should allow setting function that returns nil", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				called := atomic.Bool{}
				cfg.SetFuncCaller(func() error {
					called.Store(true)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())

				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(called.Load()).To(BeTrue())
			})

			It("should allow setting function that returns error", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				expectedErr := errors.New("test error")
				cfg.SetFuncCaller(func() error {
					return expectedErr
				})

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())

				err := pooler.Reset()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("test error"))
			})

			It("should allow updating function", func() {
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}

				count := atomic.Int32{}

				// Set first function
				cfg.SetFuncCaller(func() error {
					count.Add(1)
					return nil
				})

				pooler := queuer.New(cfg, nil)
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(count.Load()).To(Equal(int32(1)))

				// Update function
				cfg.SetFuncCaller(func() error {
					count.Add(10)
					return nil
				})

				// Note: This won't affect existing pooler
				// New pooler would get new function
				pooler2 := queuer.New(cfg, nil)
				err = pooler2.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(count.Load()).To(Equal(int32(11)))
			})
		})

		Context("function behavior", func() {
			It("should be called on reset", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  10,
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

			It("should not be called when throttling is disabled", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  0,
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

			It("should be preserved in clone", func() {
				callCount := atomic.Int32{}
				cfg := &queuer.Config{
					Max:  10,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					callCount.Add(1)
					return nil
				})

				pooler1 := queuer.New(cfg, nil)
				pooler2 := pooler1.NewPooler()

				err := pooler1.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(1)))

				err = pooler2.Reset()
				Expect(err).ToNot(HaveOccurred())
				Expect(callCount.Load()).To(Equal(int32(2)))
			})
		})
	})

	Describe("Configuration Scenarios", func() {
		Context("rate limiting configurations", func() {
			It("should configure for high throughput", func() {
				cfg := &queuer.Config{
					Max:  1000,
					Wait: 1 * time.Second,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should configure for strict limiting", func() {
				cfg := &queuer.Config{
					Max:  1,
					Wait: 5 * time.Second,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should configure for no limiting", func() {
				cfg := &queuer.Config{
					Max:  0,
					Wait: 0,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})

			It("should configure for burst with recovery", func() {
				cfg := &queuer.Config{
					Max:  100,
					Wait: 100 * time.Millisecond,
				}

				pooler := queuer.New(cfg, nil)
				Expect(pooler).ToNot(BeNil())
			})
		})

		Context("callback scenarios", func() {
			It("should support logging callback", func() {
				logs := make([]string, 0)
				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					logs = append(logs, "throttle triggered")
					return nil
				})

				pooler := queuer.New(cfg, nil)
				_ = pooler.Reset()

				Expect(logs).To(HaveLen(1))
				Expect(logs[0]).To(Equal("throttle triggered"))
			})

			It("should support counter callback", func() {
				counter := atomic.Int64{}
				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					counter.Add(1)
					return nil
				})

				pooler := queuer.New(cfg, nil)

				for i := 0; i < 5; i++ {
					_ = pooler.Reset()
				}

				Expect(counter.Load()).To(Equal(int64(5)))
			})

			It("should support error injection for testing", func() {
				injectError := atomic.Bool{}
				cfg := &queuer.Config{
					Max:  5,
					Wait: 100 * time.Millisecond,
				}
				cfg.SetFuncCaller(func() error {
					if injectError.Load() {
						return errors.New("injected error")
					}
					return nil
				})

				pooler := queuer.New(cfg, nil)

				// Normal operation
				err := pooler.Reset()
				Expect(err).ToNot(HaveOccurred())

				// Inject error
				injectError.Store(true)
				err = pooler.Reset()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("injected error"))
			})
		})
	})

	Describe("Configuration Edge Cases", func() {
		It("should handle max int values", func() {
			cfg := &queuer.Config{
				Max:  int(^uint(0) >> 1), // Max int
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())
		})

		It("should handle min int values", func() {
			cfg := &queuer.Config{
				Max:  -int(^uint(0)>>1) - 1, // Min int
				Wait: 100 * time.Millisecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())
		})

		It("should handle max duration", func() {
			cfg := &queuer.Config{
				Max:  10,
				Wait: 1<<63 - 1, // Max duration
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())
		})

		It("should handle nanosecond precision", func() {
			cfg := &queuer.Config{
				Max:  10,
				Wait: 1 * time.Nanosecond,
			}

			pooler := queuer.New(cfg, nil)
			Expect(pooler).ToNot(BeNil())
		})
	})
})
