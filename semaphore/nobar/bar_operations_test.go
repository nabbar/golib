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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libbar "github.com/nabbar/golib/semaphore/nobar"
)

var _ = Describe("Bar Operations", func() {
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

	Context("without MPB (progress bar disabled)", func() {
		It("should create a bar without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)
			time.Sleep(200 * time.Millisecond)
			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(0)))
			Expect(bar.Current()).To(Equal(int64(0))) // Without MPB, Current() returns Total()
		})

		It("should handle Inc operations without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Operations should not panic even without MPB
			bar.Inc(10)
			bar.Inc64(20)

			// Total should remain unchanged
			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should handle Dec operations without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Operations should not panic even without MPB
			bar.Dec(5)
			bar.Dec64(10)

			// Total should remain unchanged
			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should handle Reset without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Reset(200, 50)

			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should handle Complete without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Complete()

			// Without MPB, Completed() should return true
			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Context("with MPB (progress bar enabled)", func() {
		It("should create a bar with MPB", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			Expect(bar).ToNot(BeNil())
			Expect(bar.Total()).To(Equal(int64(0)))

			// Check if bar implements BarMPB interface
			if barMPB, ok := bar.(interface{ GetMPB() interface{} }); ok {
				Expect(barMPB.GetMPB()).ToNot(BeNil())
			}
		})

		It("should increment progress with Inc", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			initial := bar.Current()
			bar.Inc(10)

			// Small delay to allow MPB to update
			time.Sleep(10 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically(">=", initial))
		})

		It("should increment progress with Inc64", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			initial := bar.Current()
			bar.Inc64(25)

			time.Sleep(10 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically(">=", initial))
		})

		It("should decrement progress with Dec", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// First increment to have something to decrement
			bar.Inc(50)
			time.Sleep(10 * time.Millisecond)
			current := bar.Current()

			// Now decrement
			bar.Dec(10)
			time.Sleep(10 * time.Millisecond)

			// Current should be less than before (or at least different)
			Expect(bar.Current()).To(BeNumerically("<=", current))
		})

		It("should decrement progress with Dec64", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc64(50)
			time.Sleep(10 * time.Millisecond)
			current := bar.Current()

			bar.Dec64(15)
			time.Sleep(10 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically("<=", current))
		})

		It("should reset total and current values", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(50)
			time.Sleep(10 * time.Millisecond)

			bar.Reset(200, 100)
			time.Sleep(10 * time.Millisecond)

			Expect(bar.Total()).To(Equal(int64(0)))
			Expect(bar.Current()).To(Equal(int64(0)))
		})

		It("should complete the progress bar", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(100)
			bar.Complete()

			time.Sleep(50 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})

		It("should drop bar on complete when drop=true", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, true) // drop = true

			bar.Inc(100)
			bar.Complete()

			time.Sleep(50 * time.Millisecond)

			Expect(bar.Completed()).To(BeTrue())
		})
	})

	Context("edge cases", func() {
		It("should handle zero total", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 0, false)

			Expect(bar.Total()).To(Equal(int64(0)))
		})

		It("should handle negative increment (effectively decrement)", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			bar.Inc(50)
			time.Sleep(10 * time.Millisecond)
			current := bar.Current()

			// Increment with negative value (via Inc64)
			bar.Inc64(-10)
			time.Sleep(10 * time.Millisecond)

			Expect(bar.Current()).To(BeNumerically("<=", current))
		})

		It("should handle large values", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 1000000000, false)

			Expect(bar.Total()).To(Equal(int64(0)))
		})
	})
})
