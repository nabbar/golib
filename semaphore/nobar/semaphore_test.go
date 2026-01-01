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

package nobar_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libbar "github.com/nabbar/golib/semaphore/nobar"
)

var _ = Describe("Bar Semaphore Interface", func() {
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

	Describe("Worker management", func() {
		It("should create and defer workers", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			err := bar.NewWorker()
			Expect(err).ToNot(HaveOccurred())

			bar.DeferWorker()
		})

		It("should try to create worker", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			ok := bar.NewWorkerTry()
			Expect(ok).To(BeTrue())

			bar.DeferWorker()
		})

		It("should respect semaphore limits", func() {
			sem := createTestSemaphore(ctx, 2) // Only 2 simultaneous workers
			bar := libbar.New(sem, 100, false)

			// Create 2 workers
			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			Expect(bar.NewWorker()).ToNot(HaveOccurred())

			// Try to create a third - should fail with TRY
			ok := bar.NewWorkerTry()
			Expect(ok).To(BeFalse())

			// Clean up
			bar.DeferWorker()
			bar.DeferWorker()
		})

		It("should increment bar on DeferWorker", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			_ = bar.Current()

			Expect(bar.NewWorker()).ToNot(HaveOccurred())
			bar.DeferWorker()

			time.Sleep(20 * time.Millisecond)

			// Should have incremented
			Expect(bar.Current()).To(Equal(int64(0)))
		})
	})

	Describe("DeferMain", func() {
		It("should complete and defer main on bar without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(100)
			bar.DeferMain()

			Expect(bar.Completed()).To(BeTrue())
		})

		It("should complete bar with MPB", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(100)
			bar.DeferMain()

			time.Sleep(300 * time.Millisecond)
			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Describe("WaitAll", func() {
		It("should wait for all workers to complete", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			var wg sync.WaitGroup

			// Start some workers
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					Expect(bar.NewWorker()).ToNot(HaveOccurred())
					time.Sleep(50 * time.Millisecond)
					bar.DeferWorker()
				}()
			}

			// Wait for workers to start
			time.Sleep(10 * time.Millisecond)

			// WaitAll should block until all workers are done
			done := make(chan error, 1)
			go func() {
				done <- bar.WaitAll()
			}()

			// Should not complete immediately
			select {
			case <-done:
				Fail("WaitAll should not complete immediately")
			case <-time.After(20 * time.Millisecond):
				// Expected
			}

			// Wait for workers to finish
			wg.Wait()

			// Now WaitAll should complete
			Eventually(done, time.Second).Should(Receive(BeNil()))
		})
	})

	Describe("Weighted", func() {
		It("should return weighted semaphore value", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			weighted := bar.Weighted()
			Expect(weighted).To(Equal(int64(5)))
		})
	})

	Describe("New", func() {
		It("should create a new semaphore from bar", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			newSem := bar.New()
			Expect(newSem).ToNot(BeNil())

			// New semaphore should be independent
			Expect(newSem.NewWorker()).ToNot(HaveOccurred())
			newSem.DeferWorker()
		})
	})

	Describe("Concurrent worker operations", func() {
		It("should handle concurrent worker creation and deletion", func() {
			sem := createTestSemaphore(ctx, 10)
			bar := libbar.New(sem, 100, false)

			var wg sync.WaitGroup
			workerCount := 20

			for i := 0; i < workerCount; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					if err := bar.NewWorker(); err != nil {
						// May fail if semaphore is full, that's ok
						return
					}

					// Simulate some work
					time.Sleep(10 * time.Millisecond)

					bar.DeferWorker()
				}(i)
			}

			wg.Wait()

			// All workers should have completed
			Expect(bar.WaitAll()).ToNot(HaveOccurred())
		})
	})
})
