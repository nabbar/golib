/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package perm_test

import (
	"math"
	"os"

	. "github.com/nabbar/golib/file/perm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Permission Formatting", func() {
	Describe("String", func() {
		It("should format 0644 as string", func() {
			perm := Perm(0644)
			Expect(perm.String()).To(Equal("0644"))
		})

		It("should format 0755 as string", func() {
			perm := Perm(0755)
			Expect(perm.String()).To(Equal("0755"))
		})

		It("should format 0777 as string", func() {
			perm := Perm(0777)
			Expect(perm.String()).To(Equal("0777"))
		})

		It("should format 0 as string", func() {
			perm := Perm(0)
			Expect(perm.String()).To(Equal("0"))
		})

		It("should format permission with special bits", func() {
			perm := Perm(04755)
			Expect(perm.String()).To(Equal("04755"))
		})
	})

	Describe("FileMode", func() {
		It("should convert to os.FileMode", func() {
			perm := Perm(0644)
			fileMode := perm.FileMode()
			Expect(fileMode).To(Equal(os.FileMode(0644)))
		})

		It("should convert 0755 to os.FileMode", func() {
			perm := Perm(0755)
			fileMode := perm.FileMode()
			Expect(fileMode).To(Equal(os.FileMode(0755)))
		})

		It("should handle special permissions", func() {
			perm := Perm(04755)
			fileMode := perm.FileMode()
			Expect(fileMode).To(Equal(os.FileMode(04755)))
		})
	})

	Describe("Int64", func() {
		It("should convert to int64", func() {
			perm := Perm(0644)
			Expect(perm.Int64()).To(Equal(int64(0644)))
		})

		It("should convert 0755 to int64", func() {
			perm := Perm(0755)
			Expect(perm.Int64()).To(Equal(int64(0755)))
		})

		It("should handle large value gracefully", func() {
			// Use maximum permission value that fits in uint32
			perm := Perm(math.MaxUint32)
			result := perm.Int64()
			// MaxUint32 fits in int64, so it should convert directly
			Expect(result).To(Equal(int64(math.MaxUint32)))
		})

		It("should convert 0 to int64", func() {
			perm := Perm(0)
			Expect(perm.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("Int32", func() {
		It("should convert to int32", func() {
			perm := Perm(0644)
			Expect(perm.Int32()).To(Equal(int32(0644)))
		})

		It("should convert 0755 to int32", func() {
			perm := Perm(0755)
			Expect(perm.Int32()).To(Equal(int32(0755)))
		})

		It("should handle large value overflow gracefully", func() {
			// Use a value larger than MaxInt32
			perm := Perm(uint64(math.MaxInt32) + 1)
			Expect(perm.Int32()).To(Equal(int32(math.MaxInt32)))
		})

		It("should convert 0 to int32", func() {
			perm := Perm(0)
			Expect(perm.Int32()).To(Equal(int32(0)))
		})
	})

	Describe("Int", func() {
		It("should convert to int", func() {
			perm := Perm(0644)
			Expect(perm.Int()).To(Equal(int(0644)))
		})

		It("should convert 0755 to int", func() {
			perm := Perm(0755)
			Expect(perm.Int()).To(Equal(int(0755)))
		})

		It("should handle large value gracefully", func() {
			// Use maximum permission value
			perm := Perm(math.MaxUint32)
			result := perm.Int()
			// On most systems, this should convert correctly or cap at MaxInt
			Expect(result).To(BeNumerically("<=", int(math.MaxInt)))
			Expect(result).To(BeNumerically(">=", 0))
		})
	})

	Describe("Uint64", func() {
		It("should convert to uint64", func() {
			perm := Perm(0644)
			Expect(perm.Uint64()).To(Equal(uint64(0644)))
		})

		It("should convert 0755 to uint64", func() {
			perm := Perm(0755)
			Expect(perm.Uint64()).To(Equal(uint64(0755)))
		})

		It("should handle large value", func() {
			perm := Perm(0xFFFFFFFF)
			Expect(perm.Uint64()).To(Equal(uint64(0xFFFFFFFF)))
		})
	})

	Describe("Uint32", func() {
		It("should convert to uint32", func() {
			perm := Perm(0644)
			Expect(perm.Uint32()).To(Equal(uint32(0644)))
		})

		It("should convert 0755 to uint32", func() {
			perm := Perm(0755)
			Expect(perm.Uint32()).To(Equal(uint32(0755)))
		})

		It("should handle maximum uint32 value", func() {
			// Use MaxUint32 value
			perm := Perm(math.MaxUint32)
			Expect(perm.Uint32()).To(Equal(uint32(math.MaxUint32)))
		})
	})

	Describe("Uint", func() {
		It("should convert to uint", func() {
			perm := Perm(0644)
			Expect(perm.Uint()).To(Equal(uint(0644)))
		})

		It("should convert 0755 to uint", func() {
			perm := Perm(0755)
			Expect(perm.Uint()).To(Equal(uint(0755)))
		})

		It("should handle large value gracefully", func() {
			// Use maximum permission value
			perm := Perm(math.MaxUint32)
			result := perm.Uint()
			// On most systems, this should convert correctly or cap at MaxUint
			Expect(result).To(BeNumerically("<=", uint(math.MaxUint)))
			Expect(result).To(BeNumerically(">=", 0))
		})
	})

	Describe("Round-trip Conversions", func() {
		It("should round-trip through String and Parse", func() {
			original := Perm(0644)
			str := original.String()
			parsed, err := Parse(str)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed).To(Equal(original))
		})

		It("should round-trip through Int64 and ParseInt64", func() {
			original := Perm(0755)
			int64Val := original.Int64()
			parsed, err := ParseInt64(int64Val)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed).To(Equal(original))
		})

		It("should round-trip through Int and ParseInt", func() {
			original := Perm(0777)
			intVal := original.Int()
			parsed, err := ParseInt(intVal)
			Expect(err).ToNot(HaveOccurred())
			Expect(parsed).To(Equal(original))
		})
	})
})
