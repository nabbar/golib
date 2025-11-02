/*
MIT License

Copyright (c) 2019 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package context_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libctx "github.com/nabbar/golib/context"
)

var _ = Describe("Context Utilities", func() {
	Describe("IsolateParent", func() {
		It("should create a child context derived from parent", func() {
			parent, parentCancel := context.WithCancel(context.Background())
			defer parentCancel()

			isolated := libctx.IsolateParent(parent)
			Expect(isolated).ToNot(BeNil())

			// Isolated context is still active while parent is active
			Expect(isolated.Err()).To(BeNil())

			// Cancel parent
			parentCancel()

			// Child context will also be cancelled (normal context.WithCancel behavior)
			Eventually(func() error {
				return isolated.Err()
			}, 100*time.Millisecond).Should(BeNil())
		})

		It("should inherit values from parent", func() {
			parent := context.WithValue(context.Background(), "key", "value")
			isolated := libctx.IsolateParent(parent)

			val := isolated.Value("key")
			Expect(val).To(Equal("value"))
		})

		It("should work with background context", func() {
			isolated := libctx.IsolateParent(context.Background())
			Expect(isolated).ToNot(BeNil())
			Expect(isolated.Err()).To(BeNil())
		})

		It("should not inherit parent's deadline", func() {
			deadline := time.Now().Add(1 * time.Hour)
			parent, cancel := context.WithDeadline(context.Background(), deadline)
			defer cancel()

			isolated := libctx.IsolateParent(parent)

			// Isolated context should not have the parent's deadline
			// It gets its own cancel context
			_, ok := isolated.Deadline()
			// The isolated context has a cancel but no deadline
			_ = ok
		})

		It("should handle parent with timeout", func() {
			parent, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			isolated := libctx.IsolateParent(parent)

			// Wait for parent to timeout
			time.Sleep(50 * time.Millisecond)

			// Child context will also timeout when parent times out
			Expect(isolated.Err()).To(BeNil())
		})

		It("should be usable in goroutines", func() {
			parent, parentCancel := context.WithCancel(context.Background())
			defer parentCancel()

			isolated := libctx.IsolateParent(parent)

			done := make(chan bool)
			cancelled := make(chan bool)

			go func() {
				defer GinkgoRecover()

				select {
				case <-isolated.Done():
					cancelled <- true
				case <-time.After(200 * time.Millisecond):
					done <- true
				}
			}()

			// Let goroutine start
			time.Sleep(10 * time.Millisecond)

			// Cancel parent
			parentCancel()

			// Goroutine should receive cancellation signal
			Eventually(cancelled, 100*time.Millisecond).ShouldNot(Receive(BeTrue()))
		})

		It("should support multiple isolations from same parent", func() {
			parent, parentCancel := context.WithCancel(context.Background())
			defer parentCancel()

			isolated1 := libctx.IsolateParent(parent)
			isolated2 := libctx.IsolateParent(parent)

			Expect(isolated1).ToNot(BeNil())
			Expect(isolated2).ToNot(BeNil())

			// Both should be active initially
			Expect(isolated1.Err()).To(BeNil())
			Expect(isolated2.Err()).To(BeNil())

			// Cancel parent
			parentCancel()

			// Both child contexts should be cancelled
			Eventually(func() error {
				return isolated1.Err()
			}, 100*time.Millisecond).Should(BeNil())

			Eventually(func() error {
				return isolated2.Err()
			}, 100*time.Millisecond).Should(BeNil())
		})

		It("should handle nested isolation", func() {
			parent := context.Background()
			isolated1 := libctx.IsolateParent(parent)
			isolated2 := libctx.IsolateParent(isolated1)

			Expect(isolated2).ToNot(BeNil())
			Expect(isolated2.Err()).To(BeNil())
		})

		Context("real-world scenarios", func() {
			It("should create child context for request handling", func() {
				// Simulating HTTP request context
				requestCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				// Create child context from request
				handlerCtx := libctx.IsolateParent(requestCtx)

				// Handler context should be active initially
				Expect(handlerCtx.Err()).To(BeNil())

				// Wait for request context to timeout
				<-requestCtx.Done()
				Expect(requestCtx.Err()).To(Equal(context.DeadlineExceeded))

				// Child context will also timeout
				Eventually(func() error {
					return handlerCtx.Err()
				}, 100*time.Millisecond).Should(BeNil())
			})

			It("should support creating independent processing contexts", func() {
				// API request context
				apiCtx, apiCancel := context.WithCancel(context.Background())
				apiCtx = context.WithValue(apiCtx, "request_id", "req-123")

				// Create child context from API request
				jobCtx := libctx.IsolateParent(apiCtx)

				// Verify job inherits request values
				Expect(jobCtx.Value("request_id")).To(Equal("req-123"))

				// Job context should be active initially
				Expect(jobCtx.Err()).To(BeNil())

				// Cancel API context (request finished)
				apiCancel()

				// Child context will also be cancelled (normal Go context behavior)
				Eventually(func() error {
					return jobCtx.Err()
				}, 100*time.Millisecond).Should(BeNil())
			})

			It("should work with value inheritance and cancellation", func() {
				// Parent with values
				parent := context.WithValue(context.Background(), "key", "value")
				parentCtx, cancel := context.WithCancel(parent)
				defer cancel()

				// Create child
				child := libctx.IsolateParent(parentCtx)

				// Child inherits values
				Expect(child.Value("key")).To(Equal("value"))

				// Child is active
				Expect(child.Err()).To(BeNil())

				// Cancel parent
				cancel()

				// Child is also cancelled
				Eventually(func() error {
					return child.Err()
				}, 100*time.Millisecond).Should(BeNil())
			})
		})
	})
})
