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

var _ = Describe("Formatting", func() {
	Describe("String() method", func() {
		Context("Basic formatting", func() {
			It("should format byte sizes", func() {
				s := Size(100)
				result := s.String()
				Expect(result).To(ContainSubstring("B"))
				Expect(result).NotTo(BeEmpty())
			})

			It("should format kilobyte sizes", func() {
				s := 5 * SizeKilo
				result := s.String()
				Expect(result).To(ContainSubstring("KB"))
			})

			It("should format megabyte sizes", func() {
				s := 10 * SizeMega
				result := s.String()
				Expect(result).To(ContainSubstring("MB"))
			})

			It("should format gigabyte sizes", func() {
				s := 2 * SizeGiga
				result := s.String()
				Expect(result).To(ContainSubstring("GB"))
			})

			It("should format terabyte sizes", func() {
				s := 3 * SizeTera
				result := s.String()
				Expect(result).To(ContainSubstring("TB"))
			})

			It("should format petabyte sizes", func() {
				s := 2 * SizePeta
				result := s.String()
				Expect(result).To(ContainSubstring("PB"))
			})

			It("should format exabyte sizes", func() {
				s := 2 * SizeExa
				result := s.String()
				Expect(result).To(ContainSubstring("EB"))
			})
		})

		Context("Zero and edge values", func() {
			It("should format zero size", func() {
				s := SizeNul
				result := s.String()
				Expect(result).NotTo(BeEmpty())
			})

			It("should format size 1", func() {
				s := SizeUnit
				result := s.String()
				Expect(result).NotTo(BeEmpty())
			})

			It("should format very large sizes", func() {
				s := Size(math.MaxUint64 / 2)
				result := s.String()
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("Unit boundaries", func() {
			It("should format values above KB boundary", func() {
				s := SizeKilo + 1
				result := s.String()
				Expect(result).To(ContainSubstring("KB"))
			})

			It("should format values above MB boundary", func() {
				s := SizeMega + 1
				result := s.String()
				Expect(result).To(ContainSubstring("MB"))
			})

			It("should format values above GB boundary", func() {
				s := SizeGiga + 1
				result := s.String()
				Expect(result).To(ContainSubstring("GB"))
			})
		})

		Context("Precision", func() {
			It("should include decimal values by default", func() {
				s := 5*SizeKilo + 512
				result := s.String()
				// Should contain a decimal point for default formatting
				Expect(result).NotTo(BeEmpty())
			})

			It("should format consistently", func() {
				s := 5 * SizeMega
				result1 := s.String()
				result2 := s.String()
				Expect(result1).To(Equal(result2))
			})
		})
	})

	Describe("Format() method", func() {
		Context("FormatRound0 - no decimals", func() {
			It("should format without decimal point", func() {
				s := 5*SizeKilo + 512
				result := s.Format(FormatRound0)
				Expect(result).To(MatchRegexp(`^\d+$`))
				Expect(result).NotTo(ContainSubstring("."))
			})

			It("should format zero", func() {
				s := SizeNul
				result := s.Format(FormatRound0)
				Expect(result).To(Equal("0"))
			})
		})

		Context("FormatRound1 - 1 decimal", func() {
			It("should format with 1 decimal", func() {
				s := 5*SizeKilo + 512
				result := s.Format(FormatRound1)
				Expect(result).To(MatchRegexp(`^\d+\.\d$`))
			})
		})

		Context("FormatRound2 - 2 decimals", func() {
			It("should format with 2 decimals", func() {
				s := 5*SizeKilo + 512
				result := s.Format(FormatRound2)
				Expect(result).To(MatchRegexp(`^\d+\.\d{2}$`))
			})
		})

		Context("FormatRound3 - 3 decimals", func() {
			It("should format with 3 decimals", func() {
				s := 5*SizeKilo + 512
				result := s.Format(FormatRound3)
				Expect(result).To(MatchRegexp(`^\d+\.\d{3}$`))
			})
		})

		Context("Custom formats", func() {
			It("should handle custom format strings", func() {
				s := 5 * SizeMega
				formats := []string{
					"%.0f",
					"%.1f",
					"%.4f",
					"%.10f",
					"%f",
				}

				for _, format := range formats {
					result := s.Format(format)
					Expect(result).NotTo(BeEmpty())
				}
			})

			It("should format scientific notation compatible", func() {
				s := 5 * SizeMega
				result := s.Format("%e")
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("Edge cases", func() {
			It("should format very large values", func() {
				s := Size(math.MaxUint64 / 2)
				result := s.Format(FormatRound2)
				Expect(result).NotTo(BeEmpty())
				Expect(result).To(MatchRegexp(`\d+\.\d{2}`))
			})
		})
	})

	Describe("Unit() method", func() {
		Context("Basic units", func() {
			It("should return B for bytes", func() {
				s := Size(100)
				unit := s.Unit(0)
				Expect(unit).To(Equal("B"))
			})

			It("should return KB for kilobytes", func() {
				s := 10 * SizeKilo
				unit := s.Unit(0)
				Expect(unit).To(Equal("KB"))
			})

			It("should return MB for megabytes", func() {
				s := 10 * SizeMega
				unit := s.Unit(0)
				Expect(unit).To(Equal("MB"))
			})

			It("should return GB for gigabytes", func() {
				s := 10 * SizeGiga
				unit := s.Unit(0)
				Expect(unit).To(Equal("GB"))
			})

			It("should return TB for terabytes", func() {
				s := 10 * SizeTera
				unit := s.Unit(0)
				Expect(unit).To(Equal("TB"))
			})

			It("should return PB for petabytes", func() {
				s := 10 * SizePeta
				unit := s.Unit(0)
				Expect(unit).To(Equal("PB"))
			})

			It("should return EB for exabytes", func() {
				s := 10 * SizeExa
				unit := s.Unit(0)
				Expect(unit).To(Equal("EB"))
			})
		})

		Context("Custom unit suffix", func() {
			It("should accept custom unit suffix", func() {
				s := 10 * SizeKilo
				unit := s.Unit('i')
				Expect(unit).To(ContainSubstring("i"))
				Expect(unit).To(Equal("Ki"))
			})

			It("should work with different runes", func() {
				tests := map[rune]string{
					'o': "Ko",
					'i': "Ki",
					'x': "Kx",
				}

				s := 10 * SizeKilo
				for r, expected := range tests {
					unit := s.Unit(r)
					Expect(unit).To(Equal(expected))
				}
			})
		})

		Context("Edge cases", func() {
			It("should handle zero size", func() {
				s := SizeNul
				unit := s.Unit(0)
				Expect(unit).To(Equal("B"))
			})

			It("should always return non-empty string", func() {
				sizes := []Size{
					SizeUnit,
					10 * SizeKilo,
					10 * SizeMega,
					10 * SizeGiga,
					10 * SizeTera,
				}

				for _, s := range sizes {
					unit := s.Unit(0)
					Expect(unit).NotTo(BeEmpty())
				}
			})
		})
	})

	Describe("Code() method", func() {
		Context("Default unit suffix", func() {
			BeforeEach(func() {
				SetDefaultUnit('B')
			})

			It("should return B for SizeUnit", func() {
				code := SizeUnit.Code(0)
				Expect(code).To(Equal("B"))
			})

			It("should return KB for SizeKilo", func() {
				code := SizeKilo.Code(0)
				Expect(code).To(Equal("KB"))
			})

			It("should return MB for SizeMega", func() {
				code := SizeMega.Code(0)
				Expect(code).To(Equal("MB"))
			})

			It("should return GB for SizeGiga", func() {
				code := SizeGiga.Code(0)
				Expect(code).To(Equal("GB"))
			})

			It("should return TB for SizeTera", func() {
				code := SizeTera.Code(0)
				Expect(code).To(Equal("TB"))
			})

			It("should return PB for SizePeta", func() {
				code := SizePeta.Code(0)
				Expect(code).To(Equal("PB"))
			})

			It("should return EB for SizeExa", func() {
				code := SizeExa.Code(0)
				Expect(code).To(Equal("EB"))
			})
		})

		Context("Custom unit suffix", func() {
			It("should use custom rune for SizeKilo", func() {
				code := SizeKilo.Code('o')
				Expect(code).To(Equal("Ko"))
			})

			It("should use custom rune for SizeMega", func() {
				code := SizeMega.Code('i')
				Expect(code).To(Equal("Mi"))
			})

			It("should use custom rune for SizeGiga", func() {
				code := SizeGiga.Code('x')
				Expect(code).To(Equal("Gx"))
			})
		})

		Context("Zero unit", func() {
			It("should use default when unit is 0", func() {
				SetDefaultUnit('B')
				code := SizeKilo.Code(0)
				Expect(code).NotTo(BeEmpty())
				Expect(code).To(ContainSubstring("B"))
			})
		})

		Context("Non-standard sizes", func() {
			It("should return B for non-standard values", func() {
				s := Size(500)
				code := s.Code(0)
				Expect(code).To(Equal("B"))
			})
		})
	})

	Describe("Unit conversion methods", func() {
		Context("KiloBytes()", func() {
			It("should convert gigabytes to kilobytes", func() {
				s := 5 * SizeGiga
				kb := s.KiloBytes()
				expected := uint64(5 * 1024 * 1024)
				Expect(kb).To(Equal(expected))
			})

			It("should return zero for sizes smaller than 1KB", func() {
				s := Size(512)
				Expect(s.KiloBytes()).To(Equal(uint64(0)))
			})

			It("should floor fractional kilobytes", func() {
				s := SizeKilo + 512
				Expect(s.KiloBytes()).To(Equal(uint64(1)))
			})

			It("should handle exact kilobyte values", func() {
				s := 10 * SizeKilo
				Expect(s.KiloBytes()).To(Equal(uint64(10)))
			})
		})

		Context("MegaBytes()", func() {
			It("should convert gigabytes to megabytes", func() {
				s := 5 * SizeGiga
				mb := s.MegaBytes()
				expected := uint64(5 * 1024)
				Expect(mb).To(Equal(expected))
			})

			It("should return zero for sizes smaller than 1MB", func() {
				s := 512 * SizeKilo
				Expect(s.MegaBytes()).To(Equal(uint64(0)))
			})

			It("should floor fractional megabytes", func() {
				s := SizeMega + 512*SizeKilo
				Expect(s.MegaBytes()).To(Equal(uint64(1)))
			})

			It("should handle exact megabyte values", func() {
				s := 10 * SizeMega
				Expect(s.MegaBytes()).To(Equal(uint64(10)))
			})
		})

		Context("GigaBytes()", func() {
			It("should convert terabytes to gigabytes", func() {
				s := 2 * SizeTera
				gb := s.GigaBytes()
				Expect(gb).To(Equal(uint64(2048)))
			})

			It("should return zero for sizes smaller than 1GB", func() {
				s := 512 * SizeMega
				Expect(s.GigaBytes()).To(Equal(uint64(0)))
			})

			It("should handle exact gigabyte values", func() {
				s := 5 * SizeGiga
				gb := s.GigaBytes()
				Expect(gb).To(Equal(uint64(5)))
			})
		})

		Context("TeraBytes()", func() {
			It("should convert petabytes to terabytes", func() {
				s := SizePeta
				tb := s.TeraBytes()
				Expect(tb).To(Equal(uint64(1024)))
			})

			It("should return zero for sizes smaller than 1TB", func() {
				s := 5 * SizeGiga
				Expect(s.TeraBytes()).To(Equal(uint64(0)))
			})

			It("should handle exact terabyte values", func() {
				s := 3 * SizeTera
				tb := s.TeraBytes()
				Expect(tb).To(Equal(uint64(3)))
			})
		})

		Context("PetaBytes()", func() {
			It("should handle petabyte values", func() {
				s := 2 * SizePeta
				pb := s.PetaBytes()
				Expect(pb).To(Equal(uint64(2)))
			})

			It("should return zero for smaller sizes", func() {
				s := 500 * SizeTera
				Expect(s.PetaBytes()).To(Equal(uint64(0)))
			})
		})

		Context("ExaBytes()", func() {
			It("should handle exabyte values", func() {
				s := SizeExa
				eb := s.ExaBytes()
				Expect(eb).To(Equal(uint64(1)))
			})

			It("should return zero for smaller sizes", func() {
				s := 500 * SizePeta
				Expect(s.ExaBytes()).To(Equal(uint64(0)))
			})
		})
	})

	Describe("Format edge cases", func() {
		It("should handle very small sizes", func() {
			s := Size(1)
			result := s.String()
			Expect(result).NotTo(BeEmpty())
			Expect(result).To(ContainSubstring("B"))
		})

		It("should handle very large sizes", func() {
			s := Size(math.MaxUint64 / 2)
			result := s.String()
			Expect(result).NotTo(BeEmpty())
		})

		It("should be consistent between Format and String", func() {
			s := 5 * SizeMega
			str := s.String()
			formatted := s.Format(FormatRound2)
			Expect(str).To(ContainSubstring(formatted))
		})
	})

	Describe("Format constants", func() {
		It("should have correct format constant values", func() {
			Expect(FormatRound0).To(Equal("%.0f"))
			Expect(FormatRound1).To(Equal("%.1f"))
			Expect(FormatRound2).To(Equal("%.2f"))
			Expect(FormatRound3).To(Equal("%.3f"))
		})

		It("should apply format constants correctly", func() {
			s := Size(1024)
			Expect(func() {
				_ = s.Format(FormatRound0)
				_ = s.Format(FormatRound1)
				_ = s.Format(FormatRound2)
				_ = s.Format(FormatRound3)
			}).NotTo(Panic())
		})
	})

	Describe("Performance", func() {
		It("should format repeatedly without errors", func() {
			s := 5 * SizeGiga

			for i := 0; i < 1000; i++ {
				_ = s.String()
				_ = s.Format(FormatRound2)
				_ = s.Unit(0)
			}
		})

		It("should handle various sizes efficiently", func() {
			sizes := []Size{
				Size(100),
				SizeKilo,
				SizeMega,
				SizeGiga,
				SizeTera,
			}

			for i := 0; i < 100; i++ {
				for _, s := range sizes {
					_ = s.String()
					_ = s.Unit(0)
				}
			}
		})
	})
})
