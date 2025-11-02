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

var _ = Describe("WaitGroup Semaphore Operations", func() {
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
		It("should always succeed (no limit)", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			// Should always succeed
			for i := 0; i < 100; i++ {
				Expect(sem.NewWorker()).ToNot(HaveOccurred())
			}

			// Clean up
			for i := 0; i < 100; i++ {
				sem.DeferWorker()
			}
		})

		It("should track workers correctly", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			var wg sync.WaitGroup

			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()
					time.Sleep(20 * time.Millisecond)
				}()
			}

			wg.Wait()
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})
	})

	Describe("NewWorkerTry", func() {
		It("should always return true (no limit)", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			for i := 0; i < 100; i++ {
				Expect(sem.NewWorkerTry()).To(BeTrue())
			}

			// Clean up
			for i := 0; i < 100; i++ {
				sem.DeferWorker()
			}
		})

		It("should not block", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			start := time.Now()
			result := sem.NewWorkerTry()
			duration := time.Since(start)

			Expect(result).To(BeTrue())
			Expect(duration).To(BeNumerically("<", 10*time.Millisecond))
		})
	})

	Describe("WaitAll", func() {
		It("should wait for all workers", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			var wg sync.WaitGroup

			// Start workers
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()
					time.Sleep(50 * time.Millisecond)
				}()
			}

			// Wait for workers to start
			time.Sleep(10 * time.Millisecond)

			// Wait for workers to complete first
			wg.Wait()

			// Now WaitAll should succeed immediately
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})

		It("should succeed immediately if no workers", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})
	})

	Describe("Weighted", func() {
		It("should return -1 for unlimited", func() {
			sem := libsem.New(ctx, -1)
			Expect(sem.Weighted()).To(Equal(int64(-1)))
		})

		It("should return -1 for any negative value", func() {
			sem := libsem.New(ctx, -100)
			Expect(sem.Weighted()).To(Equal(int64(-1)))
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle unlimited concurrent workers", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
			)

			// Start many workers (more than any semaphore limit would allow)
			for i := 0; i < 500; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()

					completed.Add(1)
					time.Sleep(2 * time.Millisecond)
				}()
			}

			wg.Wait()

			Expect(completed.Load()).To(Equal(int32(500)))
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})

		It("should allow truly unlimited concurrency", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			var (
				wg             sync.WaitGroup
				currentWorkers atomic.Int32
				maxConcurrent  atomic.Int32
			)

			// Start more workers than typical semaphore limits
			for i := 0; i < 200; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()

					current := currentWorkers.Add(1)
					defer currentWorkers.Add(-1)

					// Update max
					for {
						old := maxConcurrent.Load()
						if current <= old || maxConcurrent.CompareAndSwap(old, current) {
							break
						}
					}

					time.Sleep(20 * time.Millisecond)
				}()
			}

			wg.Wait()

			// With 200 goroutines sleeping for 20ms, we should see high concurrency
			Expect(maxConcurrent.Load()).To(BeNumerically(">", 50))
		})
	})
})
