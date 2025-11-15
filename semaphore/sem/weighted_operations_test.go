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

package sem_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	libsem "github.com/nabbar/golib/semaphore/sem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Weighted Semaphore Operations", func() {
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

	Describe("NewWorker/DeferWorker", func() {
		It("should acquire and release worker slots", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()

			// Should be able to acquire again
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()
		})

		It("should respect semaphore limit", func() {
			sem := libsem.New(ctx, 2)
			defer sem.DeferMain()

			// Acquire 2 slots
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			// Try to acquire a third - should block
			done := make(chan error, 1)
			go func() {
				done <- sem.NewWorker()
			}()

			// Should not complete immediately
			select {
			case <-done:
				Fail("NewWorker should block when semaphore is full")
			case <-time.After(50 * time.Millisecond):
				// Expected
			}

			// Release one slot
			sem.DeferWorker()

			// Now the third acquisition should complete
			Eventually(done, time.Second).Should(Receive(BeNil()))

			// Cleanup
			sem.DeferWorker()
			sem.DeferWorker()
		})

		It("should block until context is cancelled", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 1)
			defer sem.DeferMain()

			// Fill the semaphore
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			// Try to acquire in goroutine
			done := make(chan error, 1)
			go func() {
				done <- sem.NewWorker()
			}()

			// Cancel context
			time.Sleep(20 * time.Millisecond)
			localCancel()

			// Should receive cancellation error
			Eventually(done, time.Second).Should(Receive(Equal(context.Canceled)))

			sem.DeferWorker()
		})
	})

	Describe("NewWorkerTry", func() {
		It("should acquire slot immediately if available", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			Expect(sem.NewWorkerTry()).To(BeTrue())
			sem.DeferWorker()
		})

		It("should return false if no slots available", func() {
			sem := libsem.New(ctx, 2)
			defer sem.DeferMain()

			// Fill semaphore
			Expect(sem.NewWorkerTry()).To(BeTrue())
			Expect(sem.NewWorkerTry()).To(BeTrue())

			// No more slots
			Expect(sem.NewWorkerTry()).To(BeFalse())

			// Cleanup
			sem.DeferWorker()
			sem.DeferWorker()
		})

		It("should not block", func() {
			sem := libsem.New(ctx, 1)
			defer sem.DeferMain()

			// Fill semaphore
			Expect(sem.NewWorkerTry()).To(BeTrue())

			// Try again should return immediately (not block)
			start := time.Now()
			result := sem.NewWorkerTry()
			duration := time.Since(start)

			Expect(result).To(BeFalse())
			Expect(duration).To(BeNumerically("<", 10*time.Millisecond))

			sem.DeferWorker()
		})
	})

	Describe("WaitAll", func() {
		It("should wait for all workers to complete", func() {
			sem := libsem.New(ctx, 3)
			defer sem.DeferMain()

			var wg sync.WaitGroup

			// Start workers
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					time.Sleep(50 * time.Millisecond)
					sem.DeferWorker()
				}()
			}

			// Wait for workers to start
			time.Sleep(10 * time.Millisecond)

			// WaitAll should block
			done := make(chan error, 1)
			go func() {
				done <- sem.WaitAll()
			}()

			// Should not complete immediately
			select {
			case <-done:
				Fail("WaitAll should block while workers are running")
			case <-time.After(20 * time.Millisecond):
				// Expected
			}

			// Wait for workers to complete
			wg.Wait()

			// Now WaitAll should complete
			Eventually(done, time.Second).Should(Receive(BeNil()))
		})

		It("should succeed if no workers are running", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			// No workers running
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})

		It("should be cancellable via context", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 2)
			defer sem.DeferMain()

			// Start a long-running worker
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			// Try to WaitAll in goroutine
			done := make(chan error, 1)
			go func() {
				done <- sem.WaitAll()
			}()

			// Cancel context
			time.Sleep(20 * time.Millisecond)
			localCancel()

			// Should receive error
			Eventually(done, time.Second).Should(Receive(HaveOccurred()))

			sem.DeferWorker()
		})
	})

	Describe("Weighted", func() {
		It("should return the configured limit", func() {
			sem := libsem.New(ctx, 10)
			Expect(sem.Weighted()).To(Equal(int64(10)))
		})

		It("should return MaxSimultaneous for nbrSimultaneous == 0", func() {
			sem := libsem.New(ctx, 0)
			expected := libsem.MaxSimultaneous()
			Expect(sem.Weighted()).To(Equal(int64(expected)))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle many concurrent workers", func() {
			sem := libsem.New(ctx, 10)
			defer sem.DeferMain()

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
			)

			// Start many workers
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := sem.NewWorker(); err != nil {
						return
					}
					defer sem.DeferWorker()

					completed.Add(1)
					time.Sleep(5 * time.Millisecond)
				}()
			}

			wg.Wait()

			Expect(completed.Load()).To(Equal(int32(100)))
		})

		It("should limit concurrency correctly", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			var (
				wg             sync.WaitGroup
				currentWorkers atomic.Int32
				maxConcurrent  atomic.Int32
			)

			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := sem.NewWorker(); err != nil {
						return
					}
					defer sem.DeferWorker()

					// Track concurrent workers
					current := currentWorkers.Add(1)
					defer currentWorkers.Add(-1)

					// Update max
					for {
						old := maxConcurrent.Load()
						if current <= old || maxConcurrent.CompareAndSwap(old, current) {
							break
						}
					}

					time.Sleep(10 * time.Millisecond)
				}()
			}

			wg.Wait()

			// Max concurrent should never exceed limit
			Expect(maxConcurrent.Load()).To(BeNumerically("<=", 5))
		})
	})
})
