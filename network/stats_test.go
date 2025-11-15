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

var _ = Describe("Stats Type", func() {
	Describe("String() method", func() {
		It("should return correct string for each stat type", func() {
			tests := map[Stats]string{
				StatBytes:   "Traffic",
				StatPackets: "Packets",
				StatFifo:    "Fifo",
				StatDrop:    "Drop",
				StatErr:     "Error",
			}

			for stat, expected := range tests {
				Expect(stat.String()).To(Equal(expected))
			}
		})

		It("should return empty string for invalid stat", func() {
			invalidStat := Stats(99)
			Expect(invalidStat.String()).To(Equal(""))
		})

		It("should return empty string for zero stat", func() {
			zeroStat := Stats(0)
			Expect(zeroStat.String()).To(Equal(""))
		})
	})

	Describe("FormatUnitInt() method", func() {
		Context("with StatBytes", func() {
			It("should format as bytes (binary units)", func() {
				n := Number(5120) // 5 KB
				result := StatBytes.FormatUnitInt(n)
				Expect(result).To(ContainSubstring("KB"))
			})

			It("should use byte formatting for traffic", func() {
				tests := map[Number]string{
					Number(1024):       "KB",
					Number(1048576):    "MB",
					Number(1073741824): "GB",
				}

				for num, expectedUnit := range tests {
					result := StatBytes.FormatUnitInt(num)
					Expect(result).To(ContainSubstring(expectedUnit))
				}
			})
		})

		Context("with StatPackets", func() {
			It("should format as decimal numbers", func() {
				n := Number(5000) // 5 K
				result := StatPackets.FormatUnitInt(n)
				Expect(result).To(ContainSubstring("K"))
				Expect(result).NotTo(ContainSubstring("KB"))
			})
		})

		Context("with StatFifo", func() {
			It("should format as decimal numbers", func() {
				n := Number(1000000) // 1 M
				result := StatFifo.FormatUnitInt(n)
				Expect(result).To(ContainSubstring("M"))
				Expect(result).NotTo(ContainSubstring("MB"))
			})
		})

		Context("with StatDrop", func() {
			It("should format as decimal numbers", func() {
				n := Number(1000) // 1 K
				result := StatDrop.FormatUnitInt(n)
				Expect(result).To(ContainSubstring("K"))
			})
		})

		Context("with StatErr", func() {
			It("should format as decimal numbers", func() {
				n := Number(500)
				result := StatErr.FormatUnitInt(n)
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("with invalid stat type", func() {
			It("should return empty string", func() {
				invalidStat := Stats(99)
				n := Number(1000)
				result := invalidStat.FormatUnitInt(n)
				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("FormatUnitFloat() method", func() {
		Context("with StatBytes", func() {
			It("should format bytes with precision", func() {
				n := Number(1536) // 1.5 KB
				result := StatBytes.FormatUnitFloat(n, 2)
				Expect(result).To(ContainSubstring("KB"))
				Expect(result).To(MatchRegexp(`\d+\.\d{2}`))
			})
		})

		Context("with StatPackets", func() {
			It("should format with decimal precision", func() {
				n := Number(1500) // 1.5 K
				result := StatPackets.FormatUnitFloat(n, 2)
				Expect(result).To(ContainSubstring("K"))
				Expect(result).To(MatchRegexp(`\d+\.\d{2}`))
			})
		})

		Context("with various precisions", func() {
			It("should respect precision parameter", func() {
				n := Number(1234567)

				result1 := StatPackets.FormatUnitFloat(n, 1)
				Expect(result1).To(MatchRegexp(`\d+\.\d{1}`))

				result2 := StatPackets.FormatUnitFloat(n, 2)
				Expect(result2).To(MatchRegexp(`\d+\.\d{2}`))

				result3 := StatPackets.FormatUnitFloat(n, 3)
				Expect(result3).To(MatchRegexp(`\d+\.\d{3}`))
			})
		})

		Context("with invalid stat type", func() {
			It("should return empty string", func() {
				invalidStat := Stats(99)
				n := Number(1000)
				result := invalidStat.FormatUnitFloat(n, 2)
				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("FormatUnit() method", func() {
		Context("with StatBytes", func() {
			It("should use float formatting with default precision", func() {
				n := Number(1536) // 1.5 KB
				result := StatBytes.FormatUnit(n)
				Expect(result).To(ContainSubstring("KB"))
				// Should have decimal places (default precision = 2)
				Expect(result).To(MatchRegexp(`\d+\.\d{2}`))
			})
		})

		Context("with StatPackets", func() {
			It("should use int formatting", func() {
				n := Number(5000)
				result := StatPackets.FormatUnit(n)
				Expect(result).To(ContainSubstring("K"))
				// Should not have decimal places
				Expect(result).NotTo(ContainSubstring("."))
			})
		})

		Context("with StatFifo", func() {
			It("should use int formatting", func() {
				n := Number(1000000)
				result := StatFifo.FormatUnit(n)
				Expect(result).To(ContainSubstring("M"))
				Expect(result).NotTo(ContainSubstring("."))
			})
		})

		Context("with StatDrop", func() {
			It("should use int formatting", func() {
				n := Number(1000)
				result := StatDrop.FormatUnit(n)
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("with StatErr", func() {
			It("should use int formatting", func() {
				n := Number(500)
				result := StatErr.FormatUnit(n)
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("with invalid stat type", func() {
			It("should return empty string", func() {
				invalidStat := Stats(99)
				n := Number(1000)
				result := invalidStat.FormatUnit(n)
				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("FormatLabelUnit() method", func() {
		It("should combine label and formatted value", func() {
			n := Number(5000)
			result := StatPackets.FormatLabelUnit(n)
			Expect(result).To(ContainSubstring("Packets:"))
			Expect(result).To(ContainSubstring("K"))
		})

		It("should format all stat types correctly", func() {
			tests := map[Stats]string{
				StatBytes:   "Traffic:",
				StatPackets: "Packets:",
				StatFifo:    "Fifo:",
				StatDrop:    "Drop:",
				StatErr:     "Error:",
			}

			n := Number(1000)
			for stat, expectedLabel := range tests {
				result := stat.FormatLabelUnit(n)
				Expect(result).To(ContainSubstring(expectedLabel))
			}
		})
	})

	Describe("FormatLabelUnitPadded() method", func() {
		It("should pad label to consistent width", func() {
			n := Number(1000)

			results := make(map[Stats]string)
			results[StatBytes] = StatBytes.FormatLabelUnitPadded(n)
			results[StatPackets] = StatPackets.FormatLabelUnitPadded(n)
			results[StatFifo] = StatFifo.FormatLabelUnitPadded(n)
			results[StatDrop] = StatDrop.FormatLabelUnitPadded(n)
			results[StatErr] = StatErr.FormatLabelUnitPadded(n)

			// All should have labels followed by consistent padding
			for stat, result := range results {
				Expect(result).To(ContainSubstring(stat.String() + ":"))
				// Check for padding spaces
				Expect(result).To(MatchRegexp(`\w+:\s+`))
			}
		})

		It("should align values for readability", func() {
			n := Number(5000)

			// Get padded results
			traffic := StatBytes.FormatLabelUnitPadded(n)
			packets := StatPackets.FormatLabelUnitPadded(n)
			fifo := StatFifo.FormatLabelUnitPadded(n)

			// Find colon positions (should be at different positions due to different label lengths)
			// But the value part should start at the same position

			// The format is: "Label: <padding> value"
			// Verify structure
			Expect(traffic).To(MatchRegexp(`\w+:\s+.+`))
			Expect(packets).To(MatchRegexp(`\w+:\s+.+`))
			Expect(fifo).To(MatchRegexp(`\w+:\s+.+`))
		})
	})

	Describe("ListStatsSort() function", func() {
		It("should return sorted list of all stat types", func() {
			list := ListStatsSort()
			Expect(list).To(HaveLen(5))
		})

		It("should include all stat types", func() {
			list := ListStatsSort()

			expectedStats := []int{
				int(StatBytes),
				int(StatPackets),
				int(StatFifo),
				int(StatDrop),
				int(StatErr),
			}

			for _, expected := range expectedStats {
				Expect(list).To(ContainElement(expected))
			}
		})

		It("should return list in sorted order", func() {
			list := ListStatsSort()

			// Verify ascending order
			for i := 0; i < len(list)-1; i++ {
				Expect(list[i]).To(BeNumerically("<", list[i+1]))
			}
		})

		It("should return consistent results", func() {
			list1 := ListStatsSort()
			list2 := ListStatsSort()
			Expect(list1).To(Equal(list2))
		})
	})

	Describe("Stat type constants", func() {
		It("should have unique values", func() {
			stats := []Stats{
				StatBytes,
				StatPackets,
				StatFifo,
				StatDrop,
				StatErr,
			}

			// Check all values are unique
			seen := make(map[Stats]bool)
			for _, stat := range stats {
				Expect(seen[stat]).To(BeFalse(), "Duplicate stat value: %v", stat)
				seen[stat] = true
			}
		})

		It("should start from 1 (iota + 1)", func() {
			Expect(StatBytes).To(Equal(Stats(1)))
			Expect(StatPackets).To(Equal(Stats(2)))
			Expect(StatFifo).To(Equal(Stats(3)))
			Expect(StatDrop).To(Equal(Stats(4)))
			Expect(StatErr).To(Equal(Stats(5)))
		})
	})

	Describe("Edge cases", func() {
		Context("with zero values", func() {
			It("should handle zero numbers correctly", func() {
				n := Number(0)

				for _, stat := range []Stats{StatBytes, StatPackets, StatFifo, StatDrop, StatErr} {
					resultInt := stat.FormatUnitInt(n)
					Expect(resultInt).NotTo(BeEmpty())

					resultFloat := stat.FormatUnitFloat(n, 2)
					Expect(resultFloat).NotTo(BeEmpty())

					resultUnit := stat.FormatUnit(n)
					Expect(resultUnit).NotTo(BeEmpty())
				}
			})
		})

		Context("with very large numbers", func() {
			It("should handle large values", func() {
				n := Number(1000000000000) // 1 trillion

				for _, stat := range []Stats{StatBytes, StatPackets, StatFifo, StatDrop, StatErr} {
					result := stat.FormatUnitInt(n)
					Expect(result).NotTo(BeEmpty())
				}
			})
		})
	})

	Describe("Formatting differences", func() {
		It("should format StatBytes differently than other stats", func() {
			n := Number(5000)

			// StatBytes uses binary units
			bytesResult := StatBytes.FormatUnitInt(n)
			Expect(bytesResult).To(ContainSubstring("KB"))

			// StatPackets uses decimal units (5000 = 5K)
			packetsResult := StatPackets.FormatUnitInt(n)
			Expect(packetsResult).To(ContainSubstring("K"))
			Expect(packetsResult).NotTo(ContainSubstring("KB"))
		})
	})

	Describe("Performance", func() {
		It("should handle repeated calls efficiently", func() {
			n := Number(12345)

			Expect(func() {
				for i := 0; i < 10000; i++ {
					for _, stat := range []Stats{StatBytes, StatPackets, StatFifo, StatDrop, StatErr} {
						_ = stat.String()
						_ = stat.FormatUnitInt(n)
						_ = stat.FormatUnitFloat(n, 2)
						_ = stat.FormatUnit(n)
						_ = stat.FormatLabelUnit(n)
						_ = stat.FormatLabelUnitPadded(n)
					}
				}
			}).NotTo(Panic())
		})
	})
})
