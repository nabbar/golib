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
	. "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Constants and Defaults", func() {
	Describe("Size constants", func() {
		Context("Basic constant values", func() {
			It("should define SizeNul as zero", func() {
				Expect(SizeNul).To(Equal(Size(0)))
			})

			It("should define SizeUnit as one byte", func() {
				Expect(SizeUnit).To(Equal(Size(1)))
			})

			It("should define SizeKilo as 1024 bytes", func() {
				Expect(SizeKilo).To(Equal(Size(1024)))
			})

			It("should define SizeMega as 1048576 bytes", func() {
				Expect(SizeMega).To(Equal(Size(1048576)))
			})

			It("should define SizeGiga as 1073741824 bytes", func() {
				Expect(SizeGiga).To(Equal(Size(1073741824)))
			})

			It("should define SizeTera as 1099511627776 bytes", func() {
				Expect(SizeTera).To(Equal(Size(1099511627776)))
			})

			It("should define SizePeta as 1125899906842624 bytes", func() {
				Expect(SizePeta).To(Equal(Size(1125899906842624)))
			})

			It("should define SizeExa as 1152921504606846976 bytes", func() {
				Expect(SizeExa).To(Equal(Size(1152921504606846976)))
			})
		})

		Context("Binary progression", func() {
			It("should follow binary powers (powers of 1024)", func() {
				Expect(SizeKilo).To(Equal(Size(1 << 10)))
				Expect(SizeMega).To(Equal(Size(1 << 20)))
				Expect(SizeGiga).To(Equal(Size(1 << 30)))
				Expect(SizeTera).To(Equal(Size(1 << 40)))
				Expect(SizePeta).To(Equal(Size(1 << 50)))
				Expect(SizeExa).To(Equal(Size(1 << 60)))
			})

			It("should maintain multiplicative relationships", func() {
				Expect(SizeKilo).To(Equal(1024 * SizeUnit))
				Expect(SizeMega).To(Equal(1024 * SizeKilo))
				Expect(SizeGiga).To(Equal(1024 * SizeMega))
				Expect(SizeTera).To(Equal(1024 * SizeGiga))
				Expect(SizePeta).To(Equal(1024 * SizeTera))
				Expect(SizeExa).To(Equal(1024 * SizePeta))
			})

			It("should have 1024 ratio between consecutive units", func() {
				Expect(uint64(SizeMega / SizeKilo)).To(Equal(uint64(1024)))
				Expect(uint64(SizeGiga / SizeMega)).To(Equal(uint64(1024)))
				Expect(uint64(SizeTera / SizeGiga)).To(Equal(uint64(1024)))
				Expect(uint64(SizePeta / SizeTera)).To(Equal(uint64(1024)))
				Expect(uint64(SizeExa / SizePeta)).To(Equal(uint64(1024)))
			})
		})

		Context("Constant uniqueness", func() {
			It("should have all unique values", func() {
				constants := []Size{
					SizeNul,
					SizeUnit,
					SizeKilo,
					SizeMega,
					SizeGiga,
					SizeTera,
					SizePeta,
					SizeExa,
				}

				seen := make(map[Size]bool)
				for _, c := range constants {
					Expect(seen[c]).To(BeFalse(), "Duplicate constant value: %d", c)
					seen[c] = true
				}
			})

			It("should be in strictly ascending order", func() {
				Expect(SizeNul).To(BeNumerically("<", SizeUnit))
				Expect(SizeUnit).To(BeNumerically("<", SizeKilo))
				Expect(SizeKilo).To(BeNumerically("<", SizeMega))
				Expect(SizeMega).To(BeNumerically("<", SizeGiga))
				Expect(SizeGiga).To(BeNumerically("<", SizeTera))
				Expect(SizeTera).To(BeNumerically("<", SizePeta))
				Expect(SizePeta).To(BeNumerically("<", SizeExa))
			})
		})

		Context("Arithmetic safety", func() {
			It("should not overflow when multiplied within range", func() {
				Expect(func() {
					_ = SizeKilo * 1024
					_ = SizeMega * 1024
					_ = SizeGiga * 1024
					_ = SizeTera * 100
				}).NotTo(Panic())
			})

			It("should handle addition within range", func() {
				Expect(func() {
					_ = SizeKilo + SizeMega
					_ = SizeMega + SizeGiga
					_ = SizeGiga + SizeTera
				}).NotTo(Panic())
			})

			It("should handle subtraction", func() {
				Expect(func() {
					_ = SizeMega - SizeKilo
					_ = SizeGiga - SizeMega
					_ = SizeTera - SizeGiga
				}).NotTo(Panic())
			})
		})
	})

	Describe("Format constants", func() {
		It("should define FormatRound0", func() {
			Expect(FormatRound0).To(Equal("%.0f"))
		})

		It("should define FormatRound1", func() {
			Expect(FormatRound1).To(Equal("%.1f"))
		})

		It("should define FormatRound2", func() {
			Expect(FormatRound2).To(Equal("%.2f"))
		})

		It("should define FormatRound3", func() {
			Expect(FormatRound3).To(Equal("%.3f"))
		})

		It("should be valid format specifiers", func() {
			Expect(func() {
				_ = Size(1024).Format(FormatRound0)
				_ = Size(1024).Format(FormatRound1)
				_ = Size(1024).Format(FormatRound2)
				_ = Size(1024).Format(FormatRound3)
			}).NotTo(Panic())
		})

		It("should produce different precision outputs", func() {
			s := Size(1536) // 1.5 KB
			r0 := s.Format(FormatRound0)
			r1 := s.Format(FormatRound1)
			r2 := s.Format(FormatRound2)
			r3 := s.Format(FormatRound3)

			// Each should be different in format
			Expect(r0).To(MatchRegexp(`^\d+$`))
			Expect(r1).To(MatchRegexp(`^\d+\.\d$`))
			Expect(r2).To(MatchRegexp(`^\d+\.\d{2}$`))
			Expect(r3).To(MatchRegexp(`^\d+\.\d{3}$`))
		})
	})

	Describe("SetDefaultUnit()", func() {
		AfterEach(func() {
			// Reset to default after each test
			SetDefaultUnit('B')
		})

		Context("Setting custom units", func() {
			It("should set custom unit 'o'", func() {
				SetDefaultUnit('o')
				Expect(SizeKilo.Code(0)).To(ContainSubstring("o"))
			})

			It("should set custom unit 'i'", func() {
				SetDefaultUnit('i')
				Expect(SizeKilo.Code(0)).To(ContainSubstring("i"))
			})

			It("should set custom unit 'x'", func() {
				SetDefaultUnit('x')
				code := SizeKilo.Code(0)
				Expect(code).To(ContainSubstring("x"))
			})

			It("should affect all size constants", func() {
				SetDefaultUnit('i')

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
					code := c.Code(0)
					Expect(code).To(ContainSubstring("i"))
				}
			})
		})

		Context("Reset to default", func() {
			It("should reset with 'B'", func() {
				SetDefaultUnit('o')
				SetDefaultUnit('B')
				Expect(SizeKilo.Code(0)).To(ContainSubstring("B"))
			})

			It("should reset with zero value", func() {
				SetDefaultUnit('o')
				SetDefaultUnit(0)
				Expect(SizeKilo.Code(0)).To(ContainSubstring("B"))
			})

			It("should reset with empty rune", func() {
				SetDefaultUnit('o')
				SetDefaultUnit(rune(0))
				code := SizeKilo.Code(0)
				Expect(code).NotTo(BeEmpty())
				Expect(code).To(ContainSubstring("B"))
			})
		})

		Context("Persistence", func() {
			It("should persist across multiple calls", func() {
				SetDefaultUnit('o')

				code1 := SizeKilo.Code(0)
				code2 := SizeKilo.Code(0)
				code3 := SizeKilo.Code(0)

				Expect(code1).To(Equal(code2))
				Expect(code2).To(Equal(code3))
			})

			It("should be overridable with explicit unit parameter", func() {
				SetDefaultUnit('o')

				defaultCode := SizeKilo.Code(0)
				explicitCode := SizeKilo.Code('x')

				Expect(defaultCode).To(ContainSubstring("o"))
				Expect(explicitCode).To(ContainSubstring("x"))
			})
		})

		Context("Edge cases", func() {
			It("should handle unicode characters", func() {
				SetDefaultUnit('€')
				code := SizeKilo.Code(0)
				Expect(code).NotTo(BeEmpty())
			})

			It("should handle space character", func() {
				SetDefaultUnit(' ')
				code := SizeKilo.Code(0)
				Expect(code).NotTo(BeEmpty())
			})

			It("should handle digit characters", func() {
				SetDefaultUnit('1')
				code := SizeKilo.Code(0)
				Expect(code).To(ContainSubstring("1"))
			})
		})
	})

	Describe("Constant usage patterns", func() {
		It("should work in arithmetic expressions", func() {
			result := 5*SizeKilo + 2*SizeMega
			expected := Size(5*1024 + 2*1024*1024)
			Expect(result).To(Equal(expected))
		})

		It("should work in comparisons", func() {
			Expect(SizeMega).To(BeNumerically(">", SizeKilo))
			Expect(SizeGiga).To(BeNumerically(">", SizeMega))
			Expect(SizeTera).To(BeNumerically(">", SizeGiga))
		})

		It("should work as function arguments", func() {
			s, err := Parse("1MB")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", SizeMega, float64(SizeMega)*0.01))
		})

		It("should work in slice initialization", func() {
			sizes := []Size{
				SizeKilo,
				SizeMega,
				SizeGiga,
			}
			Expect(sizes).To(HaveLen(3))
			Expect(sizes[0]).To(Equal(SizeKilo))
			Expect(sizes[1]).To(Equal(SizeMega))
			Expect(sizes[2]).To(Equal(SizeGiga))
		})

		It("should work in map keys", func() {
			m := map[Size]string{
				SizeKilo: "KB",
				SizeMega: "MB",
				SizeGiga: "GB",
			}
			Expect(m[SizeKilo]).To(Equal("KB"))
			Expect(m[SizeMega]).To(Equal("MB"))
			Expect(m[SizeGiga]).To(Equal("GB"))
		})
	})

	Describe("Type properties", func() {
		It("should be based on uint64", func() {
			var s Size = 1024
			u := uint64(s)
			Expect(u).To(Equal(uint64(1024)))
		})

		It("should support type conversion", func() {
			s := Size(1024)
			Expect(uint64(s)).To(Equal(uint64(1024)))
		})

		It("should have zero value of 0", func() {
			var s Size
			Expect(s).To(Equal(SizeNul))
		})

		It("should be comparable", func() {
			s1 := SizeKilo
			s2 := SizeKilo
			s3 := SizeMega
			Expect(s1 == s2).To(BeTrue())
			Expect(s1 == s3).To(BeFalse())
		})

		It("should support ordering", func() {
			Expect(SizeKilo < SizeMega).To(BeTrue())
			Expect(SizeMega > SizeKilo).To(BeTrue())
			Expect(SizeKilo <= SizeKilo).To(BeTrue())
			Expect(SizeKilo >= SizeKilo).To(BeTrue())
		})
	})
})
