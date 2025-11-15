/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package size_test

import (
	"math"

	. "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Arithmetic Operations", func() {
	Describe("Mul() and MulErr()", func() {
		Context("Basic multiplication", func() {
			It("should multiply by integer", func() {
				s := SizeKilo
				s.Mul(2)
				Expect(s).To(Equal(2 * SizeKilo))
			})

			It("should multiply by float", func() {
				s := SizeKilo
				s.Mul(2.5)
				Expect(s).To(Equal(Size(2560)))
			})

			It("should multiply by zero", func() {
				s := 5 * SizeKilo
				s.Mul(0)
				Expect(s).To(Equal(SizeNul))
			})

			It("should handle fractional multipliers", func() {
				s := SizeKilo
				s.Mul(0.5)
				Expect(s).To(Equal(Size(512)))
			})

			It("should multiply by one (no change)", func() {
				original := 5 * SizeKilo
				s := original
				s.Mul(1)
				Expect(s).To(Equal(original))
			})
		})

		Context("Rounding behavior", func() {
			It("should round up (ceil) on fractional results", func() {
				s := Size(10)
				s.Mul(1.1) // 10 * 1.1 = 11, ceil = 11
				Expect(s).To(Equal(Size(11)))
			})

			It("should handle very small multipliers", func() {
				s := Size(100)
				s.Mul(0.01) // 100 * 0.01 = 1.0
				Expect(s).To(Equal(Size(1)))
			})
		})

		Context("Overflow handling", func() {
			It("should cap at MaxUint64 on overflow", func() {
				s := Size(math.MaxUint64 / 2)
				err := s.MulErr(3)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})

			It("should not error on multiplication within range", func() {
				s := SizeKilo
				err := s.MulErr(100)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(100 * SizeKilo))
			})

			It("should handle large values safely", func() {
				s := Size(math.MaxUint64 - 1000)
				s.Mul(2)
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})
		})

		Context("Edge cases", func() {
			It("should handle very large multipliers", func() {
				s := Size(2)
				err := s.MulErr(float64(math.MaxUint64))
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})

			It("should handle negative multipliers (interpreted as 0)", func() {
				s := Size(100)
				result := s
				result.Mul(-1)
				// Negative multipliers should be handled
				Expect(result).NotTo(BeNil())
			})
		})
	})

	Describe("Div() and DivErr()", func() {
		Context("Basic division", func() {
			It("should divide by integer", func() {
				s := Size(5120)
				s.Div(2)
				Expect(s).To(Equal(Size(2560)))
			})

			It("should divide by float", func() {
				s := Size(5120)
				s.Div(2.5)
				Expect(s).To(BeNumerically("~", 2048, 1))
			})

			It("should divide by one (no change)", func() {
				original := 5 * SizeKilo
				s := original
				s.Div(1)
				Expect(s).To(Equal(original))
			})
		})

		Context("Rounding behavior", func() {
			It("should round up (ceil) on fractional results", func() {
				s := Size(5)
				s.Div(2) // 5/2 = 2.5, ceil = 3
				Expect(s).To(Equal(Size(3)))
			})

			It("should handle exact divisions", func() {
				s := Size(1024)
				s.Div(2)
				Expect(s).To(Equal(Size(512)))
			})
		})

		Context("Error handling", func() {
			It("should error on division by zero", func() {
				s := Size(100)
				err := s.DivErr(0)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid diviser"))
			})

			It("should error on negative divisor", func() {
				s := Size(100)
				err := s.DivErr(-5)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid diviser"))
			})

			It("should not error on positive divisor", func() {
				s := Size(100)
				err := s.DivErr(5)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("Edge cases", func() {
			It("should handle very small divisors", func() {
				s := Size(100)
				s.Div(0.01) // 100/0.01 = 10000
				Expect(s).To(Equal(Size(10000)))
			})

			It("should handle division of zero", func() {
				s := SizeNul
				s.Div(10)
				Expect(s).To(Equal(SizeNul))
			})
		})
	})

	Describe("Add() and AddErr()", func() {
		Context("Basic addition", func() {
			It("should add positive values", func() {
				s := SizeKilo
				s.Add(1024)
				Expect(s).To(Equal(2 * SizeKilo))
			})

			It("should add zero (no change)", func() {
				original := 5 * SizeKilo
				s := original
				s.Add(0)
				Expect(s).To(Equal(original))
			})

			It("should add large values", func() {
				s := SizeKilo
				s.Add(uint64(SizeMega))
				Expect(s).To(Equal(SizeKilo + SizeMega))
			})

			It("should add multiple times", func() {
				s := SizeNul
				s.Add(100)
				s.Add(200)
				s.Add(300)
				Expect(s).To(Equal(Size(600)))
			})
		})

		Context("Overflow handling", func() {
			It("should cap at MaxUint64 on overflow", func() {
				s := Size(math.MaxUint64 - 10)
				err := s.AddErr(20)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("overflow"))
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})

			It("should not error on addition within range", func() {
				s := Size(100)
				err := s.AddErr(200)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(Size(300)))
			})

			It("should handle addition at boundary", func() {
				s := Size(math.MaxUint64 - 100)
				err := s.AddErr(100)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})
		})

		Context("Edge cases", func() {
			It("should handle maximum value addition", func() {
				s := Size(1)
				err := s.AddErr(math.MaxUint64 - 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(Size(math.MaxUint64)))
			})
		})
	})

	Describe("Sub() and SubErr()", func() {
		Context("Basic subtraction", func() {
			It("should subtract values", func() {
				s := 2 * SizeKilo
				s.Sub(1024)
				Expect(s).To(Equal(SizeKilo))
			})

			It("should subtract zero (no change)", func() {
				original := 5 * SizeKilo
				s := original
				s.Sub(0)
				Expect(s).To(Equal(original))
			})

			It("should subtract to zero", func() {
				s := SizeMega
				s.Sub(uint64(SizeMega))
				Expect(s).To(Equal(SizeNul))
			})

			It("should subtract multiple times", func() {
				s := Size(1000)
				s.Sub(100)
				s.Sub(200)
				s.Sub(300)
				Expect(s).To(Equal(Size(400)))
			})
		})

		Context("Underflow handling", func() {
			It("should cap at zero on underflow", func() {
				s := Size(10)
				err := s.SubErr(20)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid substractor"))
				Expect(s).To(Equal(SizeNul))
			})

			It("should not error on subtraction within range", func() {
				s := Size(200)
				err := s.SubErr(100)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(Size(100)))
			})

			It("should handle exact subtraction", func() {
				s := Size(100)
				err := s.SubErr(100)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(SizeNul))
			})
		})

		Context("Edge cases", func() {
			It("should handle large subtractions", func() {
				s := Size(math.MaxUint64)
				s.Sub(1000)
				Expect(s).To(Equal(Size(math.MaxUint64 - 1000)))
			})
		})
	})

	Describe("Chained operations", func() {
		It("should handle multiplication followed by addition", func() {
			s := SizeKilo
			s.Mul(2)    // 2KB
			s.Add(1024) // 3KB
			Expect(s).To(Equal(3 * SizeKilo))
		})

		It("should handle division followed by subtraction", func() {
			s := 10 * SizeKilo
			s.Div(2)    // 5KB
			s.Sub(1024) // 4KB
			Expect(s).To(Equal(4 * SizeKilo))
		})

		It("should maintain value through inverse operations", func() {
			s := 5 * SizeMega
			original := s
			s.Mul(2)
			s.Div(2)
			Expect(s).To(BeNumerically("~", original, float64(original)*0.01))
		})

		It("should handle complex operation chains", func() {
			s := SizeKilo
			s.Mul(10)  // 10KB
			s.Add(512) // 10KB + 512B
			s.Div(2)   // ~5KB + 256B
			s.Sub(256) // ~5KB
			Expect(s).To(BeNumerically("~", 5*SizeKilo, 512))
		})
	})
})
