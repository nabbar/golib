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
	"time"

	libsem "github.com/nabbar/golib/semaphore/sem"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Semaphore Context Interface", func() {
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

	Describe("Weighted Semaphore Context", func() {
		It("should implement Deadline", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			deadline, ok := sem.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).ToNot(BeZero())
		})

		It("should implement Done", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			doneChan := sem.Done()
			Expect(doneChan).ToNot(BeNil())

			// Should not be closed initially
			select {
			case <-doneChan:
				Fail("Done channel should not be closed initially")
			default:
				// Expected
			}
		})

		It("should close Done when context cancelled", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 5)
			defer sem.DeferMain()

			doneChan := sem.Done()

			localCancel()

			Eventually(doneChan, time.Second).Should(BeClosed())
		})

		It("should implement Err", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			// Initially no error
			Expect(sem.Err()).To(BeNil())
		})

		It("should return error after cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, 5)
			defer sem.DeferMain()

			localCancel()
			time.Sleep(20 * time.Millisecond)

			Expect(sem.Err()).To(Equal(context.Canceled))
		})

		It("should implement Value", func() {
			type key string
			const testKey key = "test"

			localCtx := context.WithValue(ctx, testKey, "test-value")
			sem := libsem.New(localCtx, 5)
			defer sem.DeferMain()

			Expect(sem.Value(testKey)).To(Equal("test-value"))
		})

		It("should return nil for non-existent key", func() {
			sem := libsem.New(ctx, 5)
			defer sem.DeferMain()

			Expect(sem.Value("non-existent")).To(BeNil())
		})

		It("should respect timeout", func() {
			localCtx, localCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer localCancel()

			sem := libsem.New(localCtx, 5)
			defer sem.DeferMain()

			doneChan := sem.Done()

			Eventually(doneChan, 200*time.Millisecond).Should(BeClosed())
			Expect(sem.Err()).To(Equal(context.DeadlineExceeded))
		})
	})

	Describe("WaitGroup Semaphore Context", func() {
		It("should implement Deadline", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			deadline, ok := sem.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).ToNot(BeZero())
		})

		It("should implement Done", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			doneChan := sem.Done()
			Expect(doneChan).ToNot(BeNil())

			select {
			case <-doneChan:
				Fail("Done channel should not be closed initially")
			default:
				// Expected
			}
		})

		It("should close Done when context cancelled", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, -1)
			defer sem.DeferMain()

			doneChan := sem.Done()

			localCancel()

			Eventually(doneChan, time.Second).Should(BeClosed())
		})

		It("should implement Err", func() {
			sem := libsem.New(ctx, -1)
			defer sem.DeferMain()

			Expect(sem.Err()).To(BeNil())
		})

		It("should return error after cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := libsem.New(localCtx, -1)
			defer sem.DeferMain()

			localCancel()
			time.Sleep(20 * time.Millisecond)

			Expect(sem.Err()).To(Equal(context.Canceled))
		})

		It("should implement Value", func() {
			type key string
			const testKey key = "test"

			localCtx := context.WithValue(ctx, testKey, "test-value")
			sem := libsem.New(localCtx, -1)
			defer sem.DeferMain()

			Expect(sem.Value(testKey)).To(Equal("test-value"))
		})
	})

	Describe("DeferMain", func() {
		It("should cancel context for weighted semaphore", func() {
			sem := libsem.New(ctx, 5)

			doneChan := sem.Done()

			sem.DeferMain()

			Eventually(doneChan, time.Second).Should(BeClosed())
			Expect(sem.Err()).To(Equal(context.Canceled))
		})

		It("should cancel context for WaitGroup semaphore", func() {
			sem := libsem.New(ctx, -1)

			doneChan := sem.Done()

			sem.DeferMain()

			Eventually(doneChan, time.Second).Should(BeClosed())
			Expect(sem.Err()).To(Equal(context.Canceled))
		})

		It("should be safe to call multiple times", func() {
			sem := libsem.New(ctx, 5)

			sem.DeferMain()
			sem.DeferMain() // Should not panic
			sem.DeferMain()
		})
	})
})
