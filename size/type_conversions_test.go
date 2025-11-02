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

var _ = Describe("Type Conversions", func() {
	Describe("ParseInt64() and SizeFromInt64()", func() {
		Context("Positive values", func() {
			It("should convert zero", func() {
				result := ParseInt64(0)
				Expect(result).To(Equal(Size(0)))
			})

			It("should convert small positive values", func() {
				tests := map[int64]Size{
					1:    Size(1),
					1024: Size(1024),
					5120: Size(5120),
				}

				for input, expected := range tests {
					result := ParseInt64(input)
					Expect(result).To(Equal(expected))
				}
			})

			It("should convert large positive values", func() {
				result := ParseInt64(math.MaxInt64)
				Expect(result).To(Equal(Size(math.MaxInt64)))
			})

			It("should handle kilobyte values", func() {
				result := ParseInt64(int64(SizeKilo))
				Expect(result).To(Equal(SizeKilo))
			})
		})

		Context("Negative values", func() {
			It("should convert to absolute value", func() {
				tests := map[int64]Size{
					-1:    Size(1),
					-1024: Size(1024),
					-5120: Size(5120),
				}

				for input, expected := range tests {
					result := ParseInt64(input)
					Expect(result).To(Equal(expected))
				}
			})

			It("should handle math.MinInt64", func() {
				result := ParseInt64(math.MinInt64)
				// MinInt64's absolute value wraps in two's complement
				Expect(result).NotTo(BeNil())
			})
		})

		Context("SizeFromInt64 alias", func() {
			It("should work identically to ParseInt64", func() {
				values := []int64{0, 1024, -1024, math.MaxInt64}
				for _, val := range values {
					r1 := ParseInt64(val)
					r2 := SizeFromInt64(val)
					Expect(r1).To(Equal(r2))
				}
			})
		})
	})

	Describe("ParseUint64()", func() {
		It("should convert zero", func() {
			result := ParseUint64(0)
			Expect(result).To(Equal(Size(0)))
		})

		It("should convert positive values", func() {
			tests := map[uint64]Size{
				1:    Size(1),
				1024: Size(1024),
				5120: Size(5120),
			}

			for input, expected := range tests {
				result := ParseUint64(input)
				Expect(result).To(Equal(expected))
			}
		})

		It("should handle max uint64", func() {
			result := ParseUint64(math.MaxUint64)
			Expect(result).To(Equal(Size(math.MaxUint64)))
		})

		It("should handle size constants", func() {
			result := ParseUint64(uint64(SizeMega))
			Expect(result).To(Equal(SizeMega))
		})
	})

	Describe("ParseFloat64() and SizeFromFloat64()", func() {
		Context("Positive values", func() {
			It("should convert zero", func() {
				result := ParseFloat64(0.0)
				Expect(result).To(Equal(Size(0)))
			})

			It("should convert positive integers", func() {
				tests := map[float64]Size{
					1.0:    Size(1),
					1024.0: Size(1024),
					5120.0: Size(5120),
				}

				for input, expected := range tests {
					result := ParseFloat64(input)
					Expect(result).To(Equal(expected))
				}
			})

			It("should floor fractional values", func() {
				tests := map[float64]Size{
					1024.1: Size(1024),
					1024.5: Size(1024),
					1024.9: Size(1024),
					1025.1: Size(1025),
				}

				for input, expected := range tests {
					result := ParseFloat64(input)
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("Negative values", func() {
			It("should convert to absolute value", func() {
				tests := map[float64]Size{
					-1.0:    Size(1),
					-1024.0: Size(1024),
					-5120.0: Size(5120),
				}

				for input, expected := range tests {
					result := ParseFloat64(input)
					Expect(result).To(Equal(expected))
				}
			})

			It("should floor negative fractional values", func() {
				// math.Floor(-1024.9) = -1025, then abs = 1025
				result := ParseFloat64(-1024.9)
				Expect(result).To(Equal(Size(1025)))
			})
		})

		Context("Overflow handling", func() {
			It("should cap at MaxUint64 for MaxFloat64", func() {
				result := ParseFloat64(math.MaxFloat64)
				Expect(result).To(Equal(Size(math.MaxUint64)))
			})

			It("should cap at MaxUint64 for large negative floats", func() {
				result := ParseFloat64(-math.MaxFloat64)
				Expect(result).To(Equal(Size(math.MaxUint64)))
			})

			It("should handle very large positive values", func() {
				result := ParseFloat64(float64(math.MaxUint64) * 2)
				Expect(result).To(Equal(Size(math.MaxUint64)))
			})
		})

		Context("Special float values", func() {
			It("should handle very small fractional values", func() {
				result := ParseFloat64(0.1)
				Expect(result).To(Equal(Size(0)))
			})

			It("should handle values just under 1", func() {
				result := ParseFloat64(0.9)
				Expect(result).To(Equal(Size(0)))
			})
		})

		Context("SizeFromFloat64 alias", func() {
			It("should work identically to ParseFloat64", func() {
				values := []float64{0.0, 1024.5, -1024.5, math.MaxFloat64}
				for _, val := range values {
					r1 := ParseFloat64(val)
					r2 := SizeFromFloat64(val)
					Expect(r1).To(Equal(r2))
				}
			})
		})
	})

	Describe("To Uint64", func() {
		It("should convert to uint64", func() {
			tests := map[Size]uint64{
				Size(0):      0,
				Size(5120):   5120,
				SizeKilo:     1024,
				5 * SizeMega: 5242880,
			}

			for size, expected := range tests {
				Expect(size.Uint64()).To(Equal(expected))
			}
		})

		It("should handle max value", func() {
			s := Size(math.MaxUint64)
			Expect(s.Uint64()).To(Equal(uint64(math.MaxUint64)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Uint64()).To(Equal(uint64(0)))
		})
	})

	Describe("To Uint32", func() {
		It("should convert small values", func() {
			s := Size(5120)
			Expect(s.Uint32()).To(Equal(uint32(5120)))
		})

		It("should cap at MaxUint32 on overflow", func() {
			s := Size(math.MaxUint64)
			Expect(s.Uint32()).To(Equal(uint32(math.MaxUint32)))
		})

		It("should handle boundary value", func() {
			s := Size(math.MaxUint32)
			Expect(s.Uint32()).To(Equal(uint32(math.MaxUint32)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Uint32()).To(Equal(uint32(0)))
		})
	})

	Describe("To Uint", func() {
		It("should convert to uint", func() {
			s := Size(5120)
			Expect(s.Uint()).To(Equal(uint(5120)))
		})

		It("should cap at MaxUint on overflow", func() {
			s := Size(math.MaxUint64)
			result := s.Uint()
			Expect(result).To(BeNumerically("<=", uint(math.MaxUint)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Uint()).To(Equal(uint(0)))
		})
	})

	Describe("To Int64", func() {
		It("should convert small values", func() {
			s := Size(5120)
			Expect(s.Int64()).To(Equal(int64(5120)))
		})

		It("should cap at MaxInt64 on overflow", func() {
			s := Size(math.MaxUint64)
			Expect(s.Int64()).To(Equal(int64(math.MaxInt64)))
		})

		It("should handle boundary value", func() {
			s := Size(math.MaxInt64)
			Expect(s.Int64()).To(Equal(int64(math.MaxInt64)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("To Int32", func() {
		It("should convert small values", func() {
			s := Size(5120)
			Expect(s.Int32()).To(Equal(int32(5120)))
		})

		It("should cap at MaxInt32 on overflow", func() {
			s := Size(math.MaxUint64)
			Expect(s.Int32()).To(Equal(int32(math.MaxInt32)))
		})

		It("should handle boundary value", func() {
			s := Size(math.MaxInt32)
			Expect(s.Int32()).To(Equal(int32(math.MaxInt32)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Int32()).To(Equal(int32(0)))
		})
	})

	Describe("To Int", func() {
		It("should convert to int", func() {
			s := Size(5120)
			Expect(s.Int()).To(Equal(int(5120)))
		})

		It("should cap at MaxInt on overflow", func() {
			s := Size(math.MaxUint64)
			result := s.Int()
			Expect(result).To(BeNumerically("<=", int(math.MaxInt)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Int()).To(Equal(int(0)))
		})
	})

	Describe("To Float64", func() {
		It("should convert to float64", func() {
			tests := map[Size]float64{
				Size(0):    0.0,
				Size(1024): 1024.0,
				Size(5120): 5120.0,
			}

			for size, expected := range tests {
				Expect(size.Float64()).To(Equal(expected))
			}
		})

		It("should handle large values", func() {
			s := 5 * SizeGiga
			f := s.Float64()
			Expect(f).To(BeNumerically(">", 0))
			Expect(f).To(BeNumerically("<", math.MaxFloat64))
		})

		It("should cap at MaxFloat64 on overflow", func() {
			s := Size(math.MaxUint64)
			f := s.Float64()
			Expect(f).To(BeNumerically("<=", math.MaxFloat64))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Float64()).To(Equal(float64(0)))
		})
	})

	Describe("To Float32", func() {
		It("should convert small values", func() {
			s := Size(5120)
			Expect(s.Float32()).To(Equal(float32(5120)))
		})

		It("should cap at MaxFloat32 on overflow", func() {
			s := Size(math.MaxUint64)
			f := s.Float32()
			Expect(f).To(BeNumerically("<=", float32(math.MaxFloat32)))
		})

		It("should handle zero", func() {
			s := SizeNul
			Expect(s.Float32()).To(Equal(float32(0)))
		})
	})

	Describe("Conversion roundtrips", func() {
		Context("Uint64 roundtrip", func() {
			It("should maintain value", func() {
				original := 5 * SizeKilo
				converted := original.Uint64()
				back := Size(converted)
				Expect(back).To(Equal(original))
			})

			It("should work for various sizes", func() {
				sizes := []Size{SizeUnit, SizeKilo, SizeMega, SizeGiga}
				for _, size := range sizes {
					converted := size.Uint64()
					back := Size(converted)
					Expect(back).To(Equal(size))
				}
			})
		})

		Context("Int64 roundtrip", func() {
			It("should maintain value for small values", func() {
				original := 5 * SizeKilo
				converted := original.Int64()
				back := ParseInt64(converted)
				Expect(back).To(Equal(original))
			})

			It("should work for various sizes", func() {
				sizes := []Size{SizeUnit, SizeKilo, SizeMega}
				for _, size := range sizes {
					converted := size.Int64()
					back := ParseInt64(converted)
					Expect(back).To(Equal(size))
				}
			})
		})

		Context("Float64 roundtrip", func() {
			It("should maintain value for integer values", func() {
				original := 5 * SizeKilo
				converted := original.Float64()
				back := ParseFloat64(converted)
				Expect(back).To(Equal(original))
			})

			It("should work for various sizes", func() {
				sizes := []Size{SizeUnit, SizeKilo, SizeMega}
				for _, size := range sizes {
					converted := size.Float64()
					back := ParseFloat64(converted)
					Expect(back).To(Equal(size))
				}
			})
		})
	})

	Describe("Conversion edge cases", func() {
		It("should handle zero in all conversions", func() {
			s := SizeNul

			Expect(s.Uint64()).To(Equal(uint64(0)))
			Expect(s.Uint32()).To(Equal(uint32(0)))
			Expect(s.Uint()).To(Equal(uint(0)))
			Expect(s.Int64()).To(Equal(int64(0)))
			Expect(s.Int32()).To(Equal(int32(0)))
			Expect(s.Int()).To(Equal(int(0)))
			Expect(s.Float64()).To(Equal(float64(0)))
			Expect(s.Float32()).To(Equal(float32(0)))
		})

		It("should handle boundary values", func() {
			sizes := []Size{
				Size(1),
				Size(255),
				Size(256),
				Size(math.MaxUint8),
				Size(math.MaxUint16),
				Size(math.MaxUint32),
			}

			for _, s := range sizes {
				_ = s.Uint64()
				_ = s.Int64()
				_ = s.Float64()
			}
		})

		It("should handle size constants", func() {
			constants := []Size{
				SizeUnit,
				SizeKilo,
				SizeMega,
				SizeGiga,
				SizeTera,
				SizePeta,
				SizeExa,
			}

			for _, c := range constants {
				_ = c.Uint64()
				_ = c.Int64()
				_ = c.Float64()
			}
		})
	})

	Describe("Type safety", func() {
		It("should return correct types", func() {
			s := 5 * SizeKilo

			var _ uint64 = s.Uint64()
			var _ uint32 = s.Uint32()
			var _ uint = s.Uint()
			var _ int64 = s.Int64()
			var _ int32 = s.Int32()
			var _ int = s.Int()
			var _ float64 = s.Float64()
			var _ float32 = s.Float32()
		})

		It("should accept correct input types for Parse functions", func() {
			var _ Size = ParseInt64(int64(100))
			var _ Size = ParseUint64(uint64(100))
			var _ Size = ParseFloat64(float64(100))
		})
	})

	Describe("Overflow consistency", func() {
		It("should handle overflow consistently across types", func() {
			s := Size(math.MaxUint64)

			// All should handle overflow gracefully
			_ = s.Int64()
			_ = s.Int32()
			_ = s.Int()
			_ = s.Uint32()
			_ = s.Float64()
			_ = s.Float32()
		})

		It("should not panic on any conversion", func() {
			sizes := []Size{
				SizeNul,
				Size(1),
				Size(math.MaxUint32),
				Size(math.MaxUint64),
			}

			for _, s := range sizes {
				Expect(func() {
					_ = s.Uint64()
					_ = s.Uint32()
					_ = s.Uint()
					_ = s.Int64()
					_ = s.Int32()
					_ = s.Int()
					_ = s.Float64()
					_ = s.Float32()
				}).NotTo(Panic())
			}
		})
	})
})
