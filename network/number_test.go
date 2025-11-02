/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package network_test

import (
	. "github.com/nabbar/golib/network"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Number Type", func() {
	Describe("String() method", func() {
		It("should return numeric string representation", func() {
			tests := map[Number]string{
				Number(0):                    "0",
				Number(1):                    "1",
				Number(12345):                "12345",
				Number(9876543210):           "9876543210",
				Number(18446744073709551615): "18446744073709551615", // MaxUint64
			}

			for num, expected := range tests {
				Expect(num.String()).To(Equal(expected))
			}
		})
	})

	Describe("AsBytes() method", func() {
		It("should convert to Bytes type", func() {
			tests := map[Number]Bytes{
				Number(0):       Bytes(0),
				Number(1024):    Bytes(1024),
				Number(1048576): Bytes(1048576),
			}

			for num, expected := range tests {
				Expect(num.AsBytes()).To(Equal(expected))
			}
		})

		It("should maintain value integrity", func() {
			n := Number(123456789)
			b := n.AsBytes()
			Expect(uint64(b)).To(Equal(uint64(n)))
		})
	})

	Describe("AsUint64() method", func() {
		It("should return uint64 value", func() {
			tests := map[Number]uint64{
				Number(0):          0,
				Number(1000):       1000,
				Number(9876543210): 9876543210,
			}

			for num, expected := range tests {
				Expect(num.AsUint64()).To(Equal(expected))
			}
		})
	})

	Describe("AsFloat64() method", func() {
		It("should return float64 value", func() {
			tests := map[Number]float64{
				Number(0):    0.0,
				Number(1000): 1000.0,
				Number(5000): 5000.0,
			}

			for num, expected := range tests {
				Expect(num.AsFloat64()).To(Equal(expected))
			}
		})
	})

	Describe("FormatUnitInt() method", func() {
		Context("with values below kilo", func() {
			It("should format without unit suffix", func() {
				tests := []Number{
					Number(0),
					Number(1),
					Number(10),
					Number(100),
					Number(999),
				}

				for _, num := range tests {
					result := num.FormatUnitInt()
					Expect(result).NotTo(BeEmpty())
					Expect(result).NotTo(ContainSubstring("K"))
					Expect(result).NotTo(ContainSubstring("M"))
				}
			})
		})

		Context("with kilo values (10^3)", func() {
			It("should format with K suffix", func() {
				tests := map[Number]string{
					Number(1000):   "   1 K",
					Number(5000):   "   5 K",
					Number(15000):  "  15 K",
					Number(999000): " 999 K",
				}

				for num, expected := range tests {
					result := num.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with mega values (10^6)", func() {
			It("should format with M suffix", func() {
				tests := map[Number]string{
					Number(1000000):   "   1 M",
					Number(5000000):   "   5 M",
					Number(15000000):  "  15 M",
					Number(999000000): " 999 M",
				}

				for num, expected := range tests {
					result := num.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with giga values (10^9)", func() {
			It("should format with G suffix", func() {
				tests := map[Number]string{
					Number(1000000000):   "   1 G",
					Number(5000000000):   "   5 G",
					Number(15000000000):  "  15 G",
					Number(999000000000): " 999 G",
				}

				for num, expected := range tests {
					result := num.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with tera values (10^12)", func() {
			It("should format with T suffix", func() {
				tests := map[Number]string{
					Number(1000000000000):   "   1 T",
					Number(5000000000000):   "   5 T",
					Number(15000000000000):  "  15 T",
					Number(999000000000000): " 999 T",
				}

				for num, expected := range tests {
					result := num.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with peta values (10^15)", func() {
			It("should format with P suffix", func() {
				tests := map[Number]string{
					Number(1000000000000000):   "   1 P",
					Number(5000000000000000):   "   5 P",
					Number(15000000000000000):  "  15 P",
					Number(999000000000000000): " 999 P",
				}

				for num, expected := range tests {
					result := num.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with exa values (10^18)", func() {
			It("should format with E suffix", func() {
				n := Number(1000000000000000000) // 1 exabyte
				result := n.FormatUnitInt()
				Expect(result).To(ContainSubstring("E"))
				Expect(result).To(MatchRegexp(`\s+\d+\s+E`))
			})
		})

		Context("padding behavior", func() {
			It("should pad numbers to 4 characters", func() {
				tests := []Number{
					Number(5),    // "   5  "
					Number(50),   // "  50  "
					Number(500),  // " 500  "
					Number(5000), // "   5 K"
				}

				for _, num := range tests {
					result := num.FormatUnitInt()
					// Check minimum length (4 digits + space + unit)
					Expect(len(result)).To(BeNumerically(">=", 4))
				}
			})

			It("should handle 4-digit numbers without extra padding", func() {
				n := Number(9999)
				result := n.FormatUnitInt()
				Expect(result).To(Equal("  10 K")) // Rounds to 10K
			})
		})
	})

	Describe("FormatUnitFloat() method", func() {
		Context("with zero precision", func() {
			It("should delegate to FormatUnitInt", func() {
				tests := []Number{
					Number(5000),
					Number(5000000),
					Number(5000000000),
				}

				for _, num := range tests {
					result1 := num.FormatUnitFloat(0)
					result2 := num.FormatUnitInt()
					Expect(result1).To(Equal(result2))
				}
			})
		})

		Context("with precision 2", func() {
			It("should format with 2 decimal places", func() {
				n := Number(1500) // 1.5 K
				result := n.FormatUnitFloat(2)
				Expect(result).To(ContainSubstring("K"))
				Expect(result).To(MatchRegexp(`1\.50`))
			})

			It("should format various values correctly", func() {
				tests := map[Number]string{
					Number(1234):    "1.23 K",
					Number(5678000): "5.68 M",
				}

				for num, expectedPattern := range tests {
					result := num.FormatUnitFloat(2)
					Expect(result).To(ContainSubstring(expectedPattern))
				}
			})
		})

		Context("with various precisions", func() {
			It("should respect precision parameter", func() {
				n := Number(1234567)

				result1 := n.FormatUnitFloat(1)
				Expect(result1).To(MatchRegexp(`\d+\.\d{1}`))

				result2 := n.FormatUnitFloat(2)
				Expect(result2).To(MatchRegexp(`\d+\.\d{2}`))

				result3 := n.FormatUnitFloat(3)
				Expect(result3).To(MatchRegexp(`\d+\.\d{3}`))
			})
		})

		Context("padding behavior", func() {
			It("should pad small numbers", func() {
				n := Number(1500)
				result := n.FormatUnitFloat(2)
				// Should have padding for single digit
				Expect(result).To(MatchRegexp(`\s+1\.50`))
			})
		})
	})

	Describe("Conversion roundtrip", func() {
		It("should maintain value through conversions", func() {
			original := Number(12345)

			// Number -> Uint64 -> Number
			uint64Val := original.AsUint64()
			Expect(Number(uint64Val)).To(Equal(original))

			// Number -> Float64 -> Number (for integers)
			float64Val := original.AsFloat64()
			Expect(Number(float64Val)).To(Equal(original))

			// Number -> Bytes -> Number
			bytesVal := original.AsBytes()
			Expect(bytesVal.AsNumber()).To(Equal(original))
		})
	})

	Describe("Edge cases", func() {
		Context("with zero value", func() {
			It("should handle zero correctly", func() {
				n := Number(0)
				Expect(n.String()).To(Equal("0"))
				Expect(n.AsUint64()).To(Equal(uint64(0)))
				Expect(n.AsFloat64()).To(Equal(float64(0)))
				Expect(n.FormatUnitInt()).NotTo(BeEmpty())
				Expect(n.FormatUnitFloat(2)).NotTo(BeEmpty())
			})
		})

		Context("with maximum uint64 value", func() {
			It("should handle MaxUint64", func() {
				n := Number(18446744073709551615) // MaxUint64
				Expect(n.String()).To(Equal("18446744073709551615"))
				Expect(n.AsUint64()).To(Equal(uint64(18446744073709551615)))
				Expect(n.FormatUnitInt()).NotTo(BeEmpty())
			})
		})

		Context("with boundary values", func() {
			It("should handle power boundaries correctly", func() {
				tests := []struct {
					value   Number
					minUnit string
					notUnit string
				}{
					{Number(999), "", "K"},
					{Number(1000), "K", ""},
					{Number(999999), "K", "M"},
					{Number(1000000), "M", "K"},
					{Number(999999999), "M", "G"},
					{Number(1000000000), "G", "M"},
				}

				for _, test := range tests {
					result := test.value.FormatUnitInt()
					if test.minUnit != "" {
						Expect(result).To(ContainSubstring(test.minUnit))
					}
					if test.notUnit != "" && test.minUnit != test.notUnit {
						Expect(result).NotTo(ContainSubstring(test.notUnit))
					}
				}
			})
		})
	})

	Describe("Performance", func() {
		It("should handle repeated calls efficiently", func() {
			n := Number(12345)

			Expect(func() {
				for i := 0; i < 10000; i++ {
					_ = n.String()
					_ = n.AsBytes()
					_ = n.AsUint64()
					_ = n.AsFloat64()
					_ = n.FormatUnitInt()
					_ = n.FormatUnitFloat(2)
				}
			}).NotTo(Panic())
		})
	})
})
