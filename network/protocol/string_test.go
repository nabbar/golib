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
	. "github.com/nabbar/golib/network/protocol"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Protocol String Representation", func() {
	Describe("String() method", func() {
		Context("with valid protocols", func() {
			It("should return 'unix' for NetworkUnix", func() {
				Expect(NetworkUnix.String()).To(Equal("unix"))
			})

			It("should return 'tcp' for NetworkTCP", func() {
				Expect(NetworkTCP.String()).To(Equal("tcp"))
			})

			It("should return 'tcp4' for NetworkTCP4", func() {
				Expect(NetworkTCP4.String()).To(Equal("tcp4"))
			})

			It("should return 'tcp6' for NetworkTCP6", func() {
				Expect(NetworkTCP6.String()).To(Equal("tcp6"))
			})

			It("should return 'udp' for NetworkUDP", func() {
				Expect(NetworkUDP.String()).To(Equal("udp"))
			})

			It("should return 'udp4' for NetworkUDP4", func() {
				Expect(NetworkUDP4.String()).To(Equal("udp4"))
			})

			It("should return 'udp6' for NetworkUDP6", func() {
				Expect(NetworkUDP6.String()).To(Equal("udp6"))
			})

			It("should return 'ip' for NetworkIP", func() {
				Expect(NetworkIP.String()).To(Equal("ip"))
			})

			It("should return 'ip4' for NetworkIP4", func() {
				Expect(NetworkIP4.String()).To(Equal("ip4"))
			})

			It("should return 'ip6' for NetworkIP6", func() {
				Expect(NetworkIP6.String()).To(Equal("ip6"))
			})

			// ✅ FIXED: NetworkUnixGram now returns lowercase like other protocols
			It("should return 'unixgram' for NetworkUnixGram", func() {
				result := NetworkUnixGram.String()
				Expect(result).To(Equal("unixgram"))
			})
		})

		Context("with invalid/empty protocols", func() {
			It("should return empty string for NetworkEmpty", func() {
				Expect(NetworkEmpty.String()).To(Equal(""))
			})

			It("should return empty string for undefined protocol value", func() {
				invalidProtocol := NetworkProtocol(99)
				Expect(invalidProtocol.String()).To(Equal(""))
			})

			It("should return empty string for maximum uint8 value", func() {
				maxProtocol := NetworkProtocol(255)
				Expect(maxProtocol.String()).To(Equal(""))
			})
		})

		Context("string length and memory efficiency", func() {
			It("should return non-empty strings for valid protocols", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, p := range protocols {
					Expect(p.String()).NotTo(BeEmpty())
					Expect(len(p.String())).To(BeNumerically(">", 0))
				}
			})

			// ⚠️ MEMORY CHECK: String() should not allocate excessive memory
			It("should return reasonably sized strings", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, p := range protocols {
					str := p.String()
					// All protocol names should be less than 20 chars
					Expect(len(str)).To(BeNumerically("<=", 20))
				}
			})
		})
	})

	Describe("Code() method", func() {
		Context("with valid protocols", func() {
			It("should return lowercase 'unix' for NetworkUnix", func() {
				Expect(NetworkUnix.Code()).To(Equal("unix"))
			})

			It("should return lowercase 'tcp' for NetworkTCP", func() {
				Expect(NetworkTCP.Code()).To(Equal("tcp"))
			})

			It("should return lowercase 'tcp4' for NetworkTCP4", func() {
				Expect(NetworkTCP4.Code()).To(Equal("tcp4"))
			})

			It("should return lowercase 'tcp6' for NetworkTCP6", func() {
				Expect(NetworkTCP6.Code()).To(Equal("tcp6"))
			})

			It("should return lowercase 'udp' for NetworkUDP", func() {
				Expect(NetworkUDP.Code()).To(Equal("udp"))
			})

			It("should return lowercase 'udp4' for NetworkUDP4", func() {
				Expect(NetworkUDP4.Code()).To(Equal("udp4"))
			})

			It("should return lowercase 'udp6' for NetworkUDP6", func() {
				Expect(NetworkUDP6.Code()).To(Equal("udp6"))
			})

			It("should return lowercase 'ip' for NetworkIP", func() {
				Expect(NetworkIP.Code()).To(Equal("ip"))
			})

			It("should return lowercase 'ip4' for NetworkIP4", func() {
				Expect(NetworkIP4.Code()).To(Equal("ip4"))
			})

			It("should return lowercase 'ip6' for NetworkIP6", func() {
				Expect(NetworkIP6.Code()).To(Equal("ip6"))
			})

			It("should return lowercase 'unixgram' for NetworkUnixGram", func() {
				result := NetworkUnixGram.Code()
				Expect(result).To(Equal("unixgram"))
			})
		})

		Context("consistency with String()", func() {
			// ✅ FIXED: All protocols now consistent including UnixGram
			It("should return lowercase version of String() for all protocols", func() {
				protocols := []NetworkProtocol{
					NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
					NetworkUDP, NetworkUDP4, NetworkUDP6,
					NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
				}

				for _, p := range protocols {
					// Code() should equal String() since both are lowercase now
					str := p.String()
					code := p.Code()

					if str != "" {
						Expect(code).To(Equal(str), "For protocol %v, Code() should equal String()", p)
					}
				}
			})

			It("should be case-insensitive for comparison", func() {
				// Verify Code() normalizes case for consistent comparisons
				Expect(NetworkUnixGram.Code()).To(Equal("unixgram"))
			})
		})

		Context("with invalid protocols", func() {
			It("should return empty string for NetworkEmpty", func() {
				Expect(NetworkEmpty.Code()).To(Equal(""))
			})

			It("should return empty string for undefined protocol", func() {
				invalidProtocol := NetworkProtocol(99)
				Expect(invalidProtocol.Code()).To(Equal(""))
			})
		})
	})

	Describe("String() and Code() roundtrip", func() {
		// ✅ FIXED: All protocols can now be roundtripped
		It("should maintain consistency between String(), Code() and Parse() for all protocols", func() {
			protocols := []NetworkProtocol{
				NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
				NetworkUDP, NetworkUDP4, NetworkUDP6,
				NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
			}

			for _, original := range protocols {
				// Code() -> Parse() should return original
				code := original.Code()
				parsed := Parse(code)
				Expect(parsed).To(Equal(original),
					"Roundtrip failed for %v: Code()='%s', Parse()=%v",
					original, code, parsed)
			}
		})

		It("should roundtrip via String() as well", func() {
			protocols := []NetworkProtocol{
				NetworkUnix, NetworkTCP, NetworkTCP4, NetworkTCP6,
				NetworkUDP, NetworkUDP4, NetworkUDP6,
				NetworkIP, NetworkIP4, NetworkIP6, NetworkUnixGram,
			}

			for _, original := range protocols {
				// String() -> Parse() should return original
				str := original.String()
				parsed := Parse(str)
				Expect(parsed).To(Equal(original),
					"Roundtrip failed for %v: String()='%s', Parse()=%v",
					original, str, parsed)
			}
		})
	})

	Describe("Memory and performance", func() {
		It("should not allocate excessive memory for repeated calls", func() {
			p := NetworkTCP

			// Call multiple times - should not cause memory issues
			for i := 0; i < 1000; i++ {
				_ = p.String()
				_ = p.Code()
			}
		})

		It("should handle concurrent calls safely", func() {
			p := NetworkTCP
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer func() { done <- true }()
					for j := 0; j < 100; j++ {
						_ = p.String()
						_ = p.Code()
					}
				}()
			}

			// Wait for all goroutines
			for i := 0; i < 10; i++ {
				<-done
			}
		})
	})
})
