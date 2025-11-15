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
	"runtime"
	"time"

	libsem "github.com/nabbar/golib/semaphore/sem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semaphore Construction", func() {
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

	Describe("New with nbrSimultaneous == 0", func() {
		It("should create a weighted semaphore with MaxSimultaneous limit", func() {
			sem := libsem.New(ctx, 0)
			Expect(sem).ToNot(BeNil())

			// Should use MaxSimultaneous
			expected := libsem.MaxSimultaneous()
			Expect(sem.Weighted()).To(Equal(int64(expected)))
		})

		It("should be usable", func() {
			sem := libsem.New(ctx, 0)
			defer sem.DeferMain()

			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()
		})
	})

	Describe("New with nbrSimultaneous > 0", func() {
		It("should create a weighted semaphore with specified limit", func() {
			sem := libsem.New(ctx, 5)
			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(5)))
		})

		It("should create semaphore with limit of 1", func() {
			sem := libsem.New(ctx, 1)
			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(1)))
		})

		It("should create semaphore with large limit", func() {
			sem := libsem.New(ctx, 1000)
			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(1000)))
		})
	})

	Describe("New with nbrSimultaneous < 0", func() {
		It("should create a WaitGroup-based semaphore (unlimited)", func() {
			sem := libsem.New(ctx, -1)
			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(-1)))
		})

		It("should be usable with unlimited workers", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			// NewWorker should always succeed
			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()

			// NewWorkerTry should always return true
			Expect(sem.NewWorkerTry()).To(BeTrue())
			sem.DeferWorker()
		})

		It("should work with any negative value", func() {
			sem := libsem.New(ctx, -100)
			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(-1)))
		})
	})

	Describe("MaxSimultaneous", func() {
		It("should return GOMAXPROCS value", func() {
			expected := runtime.GOMAXPROCS(0)
			actual := libsem.MaxSimultaneous()
			Expect(actual).To(Equal(expected))
		})

		It("should return a positive value", func() {
			Expect(libsem.MaxSimultaneous()).To(BeNumerically(">", 0))
		})
	})

	Describe("SetSimultaneous", func() {
		It("should return MaxSimultaneous when n < 1", func() {
			expected := libsem.MaxSimultaneous()

			Expect(libsem.SetSimultaneous(0)).To(Equal(int64(expected)))
			Expect(libsem.SetSimultaneous(-1)).To(Equal(int64(expected)))
			Expect(libsem.SetSimultaneous(-100)).To(Equal(int64(expected)))
		})

		It("should return n when n is valid", func() {
			maxSim := libsem.MaxSimultaneous()

			if maxSim > 2 {
				Expect(libsem.SetSimultaneous(2)).To(Equal(int64(2)))
			}
			if maxSim > 5 {
				Expect(libsem.SetSimultaneous(5)).To(Equal(int64(5)))
			}
		})

		It("should return MaxSimultaneous when n > MaxSimultaneous", func() {
			expected := libsem.MaxSimultaneous()
			largeValue := expected + 1000

			Expect(libsem.SetSimultaneous(largeValue)).To(Equal(int64(expected)))
		})
	})

	Describe("New() method", func() {
		It("should create independent weighted semaphore", func() {
			sem1 := libsem.New(ctx, 5)
			defer sem1.DeferMain()

			sem2 := sem1.New()
			defer sem2.DeferMain()

			Expect(sem2).ToNot(BeNil())
			Expect(sem2.Weighted()).To(Equal(int64(5)))

			// Should be independent
			Expect(sem1.NewWorker()).ToNot(HaveOccurred())
			Expect(sem2.NewWorker()).ToNot(HaveOccurred())

			sem1.DeferWorker()
			sem2.DeferWorker()
		})

		It("should create independent WaitGroup semaphore", func() {
			sem1 := libsem.New(ctx, -1)
			defer sem1.DeferMain()

			sem2 := sem1.New()
			defer sem2.DeferMain()

			Expect(sem2).ToNot(BeNil())
			Expect(sem2.Weighted()).To(Equal(int64(-1)))
		})

		It("should inherit parent context", func() {
			parentCtx, parentCancel := context.WithCancel(ctx)
			defer parentCancel()

			sem1 := libsem.New(parentCtx, 5)
			defer sem1.DeferMain()

			sem2 := sem1.New()
			defer sem2.DeferMain()

			// Cancel parent
			parentCancel()
			time.Sleep(20 * time.Millisecond)

			// sem1 should be cancelled
			Expect(sem1.Err()).To(Equal(context.Canceled))

			// sem2 inherits from sem1, so should also be cancelled eventually
			Eventually(func() error {
				return sem2.Err()
			}, time.Second).Should(Equal(context.Canceled))
		})
	})
})
