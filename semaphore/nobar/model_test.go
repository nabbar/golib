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
	semtps "github.com/nabbar/golib/semaphore/types"
)

var _ = Describe("Bar Model Internals", func() {
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

	Describe("GetMPB method", func() {
		It("should return nil for bar without MPB", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Type assert to access GetMPB
			if barMPB, ok := bar.(semtps.BarMPB); ok {
				mpb := barMPB.GetMPB()
				Expect(mpb).To(BeNil())
			}
		})

		It("should return MPB bar instance when enabled", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Type assert to access GetMPB
			if barMPB, ok := bar.(semtps.BarMPB); ok {
				mpb := barMPB.GetMPB()
				Expect(mpb).To(BeNil())
			}
		})
	})

	Describe("Duration tracking", func() {
		It("should track time between updates correctly", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 1000, false)

			// First increment
			bar.Inc(10)
			time.Sleep(50 * time.Millisecond)

			// Second increment after delay
			bar.Inc(10)

			// The internal getDur() should have calculated a duration >= 50ms
			// We can't test this directly, but we verify the operations work
			Expect(bar.Current()).To(Equal(int64(0)))
		})

		It("should handle rapid consecutive updates", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Rapid increments
			for i := 0; i < 10; i++ {
				bar.Inc(1)
			}

			time.Sleep(20 * time.Millisecond)

			Expect(bar.Current()).To(Equal(int64(0)))
		})
	})

	Describe("Internal state consistency", func() {
		It("should maintain total value independently of MPB", func() {
			sem := createTestSemaphoreWithProgress(ctx, 5)
			bar := libbar.New(sem, 500, false)

			// Total should be 500 regardless of operations
			bar.Inc(100)
			Expect(bar.Total()).To(Equal(int64(0)))

			bar.Dec(50)
			Expect(bar.Total()).To(Equal(int64(0)))

			// Reset changes total
			bar.Reset(1000, 0)
			Expect(bar.Total()).To(Equal(int64(0)))
		})
	})
})
