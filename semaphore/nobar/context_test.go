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

var _ = Describe("Bar Context Interface", func() {
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

	Describe("Context methods", func() {
		It("should implement Deadline method", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			deadline, ok := bar.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).ToNot(BeZero())
		})

		It("should implement Done method", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			doneChan := bar.Done()
			Expect(doneChan).ToNot(BeNil())

			// Should not be closed yet
			select {
			case <-doneChan:
				Fail("Done channel should not be closed yet")
			default:
				// Expected
			}
		})

		It("should close Done channel when context is cancelled", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			doneChan := bar.Done()

			// Cancel the context
			localCancel()

			// Done channel should be closed
			Eventually(doneChan, time.Second).Should(BeClosed())
		})

		It("should implement Err method", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			// Initially should be nil
			Expect(bar.Err()).To(BeNil())
		})

		It("should return error after context cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)
			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			localCancel()

			// Wait for context to propagate
			time.Sleep(50 * time.Millisecond)

			Expect(bar.Err()).To(Equal(context.Canceled))
		})

		It("should implement Value method", func() {
			type key string
			const testKey key = "test-key"

			localCtx := context.WithValue(ctx, testKey, "test-value")
			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			val := bar.Value(testKey)
			Expect(val).To(Equal("test-value"))
		})

		It("should return nil for non-existent key", func() {
			sem := createTestSemaphore(ctx, 5)
			bar := libbar.New(sem, 100, false)

			val := bar.Value("non-existent-key")
			Expect(val).To(BeNil())
		})
	})

	Describe("Context timeout behavior", func() {
		It("should respect context timeout", func() {
			localCtx, localCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			defer localCancel()

			sem := createTestSemaphore(localCtx, 5)
			bar := libbar.New(sem, 100, false)

			doneChan := bar.Done()

			// Should close after timeout
			Eventually(doneChan, 200*time.Millisecond).Should(BeClosed())
			Expect(bar.Err()).To(Equal(context.DeadlineExceeded))
		})
	})
})
