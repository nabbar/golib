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
	DescribeTable("Cast should handle various types and zero values",
		func(src interface{}, expectedVal interface{}, expectedOk bool, targetType string) {
			switch targetType {
			case "bool":
				v, ok := libatm.Cast[bool](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "int":
				v, ok := libatm.Cast[int](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "int8":
				v, ok := libatm.Cast[int8](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "int16":
				v, ok := libatm.Cast[int16](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "int32":
				v, ok := libatm.Cast[int32](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "int64":
				v, ok := libatm.Cast[int64](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uint":
				v, ok := libatm.Cast[uint](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uint8":
				v, ok := libatm.Cast[uint8](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uint16":
				v, ok := libatm.Cast[uint16](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uint32":
				v, ok := libatm.Cast[uint32](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uint64":
				v, ok := libatm.Cast[uint64](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "uintptr":
				v, ok := libatm.Cast[uintptr](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "float32":
				v, ok := libatm.Cast[float32](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "float64":
				v, ok := libatm.Cast[float64](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "complex64":
				v, ok := libatm.Cast[complex64](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "complex128":
				v, ok := libatm.Cast[complex128](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			case "string":
				v, ok := libatm.Cast[string](src)
				Expect(ok).To(Equal(expectedOk))
				Expect(v).To(Equal(expectedVal))
			}
		},
		Entry("bool zero", false, false, false, "bool"),
		Entry("bool non-zero", true, true, true, "bool"),
		Entry("int zero", 0, 0, false, "int"),
		Entry("int non-zero", 1, 1, true, "int"),
		Entry("int8 zero", int8(0), int8(0), false, "int8"),
		Entry("int8 non-zero", int8(1), int8(1), true, "int8"),
		Entry("int16 zero", int16(0), int16(0), false, "int16"),
		Entry("int16 non-zero", int16(1), int16(1), true, "int16"),
		Entry("int32 zero", int32(0), int32(0), false, "int32"),
		Entry("int32 non-zero", int32(1), int32(1), true, "int32"),
		Entry("int64 zero", int64(0), int64(0), false, "int64"),
		Entry("int64 non-zero", int64(1), int64(1), true, "int64"),
		Entry("uint zero", uint(0), uint(0), false, "uint"),
		Entry("uint non-zero", uint(1), uint(1), true, "uint"),
		Entry("uint8 zero", uint8(0), uint8(0), false, "uint8"),
		Entry("uint8 non-zero", uint8(1), uint8(1), true, "uint8"),
		Entry("uint16 zero", uint16(0), uint16(0), false, "uint16"),
		Entry("uint16 non-zero", uint16(1), uint16(1), true, "uint16"),
		Entry("uint32 zero", uint32(0), uint32(0), false, "uint32"),
		Entry("uint32 non-zero", uint32(1), uint32(1), true, "uint32"),
		Entry("uint64 zero", uint64(0), uint64(0), false, "uint64"),
		Entry("uint64 non-zero", uint64(1), uint64(1), true, "uint64"),
		Entry("uintptr zero", uintptr(0), uintptr(0), false, "uintptr"),
		Entry("uintptr non-zero", uintptr(1), uintptr(1), true, "uintptr"),
		Entry("float32 zero", float32(0), float32(0), false, "float32"),
		Entry("float32 non-zero", float32(1.1), float32(1.1), true, "float32"),
		Entry("float64 zero", float64(0), float64(0), false, "float64"),
		Entry("float64 non-zero", float64(1.1), float64(1.1), true, "float64"),
		Entry("complex64 zero", complex64(0), complex64(0), false, "complex64"),
		Entry("complex64 non-zero", complex64(1+1i), complex64(1+1i), true, "complex64"),
		Entry("complex128 zero", complex128(0), complex128(0), false, "complex128"),
		Entry("complex128 non-zero", complex128(1+1i), complex128(1+1i), true, "complex128"),
		Entry("string zero", "", "", false, "string"),
		Entry("string non-zero", "a", "a", true, "string"),
		Entry("nil source", nil, 0, false, "int"), // Corrected expectedVal to 0
	)

	It("Cast should handle complex types with reflect.IsZero", func() {
		type st struct{ A int }
		v, ok := libatm.Cast[st](st{})
		Expect(ok).To(BeFalse())

		v, ok = libatm.Cast[st](st{A: 1})
		Expect(ok).To(BeTrue())
		Expect(v.A).To(Equal(1))

		var sl []int
		_, ok = libatm.Cast[[]int](sl)
		Expect(ok).To(BeFalse())

		sl = []int{1}
		_, ok = libatm.Cast[[]int](sl)
		Expect(ok).To(BeTrue())
	})

	It("IsEmpty should be true for zero values and false otherwise", func() {
		Expect(libatm.IsEmpty[int](0)).To(BeTrue())
		Expect(libatm.IsEmpty[int](1)).To(BeFalse())
		Expect(libatm.IsEmpty[int](nil)).To(BeTrue())

		type st struct{ A int }
		Expect(libatm.IsEmpty[st](st{})).To(BeTrue())
		Expect(libatm.IsEmpty[st](st{A: 2})).To(BeFalse())

		var s string
		Expect(libatm.IsEmpty[string](s)).To(BeTrue())
		Expect(libatm.IsEmpty[string]("a")).To(BeFalse())
	})
})
