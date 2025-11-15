/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package bar_test

import (
	"context"
	"sync"
	"time"

	libbar "github.com/nabbar/golib/semaphore/bar"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bar Edge Cases", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 5*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Boundary conditions", func() {
		It("should handle maximum int64 total", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 9223372036854775807, false) // math.MaxInt64

			Expect(bar.Total()).To(Equal(int64(9223372036854775807)))
		})

		It("should handle zero workers limit", func() {
			sem := createTestSemaphore(ctx, 0)
			bar := libbar.New(sem, 100, false)

			// Should not panic
			Expect(bar.Total()).To(Equal(int64(100)))
		})

		It("should handle very large increments", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 1000000, false)

			bar.Inc64(999999)
			time.Sleep(20 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically(">", 0))
		})

		It("should handle very large decrements", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 1000000, false)

			bar.Inc64(500000)
			time.Sleep(20 * time.Millisecond)

			bar.Dec64(250000)
			time.Sleep(20 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically(">", 0))
		})
	})

	Describe("Rapid state changes", func() {
		It("should handle rapid reset operations", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			for i := 0; i < 10; i++ {
				bar.Reset(int64((i+1)*100), 0)
			}

			Expect(bar.Total()).To(Equal(int64(1000)))
		})

		It("should handle complete/incomplete cycles", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Set progress before completing
			bar.Inc(100)
			time.Sleep(20 * time.Millisecond)

			// Multiple complete calls should be safe
			bar.Complete()
			time.Sleep(20 * time.Millisecond)
			bar.Complete()
			time.Sleep(20 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})

		It("should handle concurrent Inc and Dec operations", func() {
			sem := createTestSemaphoreWithProgress(ctx, 10)
			bar := libbar.New(sem, 10000, false)

			var wg sync.WaitGroup

			// Concurrent incrementers
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						bar.Inc(10)
					}
				}()
			}

			// Concurrent decrementers
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						bar.Dec(10)
					}
				}()
			}

			wg.Wait()

			// Should not panic and bar should still be usable
			Expect(bar.Total()).To(Equal(int64(10000)))
		})
	})

	Describe("Context cancellation scenarios", func() {
		It("should handle immediate context cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			// Cancel immediately
			localCancel()

			// Bar should still be queryable
			Expect(bar.Total()).To(Equal(int64(100)))
			Expect(bar.Completed()).To(BeTrue()) // Without MPB, always true
		})

		It("should handle context cancellation during operations", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := createTestSemaphoreWithProgress(localCtx, 5)
			bar := libbar.New(sem, 1000, false)

			var wg sync.WaitGroup

			// Start some operations
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := bar.NewWorker(); err == nil {
						defer bar.DeferWorker()
						time.Sleep(50 * time.Millisecond)
					}
				}()
			}

			// Cancel mid-operations
			time.Sleep(20 * time.Millisecond)
			localCancel()

			wg.Wait()

			// Should complete without panic
			Expect(bar.Err()).To(Equal(context.Canceled))
		})
	})

	Describe("DeferMain with various states", func() {
		It("should handle DeferMain on incomplete bar", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Don't complete it, just defer
			bar.Inc(50)
			bar.DeferMain()

			time.Sleep(50 * time.Millisecond)

			// Should have been aborted
			Expect(bar.Completed()).To(BeTrue())
		})

		It("should handle DeferMain on already completed bar", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(100)
			bar.Complete()

			time.Sleep(150 * time.Millisecond)

			bar.DeferMain() // Should not panic

			Expect(bar.Completed()).To(BeTrue())
		})

		It("should handle DeferMain with drop=true", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, true) // drop = true

			bar.Inc(50) // Incomplete
			bar.DeferMain()

			time.Sleep(150 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Describe("Worker slot exhaustion", func() {
		It("should handle worker slot exhaustion gracefully", func() {
			sem := createTestSemaphore(ctx, 2)
			bar := libbar.New(sem, 100, false)

			// Fill all slots
			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			Expect(bar.NewWorker()).ToNot(HaveOccurred())

			// Try to get another - should block or fail
			done := make(chan bool)
			go func() {
				ok := bar.NewWorkerTry()
				done <- ok
			}()

			select {
			case result := <-done:
				Expect(result).To(BeFalse()) // Should fail to acquire
			case <-time.After(150 * time.Millisecond):
				Fail("NewWorkerTry should not block")
			}

			// Clean up
			bar.DeferWorker()
			bar.DeferWorker()
		})

		It("should allow new workers after release", func() {
			sem := createTestSemaphore(ctx, 1)
			bar := libbar.New(sem, 100, false)

			// Acquire and release
			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			bar.DeferWorker()

			// Should be able to acquire again
			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			bar.DeferWorker()
		})
	})

	Describe("Zero and negative values", func() {
		It("should handle zero increment", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			initial := bar.Current()
			bar.Inc(0)
			time.Sleep(10 * time.Millisecond)

			// Current should not change significantly
			Expect(bar.Current()).To(BeNumerically(">=", initial))
		})

		It("should handle zero decrement", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(50)
			time.Sleep(10 * time.Millisecond)
			current := bar.Current()

			bar.Dec(0)
			time.Sleep(10 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically(">=", current-1)) // Allow small variance
		})
	})
})
