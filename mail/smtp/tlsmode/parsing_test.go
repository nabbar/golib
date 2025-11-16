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
	"math"

	. "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS Mode Parsing", func() {

	Describe("Parse from String", func() {
		Context("with valid string values", func() {
			It("should parse 'starttls' to TLSStartTLS", func() {
				Expect(Parse("starttls")).To(Equal(TLSStartTLS))
			})

			It("should parse 'STARTTLS' to TLSStartTLS (case insensitive)", func() {
				Expect(Parse("STARTTLS")).To(Equal(TLSStartTLS))
			})

			It("should parse 'StartTLS' to TLSStartTLS", func() {
				Expect(Parse("StartTLS")).To(Equal(TLSStartTLS))
			})

			It("should parse 'tls' to TLSStrictTLS", func() {
				Expect(Parse("tls")).To(Equal(TLSStrictTLS))
			})

			It("should parse 'TLS' to TLSStrictTLS", func() {
				Expect(Parse("TLS")).To(Equal(TLSStrictTLS))
			})

			It("should parse empty string to TLSNone", func() {
				Expect(Parse("")).To(Equal(TLSNone))
			})

			It("should parse 'none' to TLSNone", func() {
				Expect(Parse("none")).To(Equal(TLSNone))
			})
		})

		Context("with whitespace and special characters", func() {
			It("should handle leading/trailing whitespace", func() {
				Expect(Parse("  starttls  ")).To(Equal(TLSStartTLS))
				Expect(Parse("\tstarttls\t")).To(Equal(TLSStartTLS))
			})

			It("should handle newlines", func() {
				Expect(Parse("starttls\n")).To(Equal(TLSStartTLS))
				Expect(Parse("\nstarttls")).To(Equal(TLSStartTLS))
			})

			It("should handle carriage returns", func() {
				Expect(Parse("starttls\r")).To(Equal(TLSStartTLS))
				Expect(Parse("\rstarttls")).To(Equal(TLSStartTLS))
			})

			It("should handle quotes", func() {
				Expect(Parse("\"starttls\"")).To(Equal(TLSStartTLS))
				Expect(Parse("'starttls'")).To(Equal(TLSStartTLS))
			})

			It("should handle underscores and hyphens", func() {
				Expect(Parse("start_tls")).To(Equal(TLSStartTLS))
				Expect(Parse("start-tls")).To(Equal(TLSStartTLS))
			})

			It("should handle spaces in string", func() {
				Expect(Parse("start tls")).To(Equal(TLSStartTLS))
			})
		})

		Context("with invalid or unknown values", func() {
			It("should return TLSNone for unknown strings", func() {
				Expect(Parse("invalid")).To(Equal(TLSNone))
				Expect(Parse("unknown")).To(Equal(TLSNone))
				Expect(Parse("random")).To(Equal(TLSNone))
			})
		})
	})

	Describe("ParseBytes", func() {
		It("should parse byte slices correctly", func() {
			Expect(ParseBytes([]byte("starttls"))).To(Equal(TLSStartTLS))
			Expect(ParseBytes([]byte("tls"))).To(Equal(TLSStrictTLS))
			Expect(ParseBytes([]byte(""))).To(Equal(TLSNone))
		})

		It("should handle whitespace in byte slices", func() {
			Expect(ParseBytes([]byte("  starttls  "))).To(Equal(TLSStartTLS))
		})

		It("should handle special characters in byte slices", func() {
			Expect(ParseBytes([]byte("\"tls\""))).To(Equal(TLSStrictTLS))
		})
	})

	Describe("ParseInt64", func() {
		Context("with valid integer values", func() {
			It("should parse 0 to TLSNone", func() {
				Expect(ParseInt64(0)).To(Equal(TLSNone))
			})

			It("should parse 1 to TLSStartTLS", func() {
				Expect(ParseInt64(1)).To(Equal(TLSStartTLS))
			})

			It("should parse 2 to TLSStrictTLS", func() {
				Expect(ParseInt64(2)).To(Equal(TLSStrictTLS))
			})
		})

		Context("with edge case values", func() {
			It("should handle negative values as TLSNone", func() {
				Expect(ParseInt64(-1)).To(Equal(TLSNone))
				Expect(ParseInt64(-100)).To(Equal(TLSNone))
			})

			It("should handle values > MaxUint8 as TLSNone", func() {
				Expect(ParseInt64(256)).To(Equal(TLSNone))
				Expect(ParseInt64(1000)).To(Equal(TLSNone))
			})

			It("should handle invalid uint8 values as TLSNone", func() {
				Expect(ParseInt64(3)).To(Equal(TLSNone))
				Expect(ParseInt64(100)).To(Equal(TLSNone))
			})
		})
	})

	Describe("ParseUint64", func() {
		Context("with valid unsigned integer values", func() {
			It("should parse 0 to TLSNone", func() {
				Expect(ParseUint64(0)).To(Equal(TLSNone))
			})

			It("should parse 1 to TLSStartTLS", func() {
				Expect(ParseUint64(1)).To(Equal(TLSStartTLS))
			})

			It("should parse 2 to TLSStrictTLS", func() {
				Expect(ParseUint64(2)).To(Equal(TLSStrictTLS))
			})
		})

		Context("with edge case values", func() {
			It("should handle values > MaxUint8 as TLSNone", func() {
				Expect(ParseUint64(256)).To(Equal(TLSNone))
				Expect(ParseUint64(1000)).To(Equal(TLSNone))
				Expect(ParseUint64(math.MaxUint64)).To(Equal(TLSNone))
			})

			It("should handle invalid uint8 values as TLSNone", func() {
				Expect(ParseUint64(3)).To(Equal(TLSNone))
				Expect(ParseUint64(255)).To(Equal(TLSNone))
			})
		})
	})

	Describe("ParseFloat64", func() {
		Context("with valid float values", func() {
			It("should parse 0.0 to TLSNone", func() {
				Expect(ParseFloat64(0.0)).To(Equal(TLSNone))
			})

			It("should parse 1.0 to TLSStartTLS", func() {
				Expect(ParseFloat64(1.0)).To(Equal(TLSStartTLS))
			})

			It("should parse 2.0 to TLSStrictTLS", func() {
				Expect(ParseFloat64(2.0)).To(Equal(TLSStrictTLS))
			})

			It("should floor fractional values", func() {
				Expect(ParseFloat64(1.5)).To(Equal(TLSStartTLS))
				Expect(ParseFloat64(1.9)).To(Equal(TLSStartTLS))
				Expect(ParseFloat64(2.1)).To(Equal(TLSStrictTLS))
			})
		})

		Context("with edge case float values", func() {
			It("should handle large values as TLSNone", func() {
				Expect(ParseFloat64(256.0)).To(Equal(TLSNone))
				Expect(ParseFloat64(1000.0)).To(Equal(TLSNone))
				Expect(ParseFloat64(math.MaxFloat64)).To(Equal(TLSNone))
			})

			It("should handle invalid values as TLSNone", func() {
				Expect(ParseFloat64(3.0)).To(Equal(TLSNone))
				Expect(ParseFloat64(100.5)).To(Equal(TLSNone))
			})

			It("should handle floats that floor to valid values", func() {
				Expect(ParseFloat64(0.1)).To(Equal(TLSNone))
				Expect(ParseFloat64(0.9)).To(Equal(TLSNone))
				Expect(ParseFloat64(1.1)).To(Equal(TLSStartTLS))
				Expect(ParseFloat64(2.9)).To(Equal(TLSStrictTLS))
			})

			It("should handle float edge at MaxUint8 boundary", func() {
				Expect(ParseFloat64(255.0)).To(Equal(TLSNone))
				Expect(ParseFloat64(255.9)).To(Equal(TLSNone))
			})

			It("should handle edge case where floor exceeds MaxUint8", func() {
				// This tests the second condition in ParseFloat64
				// where p := math.Floor(f); p > math.MaxUint8
				Expect(ParseFloat64(math.MaxUint8 + 0.5)).To(Equal(TLSNone))
			})
		})
	})

	Describe("Deprecated Functions", func() {
		It("TLSModeFromString should work like Parse", func() {
			Expect(TLSModeFromString("starttls")).To(Equal(Parse("starttls")))
			Expect(TLSModeFromString("tls")).To(Equal(Parse("tls")))
			Expect(TLSModeFromString("")).To(Equal(Parse("")))
		})

		It("TLSModeFromInt should work like ParseInt64", func() {
			Expect(TLSModeFromInt(0)).To(Equal(ParseInt64(0)))
			Expect(TLSModeFromInt(1)).To(Equal(ParseInt64(1)))
			Expect(TLSModeFromInt(2)).To(Equal(ParseInt64(2)))
			Expect(TLSModeFromInt(-1)).To(Equal(ParseInt64(-1)))
		})
	})
})
