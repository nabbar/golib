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

package startStop_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/startStop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Error Handling tests verify that the runner properly captures and reports errors
// from start/stop functions, handles nil functions gracefully, and recovers from panics.
var _ = Describe("Error Handling", func() {
	Context("Start errors", func() {
		// Verify that errors from the start function are captured and tracked
		It("should capture error from start function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			expectedErr := errors.New("start failed")

			start := func(ctx context.Context) error {
				return expectedErr
			}
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			err := runner.Start(x)

			// Start returns nil immediately since it launches asynchronously
			Expect(err).ToNot(HaveOccurred())

			// But error should be captured in error list
			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(MatchError(expectedErr))

			errs := runner.ErrorsList()
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(MatchError(expectedErr)))
		})

		// Verify that a nil start function generates an appropriate error
		It("should handle nil start function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			stop := func(ctx context.Context) error { return nil }

			runner := New(nil, stop)
			err := runner.Start(x)

			Expect(err).ToNot(HaveOccurred())

			// Should generate "invalid start function" error
			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(HaveOccurred())

			lastErr := runner.ErrorsLast()
			Expect(lastErr).To(Not(BeNil()))
			Expect(lastErr.Error()).To(ContainSubstring("invalid start function"))
		})
	})

	Context("Stop errors", func() {
		// Verify that errors from the stop function are captured
		It("should handle error from stop function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			expectedErr := errors.New("stop failed")
			var running = new(atomic.Bool)

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return expectedErr
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			err = runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())

			// Stop should return the error
			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(MatchError(expectedErr))
		})

		// Verify that a nil stop function generates an appropriate error
		It("should handle nil stop function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running = new(atomic.Bool)

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}

			runner := New(start, nil)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			err = runner.Stop(x)

			// Should return "invalid stop function" error
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(HaveOccurred())
			Eventually(func() string {
				if err := runner.ErrorsLast(); err != nil {
					return err.Error()
				}
				return ""
			}, time.Second).Should(ContainSubstring("invalid stop function"))
		})
	})

	Context("Error tracking", func() {
		// Verify that errors from multiple start operations are tracked separately
		It("should track multiple errors", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			err1 := errors.New("error 1")
			err2 := errors.New("error 2")

			var count = new(atomic.Uint32)

			start := func(ctx context.Context) error {
				count.Add(1)
				if count.Load() == 1 {
					return err1
				}
				return err2
			}

			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)

			// First start - should return err1
			_ = runner.Start(x)

			// Wait for the start function to complete
			for count.Load() > 0 {
				time.Sleep(2 * time.Millisecond)
			}

			time.Sleep(10 * time.Millisecond)
			Expect(len(runner.ErrorsList())).To(BeNumerically("==", 1))
			Expect(runner.ErrorsLast()).To(HaveOccurred())
			Expect(runner.ErrorsLast().Error()).To(Equal(err1.Error()))

			_ = runner.Stop(x)
			time.Sleep(200 * time.Millisecond)

			// Second start - should return err2
			_ = runner.Start(x)

			// Wait for the start function to complete
			for count.Load() > 1 {
				time.Sleep(2 * time.Millisecond)
			}

			time.Sleep(10 * time.Millisecond)
			Expect(len(runner.ErrorsList())).To(BeNumerically("==", 1))
			Expect(runner.ErrorsLast()).To(HaveOccurred())
			Expect(runner.ErrorsLast().Error()).To(Equal(err2.Error()))
		})

		// Verify that calling Start() clears previous errors
		It("should clear errors on new start", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			err1 := errors.New("error 1")

			start := func(c context.Context) error {
				return err1
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)

			// First start with error - should capture the error
			_ = runner.Start(x)
			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(HaveOccurred())

			time.Sleep(200 * time.Millisecond)

			// Verify error is present
			Expect(runner.ErrorsLast()).To(HaveOccurred())

			var running = new(atomic.Bool)

			// Create a new runner with a successful start function
			start2 := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop2 := func(c context.Context) error {
				return nil
			}

			runner2 := New(start2, stop2)
			_ = runner2.Start(x)

			Eventually(func() bool {
				return running.Load() && runner2.IsRunning()
			}, time.Second).Should(BeTrue())

			// New runner should have no errors
			Consistently(func() error {
				return runner2.ErrorsLast()
			}, 200*time.Millisecond, 50*time.Millisecond).Should(BeNil())

			// Cleanup
			_ = runner2.Stop(x)
		})

		// Verify that ErrorsList() returns all captured errors
		It("should provide error list", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			err1 := errors.New("test error")

			start := func(ctx context.Context) error {
				return err1
			}
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			_ = runner.Start(x)

			Eventually(func() []error {
				return runner.ErrorsList()
			}, time.Second).ShouldNot(BeEmpty())

			errs := runner.ErrorsList()
			Expect(len(errs)).To(BeNumerically(">=", 1))
		})

		// Verify that ErrorsLast() returns the most recent error
		It("should provide last error", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			expectedErr := errors.New("last error")

			start := func(ctx context.Context) error {
				return expectedErr
			}
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			_ = runner.Start(x)

			Eventually(func() error {
				return runner.ErrorsLast()
			}, time.Second).Should(MatchError(expectedErr))
		})
	})

	Context("Panic recovery", func() {
		// Verify that the runner doesn't crash if start function has issues
		It("should recover from panic in start function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			// Function that returns an error (not a panic, but tests error handling)
			start := func(c context.Context) error {
				return fmt.Errorf("start don't panic but error reported")
			}
			stop := func(c context.Context) error { return nil }

			runner := New(start, stop)
			err := runner.Start(x)

			// Should not panic, start returns nil
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			Expect(runner.ErrorsLast()).To(HaveOccurred())

			// Runner should handle panic gracefully
			time.Sleep(200 * time.Millisecond)
			Expect(runner.IsRunning()).To(BeFalse())
		})

		// Verify that the runner doesn't crash if stop function has issues
		It("should recover from panic in stop function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running = new(atomic.Bool)

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return fmt.Errorf("stop don't panic but error reported")
			}

			runner := New(start, stop)
			_ = runner.Start(x)

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			// Should not panic - recovery will handle it
			err := runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)

			Expect(runner.ErrorsLast()).To(HaveOccurred())

			// Should eventually stop
			Eventually(runner.IsRunning, time.Second).Should(BeFalse())
		})
	})
})
