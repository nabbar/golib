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

			// Wait for start function to execute
			Eventually(func() bool {
				return started.Load()
			}, time.Second).Should(BeTrue())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

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
			}, 500*time.Millisecond).Should(BeTrue())
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
			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			initialCount := startCount.Load()

			// Second start should stop first instance and start again
			err = runner.Start(x)
			Expect(err).ToNot(HaveOccurred())
			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Should have started at least twice
			Eventually(func() int32 {
				return startCount.Load()
			}, time.Second).Should(BeNumerically(">", initialCount))

			// Cleanup
			_ = runner.Stop(x)
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
			}, time.Second).Should(BeTrue())

			err = runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())

			// Verify stop was called
			Eventually(func() bool {
				return stopped.Load()
			}, time.Second).Should(BeTrue())

			Eventually(runner.IsRunning, time.Second).Should(BeFalse())
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

		// Verify that concurrent Stop() calls are handled gracefully
		It("should handle multiple stop calls", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var stopCount atomic.Int32
			var running atomic.Bool

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				stopCount.Add(1)
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			// Call Stop() multiple times - should be safe and idempotent
			err1 := runner.Stop(x)
			err2 := runner.Stop(x)

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())

			// Only first stop should call stop function
			Consistently(func() int32 {
				return stopCount.Load()
			}, 200*time.Millisecond, 50*time.Millisecond).Should(BeNumerically("<=", 1))
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
			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			initialCount := startCount.Load()

			// Restart
			err = runner.Restart(x)
			Expect(err).ToNot(HaveOccurred())

			// Should have started again
			Eventually(func() int32 {
				return startCount.Load()
			}, time.Second).Should(BeNumerically(">", initialCount))

			// Cleanup
			_ = runner.Stop(x)
		})

		// Verify that Restart() works even when the runner is not running
		It("should handle restart when not running", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Restart(x)

			// Should succeed even when not running
			Expect(err).ToNot(HaveOccurred())

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("IsRunning", func() {
		// Verify IsRunning() returns correct state during lifecycle
		It("should return true when running", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			Expect(runner.IsRunning()).To(BeFalse())

			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			// Cleanup
			_ = runner.Stop(x)
		})

		It("should return false when stopped", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			start := func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			}
			stop := func(ctx context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, 100*time.Millisecond).Should(BeTrue())

			err = runner.Stop(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(runner.IsRunning, time.Second).Should(BeFalse())
		})
	})
})
