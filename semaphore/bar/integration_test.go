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
	"sync/atomic"
	"time"

	libbar "github.com/nabbar/golib/semaphore/bar"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bar Integration Tests", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(globalCtx, 10*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Real-world scenarios", func() {
		It("should handle a batch processing workflow", func() {
			const (
				nbrWorkers = 5
				totalTasks = 50
			)

			sem := createTestSemaphoreWithProgress(ctx, nbrWorkers)
			bar := libbar.New(sem, totalTasks, false)

			var (
				wg             sync.WaitGroup
				processedTasks atomic.Int32
			)

			// Process tasks
			for i := 0; i < totalTasks; i++ {
				wg.Add(1)
				go func(taskID int) {
					defer wg.Done()

					// Acquire worker
					if err := bar.NewWorker(); err != nil {
						return
					}
					defer bar.DeferWorker()

					// Simulate task processing
					time.Sleep(5 * time.Millisecond)

					processedTasks.Add(1)
				}(i)
			}

			wg.Wait()
			bar.DeferMain()

			Expect(processedTasks.Load()).To(Equal(int32(totalTasks)))
			Expect(bar.Completed()).To(BeTrue())
		})

		It("should track progress through multiple stages", func() {
			sem := createTestSemaphoreWithProgress(ctx, 3)
			bar := libbar.New(sem, 100, false)

			// Stage 1: Initial progress
			for i := 0; i < 3; i++ {
				Expect(bar.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer bar.DeferWorker()
					time.Sleep(20 * time.Millisecond)
				}()
			}

			time.Sleep(50 * time.Millisecond)

			// Stage 2: Reset and continue
			bar.Reset(200, 50)
			Expect(bar.Total()).To(Equal(int64(200)))

			// Stage 3: Complete remaining work
			bar.Inc(150)
			bar.Complete()

			time.Sleep(150 * time.Millisecond)
			Expect(bar.Completed()).To(BeTrue())
		})

		It("should handle rapid increment/decrement cycles", func() {
			sem := createTestSemaphoreWithProgress(ctx, 10)
			bar := libbar.New(sem, 1000, false)

			var wg sync.WaitGroup

			// Incrementers
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						bar.Inc(1)
						time.Sleep(time.Millisecond)
					}
				}()
			}

			// Decrementers (simulating rollbacks)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 50; j++ {
						bar.Dec(1)
						time.Sleep(2 * time.Millisecond)
					}
				}()
			}

			wg.Wait()

			// Bar should still be operational
			Expect(bar.Total()).To(Equal(int64(1000)))
		})

		It("should properly cleanup with drop=true", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, true) // drop = true

			// Process all tasks
			for i := 0; i < 10; i++ {
				Expect(bar.NewWorker()).ToNot(HaveOccurred())
				go func() {
					defer bar.DeferWorker()
					time.Sleep(10 * time.Millisecond)
				}()
			}

			time.Sleep(200 * time.Millisecond)

			bar.DeferMain()
			time.Sleep(100 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Describe("Error handling and recovery", func() {
		It("should continue working after context errors", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			// Cancel context
			localCancel()

			// Bar should still report values even with cancelled context
			Expect(bar.Total()).To(Equal(int64(100)))
		})

		It("should handle worker creation failures gracefully", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			defer localCancel()

			sem := createTestSemaphore(localCtx, 2)
			bar := libbar.New(sem, 100, false)

			// Fill semaphore
			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			Expect(bar.NewWorker()).ToNot(HaveOccurred())

			// Cancel context to cause failures
			localCancel()
			time.Sleep(50 * time.Millisecond)

			// Should fail gracefully
			err := bar.NewWorker()
			Expect(err).To(HaveOccurred())

			// Cleanup
			bar.DeferWorker()
			bar.DeferWorker()
		})
	})

	Describe("Performance scenarios", func() {
		It("should handle high-frequency updates efficiently", func() {
			sem := createTestSemaphoreWithProgress(ctx, 50)
			bar := libbar.New(sem, 10000, false)

			start := time.Now()

			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						bar.Inc(1)
					}
				}()
			}

			wg.Wait()
			duration := time.Since(start)

			// Should complete in reasonable time (< 2 seconds)
			Expect(duration).To(BeNumerically("<", 2*time.Second))
		})

		It("should handle many concurrent workers", func() {
			sem := createTestSemaphore(ctx, 100)
			bar := libbar.New(sem, 1000, false)

			var wg sync.WaitGroup
			workerCount := 200

			for i := 0; i < workerCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := bar.NewWorker(); err != nil {
						return
					}
					defer bar.DeferWorker()

					time.Sleep(10 * time.Millisecond)
				}()
			}

			wg.Wait()
			Expect(bar.WaitAll()).ToNot(HaveOccurred())
		})
	})

	Describe("State consistency", func() {
		It("should maintain consistent state under concurrent access", func() {
			sem := createTestSemaphore(ctx, 10)
			bar := libbar.New(sem, 1000, false)

			var wg sync.WaitGroup

			// Multiple goroutines modifying state
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					// Get values
					_ = bar.Total()
					_ = bar.Current()
					_ = bar.Completed()

					// Modify state
					bar.Inc(10)
					bar.Dec(5)
				}()
			}

			wg.Wait()

			// Should not panic and total should be unchanged
			Expect(bar.Total()).To(Equal(int64(1000)))
		})

		It("should handle Reset during active operations", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			var wg sync.WaitGroup

			// Start incrementing
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 50; i++ {
					bar.Inc(1)
					time.Sleep(2 * time.Millisecond)
				}
			}()

			// Reset in the middle
			time.Sleep(30 * time.Millisecond)
			bar.Reset(200, 0)

			wg.Wait()

			// Total should reflect the reset value
			Expect(bar.Total()).To(Equal(int64(200)))
		})
	})
})
