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

var _ = Describe("Cache Item", func() {
	It("should initialize with zero value and set duration", func() {
		itm := New[int](0, 0)
		Expect(itm.Duration()).To(Equal(time.Duration(0)))
		v, ok := itm.Load()
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(0))
		Expect(itm.Check()).To(BeTrue())
	})

	It("should store and load values with no expiration", func() {
		itm := New[string](0, "hello")
		itm.Store("hello")
		v, r, ok := itm.LoadRemain()
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal("hello"))
		Expect(r).To(Equal(time.Duration(0)))
		v2, ok2 := itm.Load()
		Expect(ok2).To(BeTrue())
		Expect(v2).To(Equal("hello"))
		Expect(itm.Check()).To(BeTrue())
	})

	It("should expire after duration", func() {
		itm := New[int](20*time.Millisecond, 123)
		v, r, ok := itm.LoadRemain()
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(123))
		Expect(r).To(BeNumerically("<=", 20*time.Millisecond))

		// wait past expiration
		time.Sleep(30 * time.Millisecond)
		v2, r2, ok2 := itm.LoadRemain()
		Expect(ok2).To(BeFalse())
		Expect(v2).To(Equal(0))
		Expect(r2).To(Equal(time.Duration(0)))
		Expect(itm.Check()).To(BeFalse())
	})

	It("Clean should reset to zero and mark as expired", func() {
		itm := New[int](0, 5)
		itm.Clean()
		v, ok := itm.Load()
		Expect(ok).To(BeFalse())
		Expect(v).To(Equal(0))
		Expect(itm.Check()).To(BeFalse())
	})

	It("Remain should reflect time left when not expired", func() {
		itm := New[int](50*time.Millisecond, 7)
		time.Sleep(10 * time.Millisecond)
		r, ok := itm.Remain()
		Expect(ok).To(BeTrue())
		Expect(r).To(BeNumerically(">", 0))
		Expect(r).To(BeNumerically("<=", 50*time.Millisecond))
	})
})
