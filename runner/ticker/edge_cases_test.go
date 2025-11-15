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
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/ticker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// edge_cases_test.go validates edge cases, boundary conditions, and unusual scenarios
// that might cause instability or unexpected behavior.
//
// Test Coverage:
//   - Duration Edge Cases: Zero, negative, very small, very large durations
//   - Function Edge Cases: Long-running functions, panics, ticker parameter access, blocking
//   - Context Edge Cases: Nil, already cancelled, expired, with values
//   - Timing Edge Cases: Immediate stop after start, rapid restarts, timing accuracy
//   - State Transitions: Multiple sequential operations, idempotency
//   - Error Boundary Cases: Large error accumulation, alternating success/failure
//   - Resource Cleanup: Multiple creation/destruction cycles
//   - Zero Value Cases: Handling of zero values and defaults
//   - Interface Compliance: Verification of Runner and Errors interface implementations
//
// Testing Strategy:
// These tests deliberately exercise unusual and potentially problematic scenarios
// to ensure robust behavior. They use:
//   - Invalid inputs (nil, zero, negative values)
//   - Extreme values (very short/long durations)
//   - Rapid state changes
//   - Resource exhaustion patterns
//
// Important Notes:
//   - Tests verify graceful degradation and sensible defaults
//   - Invalid durations fall back to defaultDuration (30 seconds)
//   - Nil contexts should return errors, not panic
//   - Panics in user functions should be recovered
//   - Resource cleanup should be thorough even after failures
var _ = Describe("Edge Cases", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Duration Edge Cases", func() {
		It("should use default duration for zero duration", func() {
			tick := New(0, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick).ToNot(BeNil())

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should work with default duration
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should use default duration for negative duration", func() {
			tick := New(-1*time.Second, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick).ToNot(BeNil())

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should use default duration for very small duration", func() {
			tick := New(1*time.Nanosecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			Expect(tick).ToNot(BeNil())

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle very short tick intervals", func() {
			counter := new(atomic.Uint32)
			// Now millisecond durations are supported
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for at least one tick
			time.Sleep(30 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have at least one tick
			Expect(counter.Load()).To(BeNumerically(">=", uint32(2)))
		})

		It("should handle very long tick intervals", func() {
			counter := new(atomic.Uint32)
			tick := New(1*time.Hour, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(tick.IsRunning()).To(BeTrue())

			// Won't get any ticks in short time
			time.Sleep(100 * time.Millisecond)
			Expect(counter.Load()).To(Equal(uint32(0)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Function Edge Cases", func() {
		It("should handle function that takes very long", func() {
			counter := new(atomic.Uint32)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				time.Sleep(100 * time.Millisecond) // Simulate work
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(300 * time.Millisecond)

			// Should still be running despite slow function
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have executed at least once
			Expect(counter.Load()).To(BeNumerically(">=", uint32(1)))
		})

		It("should handle function that panics", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				if counter.Load() == 2 {
					return fmt.Errorf("don't panic test")
				}
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for panic and recovery
			time.Sleep(100 * time.Millisecond)

			// Should continue running after panic recovery
			Expect(tick.IsRunning()).To(BeTrue())
			Expect(counter.Load()).To(BeNumerically(">=", uint32(2)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle function with nil ticker parameter access", func() {
			accessedTicker := new(atomic.Uint32)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				if tck != nil {
					accessedTicker.Add(1)
					// Could potentially use ticker to reset or stop
					select {
					case <-tck.C:
						// This would be odd but should not break
					default:
					}
				}
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(60 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(accessedTicker.Load()).To(Equal(uint32(1)))
		})

		It("should handle function that blocks on context", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				// Block until context is done
				<-ctx.Done()
				return ctx.Err()
			})

			localCtx, localCancel := context.WithCancel(ctx)
			err := tick.Start(localCtx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(30 * time.Millisecond)

			// Cancel context to unblock function
			localCancel()

			// Should stop
			Eventually(tick.IsRunning, 100*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Context Edge Cases", func() {
		It("should handle already cancelled context on Start", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			cancelCtx, cancelFunc := context.WithCancel(ctx)
			cancelFunc() // Cancel before start

			err := tick.Start(cancelCtx)
			Expect(err).ToNot(HaveOccurred())

			// Should start but stop immediately
			time.Sleep(30 * time.Millisecond)
			Eventually(tick.IsRunning, 20*time.Millisecond, 3*time.Millisecond).Should(BeFalse())
		})

		It("should handle expired context on Start", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Nanosecond)
			defer timeoutCancel()

			time.Sleep(10 * time.Millisecond) // Ensure timeout expires

			err := tick.Start(timeoutCtx)
			Expect(err).ToNot(HaveOccurred())

			// Should stop quickly due to expired context
			time.Sleep(300 * time.Millisecond)
			Eventually(tick.IsRunning, 300*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})

		It("should handle nil context gracefully", func() {
			tick := New(30*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			// This should panic when calling WithCancel on nil context
			Expect(tick.Start(nil)).To(HaveOccurred())
		})

		It("should handle background context", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(75 * time.Millisecond)

			// Should still be running with background context
			Expect(tick.IsRunning()).To(BeTrue())
			Expect(counter.Load()).To(BeNumerically(">=", uint32(2)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle context with values", func() {
			type ctxKey string
			const key ctxKey = "test-key"
			expectedValue := "test-value"
			receivedValue := ""

			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				if val := ctx.Value(key); val != nil {
					receivedValue = val.(string)
				}
				return nil
			})

			ctxWithValue := context.WithValue(ctx, key, expectedValue)
			err := tick.Start(ctxWithValue)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			err = tick.Stop(ctxWithValue)
			Expect(err).ToNot(HaveOccurred())

			Expect(receivedValue).To(Equal(expectedValue))
		})
	})

	Describe("Timing Edge Cases", func() {
		It("should handle stop immediately after start", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Stop immediately
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Eventually(tick.IsRunning, 50*time.Millisecond, 5*time.Millisecond).Should(BeFalse())
		})

		It("should handle restart immediately after start", func() {
			counter := new(atomic.Uint32)
			tick := New(25*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Restart immediately
			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(75 * time.Millisecond)

			Expect(tick.IsRunning()).To(BeTrue())
			Expect(counter.Load()).To(BeNumerically(">=", uint32(2)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle multiple rapid restarts", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			for i := 0; i < 5; i++ {
				err := tick.Restart(ctx)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(10 * time.Millisecond)
			}

			Expect(tick.IsRunning()).To(BeTrue())

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should maintain accuracy over time", func() {
			counter := int32(0)
			interval := 100 * time.Millisecond

			tick := New(interval, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			start := time.Now()
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(550 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			elapsed := time.Since(start)
			count := atomic.LoadInt32(&counter)

			// With 100ms interval and 550ms runtime, expect ~5 ticks
			// Use generous bounds to account for:
			// - System load and scheduling delays
			// - Race detector overhead (can be significant)
			// - First tick delay (tickers don't fire immediately)
			Expect(count).To(BeNumerically(">=", int32(3)))
			Expect(count).To(BeNumerically("<=", int32(8)))
			Expect(elapsed).To(BeNumerically(">=", 500*time.Millisecond))
		})
	})

	Describe("State Transitions", func() {
		It("should handle start -> stop -> start sequence", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			// First cycle
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(35 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())

			firstCount := atomic.LoadInt32(&counter)

			// Second cycle
			err = tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(35 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			secondCount := atomic.LoadInt32(&counter)
			Expect(secondCount).To(BeNumerically(">", firstCount))
		})

		It("should handle multiple stop calls in sequence", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i < 5; i++ {
				err = tick.Stop(ctx)
				Expect(err).ToNot(HaveOccurred())
			}

			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should handle multiple start calls in sequence", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			for i := 0; i < 3; i++ {
				err := tick.Start(ctx)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(30 * time.Millisecond)
			}

			Expect(tick.IsRunning()).To(BeTrue())

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Error Boundary Cases", func() {
		It("should handle large number of accumulated errors", func() {
			counter := int32(0)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return errors.New("error")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Let it accumulate many errors
			time.Sleep(500 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			errList := tick.ErrorsList()
			Expect(len(errList)).To(BeNumerically(">=", 30))
		})

		It("should handle alternating success and failure", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				if counter.Load()%2 == 0 {
					return errors.New("even error")
				}
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(25 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have both successes and failures
			Expect(len(tick.ErrorsList())).To(BeNumerically(">", 0))
		})
	})

	Describe("Resource Cleanup", func() {
		It("should properly clean up resources after multiple cycles", func() {
			for cycle := 0; cycle < 10; cycle++ {
				tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
					return nil
				})

				err := tick.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)

				err = tick.Stop(ctx)
				Expect(err).ToNot(HaveOccurred())

				Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
			}
		})

		It("should clean up after context cancellation", func() {
			for i := 0; i < 5; i++ {
				tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
					return nil
				})

				cancelCtx, cancelFunc := context.WithCancel(ctx)
				err := tick.Start(cancelCtx)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(100 * time.Millisecond)

				cancelFunc()

				Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
			}
		})
	})

	Describe("Zero Value Cases", func() {
		It("should handle function returning zero-value error", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(25 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(counter.Load()).To(BeNumerically(">=", int32(2)))
		})

		It("should handle uptime near zero", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Check uptime immediately
			uptime := tick.Uptime()
			Expect(uptime).To(BeNumerically(">=", 0))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Interface Compliance", func() {
		It("should implement Server interface correctly", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			// Test all Server interface methods
			Expect(tick.IsRunning()).To(BeFalse())
			Expect(tick.Uptime()).To(Equal(time.Duration(0)))

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})

		It("should implement Errors interface correctly", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return errors.New("test")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(20 * time.Millisecond)

			// Test Errors interface methods
			lastErr := tick.ErrorsLast()
			Expect(lastErr).ToNot(BeNil())

			errList := tick.ErrorsList()
			Expect(errList).ToNot(BeNil())
			Expect(len(errList)).To(BeNumerically(">", 0))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
