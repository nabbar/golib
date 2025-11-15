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
	"sync"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/startStop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Concurrency tests verify that the StartStop runner is thread-safe and can handle
// multiple concurrent operations without data races or deadlocks. These tests are
// especially important when run with the race detector (CGO_ENABLED=1 go test -race).
var _ = Describe("Concurrency", func() {
	Context("Concurrent Start calls", func() {
		// Verify that multiple goroutines can call Start() concurrently without races
		It("should handle multiple concurrent Start calls safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var startCount = new(atomic.Int32)

			start := func(c context.Context) error {
				startCount.Add(1)
				<-c.Done()
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)

			// Launch multiple Start calls concurrently from different goroutines
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					// Small stagger to reduce racing
					time.Sleep(time.Duration(idx) * time.Millisecond)
					_ = runner.Start(x)
				}(i)
			}

			wg.Wait()
			// Give the last Start() call time to stabilize
			time.Sleep(100 * time.Millisecond)

			// Wait for all Start() calls to have executed
			Eventually(func() int32 {
				return startCount.Load()
			}, 2*time.Second, 10*time.Millisecond).Should(BeNumerically("==", 10))

			// At this point, all Start() calls have completed safely without races.
			// The runner may or may not be running depending on timing (this is expected).
			// What matters is that no data races occurred and all starts were executed.

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("Concurrent Stop calls", func() {
		// Verify that multiple goroutines can call Stop() concurrently without races
		It("should handle multiple concurrent Stop calls safely", func() {
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

			// Launch multiple Stop calls concurrently
			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.Stop(x)
				}()
			}

			wg.Wait()

			// Should eventually be stopped
			Eventually(runner.IsRunning, time.Second).Should(BeFalse())

			// Stop should have been called, but not necessarily 10 times
			// (first one stops, others find already stopped)
			Expect(stopCount.Load()).To(BeNumerically(">=", 1))
		})
	})

	Context("Concurrent IsRunning calls", func() {
		// Verify that IsRunning() can be called from many goroutines simultaneously
		It("should handle concurrent IsRunning calls safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running atomic.Bool

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			// Launch many concurrent IsRunning calls
			var wg sync.WaitGroup
			results := make([]bool, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = runner.IsRunning()
				}(i)
			}

			wg.Wait()

			// All should have returned true (or at least most if stopped during checks)
			trueCount := 0
			for _, r := range results {
				if r {
					trueCount++
				}
			}
			Expect(trueCount).To(BeNumerically(">=", 50))

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("Concurrent Uptime calls", func() {
		// Verify that Uptime() can be called from many goroutines simultaneously
		It("should handle concurrent Uptime calls safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running atomic.Bool

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			time.Sleep(100 * time.Millisecond)

			// Launch many concurrent Uptime calls
			var wg sync.WaitGroup
			results := make([]time.Duration, 100)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = runner.Uptime()
				}(i)
			}

			wg.Wait()

			// All should have returned non-zero uptime
			for _, u := range results {
				Expect(u).To(BeNumerically(">", 0))
			}

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("Mixed concurrent operations", func() {
		// Verify that all operations can be called concurrently without deadlocks
		It("should handle concurrent Start/Stop/IsRunning/Uptime calls safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running = new(atomic.Bool)

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				time.Sleep(time.Second)
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)

			// Launch mixed operations concurrently to stress-test thread safety
			var wg sync.WaitGroup

			// Start operations from multiple goroutines
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.Start(x)
				}()
			}

			// IsRunning checks
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.IsRunning()
				}()
			}

			// Uptime checks
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.Uptime()
				}()
			}

			wg.Wait()

			// Should eventually be running
			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			// Now add Stop operations
			var wg2 sync.WaitGroup
			for i := 0; i < 5; i++ {
				wg2.Add(1)
				go func() {
					defer wg2.Done()
					_ = runner.Stop(x)
				}()
			}

			wg2.Wait()

			// Should eventually be stopped
			Eventually(runner.IsRunning, time.Second).Should(BeFalse())
		})

		// Verify that rapid start/stop cycles work without race conditions
		It("should handle rapid Start/Stop cycles safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var generation atomic.Int32

			start := func(c context.Context) error {
				generation.Add(1)
				<-c.Done()
				return nil
			}

			runner := New(start, func(c context.Context) error { return nil })

			// Rapid cycles
			for i := 0; i < 5; i++ {
				_ = runner.Start(x)
				time.Sleep(20 * time.Millisecond)
				_ = runner.Stop(x)
				time.Sleep(20 * time.Millisecond)
			}

			// Should have started multiple times
			Expect(generation.Load()).To(BeNumerically(">=", 1))

			// Should not be running at the end
			Eventually(runner.IsRunning, time.Second).Should(BeFalse())
		})
	})

	Context("Concurrent error tracking", func() {
		// Verify that error tracking methods are thread-safe
		It("should handle concurrent error access safely", func() {
			x, n := context.WithTimeout(context.Background(), 5*time.Second)
			defer n()

			var running atomic.Bool

			start := func(c context.Context) error {
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			// Concurrent error list access
			var wg sync.WaitGroup
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.ErrorsList()
				}()
			}

			// Concurrent error last access
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.ErrorsLast()
				}()
			}

			wg.Wait()

			// Cleanup
			_ = runner.Stop(x)
		})
	})

	Context("Concurrent Restart calls", func() {
		// Verify that multiple Restart() calls can be made concurrently
		It("should handle multiple concurrent Restart calls safely", func() {
			x, n := context.WithTimeout(context.Background(), 10*time.Second)
			defer n()

			var generation atomic.Int32
			var running atomic.Bool

			start := func(c context.Context) error {
				generation.Add(1)
				running.Store(true)
				<-c.Done()
				running.Store(false)
				return nil
			}
			stop := func(c context.Context) error {
				return nil
			}

			runner := New(start, stop)

			// Initial start
			err := runner.Start(x)
			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return running.Load() && runner.IsRunning()
			}, time.Second).Should(BeTrue())

			initialGen := generation.Load()

			// Concurrent restarts
			var wg sync.WaitGroup
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = runner.Restart(x)
				}()
			}

			wg.Wait()

			// Should have restarted at least once
			Eventually(func() int32 {
				return generation.Load()
			}, time.Second).Should(BeNumerically(">", initialGen))

			// Cleanup
			_ = runner.Stop(x)
		})
	})
})
