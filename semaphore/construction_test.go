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
	"time"

	libsem "github.com/nabbar/golib/semaphore"
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

	Describe("New without progress", func() {
		It("should create a semaphore without MPB", func() {
			sem := libsem.New(ctx, 5, false)
			defer sem.DeferMain()

			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(5)))

			// Type assert to access GetMPB
			if semPgb, ok := sem.(interface{ GetMPB() interface{} }); ok {
				Expect(semPgb.GetMPB()).To(BeNil())
			}
		})

		It("should be usable for worker management", func() {
			sem := libsem.New(ctx, 3, false)
			defer sem.DeferMain()

			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()
		})
	})

	Describe("New with progress", func() {
		It("should create a semaphore with MPB", func() {
			sem := libsem.New(ctx, 5, true)
			defer sem.DeferMain()

			Expect(sem).ToNot(BeNil())
			Expect(sem.Weighted()).To(Equal(int64(5)))

			// Type assert to access GetMPB
			if semPgb, ok := sem.(interface{ GetMPB() interface{} }); ok {
				Expect(semPgb.GetMPB()).ToNot(BeNil())
			}
		})

		It("should be usable for worker management with progress", func() {
			sem := libsem.New(ctx, 3, true)
			defer sem.DeferMain()

			Expect(sem.NewWorker()).ToNot(HaveOccurred())
			sem.DeferWorker()
		})
	})

	Describe("MaxSimultaneous", func() {
		It("should return a positive value", func() {
			max := libsem.MaxSimultaneous()
			Expect(max).To(BeNumerically(">", 0))
		})
	})

	Describe("SetSimultaneous", func() {
		It("should return MaxSimultaneous for invalid values", func() {
			expected := int64(libsem.MaxSimultaneous())
			Expect(libsem.SetSimultaneous(0)).To(Equal(expected))
			Expect(libsem.SetSimultaneous(-1)).To(Equal(expected))
		})

		It("should return the value when valid", func() {
			maxSim := libsem.MaxSimultaneous()
			if maxSim > 2 {
				Expect(libsem.SetSimultaneous(2)).To(Equal(int64(2)))
			}
		})
	})

	Describe("Clone", func() {
		It("should create an independent clone without progress", func() {
			sem1 := libsem.New(ctx, 5, false)
			defer sem1.DeferMain()

			sem2 := sem1.Clone()
			defer sem2.DeferMain()

			Expect(sem2).ToNot(BeNil())
			Expect(sem2.Weighted()).To(Equal(int64(5)))

			// Should be independent
			Expect(sem1.NewWorker()).ToNot(HaveOccurred())
			Expect(sem2.NewWorker()).ToNot(HaveOccurred())

			sem1.DeferWorker()
			sem2.DeferWorker()
		})

		It("should share MPB container when cloning with progress", func() {
			sem1 := libsem.New(ctx, 5, true)
			defer sem1.DeferMain()

			sem2 := sem1.Clone()
			defer sem2.DeferMain()

			Expect(sem2).ToNot(BeNil())

			// Type assert both to compare MPB containers
			semPgb1, ok1 := sem1.(interface{ GetMPB() interface{} })
			semPgb2, ok2 := sem2.(interface{ GetMPB() interface{} })

			if ok1 && ok2 {
				Expect(semPgb2.GetMPB()).To(Equal(semPgb1.GetMPB()))
			}
		})
	})

	Describe("New() method", func() {
		It("should create independent semaphore", func() {
			sem1 := libsem.New(ctx, 5, false)
			defer sem1.DeferMain()

			sem2 := sem1.New()
			Expect(sem2).ToNot(BeNil())
			Expect(sem2.Weighted()).To(Equal(int64(5)))
		})
	})
})
