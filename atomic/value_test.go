/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package atomic_test

import (
	libatm "github.com/nabbar/golib/atomic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Value", func() {
	Context("with default NewValue (no defaults set)", func() {
		var v libatm.Value[int]

		BeforeEach(func() {
			v = libatm.NewValue[int]()
		})

		It("should return zero value when empty", func() {
			Expect(v.Load()).To(Equal(0))
		})

		It("should store and load values", func() {
			v.Store(42)
			Expect(v.Load()).To(Equal(42))
		})

		It("should return zero value if stored value is zero", func() {
			v.Store(0)
			Expect(v.Load()).To(Equal(0))
		})
	})

	Context("with NewValueDefault", func() {
		var v libatm.Value[int]

		BeforeEach(func() {
			v = libatm.NewValueDefault[int](10, 20)
		})

		It("should return default load value when empty", func() {
			Expect(v.Load()).To(Equal(10))
		})

		It("should use default store value when storing zero", func() {
			v.Store(0)
			Expect(v.Load()).To(Equal(20))
		})

		It("should store and load normal values", func() {
			v.Store(42)
			Expect(v.Load()).To(Equal(42))
		})
	})

	Context("with SetDefaultLoad and SetDefaultStore", func() {
		var v libatm.Value[string]

		BeforeEach(func() {
			v = libatm.NewValue[string]()
		})

		It("should update default load value", func() {
			v.SetDefaultLoad("def-load")
			Expect(v.Load()).To(Equal("def-load"))
			v.Store("real")
			Expect(v.Load()).To(Equal("real"))
		})

		It("should update default store value", func() {
			v.SetDefaultStore("def-store")
			v.Store("")
			Expect(v.Load()).To(Equal("def-store"))
		})
	})

	Describe("Swap", func() {
		It("should swap values without defaults", func() {
			v := libatm.NewValue[int]()
			v.Store(1)
			old := v.Swap(2)
			Expect(old).To(Equal(1))
			Expect(v.Load()).To(Equal(2))
		})

		It("should swap first value correctly", func() {
			v := libatm.NewValue[int]()
			old := v.Swap(1)
			Expect(old).To(Equal(0))
			Expect(v.Load()).To(Equal(1))
		})

		It("should respect defaults during swap", func() {
			v := libatm.NewValueDefault[int](10, 20)
			old := v.Swap(0) // stores 20
			Expect(old).To(Equal(10))
			Expect(v.Load()).To(Equal(20))
		})

		It("should handle failed cast in swap", func() {
			v := libatm.NewValueDefault[string]("def-load", "def-store")
			// swap with something that is zero, triggering default store
			old := v.Swap("")
			Expect(old).To(Equal("def-load")) // initial load default is nil/0 in atomic.Value, so returns Load default
		})
	})

	Describe("CompareAndSwap", func() {
		It("should CAS values without defaults", func() {
			v := libatm.NewValue[int]()
			v.Store(1)
			swapped := v.CompareAndSwap(1, 2)
			Expect(swapped).To(BeTrue())
			Expect(v.Load()).To(Equal(2))

			swapped = v.CompareAndSwap(1, 3)
			Expect(swapped).To(BeFalse())
			Expect(v.Load()).To(Equal(2))
		})

		It("should respect defaults during CAS", func() {
			v := libatm.NewValueDefault[int](10, 20)
			v.Store(20)
			swapped := v.CompareAndSwap(0, 30) // 0 -> 20, so CAS(20, 30)
			Expect(swapped).To(BeTrue())
			Expect(v.Load()).To(Equal(30))
		})
	})
})
