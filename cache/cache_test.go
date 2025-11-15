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

var _ = Describe("Cache", func() {
	It("New should create a cache and ticker goroutine without panic", func() {
		c := New[string, int](context.Background(), 10*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })
		Expect(c).NotTo(BeNil())
	})

	It("Store and Load should work with no expiration", func() {
		c := New[string, string](context.Background(), 0)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("k1", "v1")
		v, r, ok := c.Load("k1")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal("v1"))
		Expect(r).To(Equal(time.Duration(0)))
	})

	It("Load should return false for missing or expired keys", func() {
		c := New[string, int](context.Background(), 20*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })

		// missing
		_, _, ok := c.Load("missing")
		Expect(ok).To(BeFalse())

		// expired
		c.Store("k", 1)
		time.Sleep(30 * time.Millisecond)
		_, _, ok2 := c.Load("k")
		Expect(ok2).To(BeFalse())
	})

	It("LoadOrStore should store on miss and return existing on hit", func() {
		c := New[string, int](context.Background(), 50*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })

		// miss => store
		v, r, ok := c.LoadOrStore("a", 10)
		Expect(ok).To(BeFalse())
		Expect(v).To(Equal(0))
		Expect(r).To(Equal(time.Duration(0)))

		// hit => return existing
		v2, r2, ok2 := c.LoadOrStore("a", 20)
		Expect(ok2).To(BeTrue())
		Expect(v2).To(Equal(10))
		Expect(r2).To(BeNumerically(">=", 0))
	})

	It("LoadAndDelete should return value then remove it", func() {
		c := New[string, string](context.Background(), 0)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("x", "y")
		v, ok := c.LoadAndDelete("x")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal("y"))

		_, _, ok2 := c.Load("x")
		Expect(ok2).To(BeFalse())
	})

	It("Swap should replace value and return old one if valid", func() {
		c := New[string, int](context.Background(), 100*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("s", 1)
		v, r, ok := c.Swap("s", 2)
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(1))
		Expect(r).To(BeNumerically(">=", 0))

		// swap on missing => false
		v2, r2, ok2 := c.Swap("missing", 3)
		Expect(ok2).To(BeFalse())
		Expect(v2).To(Equal(0))
		Expect(r2).To(Equal(time.Duration(0)))
	})

	It("Walk should iterate only valid items", func() {
		c := New[string, int](context.Background(), 20*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("a", 1)
		c.Store("b", 2)
		time.Sleep(25 * time.Millisecond) // expire
		c.Store("c", 3)

		var keys []string
		c.Walk(func(k string, v int, r time.Duration) bool {
			keys = append(keys, k)
			Expect(v).NotTo(Equal(0))
			Expect(r).To(BeNumerically(">=", 0))
			return true
		})
		Expect(keys).To(ContainElement("c"))
		Expect(keys).NotTo(ContainElement("a"))
		Expect(keys).NotTo(ContainElement("b"))
	})

	It("Merge should import items from another cache", func() {
		c1 := New[string, int](context.Background(), 0)
		c2 := New[string, int](context.Background(), 0)
		DeferCleanup(func() { _ = c1.Close() })
		DeferCleanup(func() { _ = c2.Close() })

		c1.Store("k1", 1)
		c2.Merge(c1)
		v, _, ok := c2.Load("k1")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(1))
	})

	It("Clone should deep copy data and keep independence", func() {
		c := New[string, int](context.Background(), 0)
		DeferCleanup(func() { _ = c.Close() })
		c.Store("k", 9)

		cloned, err := c.Clone(context.Background())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = cloned.Close() })

		// both contain the item initially
		v, _, ok := cloned.Load("k")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(9))

		// mutate original, cloned should stay consistent
		c.Store("k", 10)
		v2, _, ok2 := cloned.Load("k")
		Expect(ok2).To(BeTrue())
		Expect(v2).To(Equal(9))
	})

	It("Clean should remove all items and Expire should drop expired ones", func() {
		c := New[string, int](context.Background(), 10*time.Millisecond)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("a", 1)
		c.Store("b", 2)
		c.Clean()
		_, _, oka := c.Load("a")
		_, _, okb := c.Load("b")
		Expect(oka).To(BeFalse())
		Expect(okb).To(BeFalse())

		c.Store("x", 1)
		time.Sleep(20 * time.Millisecond)
		c.Expire()
		_, _, okx := c.Load("x")
		Expect(okx).To(BeFalse())
	})

	It("should proxy Context methods", func() {
		ctx, cancel := context.WithCancel(context.Background())
		c := New[string, int](ctx, 0)
		DeferCleanup(func() { _ = c.Close() })

		// Value falls back to underlying context
		type keyT string
		ctx2 := context.WithValue(ctx, keyT("k"), "v")
		_ = cancel // avoid unused, we will cancel new derived below
		c2 := New[string, int](ctx2, 0)
		DeferCleanup(func() { _ = c2.Close() })
		Expect(c2.Value(keyT("k")).(string)).To(Equal("v"))

		// Done/Err work
		c3 := New[string, int](context.Background(), 0)
		DeferCleanup(func() { _ = c3.Close() })
		// cancel by closing the cache's internal cancel func via Close
		Expect(c3.Err()).To(BeNil())
	})

	It("Delete should remove a key", func() {
		c := New[string, int](context.Background(), 0)
		DeferCleanup(func() { _ = c.Close() })

		c.Store("del", 42)
		v, _, ok := c.Load("del")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(42))

		c.Delete("del")
		_, _, ok2 := c.Load("del")
		Expect(ok2).To(BeFalse())
	})

	It("Close should cancel the context and set Err", func() {
		c := New[string, int](context.Background(), 0)
		// Closing should cancel internal context
		_ = c.Close()
		// give scheduler a moment
		Eventually(func() error { return c.Err() }).ShouldNot(BeNil())
	})
})
