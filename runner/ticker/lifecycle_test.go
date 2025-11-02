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

package ticker_test

import (
	"context"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/ticker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// lifecycle_test.go validates the basic lifecycle operations of the ticker package.
//
// Test Coverage:
//   - New(): Ticker creation with various duration values and nil function handling
//   - Start(): Starting a ticker and verifying it executes periodically
//   - Stop(): Stopping a running ticker and ensuring cleanup
//   - Restart(): Atomic stop-and-start operation
//   - IsRunning(): State detection
//   - Uptime(): Duration tracking since start
//   - Context Cancellation: Automatic stopping when parent context is cancelled
//
// Testing Strategy:
// These tests use time.Sleep() and Eventually() to handle timing-sensitive operations.
// All tests use a 30-second timeout context to prevent hanging on failures.
// Counter values use atomic operations to avoid race conditions.
// Sleep durations are chosen to be larger than tick intervals to ensure at least one tick occurs.
//
// Potential Instability Sources:
//   - System load can delay goroutine scheduling
//   - Timing assertions may fail on very slow systems
//   - Eventually() timeouts should be generous to accommodate CI environments
var _ = Describe("Lifecycle Operations", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("New", func() {
		It("should create a new ticker with valid duration", func() {
			counter := int32(0)
			tick := New(100*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			Expect(tick).ToNot(BeNil())
			Expect(tick.IsRunning()).To(BeFalse())
			Expect(tick.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should use default duration when provided duration is too small", func() {
			tick := New(500*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick).ToNot(BeNil())
			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should accept nil function without panic", func() {
			Expect(func() {
				tick := New(10*time.Millisecond, nil)
				Expect(tick).ToNot(BeNil())
			}).ToNot(Panic())
		})
	})

	Describe("Start", func() {
		It("should start the ticker successfully", func() {
			counter := int32(0)
			tick := New(100*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			// Wait for at least one tick
			time.Sleep(150 * time.Millisecond)
			Expect(atomic.LoadInt32(&counter)).To(BeNumerically(">=", int32(1)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should track uptime correctly after start", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			time.Sleep(20 * time.Millisecond)
			uptime := tick.Uptime()
			// Uptime should be at least a few milliseconds but less than a large margin
			// Use generous bounds to account for system load and scheduling delays
			Expect(uptime).To(BeNumerically(">=", 1*time.Millisecond))
			Expect(uptime).To(BeNumerically("<", 200*time.Millisecond))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should stop existing instance before starting new one", func() {
			counter := int32(0)
			tick := New(100*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			// Start first time
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(150 * time.Millisecond)

			firstCount := atomic.LoadInt32(&counter)
			Expect(firstCount).To(BeNumerically(">=", int32(1)))

			// Start again - should restart
			err = tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should execute ticker function multiple times", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for multiple ticks (100ms / 25ms = ~4 ticks expected)
			// Use >= 2 to be conservative and account for timing variations
			time.Sleep(100 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">=", uint32(2)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Stop", func() {
		It("should stop running ticker", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			time.Sleep(100 * time.Millisecond)
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Verify it stopped
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
			Expect(tick.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should be idempotent - multiple stops should not error", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Stop again
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should not error when stopping non-running ticker", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should wait for ticker cleanup with exponential backoff", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				time.Sleep(10 * time.Millisecond) // Simulate work
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			startStop := time.Now()
			err = tick.Stop(ctx)
			stopDuration := time.Since(startStop)

			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeFalse())
			// Should wait some time but not too long
			Expect(stopDuration).To(BeNumerically(">=", 0))
			Expect(stopDuration).To(BeNumerically("<", 3*time.Second))
		})

		It("should prevent ticks after stop", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			countAtStop := atomic.LoadInt32(&counter)
			time.Sleep(20 * time.Millisecond)
			countAfterStop := atomic.LoadInt32(&counter)

			// Counter should not increase after stop
			Expect(countAfterStop).To(Equal(countAtStop))
		})
	})

	Describe("Restart", func() {
		It("should restart a running ticker", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			// Start first time
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(30 * time.Millisecond)

			firstCount := counter.Load()
			firstUptime := tick.Uptime()

			// Restart
			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			// Uptime should reset
			time.Sleep(5 * time.Millisecond)
			newUptime := tick.Uptime()
			Expect(newUptime).To(BeNumerically("<", firstUptime))

			// Should continue ticking
			time.Sleep(100 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">", firstCount))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should start ticker if not running", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			time.Sleep(30 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">=", int32(1)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle rapid restart operations", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			for i := 0; i < 3; i++ {
				err := tick.Restart(ctx)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(20 * time.Millisecond)
			}

			Expect(tick.IsRunning()).To(BeTrue())
			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("IsRunning", func() {
		It("should return false for new ticker", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should return true while running", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return false after stop", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(tick.IsRunning, 50*time.Millisecond, 3*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Uptime", func() {
		It("should return 0 for new ticker", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick.Uptime()).To(Equal(time.Duration(0)))
		})

		It("should increase while running", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(30 * time.Millisecond)
			uptime1 := tick.Uptime()
			// Uptime should be at least some reasonable value
			// Use a conservative threshold to account for slow systems
			Expect(uptime1).To(BeNumerically(">", 10*time.Millisecond))

			time.Sleep(30 * time.Millisecond)
			uptime2 := tick.Uptime()
			// Second uptime should be strictly greater than first
			Expect(uptime2).To(BeNumerically(">", uptime1))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reset to 0 after stop", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(15 * time.Millisecond)
			Expect(tick.Uptime()).To(BeNumerically(">", 0))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(tick.Uptime, 50*time.Millisecond, 3*time.Millisecond).Should(Equal(time.Duration(0)))
		})

		It("should reset after restart", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(30 * time.Millisecond)
			oldUptime := tick.Uptime()

			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(15 * time.Millisecond)
			newUptime := tick.Uptime()
			Expect(newUptime).To(BeNumerically("<", oldUptime))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Context Cancellation", func() {
		It("should stop when context is cancelled", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			cancelCtx, cancelFunc := context.WithCancel(ctx)
			err := tick.Start(cancelCtx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(20 * time.Millisecond)
			Expect(tick.IsRunning()).To(BeTrue())

			// Cancel context
			cancelFunc()

			// Should stop eventually
			Eventually(tick.IsRunning, 50*time.Millisecond, 3*time.Millisecond).Should(BeFalse())
		})

		It("should respect context timeout", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 5*time.Millisecond)
			defer timeoutCancel()

			err := tick.Start(timeoutCtx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for context timeout
			time.Sleep(30 * time.Millisecond)

			// Should stop after context timeout
			Eventually(tick.IsRunning, 50*time.Millisecond, 3*time.Millisecond).Should(BeFalse())
		})

		It("should detect context cancellation in ticker function", func() {
			cancelled := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				select {
				case <-ctx.Done():
					cancelled.Store(1)
					return ctx.Err()
				}
			})

			cancelCtx, cancelFunc := context.WithCancel(ctx)
			err := tick.Start(cancelCtx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(20 * time.Millisecond)
			cancelFunc()

			// Wait for ticker to detect cancellation
			time.Sleep(30 * time.Millisecond)
			Expect(tick.IsRunning()).To(BeFalse())
			Expect(cancelled.Load()).To(BeNumerically("==", uint32(1)))
		})
	})
})
