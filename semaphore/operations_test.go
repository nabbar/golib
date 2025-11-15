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

package semaphore_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	libsem "github.com/nabbar/golib/semaphore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semaphore Operations", func() {
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

	Describe("Worker management without progress", func() {
		It("should acquire and release workers", func() {
			sem := libsem.New(ctx, 3, false)
			defer sem.DeferMain()

			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			sem.DeferWorker()
			sem.DeferWorker()
			sem.DeferWorker()
		})

		It("should respect concurrency limits", func() {
			sem := libsem.New(ctx, 2, false)
			defer sem.DeferMain()

			// Fill semaphore
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			Expect(sem.NewWorker()).ToNot(HaveOccurred())

			// Try should fail
			Expect(sem.NewWorkerTry()).To(BeFalse())

			// Release and try again
			sem.DeferWorker()
			Expect(sem.NewWorkerTry()).To(BeTrue())

			sem.DeferWorker()
		})

		It("should handle WaitAll", func() {
			sem := libsem.New(ctx, 3, false)
			defer sem.DeferMain()

			var wg sync.WaitGroup

			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := sem.NewWorker(); err == nil {
						defer sem.DeferWorker()
						time.Sleep(20 * time.Millisecond)
					}
				}()
			}

			wg.Wait()
			Expect(sem.WaitAll()).ToNot(HaveOccurred())
		})
	})

	Describe("Worker management with progress", func() {
		It("should work with progress bars", func() {
			sem := libsem.New(ctx, 3, true)
			defer sem.DeferMain()

			bar := sem.BarNumber("Tasks", "processing", 10, false, nil)

			var wg sync.WaitGroup
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := bar.NewWorker(); err == nil {
						defer bar.DeferWorker() // bar.DeferWorker calls Inc(1) then releases
						time.Sleep(10 * time.Millisecond)
					}
				}()
			}

			wg.Wait()
			time.Sleep(20 * time.Millisecond)
		})
	})

	Describe("Weighted", func() {
		It("should return correct weight", func() {
			sem := libsem.New(ctx, 7, false)
			defer sem.DeferMain()

			Expect(sem.Weighted()).To(Equal(int64(7)))
		})

		It("should return -1 for unlimited", func() {
			sem := libsem.New(ctx, -1, false)
			defer sem.DeferMain()

			Expect(sem.Weighted()).To(Equal(int64(-1)))
		})
	})

	Describe("Context interface", func() {
		It("should implement Deadline", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			deadline, ok := sem.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).ToNot(BeZero())
		})

		It("should implement Done", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			doneChan := sem.Done()
			Expect(doneChan).ToNot(BeNil())

			select {
			case <-doneChan:
				Fail("Should not be closed initially")
			default:
				// Expected
			}
		})

		It("should implement Err", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			Expect(sem.Err()).To(BeNil())
		})

		It("should implement Value", func() {
			type key string
			const testKey key = "test"

			localCtx := context.WithValue(ctx, testKey, "value")
			sem := libsem.New(localCtx, 5, false)
			defer sem.DeferMain()

			Expect(sem.Value(testKey)).To(Equal("value"))
		})
	})

	Describe("DeferMain", func() {
		It("should cleanup without progress", func() {
			sem := libsem.New(ctx, 5, false)

			doneChan := sem.Done()
			sem.DeferMain()

			Eventually(doneChan, time.Second).Should(BeClosed())
		})

		It("should cleanup with progress", func() {
			sem := libsem.New(ctx, 5, true)

			doneChan := sem.Done()
			sem.DeferMain()

			Eventually(doneChan, time.Second).Should(BeClosed())
		})
	})

	Describe("Concurrent operations", func() {
		It("should handle many concurrent workers", func() {
			sem := libsem.New(ctx, 10, false)
			defer sem.DeferMain()

			var (
				wg        sync.WaitGroup
				completed atomic.Int32
			)

			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := sem.NewWorker(); err == nil {
						defer sem.DeferWorker()
						completed.Add(1)
						time.Sleep(5 * time.Millisecond)
					}
				}()
			}

			wg.Wait()
			Expect(completed.Load()).To(Equal(int32(100)))
		})
	})
})
