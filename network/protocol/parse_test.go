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

package protocol_test

import (
	"math"

	. "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Protocol Parsing", func() {
	Describe("Parse function", func() {
		Context("with valid protocol strings", func() {
			It("should parse tcp (lowercase)", func() {
				result := Parse("tcp")
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should parse TCP (uppercase)", func() {
				result := Parse("TCP")
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should parse tcp4", func() {
				result := Parse("tcp4")
				Expect(result).To(Equal(NetworkTCP4))
			})

			It("should parse tcp6", func() {
				result := Parse("tcp6")
				Expect(result).To(Equal(NetworkTCP6))
			})

			It("should parse udp", func() {
				result := Parse("udp")
				Expect(result).To(Equal(NetworkUDP))
			})

			It("should parse unix", func() {
				result := Parse("unix")
				Expect(result).To(Equal(NetworkUnix))
			})

			It("should parse unixgram (case insensitive)", func() {
				result := Parse("unixgram")
				Expect(result).To(Equal(NetworkUnixGram))
			})

			It("should parse UnixGram with mixed case", func() {
				result := Parse("UnixGram")
				Expect(result).To(Equal(NetworkUnixGram))
			})
		})

		Context("with all protocol strings", func() {
			// ✅ FIXED: Parse() now handles all protocols
			It("should parse udp4", func() {
				result := Parse("udp4")
				Expect(result).To(Equal(NetworkUDP4))
			})

			It("should parse udp6", func() {
				result := Parse("udp6")
				Expect(result).To(Equal(NetworkUDP6))
			})

			It("should parse ip", func() {
				result := Parse("ip")
				Expect(result).To(Equal(NetworkIP))
			})

			It("should parse ip4", func() {
				result := Parse("ip4")
				Expect(result).To(Equal(NetworkIP4))
			})

			It("should parse ip6", func() {
				result := Parse("ip6")
				Expect(result).To(Equal(NetworkIP6))
			})

			It("should parse all protocols case-insensitively", func() {
				tests := map[string]NetworkProtocol{
					"TCP":      NetworkTCP,
					"udp":      NetworkUDP,
					"UnixGram": NetworkUnixGram,
					"IP4":      NetworkIP4,
					"UDP6":     NetworkUDP6,
				}

				for input, expected := range tests {
					result := Parse(input)
					Expect(result).To(Equal(expected), "Failed for: %s", input)
				}
			})

			It("should return NetworkEmpty for unknown protocol", func() {
				result := Parse("invalid")
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should return NetworkEmpty for empty string", func() {
				result := Parse("")
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should return NetworkEmpty for random text", func() {
				result := Parse("http")
				Expect(result).To(Equal(NetworkEmpty))
			})
		})

		Context("with edge cases", func() {
			// ✅ FIXED: Parse() now trims whitespace
			It("should handle whitespace (trimmed)", func() {
				result := Parse(" tcp ")
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should handle leading whitespace", func() {
				result := Parse("  udp")
				Expect(result).To(Equal(NetworkUDP))
			})

			It("should handle trailing whitespace", func() {
				result := Parse("unix  ")
				Expect(result).To(Equal(NetworkUnix))
			})

			It("should handle tabs and newlines", func() {
				result := Parse("\ttcp\n")
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should handle protocols with different casing", func() {
				result := Parse("TcP")
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should handle very long strings without panic", func() {
				longString := string(make([]byte, 10000))
				Expect(func() { Parse(longString) }).NotTo(Panic())
			})

			// ✅ NEW: Parse() now handles quoted strings
			It("should handle double-quoted strings", func() {
				result := Parse(`"tcp"`)
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should handle raw string with backslash-escaped quotes", func() {
				// Note: `\"udp\"` is literally the 7 characters: \, ", u, d, p, ", \
				// After trimming quotes, it becomes: \, ", u, d, p, "
				// This doesn't match "udp" so returns NetworkEmpty
				result := Parse(`\"udp\"`)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should handle Go-style quoted strings", func() {
				result := Parse("`unix`")
				Expect(result).To(Equal(NetworkUnix))
			})

			It("should handle whitespace and quotes together", func() {
				result := Parse(`  "tcp"  `)
				Expect(result).To(Equal(NetworkTCP))
			})
		})
	})

	Describe("ParseBytes function", func() {
		It("should parse valid bytes to NetworkTCP", func() {
			result := ParseBytes([]byte("tcp"))
			Expect(result).To(Equal(NetworkTCP))
		})

		It("should parse valid bytes to NetworkUDP", func() {
			result := ParseBytes([]byte("udp"))
			Expect(result).To(Equal(NetworkUDP))
		})

		It("should handle empty bytes", func() {
			result := ParseBytes([]byte{})
			Expect(result).To(Equal(NetworkEmpty))
		})

		It("should handle nil bytes without panic", func() {
			Expect(func() { ParseBytes(nil) }).NotTo(Panic())
		})

		// ⚠️ POTENTIAL MEMORY ISSUE: Large byte arrays
		It("should handle large byte arrays without memory overflow", func() {
			largeBytes := make([]byte, 1024*1024) // 1MB
			Expect(func() { ParseBytes(largeBytes) }).NotTo(Panic())
		})

		It("should handle bytes with null characters", func() {
			result := ParseBytes([]byte("tcp\x00"))
			// Null character in string causes parse failure
			// The string becomes "tcp\x00" which doesn't match "tcp"
			Expect(result).To(Equal(NetworkEmpty))
		})
	})

	Describe("ParseInt64 function", func() {
		Context("with valid NetworkProtocol values", func() {
			It("should parse 1 to NetworkUnix", func() {
				result := ParseInt64(1)
				Expect(result).To(Equal(NetworkUnix))
			})

			It("should parse 2 to NetworkTCP", func() {
				result := ParseInt64(2)
				Expect(result).To(Equal(NetworkTCP))
			})

			It("should parse 3 to NetworkTCP4", func() {
				result := ParseInt64(3)
				Expect(result).To(Equal(NetworkTCP4))
			})

			It("should parse 4 to NetworkTCP6", func() {
				result := ParseInt64(4)
				Expect(result).To(Equal(NetworkTCP6))
			})

			It("should parse 5 to NetworkUDP", func() {
				result := ParseInt64(5)
				Expect(result).To(Equal(NetworkUDP))
			})

			It("should parse 11 to NetworkUnixGram", func() {
				result := ParseInt64(11)
				Expect(result).To(Equal(NetworkUnixGram))
			})
		})

		Context("with all valid protocol values", func() {
			// ✅ FIXED: ParseInt64 now handles all protocols
			It("should handle NetworkUDP4 value", func() {
				result := ParseInt64(6) // NetworkUDP4 value
				Expect(result).To(Equal(NetworkUDP4))
			})

			It("should handle NetworkUDP6 value", func() {
				result := ParseInt64(7)
				Expect(result).To(Equal(NetworkUDP6))
			})

			It("should handle NetworkIP value", func() {
				result := ParseInt64(8)
				Expect(result).To(Equal(NetworkIP))
			})

			It("should handle NetworkIP4 value", func() {
				result := ParseInt64(9)
				Expect(result).To(Equal(NetworkIP4))
			})

			It("should handle NetworkIP6 value", func() {
				result := ParseInt64(10)
				Expect(result).To(Equal(NetworkIP6))
			})

			It("should return NetworkEmpty for 0", func() {
				result := ParseInt64(0)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should return NetworkEmpty for invalid constant value", func() {
				result := ParseInt64(99)
				Expect(result).To(Equal(NetworkEmpty))
			})
		})

		Context("with edge cases and boundary values", func() {
			// ✅ FIXED: Negative values are now explicitly rejected
			It("should reject negative values", func() {
				result := ParseInt64(-1)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should handle large negative values", func() {
				result := ParseInt64(-1000)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should handle math.MinInt64", func() {
				result := ParseInt64(math.MinInt64)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should cap at MaxUint8 for values > 255", func() {
				result := ParseInt64(256)
				// Should be capped at 255 (math.MaxUint8)
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should cap at MaxUint8 for very large values", func() {
				result := ParseInt64(math.MaxInt64)
				// Should be capped at 255
				Expect(result).To(Equal(NetworkEmpty))
			})

			It("should handle MaxUint8 boundary", func() {
				result := ParseInt64(255)
				// 255 is not a valid protocol, should return NetworkEmpty
				Expect(result).To(Equal(NetworkEmpty))
			})

			// ⚠️ MEMORY SAFETY: These tests verify no panic occurs with extreme values
			It("should not panic with extreme positive value", func() {
				Expect(func() { ParseInt64(math.MaxInt64) }).NotTo(Panic())
			})

			It("should not panic with extreme negative value", func() {
				Expect(func() { ParseInt64(math.MinInt64) }).NotTo(Panic())
			})
		})

		Context("conversion accuracy", func() {
			It("should preserve valid uint8 range values", func() {
				for i := int64(0); i <= 11; i++ {
					Expect(func() { ParseInt64(i) }).NotTo(Panic())
				}
			})

			// Verify that the function doesn't incorrectly convert values
			It("should not mistake large values for valid protocols", func() {
				result := ParseInt64(1000)
				Expect(result).To(Equal(NetworkEmpty))
			})
		})
	})

	Describe("Protocol constant integrity", func() {
		It("should have unique values for each protocol", func() {
			protocols := []NetworkProtocol{
				NetworkEmpty,
				NetworkUnix,
				NetworkTCP,
				NetworkTCP4,
				NetworkTCP6,
				NetworkUDP,
				NetworkUDP4,
				NetworkUDP6,
				NetworkIP,
				NetworkIP4,
				NetworkIP6,
				NetworkUnixGram,
			}

			// Check for duplicates
			seen := make(map[NetworkProtocol]bool)
			for _, p := range protocols {
				Expect(seen[p]).To(BeFalse(), "Duplicate protocol value: %v", p)
				seen[p] = true
			}
		})

		It("should have NetworkEmpty as zero value", func() {
			var p NetworkProtocol
			Expect(p).To(Equal(NetworkEmpty))
		})

		It("should fit within uint8 range", func() {
			// NetworkProtocol is uint8, so all values should be <= 255
			Expect(uint8(NetworkUnixGram)).To(BeNumerically("<=", 255))
		})
	})
})
