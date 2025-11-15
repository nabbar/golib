/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package protocol_test

import (
	. "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Protocol Format Conversions", func() {
	Describe("Int() method", func() {
		Context("with valid protocols", func() {
			It("should return correct int for NetworkTCP", func() {
				result := NetworkTCP.Int()
				Expect(result).To(Equal(2))
			})

			It("should return correct int for NetworkUDP", func() {
				result := NetworkUDP.Int()
				Expect(result).To(Equal(5))
			})

			It("should return correct int for NetworkUnix", func() {
				result := NetworkUnix.Int()
				Expect(result).To(Equal(1))
			})

			It("should return correct int for all protocols", func() {
				tests := map[NetworkProtocol]int{
					NetworkUnix:     1,
					NetworkTCP:      2,
					NetworkTCP4:     3,
					NetworkTCP6:     4,
					NetworkUDP:      5,
					NetworkUDP4:     6,
					NetworkUDP6:     7,
					NetworkIP:       8,
					NetworkIP4:      9,
					NetworkIP6:      10,
					NetworkUnixGram: 11,
				}

				for protocol, expected := range tests {
					result := protocol.Int()
					Expect(result).To(Equal(expected), "Failed for protocol %v", protocol)
				}
			})
		})

		Context("with invalid protocols", func() {
			It("should return 0 for NetworkEmpty", func() {
				result := NetworkEmpty.Int()
				Expect(result).To(Equal(0))
			})

			It("should return 0 for undefined protocol value", func() {
				invalidProtocol := NetworkProtocol(99)
				result := invalidProtocol.Int()
				Expect(result).To(Equal(0))
			})

			It("should return 0 for out of range protocol", func() {
				invalidProtocol := NetworkProtocol(255)
				result := invalidProtocol.Int()
				Expect(result).To(Equal(0))
			})
		})
	})

	Describe("Int64() method", func() {
		Context("with valid protocols", func() {
			It("should return correct int64 for NetworkTCP", func() {
				result := NetworkTCP.Int64()
				Expect(result).To(Equal(int64(2)))
			})

			It("should return correct int64 for NetworkUDP", func() {
				result := NetworkUDP.Int64()
				Expect(result).To(Equal(int64(5)))
			})

			It("should return correct int64 for NetworkUnix", func() {
				result := NetworkUnix.Int64()
				Expect(result).To(Equal(int64(1)))
			})

			It("should return correct int64 for all protocols", func() {
				tests := map[NetworkProtocol]int64{
					NetworkUnix:     1,
					NetworkTCP:      2,
					NetworkTCP4:     3,
					NetworkTCP6:     4,
					NetworkUDP:      5,
					NetworkUDP4:     6,
					NetworkUDP6:     7,
					NetworkIP:       8,
					NetworkIP4:      9,
					NetworkIP6:      10,
					NetworkUnixGram: 11,
				}

				for protocol, expected := range tests {
					result := protocol.Int64()
					Expect(result).To(Equal(expected), "Failed for protocol %v", protocol)
				}
			})
		})

		Context("with invalid protocols", func() {
			It("should return 0 for NetworkEmpty", func() {
				result := NetworkEmpty.Int64()
				Expect(result).To(Equal(int64(0)))
			})

			It("should return 0 for undefined protocol value", func() {
				invalidProtocol := NetworkProtocol(99)
				result := invalidProtocol.Int64()
				Expect(result).To(Equal(int64(0)))
			})

			It("should return 0 for out of range protocol", func() {
				invalidProtocol := NetworkProtocol(255)
				result := invalidProtocol.Int64()
				Expect(result).To(Equal(int64(0)))
			})
		})

		Context("roundtrip with ParseInt64", func() {
			It("should maintain consistency with ParseInt64", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, original := range protocols {
					// Protocol -> Int64 -> ParseInt64 -> Protocol
					val := original.Int64()
					parsed := ParseInt64(val)
					Expect(parsed).To(Equal(original),
						"Roundtrip failed for %v: Int64()=%d, ParseInt64()=%v",
						original, val, parsed)
				}
			})
		})
	})

	Describe("Uint() method", func() {
		Context("with valid protocols", func() {
			It("should return correct uint for NetworkTCP", func() {
				result := NetworkTCP.Uint()
				Expect(result).To(Equal(uint(2)))
			})

			It("should return correct uint for NetworkUDP", func() {
				result := NetworkUDP.Uint()
				Expect(result).To(Equal(uint(5)))
			})

			It("should return correct uint for NetworkUnix", func() {
				result := NetworkUnix.Uint()
				Expect(result).To(Equal(uint(1)))
			})

			It("should return correct uint for all protocols", func() {
				tests := map[NetworkProtocol]uint{
					NetworkUnix:     1,
					NetworkTCP:      2,
					NetworkTCP4:     3,
					NetworkTCP6:     4,
					NetworkUDP:      5,
					NetworkUDP4:     6,
					NetworkUDP6:     7,
					NetworkIP:       8,
					NetworkIP4:      9,
					NetworkIP6:      10,
					NetworkUnixGram: 11,
				}

				for protocol, expected := range tests {
					result := protocol.Uint()
					Expect(result).To(Equal(expected), "Failed for protocol %v", protocol)
				}
			})
		})

		Context("with invalid protocols", func() {
			It("should return 0 for NetworkEmpty", func() {
				result := NetworkEmpty.Uint()
				Expect(result).To(Equal(uint(0)))
			})

			It("should return 0 for undefined protocol value", func() {
				invalidProtocol := NetworkProtocol(99)
				result := invalidProtocol.Uint()
				Expect(result).To(Equal(uint(0)))
			})

			It("should return 0 for out of range protocol", func() {
				invalidProtocol := NetworkProtocol(255)
				result := invalidProtocol.Uint()
				Expect(result).To(Equal(uint(0)))
			})
		})
	})

	Describe("Uint64() method", func() {
		Context("with valid protocols", func() {
			It("should return correct uint64 for NetworkTCP", func() {
				result := NetworkTCP.Uint64()
				Expect(result).To(Equal(uint64(2)))
			})

			It("should return correct uint64 for NetworkUDP", func() {
				result := NetworkUDP.Uint64()
				Expect(result).To(Equal(uint64(5)))
			})

			It("should return correct uint64 for NetworkUnix", func() {
				result := NetworkUnix.Uint64()
				Expect(result).To(Equal(uint64(1)))
			})

			It("should return correct uint64 for all protocols", func() {
				tests := map[NetworkProtocol]uint64{
					NetworkUnix:     1,
					NetworkTCP:      2,
					NetworkTCP4:     3,
					NetworkTCP6:     4,
					NetworkUDP:      5,
					NetworkUDP4:     6,
					NetworkUDP6:     7,
					NetworkIP:       8,
					NetworkIP4:      9,
					NetworkIP6:      10,
					NetworkUnixGram: 11,
				}

				for protocol, expected := range tests {
					result := protocol.Uint64()
					Expect(result).To(Equal(expected), "Failed for protocol %v", protocol)
				}
			})
		})

		Context("with invalid protocols", func() {
			It("should return 0 for NetworkEmpty", func() {
				result := NetworkEmpty.Uint64()
				Expect(result).To(Equal(uint64(0)))
			})

			It("should return 0 for undefined protocol value", func() {
				invalidProtocol := NetworkProtocol(99)
				result := invalidProtocol.Uint64()
				Expect(result).To(Equal(uint64(0)))
			})

			It("should return 0 for out of range protocol", func() {
				invalidProtocol := NetworkProtocol(255)
				result := invalidProtocol.Uint64()
				Expect(result).To(Equal(uint64(0)))
			})
		})
	})

	Describe("Type conversion consistency", func() {
		Context("comparing different conversion methods", func() {
			It("should have Int() equal to int(protocol) for valid protocols", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix,
				}

				for _, p := range protocols {
					Expect(p.Int()).To(Equal(int(p)))
				}
			})

			It("should have consistent values across all numeric types", func() {
				p := NetworkTCP

				intVal := p.Int()
				int64Val := p.Int64()
				uintVal := p.Uint()
				uint64Val := p.Uint64()

				// All should represent the same value
				Expect(int64Val).To(Equal(int64(intVal)))
				Expect(uint64Val).To(Equal(uint64(intVal)))
				Expect(uintVal).To(Equal(uint(intVal)))
			})

			It("should have zero values for invalid protocols across all types", func() {
				invalid := NetworkProtocol(99)

				Expect(invalid.Int()).To(Equal(0))
				Expect(invalid.Int64()).To(Equal(int64(0)))
				Expect(invalid.Uint()).To(Equal(uint(0)))
				Expect(invalid.Uint64()).To(Equal(uint64(0)))
			})
		})

		Context("with NetworkEmpty", func() {
			It("should return zero for all numeric conversions", func() {
				Expect(NetworkEmpty.Int()).To(Equal(0))
				Expect(NetworkEmpty.Int64()).To(Equal(int64(0)))
				Expect(NetworkEmpty.Uint()).To(Equal(uint(0)))
				Expect(NetworkEmpty.Uint64()).To(Equal(uint64(0)))
			})

			It("should return empty string", func() {
				Expect(NetworkEmpty.String()).To(Equal(""))
				Expect(NetworkEmpty.Code()).To(Equal(""))
			})
		})
	})

	Describe("Performance and memory", func() {
		It("should not panic with repeated conversions", func() {
			p := NetworkTCP

			Expect(func() {
				for i := 0; i < 10000; i++ {
					_ = p.Int()
					_ = p.Int64()
					_ = p.Uint()
					_ = p.Uint64()
				}
			}).NotTo(Panic())
		})

		It("should handle all protocols without panic", func() {
			protocols := []NetworkProtocol{
				NetworkEmpty, NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
				NetworkUDP, NetworkUDP4, NetworkUDP6,
				NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
			}

			Expect(func() {
				for _, p := range protocols {
					_ = p.Int()
					_ = p.Int64()
					_ = p.Uint()
					_ = p.Uint64()
				}
			}).NotTo(Panic())
		})
	})

	Describe("Edge cases", func() {
		Context("with boundary values", func() {
			It("should handle protocol value 1 (NetworkUnix)", func() {
				p := NetworkProtocol(1)
				Expect(p.Int()).To(Equal(1))
				Expect(p.Int64()).To(Equal(int64(1)))
				Expect(p.Uint()).To(Equal(uint(1)))
				Expect(p.Uint64()).To(Equal(uint64(1)))
			})

			It("should handle protocol value 11 (NetworkUnixGram)", func() {
				p := NetworkProtocol(11)
				Expect(p.Int()).To(Equal(11))
				Expect(p.Int64()).To(Equal(int64(11)))
				Expect(p.Uint()).To(Equal(uint(11)))
				Expect(p.Uint64()).To(Equal(uint64(11)))
			})

			It("should return 0 for protocol value 12 (invalid)", func() {
				p := NetworkProtocol(12)
				Expect(p.Int()).To(Equal(0))
				Expect(p.Int64()).To(Equal(int64(0)))
				Expect(p.Uint()).To(Equal(uint(0)))
				Expect(p.Uint64()).To(Equal(uint64(0)))
			})
		})

		Context("with unusual values", func() {
			It("should handle protocol value 50 (undefined)", func() {
				p := NetworkProtocol(50)
				Expect(p.Int()).To(Equal(0))
				Expect(p.Uint()).To(Equal(uint(0)))
			})

			It("should handle protocol value 200 (undefined)", func() {
				p := NetworkProtocol(200)
				Expect(p.Int64()).To(Equal(int64(0)))
				Expect(p.Uint64()).To(Equal(uint64(0)))
			})
		})
	})

	Describe("Integration with other methods", func() {
		Context("combined with String()", func() {
			It("should have valid string for protocols with non-zero int", func() {
				protocols := []NetworkProtocol{
					NetworkTCP, NetworkUDP, NetworkUnix,
				}

				for _, p := range protocols {
					if p.Int() > 0 {
						Expect(p.String()).NotTo(BeEmpty())
					}
				}
			})

			It("should have empty string for protocols with zero int", func() {
				if NetworkEmpty.Int() == 0 {
					Expect(NetworkEmpty.String()).To(BeEmpty())
				}
			})
		})

		Context("validation pattern", func() {
			It("should validate protocol by checking Int() != 0", func() {
				// Valid protocols should have Int() != 0
				Expect(NetworkTCP.Int()).NotTo(Equal(0))
				Expect(NetworkUDP.Int()).NotTo(Equal(0))

				// Invalid protocols should have Int() == 0
				invalid := NetworkProtocol(99)
				Expect(invalid.Int()).To(Equal(0))
			})

			It("should validate protocol by checking String() != empty", func() {
				// Valid protocols should have non-empty string
				if NetworkTCP.Int() != 0 {
					Expect(NetworkTCP.String()).NotTo(BeEmpty())
				}

				// Invalid protocols should have empty string
				invalid := NetworkProtocol(99)
				if invalid.Int() == 0 {
					Expect(invalid.String()).To(BeEmpty())
				}
			})
		})
	})
})
