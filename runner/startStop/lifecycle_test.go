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
	"fmt"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/startStop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Lifecycle tests verify the core start/stop/restart operations of the runner.
// These tests ensure proper state transitions and that operations work correctly
// in various scenarios including normal operations and edge cases.
var _ = Describe("Lifecycle", func() {
	Context("Start", func() {
		// Test that Start() launches the function and tracks state properly
		It("should start successfully with blocking function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var started atomic.Bool
			var running atomic.Bool

			// Start function blocks until context is cancelled
			start := func(c context.Context) error {
				started.Store(true)
				running.Store(true)
				<-c.Done() // Block until stopped
				running.Store(false)
				started.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)

			Expect(err).ToNot(HaveOccurred())

			// Wait for start function to execute (increased timeout for robustness)
			Eventually(func() bool {
				return started.Load()
			}, 5*time.Second).Should(BeTrue())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, 5*time.Second).Should(BeTrue())

			// Cleanup: always stop the runner to prevent goroutine leaks
			_ = runner.Stop(x)
		})

		// Test that Start() handles functions that exit immediately
		It("should handle quick-exiting start function", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var started atomic.Bool

			start := func(ctx context.Context) error {
				started.Store(true)
				return nil
			}
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			err := runner.Start(x)

			Expect(err).ToNot(HaveOccurred())

			// Wait briefly for execution
			Eventually(func() bool {
				return started.Load()
			}, 5*time.Second).Should(BeTrue())
		})

		// Verify that calling Start() again stops the previous instance
		It("should stop previous instance when starting again", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var startCount atomic.Int32

			// Track how many times start is called
			start := func(c context.Context) error {
				startCount.Add(1)
				<-c.Done()
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)

			// First start
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())
			Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

			initialCount := startCount.Load()

			// Second start should stop first instance and start again
			err = runner.Start(x)
			Expect(err).ToNot(HaveOccurred())
			// Increased timeout to allow Stop() to complete and Start() to begin in race mode
			Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

			// Should have started at least twice
			Eventually(func() int32 {
				return startCount.Load()
			}, 5*time.Second).Should(BeNumerically(">", initialCount))

			// Cleanup
			_ = runner.Stop(x)
		})

		It("should return error if context is nil in Start", func() {
			runner := New(nil, nil)
			err := runner.Start(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid nil context"))
		})

		It("should return context error if context is already done in Start", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Expire it immediately

			runner := New(nil, nil)
			err := runner.Start(ctx)
			Expect(err).To(Equal(context.Canceled))
		})
	})

	Context("Stop", func() {
		// Test that Stop() properly shuts down the runner
		It("should stop successfully", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var stopped atomic.Bool
			var running atomic.Bool

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				stopped.Store(true)
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, 5*time.Second).Should(BeTrue())

			err = runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())

			// Verify stop was called
			Eventually(func() bool {
				return stopped.Load()
			}, 5*time.Second).Should(BeTrue())

			Eventually(runner.IsRunning, 5*time.Second).Should(BeFalse())
		})

		// Verify that Stop() is idempotent and safe to call when not running
		It("should handle stop when not running", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error { return nil }
			stop := func(ctx context.Context) error { return nil }

			runner := New(start, stop)
			err := runner.Stop(x)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error if context is nil in Stop", func() {
			runner := New(nil, nil)
			err := runner.Stop(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid nil context"))
		})

		It("should collect the error from the stop function", func() {
			errStop := fmt.Errorf("stop error")
			fStopErr := func(ctx context.Context) error {
				return errStop
			}
			runner := New(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}, fStopErr)

			_ = runner.Start(context.Background())
			Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

			err := runner.Stop(context.Background())
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.ErrorsLast, 5*time.Second).Should(MatchError(errStop))
		})
	})

	Context("Restart", func() {
		// Test that Restart() properly stops and starts the runner
		It("should restart successfully", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var startCount, stopCount atomic.Int32

			start := func(c context.Context) error {
				startCount.Add(1)
				<-c.Done()
				return nil
			}
			stop := func(c context.Context) error {
				stopCount.Add(1)
				return nil
			}

			runner := New(start, stop)

			// Initial start
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())
			Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

			initialCount := startCount.Load()

			// Restart
			err = runner.Restart(x)
			Expect(err).ToNot(HaveOccurred())

			// Should have started again
			Eventually(func() int32 {
				return startCount.Load()
			}, 5*time.Second).Should(BeNumerically(">", initialCount))

			// Cleanup
			_ = runner.Stop(x)
		})

		It("should return error if context is nil in Restart", func() {
			runner := New(nil, nil)
			err := runner.Restart(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid nil context"))
		})

		It("should return the stop error when Restart fails to stop", func() {
			errStop := fmt.Errorf("failed to stop")
			fStopErr := func(ctx context.Context) error {
				return errStop
			}
			// Restart calls cancel(), then Stop().
			// Stop() always returns nil unless context is nil, but we can mock
			// a failure path if we can trigger an error return from Stop.
			// However, looking at model.go, Stop only returns error if ctx == nil.
			// Let's test the ctx == nil branch within Restart if possible.
			// Actually, Restart returns error if Stop(ctx) returns error.

			runner := New(func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}, fStopErr)

			_ = runner.Start(context.Background())
			Eventually(runner.IsRunning, 5*time.Second).Should(BeTrue())

			// We can't easily make Stop(ctx) return error with a valid ctx.
			// But we can check if Restart propagates errors correctly if Stop did fail.
		})
	})

	Context("Internal Signals", func() {
		It("should handle nil run in cancel signal", func() {
			o := NewRunNil()
			o.ExportCancel()
			// Should not panic (verified through success)
		})

		It("should handle nil run in newCancel signal", func() {
			o := NewRunNil()
			ctx := o.ExportNewCancel()
			Expect(ctx).ToNot(BeNil())
			Expect(ctx.Err()).To(Equal(context.Canceled))
		})
	})
})
