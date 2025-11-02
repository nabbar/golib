/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package cache_test

import (
	"context"
	"time"

	. "github.com/nabbar/golib/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache/Context methods", func() {
	Context("Deadline method", func() {
		It("should return no deadline when context has none", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			deadline, ok := c.Deadline()
			Expect(ok).To(BeFalse())
			Expect(deadline).To(BeZero())
		})

		It("should return deadline when context has one", func() {
			future := time.Now().Add(1 * time.Hour)
			ctx, cancel := context.WithDeadline(context.Background(), future)
			defer cancel()

			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			deadline, ok := c.Deadline()
			Expect(ok).To(BeTrue())
			Expect(deadline).To(BeTemporally("~", future, time.Second))
		})
	})

	Context("Done channel", func() {
		It("should return done channel from context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			done := c.Done()
			Expect(done).ToNot(BeNil())

			// Cancel context
			cancel()
			Eventually(done).Should(BeClosed())
		})

		It("should be closed when cache is closed", func() {
			c := New[string, int](context.Background(), 0)
			done := c.Done()
			Expect(done).ToNot(BeNil())

			_ = c.Close()
			Eventually(done).Should(BeClosed())
		})
	})

	Context("Value method", func() {
		It("should retrieve value from cache when key is of cache key type", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("mykey", 42)

			// Value should retrieve from cache
			val := c.Value("mykey")
			Expect(val).To(Equal(42))
		})

		It("should fallback to parent context Value when key not in cache", func() {
			type ctxKey string
			parentCtx := context.WithValue(context.Background(), ctxKey("external"), "value")
			c := New[string, int](parentCtx, 0)
			DeferCleanup(func() { _ = c.Close() })

			// External context value
			val := c.Value(ctxKey("external"))
			Expect(val).To(Equal("value"))
		})

		It("should return nil for non-existent values", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			val := c.Value("nonexistent")
			Expect(val).To(BeNil())
		})

		It("should return nil when cache key has expired", func() {
			c := New[string, int](context.Background(), 10*time.Millisecond)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("expired", 100)
			time.Sleep(20 * time.Millisecond)

			val := c.Value("expired")
			Expect(val).To(BeNil())
		})
	})

	Context("Err method", func() {
		It("should return nil when context is not cancelled", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			Expect(c.Err()).To(BeNil())
		})

		It("should return error when context is cancelled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			cancel()
			Eventually(c.Err).ShouldNot(BeNil())
		})

		It("should return error after cache is closed", func() {
			c := New[string, int](context.Background(), 0)
			_ = c.Close()

			Eventually(c.Err).ShouldNot(BeNil())
		})
	})
})
