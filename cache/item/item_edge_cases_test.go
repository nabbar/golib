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

package item_test

import (
	"time"

	. "github.com/nabbar/golib/cache/item"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache Item/Edge cases", func() {
	Context("LoadRemain with various states", func() {
		It("should handle LoadRemain after Clean", func() {
			itm := New[int](50*time.Millisecond, 100)
			itm.Clean()

			v, r, ok := itm.LoadRemain()
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))
		})

		It("should return correct elapsed time", func() {
			itm := New[int](100*time.Millisecond, 42)

			// Immediately after creation
			v1, r1, ok1 := itm.LoadRemain()
			Expect(ok1).To(BeTrue())
			Expect(v1).To(Equal(42))
			Expect(r1).To(BeNumerically(">=", 0))
			Expect(r1).To(BeNumerically("<", 100*time.Millisecond))

			// Wait a bit
			time.Sleep(20 * time.Millisecond)
			v2, r2, ok2 := itm.LoadRemain()
			Expect(ok2).To(BeTrue())
			Expect(v2).To(Equal(42))
			// r2 should be greater than r1 (more time has passed)
			Expect(r2).To(BeNumerically(">=", 20*time.Millisecond))
			Expect(r2).To(BeNumerically(">", r1))
		})

		It("should return false after exact expiration boundary", func() {
			itm := New[int](50*time.Millisecond, 99)

			// Wait past expiration
			time.Sleep(60 * time.Millisecond)

			v, r, ok := itm.LoadRemain()
			Expect(ok).To(BeFalse())
			Expect(v).To(Equal(0))
			Expect(r).To(Equal(time.Duration(0)))
		})

		It("should handle zero duration (never expires)", func() {
			itm := New[string](0, "permanent")

			// Multiple loads should always succeed
			for i := 0; i < 3; i++ {
				v, r, ok := itm.LoadRemain()
				Expect(ok).To(BeTrue())
				Expect(v).To(Equal("permanent"))
				Expect(r).To(Equal(time.Duration(0)))
				time.Sleep(10 * time.Millisecond)
			}
		})

		It("should handle Store after expiration", func() {
			itm := New[int](20*time.Millisecond, 1)

			// Let it expire
			time.Sleep(30 * time.Millisecond)
			v1, _, ok1 := itm.LoadRemain()
			Expect(ok1).To(BeFalse())
			Expect(v1).To(Equal(0))

			// Store new value
			itm.Store(2)
			v2, _, ok2 := itm.LoadRemain()
			Expect(ok2).To(BeTrue())
			Expect(v2).To(Equal(2))
		})
	})

	Context("Check method", func() {
		It("should return false after expiration", func() {
			itm := New[bool](10*time.Millisecond, true)

			Expect(itm.Check()).To(BeTrue())

			time.Sleep(15 * time.Millisecond)
			Expect(itm.Check()).To(BeFalse())
		})

		It("should return false after Clean", func() {
			itm := New[bool](0, true)
			Expect(itm.Check()).To(BeTrue())

			itm.Clean()
			Expect(itm.Check()).To(BeFalse())
		})
	})

	Context("Duration method", func() {
		It("should return configured duration", func() {
			dur := 123 * time.Millisecond
			itm := New[int](dur, 0)

			Expect(itm.Duration()).To(Equal(dur))
		})

		It("should return zero for no expiration", func() {
			itm := New[int](0, 0)

			Expect(itm.Duration()).To(Equal(time.Duration(0)))
		})
	})

	Context("Different data types", func() {
		It("should handle string values", func() {
			itm := New[string](0, "hello")

			v, ok := itm.Load()
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal("hello"))

			itm.Store("world")
			v2, ok2 := itm.Load()
			Expect(ok2).To(BeTrue())
			Expect(v2).To(Equal("world"))
		})

		It("should handle struct values", func() {
			type TestStruct struct {
				Name  string
				Value int
			}

			original := TestStruct{Name: "test", Value: 42}
			itm := New[TestStruct](0, original)

			v, ok := itm.Load()
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(original))
		})

		It("should handle pointer values", func() {
			value := 100
			itm := New[*int](0, &value)

			v, ok := itm.Load()
			Expect(ok).To(BeTrue())
			Expect(*v).To(Equal(100))
		})
	})

	Context("Concurrent access patterns", func() {
		It("should handle rapid Store and Load cycles", func() {
			itm := New[int](50*time.Millisecond, 0)

			for i := 0; i < 10; i++ {
				itm.Store(i)
				v, ok := itm.Load()
				Expect(ok).To(BeTrue())
				Expect(v).To(Equal(i))
			}
		})

		It("should handle Store updates before expiration", func() {
			itm := New[int](50*time.Millisecond, 1)

			time.Sleep(10 * time.Millisecond)
			itm.Store(2) // Update before expiration

			time.Sleep(10 * time.Millisecond)
			itm.Store(3) // Update again

			// Should still be valid as we updated
			time.Sleep(15 * time.Millisecond)
			v, ok := itm.Load()
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal(3))
		})
	})

	Context("Remain method", func() {
		It("should return accurate elapsed time", func() {
			itm := New[int](100*time.Millisecond, 5)

			r1, ok1 := itm.Remain()
			Expect(ok1).To(BeTrue())
			Expect(r1).To(BeNumerically(">=", 0))

			time.Sleep(30 * time.Millisecond)
			r2, ok2 := itm.Remain()
			Expect(ok2).To(BeTrue())
			// r2 should be greater than r1 (more time has elapsed)
			Expect(r2).To(BeNumerically(">", r1))
		})

		It("should return false after expiration", func() {
			itm := New[int](20*time.Millisecond, 5)

			time.Sleep(25 * time.Millisecond)
			r, ok := itm.Remain()
			Expect(ok).To(BeFalse())
			Expect(r).To(Equal(time.Duration(0)))
		})
	})
})
