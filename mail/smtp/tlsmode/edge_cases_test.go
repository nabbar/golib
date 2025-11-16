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

package tlsmode_test

import (
	"encoding/json"
	"math"
	"strings"
	"sync"

	. "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS Mode Edge Cases and Error Handling", func() {

	Describe("Boundary Value Testing", func() {
		Context("Integer boundaries", func() {
			It("should handle MaxInt64", func() {
				Expect(ParseInt64(math.MaxInt64)).To(Equal(TLSNone))
			})

			It("should handle MinInt64", func() {
				Expect(ParseInt64(math.MinInt64)).To(Equal(TLSNone))
			})

			It("should handle MaxUint64", func() {
				Expect(ParseUint64(math.MaxUint64)).To(Equal(TLSNone))
			})

			It("should handle values at MaxUint8", func() {
				Expect(ParseUint64(255)).To(Equal(TLSNone))
			})

			It("should handle values at MaxUint8 + 1", func() {
				Expect(ParseUint64(256)).To(Equal(TLSNone))
			})
		})

		Context("Float boundaries", func() {
			It("should handle MaxFloat64", func() {
				Expect(ParseFloat64(math.MaxFloat64)).To(Equal(TLSNone))
			})

			It("should handle negative floats", func() {
				Expect(ParseFloat64(-1.0)).To(Equal(TLSNone))
			})

			It("should handle NaN", func() {
				result := ParseFloat64(math.NaN())
				Expect(result).To(Equal(TLSNone))
			})

			It("should handle Inf", func() {
				Expect(ParseFloat64(math.Inf(1))).To(Equal(TLSNone))
				Expect(ParseFloat64(math.Inf(-1))).To(Equal(TLSNone))
			})

			It("should handle very small positive floats", func() {
				Expect(ParseFloat64(0.1)).To(Equal(TLSNone))
				Expect(ParseFloat64(0.9)).To(Equal(TLSNone))
			})
		})
	})

	Describe("String Edge Cases", func() {
		Context("Unicode and special characters", func() {
			It("should handle Unicode characters", func() {
				Expect(Parse("starttls™")).To(Equal(TLSNone))
				Expect(Parse("tls™")).To(Equal(TLSNone))
			})

			It("should handle mixed case with special chars", func() {
				Expect(Parse("StArTtLs")).To(Equal(TLSStartTLS))
			})

			It("should handle tabs", func() {
				Expect(Parse("\tstarttls\t")).To(Equal(TLSStartTLS))
			})

			It("should handle multiple types of whitespace", func() {
				Expect(Parse("  \t\n\rstarttls\r\n\t  ")).To(Equal(TLSStartTLS))
			})
		})

		Context("Very long strings", func() {
			It("should handle very long strings", func() {
				longString := strings.Repeat("a", 10000) + "starttls"
				Expect(Parse(longString)).To(Equal(TLSNone))
			})

			It("should handle empty-like strings", func() {
				Expect(Parse("     ")).To(Equal(TLSNone))
				Expect(Parse("\n\r\t")).To(Equal(TLSNone))
			})
		})

		Context("Similar but invalid strings", func() {
			It("should reject close matches", func() {
				Expect(Parse("start")).To(Equal(TLSNone))
				Expect(Parse("startls")).To(Equal(TLSNone))
				Expect(Parse("starttl")).To(Equal(TLSNone))
				Expect(Parse("tl")).To(Equal(TLSNone))
				Expect(Parse("tlss")).To(Equal(TLSNone))
			})
		})
	})

	Describe("Concurrent Access", func() {
		It("should handle concurrent Parse calls", func() {
			var wg sync.WaitGroup
			iterations := 100
			wg.Add(iterations * 3)

			// Test concurrent parsing
			for i := 0; i < iterations; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					Expect(Parse("starttls")).To(Equal(TLSStartTLS))
				}()

				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					Expect(Parse("tls")).To(Equal(TLSStrictTLS))
				}()

				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					Expect(Parse("")).To(Equal(TLSNone))
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent format conversions", func() {
			var wg sync.WaitGroup
			iterations := 100
			wg.Add(iterations * 3)

			modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}

			for i := 0; i < iterations; i++ {
				for _, mode := range modes {
					m := mode
					go func() {
						defer GinkgoRecover()
						defer wg.Done()
						_ = m.String()
						_ = m.Int()
						_ = m.Uint64()
						_ = m.Float64()
					}()
				}
			}

			wg.Wait()
		})

		It("should handle concurrent encoding/decoding", func() {
			var wg sync.WaitGroup
			iterations := 50
			wg.Add(iterations)

			for i := 0; i < iterations; i++ {
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					mode := TLSStartTLS
					data, err := json.Marshal(mode)
					Expect(err).ToNot(HaveOccurred())

					var decoded TLSMode
					err = json.Unmarshal(data, &decoded)
					Expect(err).ToNot(HaveOccurred())
					Expect(decoded).To(Equal(mode))
				}()
			}

			wg.Wait()
		})
	})

	Describe("Memory and Resource Tests", func() {
		It("should handle many parse operations", func() {
			for i := 0; i < 10000; i++ {
				Parse("starttls")
				Parse("tls")
				Parse("")
			}
		})

		It("should handle many conversions", func() {
			mode := TLSStartTLS
			for i := 0; i < 10000; i++ {
				_ = mode.String()
				_ = mode.Int64()
				_ = mode.Float64()
			}
		})

		It("should handle many encoding operations", func() {
			mode := TLSStrictTLS
			for i := 0; i < 1000; i++ {
				data, _ := json.Marshal(mode)
				var decoded TLSMode
				_ = json.Unmarshal(data, &decoded)
			}
		})
	})

	Describe("Type Safety", func() {
		It("should maintain type safety across conversions", func() {
			original := TLSStartTLS

			// Through various conversions
			asInt := original.Int64()
			fromInt := ParseInt64(asInt)
			Expect(fromInt).To(Equal(original))

			asStr := original.String()
			fromStr := Parse(asStr)
			Expect(fromStr).To(Equal(original))

			asFloat := original.Float64()
			fromFloat := ParseFloat64(asFloat)
			Expect(fromFloat).To(Equal(original))
		})

		It("should handle zero values correctly", func() {
			var mode TLSMode
			Expect(mode).To(Equal(TLSNone))
			Expect(mode.String()).To(Equal(""))
			Expect(mode.Int()).To(Equal(0))
		})
	})

	Describe("Error Propagation in Encoding", func() {
		Context("TOML Unmarshal errors", func() {
			It("should return error for unsupported types", func() {
				var mode TLSMode
				err := mode.UnmarshalTOML(make(chan int))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("CBOR Unmarshal errors", func() {
			It("should return error for invalid CBOR data", func() {
				var mode TLSMode
				err := mode.UnmarshalCBOR([]byte{0xFF, 0xFF, 0xFF})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("JSON Unmarshal errors", func() {
			It("should return error for invalid JSON types", func() {
				var mode TLSMode
				err := json.Unmarshal([]byte(`{"invalid":"object"}`), &mode)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for malformed JSON", func() {
				var mode TLSMode
				err := json.Unmarshal([]byte(`{invalid`), &mode)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Consistency Across Operations", func() {
		It("should maintain consistency for all valid modes", func() {
			modes := []TLSMode{TLSNone, TLSStartTLS, TLSStrictTLS}

			for _, mode := range modes {
				// String roundtrip
				parsed := Parse(mode.String())
				Expect(parsed).To(Equal(mode))

				// Int64 roundtrip
				parsed = ParseInt64(mode.Int64())
				Expect(parsed).To(Equal(mode))

				// Uint64 roundtrip
				parsed = ParseUint64(mode.Uint64())
				Expect(parsed).To(Equal(mode))

				// Float64 roundtrip
				parsed = ParseFloat64(mode.Float64())
				Expect(parsed).To(Equal(mode))

				// Bytes roundtrip
				parsed = ParseBytes([]byte(mode.String()))
				Expect(parsed).To(Equal(mode))
			}
		})

		It("should maintain consistency across all conversion methods", func() {
			mode := TLSStartTLS

			Expect(int(mode.Uint())).To(Equal(mode.Int()))
			Expect(int64(mode.Uint64())).To(Equal(mode.Int64()))
			Expect(int32(mode.Uint32())).To(Equal(mode.Int32()))
			Expect(int(mode.Float32())).To(Equal(mode.Int()))
			Expect(int(mode.Float64())).To(Equal(mode.Int()))
		})
	})

	Describe("Invalid TLSMode Values", func() {
		It("should handle casting invalid uint8 values", func() {
			// Directly cast invalid values
			invalidMode := TLSMode(99)

			// Should still be able to call methods
			str := invalidMode.String()
			Expect(str).To(Equal("")) // Unknown values map to TLSNone string

			// Conversions should work
			Expect(invalidMode.Int()).To(Equal(99))
			Expect(invalidMode.Uint()).To(Equal(uint8(99)))
		})
	})
})
