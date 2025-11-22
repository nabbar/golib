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
 */

package aggregator_test

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner Operations", func() {
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

	Describe("Start()", func() {
		It("should start successfully", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should be idempotent", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			// Start multiple times
			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 5*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should trigger async function", func() {
			writer := newTestWriter()
			counter := newTestCounter()

			cfg := aggregator.Config{
				AsyncTimer: 100 * time.Millisecond,
				AsyncMax:   5,
				AsyncFct: func(ctx context.Context) {
					counter.Inc()
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for async function to be called
			Eventually(func() int {
				return counter.Get()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should trigger sync function", func() {
			writer := newTestWriter()
			counter := newTestCounter()

			cfg := aggregator.Config{
				SyncTimer: 100 * time.Millisecond,
				SyncFct: func(ctx context.Context) {
					counter.Inc()
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for sync function to be called
			Eventually(func() int {
				return counter.Get()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should trigger both async and sync functions", func() {
			writer := newTestWriter()
			asyncCounter := newTestCounter()
			syncCounter := newTestCounter()

			cfg := aggregator.Config{
				AsyncTimer: 50 * time.Millisecond,
				AsyncMax:   5,
				AsyncFct: func(ctx context.Context) {
					asyncCounter.Inc()
				},
				SyncTimer: 50 * time.Millisecond,
				SyncFct: func(ctx context.Context) {
					syncCounter.Inc()
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for both functions to be called
			Eventually(func() int {
				return asyncCounter.Get()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			Eventually(func() int {
				return syncCounter.Get()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should respect AsyncMax limit", func() {
			writer := newTestWriter()
			var activeCount atomic.Int32
			var maxConcurrent atomic.Int32

			cfg := aggregator.Config{
				AsyncTimer: 10 * time.Millisecond,
				AsyncMax:   3,
				AsyncFct: func(ctx context.Context) {
					current := activeCount.Add(1)

					// Update max if needed
					for {
						mx := maxConcurrent.Load()
						if current <= mx {
							break
						}
						if maxConcurrent.CompareAndSwap(mx, current) {
							break
						}
					}

					time.Sleep(100 * time.Millisecond)
					activeCount.Add(-1)
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for some async calls
			time.Sleep(500 * time.Millisecond)

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			// Check that we never exceeded the limit
			mx := maxConcurrent.Load()
			Expect(mx).To(BeNumerically("<=", 3))
		})
	})

	Describe("Stop()", func() {
		It("should stop successfully", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())
		})

		It("should stop when not running", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())
		})

		It("should be idempotent", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Stop multiple times
			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())
		})

		It("should stop async functions", func() {
			writer := newTestWriter()
			counter := newTestCounter()

			cfg := aggregator.Config{
				AsyncTimer: 50 * time.Millisecond,
				AsyncMax:   5,
				AsyncFct: func(ctx context.Context) {
					counter.Inc()
				},
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for some calls
			time.Sleep(200 * time.Millisecond)

			countBeforeStop := counter.Get()
			Expect(countBeforeStop).To(BeNumerically(">=", 2))

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait a bit
			time.Sleep(300 * time.Millisecond)

			// Counter should not increase much after stop
			countAfterStop := counter.Get()
			Expect(countAfterStop - countBeforeStop).To(BeNumerically("<=", 2))
		})
	})

	Describe("Restart()", func() {
		It("should restart successfully", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 3*time.Second, 50*time.Millisecond).Should(BeTrue())

			// Wait a bit before restart to ensure it's fully started
			time.Sleep(200 * time.Millisecond)

			uptime1 := agg.Uptime()

			err = agg.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			// After restart, it should be running again (maybe not immediately)
			time.Sleep(time.Second)

			// Check it's running OR will be running soon
			Eventually(func() bool {
				return agg.IsRunning()
			}, 10*time.Second, 200*time.Millisecond).Should(BeTrue())

			// Uptime should be less than before restart
			uptime2 := agg.Uptime()
			Expect(uptime2).To(BeNumerically("<", uptime1+time.Second))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should restart when not running", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())

			err = agg.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset uptime on restart", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait a bit
			time.Sleep(200 * time.Millisecond)

			uptime1 := agg.Uptime()
			Expect(uptime1).To(BeNumerically(">", 0))

			err = agg.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Uptime should be reset
			uptime2 := agg.Uptime()
			Expect(uptime2).To(BeNumerically("<", uptime1))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("IsRunning()", func() {
		It("should return false initially", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.IsRunning()).To(BeFalse())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return true when running", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return false after stop", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Uptime()", func() {
		It("should return 0 initially", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			Expect(agg.Uptime()).To(Equal(time.Duration(0)))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should increase while running", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			uptime1 := agg.Uptime()

			time.Sleep(100 * time.Millisecond)
			uptime2 := agg.Uptime()

			Expect(uptime2).To(BeNumerically(">", uptime1))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop increasing after stop", func() {
			writer := newTestWriter()
			cfg := aggregator.Config{
				FctWriter: writer.Write,
			}

			agg, err := aggregator.New(ctx, cfg, globalLog)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			uptime1 := agg.Uptime()
			time.Sleep(100 * time.Millisecond)
			uptime2 := agg.Uptime()

			// Uptime should not change significantly after stop
			Expect(uptime2).To(BeNumerically("~", uptime1, 10*time.Millisecond))
		})
	})
})
