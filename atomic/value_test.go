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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libatm "github.com/nabbar/golib/atomic"
)

var _ = Describe("Value[T]", func() {
	It("Load should return default load when not set", func() {
		v := libatm.NewValueDefault[int](42, 99)
		Expect(v.Load()).To(Equal(42))
	})

	It("Store should use default store for zero values", func() {
		v := libatm.NewValueDefault[int](1, 7)
		v.Store(0)
		Expect(v.Load()).To(Equal(7))
		v.Store(10)
		Expect(v.Load()).To(Equal(10))
	})

	It("Swap should return previous (default-load if unset) and store respecting default-store for zero", func() {
		v := libatm.NewValueDefault[string]("L", "S")
		// first swap: was unset -> returns default load "L", sets new respecting default-store
		old := v.Swap("")
		Expect(old).To(Equal("L"))
		Expect(v.Load()).To(Equal("S"))
		// second swap with non-empty
		old = v.Swap("B")
		Expect(old).To(Equal("S"))
		Expect(v.Load()).To(Equal("B"))
	})

	It("CompareAndSwap should treat zero old/new as default-store", func() {
		v := libatm.NewValueDefault[int](100, 5)
		// initial Store(0) -> default-store
		v.Store(0)
		Expect(v.Load()).To(Equal(5))
		// compare with old=0 should map to 5 and succeed, new=0 maps to 5 (no visible change)
		ok := v.CompareAndSwap(0, 0)
		Expect(ok).To(BeTrue())
		Expect(v.Load()).To(Equal(5))
		// now change to 8
		ok = v.CompareAndSwap(5, 8)
		Expect(ok).To(BeTrue())
		Expect(v.Load()).To(Equal(8))
		// failing case
		ok = v.CompareAndSwap(5, 9)
		Expect(ok).To(BeFalse())
		Expect(v.Load()).To(Equal(8))
	})

	It("SetDefaultLoad/SetDefaultStore should alter behavior", func() {
		v := libatm.NewValueDefault[int](0, 0)
		v.SetDefaultLoad(11)
		v.SetDefaultStore(22)
		Expect(v.Load()).To(Equal(11))
		v.Store(0)
		Expect(v.Load()).To(Equal(22))
	})
})
