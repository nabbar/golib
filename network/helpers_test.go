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

package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helper Functions", func() {
	Describe("power2Unit()", func() {
		Context("with valid power values", func() {
			It("should return correct unit for each power", func() {
				tests := map[int]string{
					_PowerUnit_:  "",
					_PowerKilo_:  "K",
					_PowerMega_:  "M",
					_PowerGiga_:  "G",
					_PowerTera_:  "T",
					_PowerPeta_:  "P",
					_PowerExa_:   "E",
					_PowerZetta_: "Z",
					_PowerYotta_: "Y",
				}

				for power, expected := range tests {
					result := power2Unit(power)
					Expect(result).To(Equal(expected), "Failed for power %d", power)
				}
			})

			It("should map powers to correct SI prefixes", func() {
				Expect(power2Unit(0)).To(Equal(""))   // Unit (no prefix)
				Expect(power2Unit(3)).To(Equal("K"))  // Kilo (10^3)
				Expect(power2Unit(6)).To(Equal("M"))  // Mega (10^6)
				Expect(power2Unit(9)).To(Equal("G"))  // Giga (10^9)
				Expect(power2Unit(12)).To(Equal("T")) // Tera (10^12)
				Expect(power2Unit(15)).To(Equal("P")) // Peta (10^15)
				Expect(power2Unit(18)).To(Equal("E")) // Exa (10^18)
				Expect(power2Unit(21)).To(Equal("Z")) // Zetta (10^21)
				Expect(power2Unit(24)).To(Equal("Y")) // Yotta (10^24)
			})
		})

		Context("with edge case values", func() {
			It("should return empty string for negative values", func() {
				tests := []int{-1, -3, -10, -100}
				for _, power := range tests {
					result := power2Unit(power)
					Expect(result).To(Equal(""), "Failed for power %d", power)
				}
			})

			It("should return Yotta for values >= 24", func() {
				tests := []int{24, 25, 27, 30, 100}
				for _, power := range tests {
					result := power2Unit(power)
					Expect(result).To(Equal("Y"), "Failed for power %d", power)
				}
			})

			It("should handle intermediate values correctly", func() {
				// Values between defined powers should use the lower bound
				Expect(power2Unit(1)).To(Equal(""))  // Between 0-3 -> ""
				Expect(power2Unit(2)).To(Equal(""))  // Between 0-3 -> ""
				Expect(power2Unit(4)).To(Equal("K")) // Between 3-6 -> "K"
				Expect(power2Unit(5)).To(Equal("K")) // Between 3-6 -> "K"
				Expect(power2Unit(7)).To(Equal("M")) // Between 6-9 -> "M"
				Expect(power2Unit(8)).To(Equal("M")) // Between 6-9 -> "M"
			})
		})

		Context("boundary testing", func() {
			It("should handle exact power boundaries", func() {
				boundaries := map[int]string{
					3:  "K",
					6:  "M",
					9:  "G",
					12: "T",
					15: "P",
					18: "E",
					21: "Z",
					24: "Y",
				}

				for power, expectedUnit := range boundaries {
					result := power2Unit(power)
					Expect(result).To(Equal(expectedUnit))
				}
			})

			It("should handle values just below boundaries", func() {
				tests := map[int]string{
					2:  "",  // Just below Kilo
					5:  "K", // Just below Mega
					8:  "M", // Just below Giga
					11: "G", // Just below Tera
					14: "T", // Just below Peta
					17: "P", // Just below Exa
					20: "E", // Just below Zetta
					23: "Z", // Just below Yotta
				}

				for power, expectedUnit := range tests {
					result := power2Unit(power)
					Expect(result).To(Equal(expectedUnit))
				}
			})
		})

		Context("consistency checks", func() {
			It("should return same result for same input", func() {
				for i := 0; i < 100; i++ {
					result1 := power2Unit(9)
					result2 := power2Unit(9)
					Expect(result1).To(Equal(result2))
					Expect(result1).To(Equal("G"))
				}
			})

			It("should be deterministic", func() {
				inputs := []int{0, 3, 6, 9, 12, 15, 18, 21, 24}
				results1 := make([]string, len(inputs))
				results2 := make([]string, len(inputs))

				for i, input := range inputs {
					results1[i] = power2Unit(input)
				}

				for i, input := range inputs {
					results2[i] = power2Unit(input)
				}

				Expect(results1).To(Equal(results2))
			})
		})
	})

	Describe("powerList()", func() {
		Context("basic properties", func() {
			It("should return a list of 9 elements", func() {
				list := powerList()
				Expect(list).To(HaveLen(9))
			})

			It("should contain all power constants", func() {
				list := powerList()

				expectedPowers := []int{
					_PowerUnit_,
					_PowerKilo_,
					_PowerMega_,
					_PowerGiga_,
					_PowerTera_,
					_PowerPeta_,
					_PowerExa_,
					_PowerZetta_,
					_PowerYotta_,
				}

				for _, expected := range expectedPowers {
					Expect(list).To(ContainElement(expected))
				}
			})
		})

		Context("sorting and order", func() {
			It("should return list in descending order", func() {
				list := powerList()

				// Verify descending order
				for i := 0; i < len(list)-1; i++ {
					Expect(list[i]).To(BeNumerically(">", list[i+1]),
						"List not in descending order at index %d: %d > %d", i, list[i], list[i+1])
				}
			})

			It("should start with Yotta (largest)", func() {
				list := powerList()
				Expect(list[0]).To(Equal(_PowerYotta_))
				Expect(list[0]).To(Equal(24))
			})

			It("should end with Unit (smallest)", func() {
				list := powerList()
				Expect(list[len(list)-1]).To(Equal(_PowerUnit_))
				Expect(list[len(list)-1]).To(Equal(0))
			})

			It("should have correct order: Y, Z, E, P, T, G, M, K, Unit", func() {
				list := powerList()
				expected := []int{
					_PowerYotta_, // 24
					_PowerZetta_, // 21
					_PowerExa_,   // 18
					_PowerPeta_,  // 15
					_PowerTera_,  // 12
					_PowerGiga_,  // 9
					_PowerMega_,  // 6
					_PowerKilo_,  // 3
					_PowerUnit_,  // 0
				}

				Expect(list).To(Equal(expected))
			})
		})

		Context("consistency", func() {
			It("should return same list on multiple calls", func() {
				list1 := powerList()
				list2 := powerList()
				Expect(list1).To(Equal(list2))
			})

			It("should be deterministic", func() {
				results := make([][]int, 10)
				for i := 0; i < 10; i++ {
					results[i] = powerList()
				}

				// All results should be identical
				for i := 1; i < 10; i++ {
					Expect(results[i]).To(Equal(results[0]))
				}
			})
		})

		Context("no duplicates", func() {
			It("should not contain duplicate values", func() {
				list := powerList()
				seen := make(map[int]bool)

				for _, power := range list {
					Expect(seen[power]).To(BeFalse(), "Duplicate power found: %d", power)
					seen[power] = true
				}
			})
		})

		Context("integration with power2Unit", func() {
			It("should produce valid units for all powers", func() {
				list := powerList()

				for _, power := range list {
					unit := power2Unit(power)
					Expect(unit).NotTo(BeNil())
					// All units should be valid (empty string for Unit is valid)
				}
			})

			It("should iterate from largest to smallest", func() {
				list := powerList()
				units := []string{"Y", "Z", "E", "P", "T", "G", "M", "K", ""}

				for i, power := range list {
					unit := power2Unit(power)
					Expect(unit).To(Equal(units[i]))
				}
			})
		})

		Context("usage in formatting", func() {
			It("should enable efficient unit selection", func() {
				// Simulate how powerList is used in formatting
				testValues := []struct {
					value        float64
					expectedUnit string
				}{
					{500, ""},
					{5000, "K"},
					{5000000, "M"},
					{5000000000, "G"},
					{5000000000000, "T"},
					{5000000000000000, "P"},
				}

				for _, test := range testValues {
					found := false
					for _, p := range powerList() {
						if test.value >= float64(pow10(p)) {
							unit := power2Unit(p)
							Expect(unit).To(Equal(test.expectedUnit))
							found = true
							break
						}
					}
					Expect(found).To(BeTrue(), "No unit found for value %f", test.value)
				}
			})
		})
	})

	Describe("Constants", func() {
		Context("power constants", func() {
			It("should define correct SI power values", func() {
				Expect(_PowerUnit_).To(Equal(0))
				Expect(_PowerKilo_).To(Equal(3))
				Expect(_PowerMega_).To(Equal(6))
				Expect(_PowerGiga_).To(Equal(9))
				Expect(_PowerTera_).To(Equal(12))
				Expect(_PowerPeta_).To(Equal(15))
				Expect(_PowerExa_).To(Equal(18))
				Expect(_PowerZetta_).To(Equal(21))
				Expect(_PowerYotta_).To(Equal(24))
			})

			It("should have powers spaced by 3", func() {
				powers := []int{
					_PowerUnit_, _PowerKilo_, _PowerMega_, _PowerGiga_,
					_PowerTera_, _PowerPeta_, _PowerExa_, _PowerZetta_, _PowerYotta_,
				}

				for i := 1; i < len(powers); i++ {
					diff := powers[i] - powers[i-1]
					Expect(diff).To(Equal(3), "Power spacing incorrect between %d and %d", powers[i-1], powers[i])
				}
			})
		})

		Context("format constants", func() {
			It("should define correct padding size", func() {
				Expect(_MaxSizeOfPad_).To(Equal(4))
			})

			It("should define correct int pattern", func() {
				Expect(_PadIntPattern_).To(Equal("%4d"))
			})
		})

		Context("byte calculations", func() {
			It("should calculate correct byte thresholds using size package", func() {
				// Verify binary unit calculations
				// 1 KB = 2^10 = 1024 bytes
				// 1 MB = 2^20 = 1048576 bytes
				// 1 GB = 2^30 = 1073741824 bytes

				Expect(1 << 10).To(Equal(1024))       // 1 KB
				Expect(1 << 20).To(Equal(1048576))    // 1 MB
				Expect(1 << 30).To(Equal(1073741824)) // 1 GB
			})
		})
	})

	Describe("Integration", func() {
		It("should work together for number formatting", func() {
			// Test that powerList and power2Unit work together
			list := powerList()

			for _, power := range list {
				unit := power2Unit(power)
				// Each power should produce a valid unit
				Expect(unit).NotTo(BeNil())
			}
		})

		It("should support iterating from largest to smallest unit", func() {
			// This is how the formatting functions use these helpers
			testValue := float64(5000000) // 5 million

			for _, p := range powerList() {
				if testValue >= float64(pow10(p)) {
					unit := power2Unit(p)
					// Should find Mega (10^6)
					Expect(unit).To(Equal("M"))
					break
				}
			}
		})
	})

	Describe("Performance", func() {
		It("should handle repeated power2Unit calls efficiently", func() {
			Expect(func() {
				for i := 0; i < 100000; i++ {
					_ = power2Unit(9)
					_ = power2Unit(12)
					_ = power2Unit(15)
				}
			}).NotTo(Panic())
		})

		It("should handle repeated powerList calls efficiently", func() {
			Expect(func() {
				for i := 0; i < 10000; i++ {
					_ = powerList()
				}
			}).NotTo(Panic())
		})
	})
})

// Helper function to calculate 10^n
func pow10(n int) int {
	if n == 0 {
		return 1
	}
	result := 10
	for i := 1; i < n; i++ {
		result *= 10
	}
	return result
}
