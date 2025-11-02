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

var _ = Describe("Bytes Type", func() {
	Describe("String() method", func() {
		It("should return numeric string representation", func() {
			tests := map[Bytes]string{
				Bytes(0):          "0",
				Bytes(1):          "1",
				Bytes(1024):       "1024",
				Bytes(67890):      "67890",
				Bytes(1234567890): "1234567890",
			}

			for b, expected := range tests {
				Expect(b.String()).To(Equal(expected))
			}
		})
	})

	Describe("AsNumber() method", func() {
		It("should convert to Number type", func() {
			tests := map[Bytes]Number{
				Bytes(0):       Number(0),
				Bytes(2048):    Number(2048),
				Bytes(1048576): Number(1048576),
			}

			for b, expected := range tests {
				Expect(b.AsNumber()).To(Equal(expected))
			}
		})

		It("should maintain value integrity", func() {
			b := Bytes(987654321)
			n := b.AsNumber()
			Expect(uint64(n)).To(Equal(uint64(b)))
		})
	})

	Describe("AsUint64() method", func() {
		It("should return uint64 value", func() {
			tests := map[Bytes]uint64{
				Bytes(0):          0,
				Bytes(1024):       1024,
				Bytes(1234567890): 1234567890,
			}

			for b, expected := range tests {
				Expect(b.AsUint64()).To(Equal(expected))
			}
		})
	})

	Describe("AsFloat64() method", func() {
		It("should return float64 value", func() {
			tests := map[Bytes]float64{
				Bytes(0):    0.0,
				Bytes(1024): 1024.0,
				Bytes(5000): 5000.0,
			}

			for b, expected := range tests {
				Expect(b.AsFloat64()).To(Equal(expected))
			}
		})
	})

	Describe("FormatUnitInt() method", func() {
		Context("with values below kilobyte", func() {
			It("should format without unit suffix", func() {
				tests := []Bytes{
					Bytes(0),
					Bytes(1),
					Bytes(10),
					Bytes(100),
					Bytes(1023),
				}

				for _, b := range tests {
					result := b.FormatUnitInt()
					Expect(result).NotTo(BeEmpty())
					Expect(result).NotTo(ContainSubstring("KB"))
					Expect(result).NotTo(ContainSubstring("MB"))
				}
			})
		})

		Context("with kilobyte values (2^10)", func() {
			It("should format with KB suffix", func() {
				tests := map[Bytes]string{
					Bytes(1024):    "   1 KB", // 1 KB
					Bytes(5120):    "   5 KB", // 5 KB
					Bytes(10240):   "  10 KB", // 10 KB
					Bytes(102400):  " 100 KB", // 100 KB
					Bytes(1048575): "1024 KB", // ~1024 KB (just before 1 MB)
				}

				for b, expected := range tests {
					result := b.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with megabyte values (2^20)", func() {
			It("should format with MB suffix", func() {
				tests := map[Bytes]string{
					Bytes(1048576):   "   1 MB", // 1 MB
					Bytes(5242880):   "   5 MB", // 5 MB
					Bytes(10485760):  "  10 MB", // 10 MB
					Bytes(104857600): " 100 MB", // 100 MB
				}

				for b, expected := range tests {
					result := b.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with gigabyte values (2^30)", func() {
			It("should format with GB suffix", func() {
				tests := map[Bytes]string{
					Bytes(1073741824):   "   1 GB", // 1 GB
					Bytes(5368709120):   "   5 GB", // 5 GB
					Bytes(10737418240):  "  10 GB", // 10 GB
					Bytes(107374182400): " 100 GB", // 100 GB
				}

				for b, expected := range tests {
					result := b.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with terabyte values (2^40)", func() {
			It("should format with TB suffix", func() {
				tests := map[Bytes]string{
					Bytes(1099511627776):   "   1 TB", // 1 TB
					Bytes(5497558138880):   "   5 TB", // 5 TB
					Bytes(10995116277760):  "  10 TB", // 10 TB
					Bytes(109951162777600): " 100 TB", // 100 TB
				}

				for b, expected := range tests {
					result := b.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with petabyte values (2^50)", func() {
			It("should format with PB suffix", func() {
				tests := map[Bytes]string{
					Bytes(1125899906842624):  "   1 PB", // 1 PB
					Bytes(5629499534213120):  "   5 PB", // 5 PB
					Bytes(11258999068426240): "  10 PB", // 10 PB
				}

				for b, expected := range tests {
					result := b.FormatUnitInt()
					Expect(result).To(Equal(expected))
				}
			})
		})

		Context("with exabyte values (2^60)", func() {
			It("should format with EB suffix", func() {
				b := Bytes(1152921504606846976) // 1 EB
				result := b.FormatUnitInt()
				Expect(result).To(ContainSubstring("EB"))
				Expect(result).To(MatchRegexp(`\s+\d+\s+EB`))
			})
		})

		Context("padding behavior", func() {
			It("should pad numbers to 4 characters", func() {
				tests := []Bytes{
					Bytes(5),
					Bytes(50),
					Bytes(500),
					Bytes(5000),
				}

				for _, b := range tests {
					result := b.FormatUnitInt()
					Expect(len(result)).To(BeNumerically(">=", 4))
				}
			})
		})
	})

	Describe("FormatUnitFloat() method", func() {
		Context("with zero precision", func() {
			It("should delegate to FormatUnitInt", func() {
				tests := []Bytes{
					Bytes(5120),
					Bytes(5242880),
					Bytes(5368709120),
				}

				for _, b := range tests {
					result1 := b.FormatUnitFloat(0)
					result2 := b.FormatUnitInt()
					Expect(result1).To(Equal(result2))
				}
			})
		})

		Context("with precision 2", func() {
			It("should format with 2 decimal places", func() {
				b := Bytes(1536) // 1.5 KB
				result := b.FormatUnitFloat(2)
				Expect(result).To(ContainSubstring("KB"))
				Expect(result).To(MatchRegexp(`1\.50`))
			})

			It("should format various byte values correctly", func() {
				tests := []struct {
					bytes   Bytes
					pattern string
				}{
					{Bytes(2560), "2.50 KB"},    // 2.5 KB
					{Bytes(3670016), "3.50 MB"}, // 3.5 MB
				}

				for _, test := range tests {
					result := test.bytes.FormatUnitFloat(2)
					Expect(result).To(ContainSubstring(test.pattern))
				}
			})
		})

		Context("with various precisions", func() {
			It("should respect precision parameter", func() {
				b := Bytes(1234567)

				result1 := b.FormatUnitFloat(1)
				Expect(result1).To(MatchRegexp(`\d+\.\d{1}`))

				result2 := b.FormatUnitFloat(2)
				Expect(result2).To(MatchRegexp(`\d+\.\d{2}`))

				result3 := b.FormatUnitFloat(3)
				Expect(result3).To(MatchRegexp(`\d+\.\d{3}`))
			})
		})
	})

	Describe("Byte unit calculations", func() {
		It("should use binary (base-2) calculations", func() {
			// Verify that 1 KB = 1024 bytes (not 1000)
			b := Bytes(1024)
			result := b.FormatUnitInt()
			Expect(result).To(ContainSubstring("1 KB"))

			// Verify that 1 MB = 1024 KB = 1048576 bytes
			b2 := Bytes(1048576)
			result2 := b2.FormatUnitInt()
			Expect(result2).To(ContainSubstring("1 MB"))

			// Verify that 1 GB = 1024 MB = 1073741824 bytes
			b3 := Bytes(1073741824)
			result3 := b3.FormatUnitInt()
			Expect(result3).To(ContainSubstring("1 GB"))
		})
	})

	Describe("Conversion roundtrip", func() {
		It("should maintain value through conversions", func() {
			original := Bytes(67890)

			// Bytes -> Uint64 -> Bytes
			uint64Val := original.AsUint64()
			Expect(Bytes(uint64Val)).To(Equal(original))

			// Bytes -> Float64 -> Bytes (for integers)
			float64Val := original.AsFloat64()
			Expect(Bytes(float64Val)).To(Equal(original))

			// Bytes -> Number -> Bytes
			numberVal := original.AsNumber()
			Expect(numberVal.AsBytes()).To(Equal(original))
		})
	})

	Describe("Edge cases", func() {
		Context("with zero value", func() {
			It("should handle zero correctly", func() {
				b := Bytes(0)
				Expect(b.String()).To(Equal("0"))
				Expect(b.AsUint64()).To(Equal(uint64(0)))
				Expect(b.AsFloat64()).To(Equal(float64(0)))
				Expect(b.FormatUnitInt()).NotTo(BeEmpty())
				Expect(b.FormatUnitFloat(2)).NotTo(BeEmpty())
			})
		})

		Context("with maximum uint64 value", func() {
			It("should handle MaxUint64", func() {
				b := Bytes(18446744073709551615) // MaxUint64
				Expect(b.String()).To(Equal("18446744073709551615"))
				Expect(b.AsUint64()).To(Equal(uint64(18446744073709551615)))
				Expect(b.FormatUnitInt()).NotTo(BeEmpty())
			})
		})

		Context("with boundary values", func() {
			It("should handle power-of-2 boundaries correctly", func() {
				tests := []struct {
					value   Bytes
					minUnit string
					notUnit string
				}{
					{Bytes(1023), "", "KB"},
					{Bytes(1024), "KB", ""},
					{Bytes(1048575), "KB", "MB"},
					{Bytes(1048576), "MB", "KB"},
					{Bytes(1073741823), "MB", "GB"},
					{Bytes(1073741824), "GB", "MB"},
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

		Context("with fractional kilobytes", func() {
			It("should format fractional values correctly", func() {
				// Test values that don't divide evenly
				tests := []struct {
					bytes   Bytes
					minUnit string
				}{
					{Bytes(1536), "KB"}, // 1.5 KB
					{Bytes(2048), "KB"}, // 2.0 KB
					{Bytes(2560), "KB"}, // 2.5 KB
				}

				for _, test := range tests {
					result := test.bytes.FormatUnitFloat(2)
					Expect(result).To(ContainSubstring(test.minUnit))
				}
			})
		})
	})

	Describe("Performance", func() {
		It("should handle repeated calls efficiently", func() {
			b := Bytes(67890)

			Expect(func() {
				for i := 0; i < 10000; i++ {
					_ = b.String()
					_ = b.AsNumber()
					_ = b.AsUint64()
					_ = b.AsFloat64()
					_ = b.FormatUnitInt()
					_ = b.FormatUnitFloat(2)
				}
			}).NotTo(Panic())
		})
	})

	Describe("Comparison with Number type", func() {
		It("should handle same numeric values differently", func() {
			// Same numeric value but different formatting
			n := Number(5000)
			b := Bytes(5000)

			// Number uses decimal (K = 10^3), so 5000 = 5K
			numResult := n.FormatUnitInt()
			Expect(numResult).To(ContainSubstring("K"))
			Expect(numResult).NotTo(ContainSubstring("KB"))

			// Bytes uses binary (KB = 2^10), so 5000 bytes ≈ 4.88 KB
			bytesResult := b.FormatUnitInt()
			Expect(bytesResult).To(ContainSubstring("KB"))
		})

		It("should convert between Number and Bytes correctly", func() {
			n := Number(5000)
			b := n.AsBytes()
			Expect(b).To(Equal(Bytes(5000)))

			b2 := Bytes(5000)
			n2 := b2.AsNumber()
			Expect(n2).To(Equal(Number(5000)))
		})
	})
})
