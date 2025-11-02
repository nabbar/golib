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

var _ = Describe("Cache/Advanced operations", func() {
	Context("Merge operation", func() {
		It("should merge all items from source cache", func() {
			c1 := New[string, int](context.Background(), 0)
			c2 := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c1.Close() })
			DeferCleanup(func() { _ = c2.Close() })

			c1.Store("a", 1)
			c1.Store("b", 2)
			c1.Store("c", 3)

			c2.Merge(c1)

			va, _, oka := c2.Load("a")
			vb, _, okb := c2.Load("b")
			vc, _, okc := c2.Load("c")

			Expect(oka).To(BeTrue())
			Expect(okb).To(BeTrue())
			Expect(okc).To(BeTrue())
			Expect(va).To(Equal(1))
			Expect(vb).To(Equal(2))
			Expect(vc).To(Equal(3))
		})

		It("should replace existing items during merge", func() {
			c1 := New[string, int](context.Background(), 0)
			c2 := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c1.Close() })
			DeferCleanup(func() { _ = c2.Close() })

			c2.Store("key", 100)
			c1.Store("key", 200)

			c2.Merge(c1)

			v, _, ok := c2.Load("key")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(200))
		})

		It("should skip expired items during merge", func() {
			c1 := New[string, int](context.Background(), 10*time.Millisecond)
			c2 := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c1.Close() })
			DeferCleanup(func() { _ = c2.Close() })

			c1.Store("expired", 2)
			time.Sleep(15 * time.Millisecond) // Let "expired" expire
			c1.Store("valid", 1)              // Add after expiration

			c2.Merge(c1)

			// Expired item should not be in c2
			_, _, expired := c2.Load("expired")
			v1, _, ok1 := c2.Load("valid")

			Expect(expired).To(BeFalse())
			Expect(ok1).To(BeTrue())
			Expect(v1).To(Equal(1))
		})

		It("should handle merge with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c1 := New[string, int](ctx, 0)
			c2 := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c1.Close() })
			DeferCleanup(func() { _ = c2.Close() })

			c1.Store("a", 1)
			cancel()

			// Wait for context cancellation to propagate
			Eventually(c1.Err).ShouldNot(BeNil())

			// Merge should handle cancelled source
			c2.Merge(c1)

			// The merge may or may not succeed depending on timing
			// but it should not panic
		})
	})

	Context("LoadOrStore operation", func() {
		It("should store new value when key does not exist", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			v, r, loaded := c.LoadOrStore("newkey", 42)
			Expect(loaded).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))

			// Verify it was stored
			v2, _, ok := c.Load("newkey")
			Expect(ok).To(BeTrue())
			Expect(v2).To(Equal(42))
		})

		It("should load existing value when key exists", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("existing", 100)
			v, _, loaded := c.LoadOrStore("existing", 200)

			Expect(loaded).To(BeTrue())
			Expect(v).To(Equal(100)) // Original value, not new one
		})

		It("should store when existing item has expired", func() {
			c := New[string, int](context.Background(), 10*time.Millisecond)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 100)
			time.Sleep(15 * time.Millisecond) // Expire

			v, r, loaded := c.LoadOrStore("key", 200)
			Expect(loaded).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))

			// Verify new value was stored
			v2, _, ok := c.Load("key")
			Expect(ok).To(BeTrue())
			Expect(v2).To(Equal(200))
		})

		It("should handle LoadOrStore with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			v, r, loaded := c.LoadOrStore("key", 42)
			Expect(loaded).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))
		})
	})

	Context("LoadAndDelete operation", func() {
		It("should return false for non-existent key", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			v, ok := c.LoadAndDelete("missing")
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
		})

		It("should load and delete existing value", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 42)
			v, ok := c.LoadAndDelete("key")

			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(42))

			// Verify it was deleted
			_, _, exists := c.Load("key")
			Expect(exists).To(BeFalse())
		})

		It("should return false for expired item", func() {
			c := New[string, int](context.Background(), 10*time.Millisecond)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 42)
			time.Sleep(15 * time.Millisecond) // Expire

			v, ok := c.LoadAndDelete("key")
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
		})

		It("should handle LoadAndDelete with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 42)
			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			v, ok := c.LoadAndDelete("key")
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
		})
	})

	Context("Clone operation", func() {
		It("should return error when source context is cancelled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			cloned, err := c.Clone(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(cloned).To(BeNil())
		})

		It("should use parent context when nil context is provided", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 42)

			cloned, err := c.Clone(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloned).ToNot(BeNil())
			DeferCleanup(func() { _ = cloned.Close() })

			v, _, ok := cloned.Load("key")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(42))
		})
	})

	Context("Walk operation", func() {
		It("should stop walking when callback returns false", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("a", 1)
			c.Store("b", 2)
			c.Store("c", 3)

			count := 0
			c.Walk(func(k string, v int, d time.Duration) bool {
				count++
				return count < 2 // Stop after 2 iterations
			})

			Expect(count).To(Equal(2))
		})

		It("should handle walk with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("a", 1)
			c.Store("b", 2)

			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			count := 0
			c.Walk(func(k string, v int, d time.Duration) bool {
				count++
				return true
			})

			// Walk should stop early due to cancelled context
			Expect(count).To(BeNumerically("<=", 2))
		})
	})

	Context("Load operation", func() {
		It("should handle Load with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 42)
			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			v, r, ok := c.Load("key")
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))
		})
	})

	Context("Swap operation", func() {
		It("should return zero value when no existing item", func() {
			c := New[string, int](context.Background(), 0)
			DeferCleanup(func() { _ = c.Close() })

			v, r, ok := c.Swap("newkey", 42)
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))

			// Verify new value was stored
			v2, _, ok2 := c.Load("newkey")
			Expect(ok2).To(BeTrue())
			Expect(v2).To(Equal(42))
		})

		It("should handle Swap with cancelled context", func() {
			ctx, cancel := context.WithCancel(context.Background())
			c := New[string, int](ctx, 0)
			DeferCleanup(func() { _ = c.Close() })

			c.Store("key", 100)
			cancel()
			Eventually(c.Err).ShouldNot(BeNil())

			v, r, ok := c.Swap("key", 200)
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))
		})
	})
})
