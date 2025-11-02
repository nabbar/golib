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

var _ = Describe("Cast and IsEmpty", func() {
	It("Cast should return false for zero value comparison branch", func() {
		// For zero int
		var zero interface{} = int(0)
		v, ok := libatm.Cast[int](zero)
		Expect(ok).To(BeFalse())
		Expect(v).To(Equal(0))

		// For non-zero
		var one interface{} = int(1)
		v, ok = libatm.Cast[int](one)
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal(1))
	})

	It("IsEmpty should be true for zero values and false otherwise", func() {
		Expect(libatm.IsEmpty[int](0)).To(BeTrue())
		Expect(libatm.IsEmpty[int](1)).To(BeFalse())

		type st struct{ A int }
		Expect(libatm.IsEmpty[st](st{})).To(BeTrue())
		Expect(libatm.IsEmpty[st](st{A: 2})).To(BeFalse())

		var s string
		Expect(libatm.IsEmpty[string](s)).To(BeTrue())
		Expect(libatm.IsEmpty[string]("a")).To(BeFalse())
	})
})
