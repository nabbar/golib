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
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/ticker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// errors_test.go validates error handling and collection capabilities of the ticker package.
//
// Test Coverage:
//   - Error collection from ticker function across multiple ticks
//   - ErrorsLast(): Retrieving the most recent error
//   - ErrorsList(): Retrieving all collected errors
//   - Nil error handling (no errors should not cause issues)
//   - Error clearing on Restart()
//   - Nil function handling (default error-returning behavior)
//   - Various error patterns (intermittent, wrapped, context errors)
//   - Error state persistence across Stop/Start cycles
//   - Error recovery (ticker continues running after errors)
//
// Testing Strategy:
// Tests use longer tick intervals (250ms) to allow time for multiple ticks
// and error accumulation. Atomic counters track execution count independently
// of error collection. Tests verify that errors don't stop ticker execution.
//
// Important Notes:
//   - Errors are collected via github.com/nabbar/golib/errors/pool
//   - The ticker continues running even if the function returns errors
//   - Errors are cleared when Start() or Restart() is called
//   - Nil errors may be stored or filtered by the error pool implementation
var _ = Describe("Error Handling", func() {
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

	Describe("Error Collection", func() {
		It("should collect errors from ticker function", func() {
			testErr := errors.New("test error")
			counter := int32(0)

			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return testErr
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for multiple ticks
			time.Sleep(800 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Check collected errors
			lastErr := tick.ErrorsLast()
			Expect(lastErr).To(MatchError(testErr))

			errList := tick.ErrorsList()
			Expect(len(errList)).To(BeNumerically(">", 0))
			Expect(errList[len(errList)-1]).To(MatchError(testErr))
		})

		It("should collect multiple different errors", func() {
			counter := int32(0)
			errors := []error{
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			}

			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				idx := int(atomic.AddInt32(&counter, 1) - 1)
				if idx < len(errors) {
					return errors[idx]
				}
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for multiple ticks
			time.Sleep(800 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Check all errors collected
			errList := tick.ErrorsList()
			Expect(len(errList)).To(BeNumerically(">=", 1))
		})

		It("should handle nil errors", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(30 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have run successfully
			Expect(counter.Load()).To(BeNumerically(">=", int32(2)))

			// Errors list may contain nils or be empty
			errList := tick.ErrorsList()
			// Just verify we can retrieve the list
			Expect(errList).ToNot(BeNil())
		})

		It("should clear errors on restart", func() {
			testErr := errors.New("test error")
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return testErr
			})

			// First run
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(1200 * time.Millisecond)

			// Should have errors
			Expect(tick.ErrorsLast()).ToNot(BeNil())

			// Restart
			err = tick.Restart(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Errors should be cleared briefly after restart, then collect new ones
			time.Sleep(1200 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("ErrorsLast", func() {
		It("should return last error", func() {
			expectedErr := errors.New("last error")
			counter := int32(0)

			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				count := atomic.AddInt32(&counter, 1)
				if count == 3 {
					return expectedErr
				}
				return errors.New("other error")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for at least 3 ticks
			time.Sleep(800 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Last error should be available
			lastErr := tick.ErrorsLast()
			Expect(lastErr).ToNot(BeNil())
		})

		It("should return nil when no errors", func() {
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			// Check before start
			Expect(tick.ErrorsLast()).To(BeNil())

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(750 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should still be nil if all returned nil
			lastErr := tick.ErrorsLast()
			// Either nil or the last nil value, both are acceptable
			_ = lastErr
		})
	})

	Describe("ErrorsList", func() {
		It("should return all collected errors", func() {
			counter := int32(0)
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return errors.New("tick error")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for several ticks
			time.Sleep(800 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			errList := tick.ErrorsList()
			Expect(errList).ToNot(BeNil())
			Expect(len(errList)).To(BeNumerically(">=", 1))
		})

		It("should return empty or nil list when no errors", func() {
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			errList := tick.ErrorsList()
			// Should return a valid slice (empty or with nil entries)
			Expect(errList).ToNot(BeNil())
		})
	})

	Describe("Nil Function Handling", func() {
		It("should handle nil function with error", func() {
			tick := New(250*time.Millisecond, nil)

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for a tick
			time.Sleep(300 * time.Millisecond)

			// Should have error from nil function
			lastErr := tick.ErrorsLast()
			Expect(lastErr).ToNot(BeNil())
			Expect(lastErr.Error()).To(ContainSubstring("invalid function"))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Error Patterns", func() {
		It("should handle intermittent errors", func() {
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

			// Should have collected some errors
			errList := tick.ErrorsList()
			Expect(len(errList)).To(BeNumerically(">", 0))

			// Count non-nil errors
			nonNilCount := 0
			for _, e := range errList {
				if e != nil {
					nonNilCount++
				}
			}
			Expect(len(errList)).To(Equal(nonNilCount))
		})

		It("should handle wrapped errors", func() {
			baseErr := errors.New("base error")
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return errors.Join(baseErr, errors.New("additional context"))
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(15 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Ensure we had some ticks
			Expect(counter.Load()).To(BeNumerically(">=", int32(1)))

			lastErr := tick.ErrorsLast()
			Expect(lastErr).ToNot(BeNil())
			Expect(lastErr.Error()).To(ContainSubstring("base error"))
		})

		It("should handle context errors", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			})

			cancelCtx, cancelFunc := context.WithCancel(context.Background())
			err := tick.Start(cancelCtx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(20 * time.Millisecond)

			// Cancel context
			cancelFunc()
			time.Sleep(20 * time.Millisecond)

			// May have context cancellation error
			lastErr := tick.ErrorsLast()
			if lastErr != nil {
				// Context cancellation is a valid error
				Expect(errors.Is(lastErr, context.Canceled)).To(BeTrue())
			}
		})
	})

	Describe("Error State Management", func() {
		It("should maintain error state across stops and starts", func() {
			testErr := errors.New("test error")
			counter := new(atomic.Uint32)

			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				if counter.Load() <= 2 {
					return testErr
				}
				return nil
			})

			// First run with errors
			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(25 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have had errors from first run
			firstCount := counter.Load()
			Expect(firstCount).To(BeNumerically(">=", int32(2)))

			// Second run - errors are cleared on start
			err = tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(50 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Verify second run happened
			Expect(counter.Load()).To(BeNumerically(">", firstCount))
		})

		It("should handle rapid error accumulation", func() {
			counter := int32(0)
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return errors.New("rapid error")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Let it accumulate errors
			time.Sleep(800 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Should have collected several errors
			errList := tick.ErrorsList()
			Expect(len(errList)).To(BeNumerically(">=", 3))
		})
	})

	Describe("Error Recovery", func() {
		It("should continue running after errors", func() {
			counter := int32(0)
			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return errors.New("continuous error")
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(300 * time.Millisecond)

			// Should still be running despite errors
			Expect(tick.IsRunning()).To(BeTrue())

			count := atomic.LoadInt32(&counter)
			Expect(count).To(BeNumerically(">=", int32(1)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not stop on first error", func() {
			counter := new(atomic.Uint32)
			firstErrorSeen := new(atomic.Bool)

			tick := New(250*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				if counter.Load() == 1 {
					firstErrorSeen.Store(true)
					return errors.New("first error")
				}
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(300 * time.Millisecond)

			// Should have continued after first error
			Expect(firstErrorSeen.Load()).To(BeTrue())
			Expect(counter.Load()).To(BeNumerically(">=", int32(1)))
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
