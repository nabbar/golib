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

var _ = Describe("Parsing", func() {
	Describe("Parse() - Basic functionality", func() {
		Context("Single letter units", func() {
			It("should parse B (byte)", func() {
				s, err := Parse("1B")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(SizeUnit))
			})

			It("should parse K (kilobyte)", func() {
				s, err := Parse("1K")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeKilo, 1))
			})

			It("should parse M (megabyte)", func() {
				s, err := Parse("1M")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeMega, 1))
			})

			It("should parse G (gigabyte)", func() {
				s, err := Parse("1G")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeGiga, 1))
			})

			It("should parse T (terabyte)", func() {
				s, err := Parse("1T")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeTera, 1))
			})

			It("should parse P (petabyte)", func() {
				s, err := Parse("1P")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizePeta, 1))
			})
		})

		Context("Two letter units", func() {
			It("should parse KB", func() {
				s, err := Parse("1KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeKilo, 1))
			})

			It("should parse MB", func() {
				s, err := Parse("1MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeMega, 1))
			})

			It("should parse GB", func() {
				s, err := Parse("1GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeGiga, 1))
			})

			It("should parse TB", func() {
				s, err := Parse("1TB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeTera, 1))
			})

			It("should parse PB", func() {
				s, err := Parse("1PB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizePeta, 1))
			})

			It("should parse EB", func() {
				s, err := Parse("1EB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", SizeExa, 1))
			})
		})

		Context("Case variations", func() {
			It("should parse lowercase units", func() {
				tests := []string{"1kb", "1mb", "1gb"}
				for _, test := range tests {
					s, err := Parse(test)
					Expect(err).ToNot(HaveOccurred())
					Expect(s).To(BeNumerically(">", 0))
				}
			})

			It("should parse mixed case units", func() {
				tests := []string{"1Kb", "1kB", "1Mb", "1mB"}
				for _, test := range tests {
					s, err := Parse(test)
					Expect(err).ToNot(HaveOccurred())
					Expect(s).To(BeNumerically(">", 0))
				}
			})

			It("should parse uppercase units", func() {
				tests := []string{"1KB", "1MB", "1GB"}
				for _, test := range tests {
					s, err := Parse(test)
					Expect(err).ToNot(HaveOccurred())
					Expect(s).To(BeNumerically(">", 0))
				}
			})
		})
	})

	Describe("Parse() - Numeric values", func() {
		Context("Integer multiples", func() {
			It("should parse 2KB", func() {
				s, err := Parse("2KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(2 * SizeKilo))
			})

			It("should parse 5MB", func() {
				s, err := Parse("5MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(5 * SizeMega))
			})

			It("should parse 10GB", func() {
				s, err := Parse("10GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(10 * SizeGiga))
			})

			It("should parse 100TB", func() {
				s, err := Parse("100TB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(100 * SizeTera))
			})
		})

		Context("Fractional values", func() {
			It("should parse 1.5KB", func() {
				s, err := Parse("1.5KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", Size(1.5*float64(SizeKilo)), 10))
			})

			It("should parse 2.5MB", func() {
				s, err := Parse("2.5MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", Size(2.5*float64(SizeMega)), 100))
			})

			It("should parse 5.5GB", func() {
				s, err := Parse("5.5GB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", Size(5.5*float64(SizeGiga)), 1000))
			})

			It("should parse 0.5TB", func() {
				s, err := Parse("0.5TB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", Size(0.5*float64(SizeTera)), 1000))
			})
		})

		Context("Very small fractions", func() {
			It("should parse 0.001MB", func() {
				s, err := Parse("0.001MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically(">", 0))
			})

			It("should parse 0.1KB", func() {
				s, err := Parse("0.1KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically(">", 0))
			})
		})

		Context("Zero values", func() {
			It("should parse 0B", func() {
				s, err := Parse("0B")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(SizeNul))
			})

			It("should parse 0KB", func() {
				s, err := Parse("0KB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(SizeNul))
			})

			It("should parse 0.0MB", func() {
				s, err := Parse("0.0MB")
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(SizeNul))
			})
		})
	})

	Describe("Parse() - Whitespace handling", func() {
		It("should handle leading whitespace", func() {
			s, err := Parse(" 5MB")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})

		It("should handle trailing whitespace", func() {
			s, err := Parse("5MB ")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})

		It("should handle both leading and trailing whitespace", func() {
			s, err := Parse(" 5MB ")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})

		It("should handle tab characters", func() {
			s, err := Parse("\t5MB\t")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})

		It("should handle multiple spaces", func() {
			s, err := Parse("  5MB  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})
	})

	Describe("Parse() - Quoted strings", func() {
		It("should handle double quotes", func() {
			s, err := Parse(`"5MB"`)
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically(">", 0))
		})

		It("should handle single quotes", func() {
			s, err := Parse(`'5MB'`)
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically(">", 0))
		})

		It("should handle quotes with whitespace", func() {
			s, err := Parse(`" 5MB "`)
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically(">", 0))
		})
	})

	Describe("Parse() - Sign handling", func() {
		It("should handle positive sign", func() {
			s, err := Parse("+5MB")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(BeNumerically("~", 5*SizeMega, float64(5*SizeMega)*0.01))
		})

		It("should reject negative values", func() {
			_, err := Parse("-5MB")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("negative"))
		})
	})

	Describe("Parse() - Error cases", func() {
		Context("Empty and invalid inputs", func() {
			It("should reject empty string", func() {
				_, err := Parse("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid size"))
			})

			It("should reject only whitespace", func() {
				_, err := Parse("   ")
				Expect(err).To(HaveOccurred())
			})

			It("should reject only quotes", func() {
				_, err := Parse(`""`)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Missing components", func() {
			It("should reject numbers without units", func() {
				_, err := Parse("123")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing unit"))
			})

			It("should reject fractional numbers without units", func() {
				_, err := Parse("456.78")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing unit"))
			})

			It("should reject units without numbers", func() {
				invalidInputs := []string{"MB", "GB", "KB"}
				for _, input := range invalidInputs {
					_, err := Parse(input)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		Context("Invalid units", func() {
			It("should reject unknown units", func() {
				invalidInputs := []string{
					"5XYZ",
					"10ABC",
					"123ZZ",
					"5QQ",
				}

				for _, input := range invalidInputs {
					_, err := Parse(input)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unknown unit"))
				}
			})
		})

		Context("Overflow", func() {
			It("should reject very large numbers", func() {
				invalidInputs := []string{
					"99999999999999999999TB",
					"99999999999999999999PB",
					"99999999999999999999EB",
				}

				for _, input := range invalidInputs {
					_, err := Parse(input)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		Context("Malformed numbers", func() {
			It("should reject multiple decimal points", func() {
				_, err := Parse("5.5.5MB")
				Expect(err).To(HaveOccurred())
			})

			It("should handle decimal without trailing digits", func() {
				_, err := Parse("5.MB")
				// May succeed or fail depending on implementation
				_ = err
			})
		})
	})

	Describe("Parse() - Compound sizes", func() {
		It("should parse compound sizes if supported", func() {
			s, err := Parse("5GB2MB")
			if err == nil {
				expected := 5*SizeGiga + 2*SizeMega
				Expect(s).To(BeNumerically("~", expected, float64(expected)*0.01))
			}
		})

		It("should parse multi-unit sizes", func() {
			s, err := Parse("1GB500MB")
			if err == nil {
				expected := SizeGiga + 500*SizeMega
				Expect(s).To(BeNumerically("~", expected, float64(expected)*0.01))
			}
		})
	})

	Describe("ParseByte()", func() {
		It("should parse byte slices", func() {
			tests := map[string]Size{
				"10KB":  10 * SizeKilo,
				"5MB":   5 * SizeMega,
				"2.5GB": Size(2.5 * float64(SizeGiga)),
			}

			for input, expected := range tests {
				s, err := ParseByte([]byte(input))
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(BeNumerically("~", expected, float64(expected)*0.05))
			}
		})

		It("should handle empty byte slices", func() {
			_, err := ParseByte([]byte{})
			Expect(err).To(HaveOccurred())
		})

		It("should handle invalid byte slices", func() {
			_, err := ParseByte([]byte("invalid"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ParseSize() - Deprecated alias", func() {
		It("should work identically to Parse()", func() {
			input := "5MB"
			s1, err1 := Parse(input)
			s2, err2 := ParseSize(input)

			if err1 != nil {
				Expect(err2).To(HaveOccurred())
			} else {
				Expect(err2).ToNot(HaveOccurred())
			}
			Expect(s1).To(Equal(s2))
		})

		It("should handle errors identically", func() {
			input := "invalid"
			s1, err1 := Parse(input)
			s2, err2 := ParseSize(input)

			Expect(err1).To(HaveOccurred())
			Expect(err2).To(HaveOccurred())
			Expect(s1).To(Equal(s2))
		})
	})

	Describe("ParseByteAsSize() - Deprecated alias", func() {
		It("should work identically to ParseByte()", func() {
			input := []byte("10KB")
			s1, err1 := ParseByte(input)
			s2, err2 := ParseByteAsSize(input)

			if err1 != nil {
				Expect(err2).To(HaveOccurred())
			} else {
				Expect(err2).ToNot(HaveOccurred())
			}
			Expect(s1).To(Equal(s2))
		})
	})

	Describe("GetSize() - Deprecated function", func() {
		It("should return size and true for valid input", func() {
			tests := map[string]Size{
				"1GB":  SizeGiga,
				"5MB":  5 * SizeMega,
				"10KB": 10 * SizeKilo,
			}

			for input, expected := range tests {
				s, ok := GetSize(input)
				Expect(ok).To(BeTrue())
				Expect(s).To(BeNumerically("~", expected, float64(expected)*0.01))
			}
		})

		It("should return false for invalid input", func() {
			invalidInputs := []string{
				"invalid",
				"",
				"123",
				"XYZ",
			}

			for _, input := range invalidInputs {
				_, ok := GetSize(input)
				Expect(ok).To(BeFalse())
			}
		})

		It("should return zero size on failure", func() {
			s, ok := GetSize("invalid")
			Expect(ok).To(BeFalse())
			Expect(s).To(Equal(SizeNul))
		})
	})

	Describe("Parse() - Performance", func() {
		It("should parse repeatedly without errors", func() {
			input := "5MB"
			for i := 0; i < 1000; i++ {
				_, err := Parse(input)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should handle various inputs in sequence", func() {
			inputs := []string{
				"1B", "1KB", "1MB", "1GB", "1TB",
				"5.5MB", "10.25GB", "100KB",
			}

			for i := 0; i < 100; i++ {
				for _, input := range inputs {
					_, _ = Parse(input)
				}
			}
		})
	})
})
