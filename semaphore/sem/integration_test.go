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

var _ = Describe("Semaphore Integration Tests", func() {
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
		It("should handle batch processing with weighted semaphore", func() {
			const (
				workers    = 5
				totalTasks = 50
			)

			sem := libsem.New(ctx, workers)
			defer sem.DeferMain()

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
			)

			for i := 0; i < totalTasks; i++ {
				wg.Add(1)
				go func(taskID int) {
					defer wg.Done()

					if err := sem.NewWorker(); err != nil {
						return
					}
					defer sem.DeferWorker()

					// Simulate task
					time.Sleep(10 * time.Millisecond)
					completed.Add(1)
				}(i)
			}

			wg.Wait()

			Expect(completed.Load()).To(Equal(int32(totalTasks)))
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})

		It("should handle batch processing with unlimited semaphore", func() {
			const totalTasks = 100

			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
			)

			for i := 0; i < totalTasks; i++ {
				wg.Add(1)
				go func(taskID int) {
					defer wg.Done()

					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()

					time.Sleep(5 * time.Millisecond)
					completed.Add(1)
				}(i)
			}

			wg.Wait()

			Expect(completed.Load()).To(Equal(int32(totalTasks)))
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})

		It("should handle mixed Try and blocking acquisitions", func() {
			sem := libsem.New(ctx, 3)
			defer sem.DeferMain()

			var (
				wg                sync.WaitGroup
				trySuccesses      atomic.Int32
				blockingSuccesses atomic.Int32
			)

			// Try acquisitions (may fail if semaphore is full)
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if sem.NewWorkerTry() {
						defer sem.DeferWorker()
						trySuccesses.Add(1)
						time.Sleep(20 * time.Millisecond)
					}
				}()
			}

			// Blocking acquisitions (will eventually succeed)
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := sem.NewWorker(); err == nil {
						defer sem.DeferWorker()
						blockingSuccesses.Add(1)
						time.Sleep(20 * time.Millisecond)
					}
				}()
			}

			wg.Wait()

			// Blocking should succeed, Try may partially succeed
			Expect(blockingSuccesses.Load()).To(Equal(int32(20)))
			Expect(trySuccesses.Load()).To(BeNumerically(">", 0))

			// Total should be reasonable
			total := trySuccesses.Load() + blockingSuccesses.Load()
			Expect(total).To(BeNumerically(">=", 20))
			Expect(total).To(BeNumerically("<=", 40))
		})

		It("should handle graceful shutdown", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 5)

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
				started   atomic.Int32
			)

			// Start many tasks
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := sem.NewWorker(); err != nil {
						return
					}
					defer sem.DeferWorker()

					started.Add(1)
					time.Sleep(50 * time.Millisecond)
					completed.Add(1)
				}()
			}

			// Let some tasks start
			time.Sleep(20 * time.Millisecond)

			// Cancel context (graceful shutdown)
			localCancel()

			wg.Wait()

			// Some tasks should have started
			Expect(started.Load()).To(BeNumerically(">", 0))

			// Completed should be <= started
			Expect(completed.Load()).To(BeNumerically("<=", started.Load()))

			sem.DeferMain()
		})
	})

	Describe("Performance scenarios", func() {
		It("should handle high-throughput with weighted semaphore", func() {
			sem := libsem.New(ctx, 20)
			defer sem.DeferMain()

			start := time.Now()

			var wg sync.WaitGroup
			for i := 0; i < 1000; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					if err := sem.NewWorker(); err == nil {
						defer sem.DeferWorker()
						time.Sleep(time.Millisecond)
					}
				}()
			}

			wg.Wait()
			duration := time.Since(start)

			// Should complete in reasonable time
			Expect(duration).To(BeNumerically("<", 5*time.Second))
		})

		It("should handle high-throughput with unlimited semaphore", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			start := time.Now()

			var wg sync.WaitGroup
			for i := 0; i < 1000; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					Expect(sem.NewWorker()).ToNot(HaveOccurred())
					defer sem.DeferWorker()
					time.Sleep(time.Millisecond)
				}()
			}

			wg.Wait()
			duration := time.Since(start)

			// Should be faster than weighted (all concurrent)
			Expect(duration).To(BeNumerically("<", 2*time.Second))
		})
	})

	Describe("Error recovery", func() {
		It("should recover from context cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 3)
			defer sem.DeferMain()

			// Fill semaphore
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			// Try to acquire in goroutine
			done := make(chan error, 1)
			go func() {
				done <- sem.NewWorker()
			}()

			// Cancel context
			localCancel()

			// Should receive error
			Eventually(done, time.Second).Should(Receive(HaveOccurred()))

			// Clean up
			sem.DeferWorker()
			sem.DeferWorker()
			sem.DeferWorker()
		})

		It("should handle rapid worker churn", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			for i := 0; i < 100; i++ {
				Expect(sem.NewWorker()).ToNot(HaveOccurred())
				sem.DeferWorker()
			}

			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})
	})

	Describe("Nested semaphores", func() {
		It("should work with nested weighted semaphores", func() {
			parent := libsem.New(ctx, 5)
			defer parent.DeferMain()

			child := parent.New()
			defer child.DeferMain()

			// Both should work independently
			Expect(parent.NewWorker()).ToNot(HaveOccurred())
			Expect(child.NewWorker()).ToNot(HaveOccurred())

			parent.DeferWorker()
			child.DeferWorker()

			// Both should have same weight
			Expect(child.Weighted()).To(Equal(parent.Weighted()))
		})

		It("should work with nested WaitGroup semaphores", func() {
			parent := libsem.New(ctx, -1)
			defer parent.DeferMain()

			child := parent.New()
			defer child.DeferMain()

			Expect(parent.NewWorker()).ToNot(HaveOccurred())
			Expect(child.NewWorker()).ToNot(HaveOccurred())

			parent.DeferWorker()
			child.DeferWorker()

			Expect(child.Weighted()).To(Equal(int64(-1)))
		})
	})
})
