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

package status_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/monitor/status"
)

var _ = Describe("Parse Functions", func() {
	Describe("Parse", func() {
		Context("with valid status strings", func() {
			It("should parse 'OK' to OK status", func() {
				s := status.Parse("OK")
				Expect(s).To(Equal(status.OK))
			})

			It("should parse 'ok' to OK status", func() {
				s := status.Parse("ok")
				Expect(s).To(Equal(status.OK))
			})

			It("should parse 'Warn' to Warn status", func() {
				s := status.Parse("Warn")
				Expect(s).To(Equal(status.Warn))
			})

			It("should parse 'warn' to Warn status", func() {
				s := status.Parse("warn")
				Expect(s).To(Equal(status.Warn))
			})

			It("should parse 'KO' to KO status", func() {
				s := status.Parse("KO")
				Expect(s).To(Equal(status.KO))
			})

			It("should parse 'ko' to KO status", func() {
				s := status.Parse("ko")
				Expect(s).To(Equal(status.KO))
			})
		})

		Context("with whitespace and quotes", func() {
			It("should handle leading/trailing spaces", func() {
				Expect(status.Parse(" OK ")).To(Equal(status.OK))
				Expect(status.Parse("  Warn  ")).To(Equal(status.Warn))
			})

			It("should handle double quotes", func() {
				Expect(status.Parse(`"OK"`)).To(Equal(status.OK))
				Expect(status.Parse(`"Warn"`)).To(Equal(status.Warn))
			})

			It("should handle single quotes", func() {
				Expect(status.Parse("'OK'")).To(Equal(status.OK))
				Expect(status.Parse("'Warn'")).To(Equal(status.Warn))
			})

			It("should handle 'status' prefix", func() {
				Expect(status.Parse("statusOK")).To(Equal(status.OK))
				Expect(status.Parse("status Warn")).To(Equal(status.Warn))
			})
		})

		Context("with invalid strings", func() {
			It("should return KO for empty string", func() {
				Expect(status.Parse("")).To(Equal(status.KO))
			})

			It("should return KO for unknown string", func() {
				Expect(status.Parse("unknown")).To(Equal(status.KO))
			})
		})
	})

	Describe("ParseByte", func() {
		It("should parse byte slice to OK", func() {
			Expect(status.ParseByte([]byte("OK"))).To(Equal(status.OK))
		})

		It("should parse byte slice to Warn", func() {
			Expect(status.ParseByte([]byte("Warn"))).To(Equal(status.Warn))
		})

		It("should parse byte slice to KO", func() {
			Expect(status.ParseByte([]byte("KO"))).To(Equal(status.KO))
		})

		It("should handle empty byte slice", func() {
			Expect(status.ParseByte([]byte{})).To(Equal(status.KO))
		})
	})

	Describe("ParseUint", func() {
		It("should parse 0 to KO", func() {
			Expect(status.ParseUint(0)).To(Equal(status.KO))
		})

		It("should parse 1 to Warn", func() {
			Expect(status.ParseUint(1)).To(Equal(status.Warn))
		})

		It("should parse 2 to OK", func() {
			Expect(status.ParseUint(2)).To(Equal(status.OK))
		})

		It("should parse 3 to KO (invalid)", func() {
			Expect(status.ParseUint(3)).To(Equal(status.KO))
		})

		It("should parse large values to KO", func() {
			Expect(status.ParseUint(256)).To(Equal(status.KO))
			Expect(status.ParseUint(1000)).To(Equal(status.KO))
		})
	})

	Describe("ParseUint8", func() {
		It("should parse 0 to KO", func() {
			Expect(status.ParseUint8(0)).To(Equal(status.KO))
		})

		It("should parse 1 to Warn", func() {
			Expect(status.ParseUint8(1)).To(Equal(status.Warn))
		})

		It("should parse 2 to OK", func() {
			Expect(status.ParseUint8(2)).To(Equal(status.OK))
		})

		It("should parse 3 to KO", func() {
			Expect(status.ParseUint8(3)).To(Equal(status.KO))
		})

		It("should parse 255 to KO", func() {
			Expect(status.ParseUint8(255)).To(Equal(status.KO))
		})
	})

	Describe("ParseUint64", func() {
		It("should parse 0 to KO", func() {
			Expect(status.ParseUint64(0)).To(Equal(status.KO))
		})

		It("should parse 1 to Warn", func() {
			Expect(status.ParseUint64(1)).To(Equal(status.Warn))
		})

		It("should parse 2 to OK", func() {
			Expect(status.ParseUint64(2)).To(Equal(status.OK))
		})

		It("should parse values > MaxUint8 to KO", func() {
			Expect(status.ParseUint64(256)).To(Equal(status.KO))
			Expect(status.ParseUint64(math.MaxUint64)).To(Equal(status.KO))
		})
	})

	Describe("ParseInt", func() {
		It("should parse 0 to KO", func() {
			Expect(status.ParseInt(0)).To(Equal(status.KO))
		})

		It("should parse 1 to Warn", func() {
			Expect(status.ParseInt(1)).To(Equal(status.Warn))
		})

		It("should parse 2 to OK", func() {
			Expect(status.ParseInt(2)).To(Equal(status.OK))
		})

		It("should parse negative values to KO", func() {
			Expect(status.ParseInt(-1)).To(Equal(status.KO))
			Expect(status.ParseInt(-100)).To(Equal(status.KO))
		})

		It("should parse large values to KO", func() {
			Expect(status.ParseInt(256)).To(Equal(status.KO))
			Expect(status.ParseInt(1000)).To(Equal(status.KO))
		})
	})

	Describe("ParseInt64", func() {
		It("should parse 0 to KO", func() {
			Expect(status.ParseInt64(0)).To(Equal(status.KO))
		})

		It("should parse 1 to Warn", func() {
			Expect(status.ParseInt64(1)).To(Equal(status.Warn))
		})

		It("should parse 2 to OK", func() {
			Expect(status.ParseInt64(2)).To(Equal(status.OK))
		})

		It("should parse negative values to KO", func() {
			Expect(status.ParseInt64(-1)).To(Equal(status.KO))
			Expect(status.ParseInt64(math.MinInt64)).To(Equal(status.KO))
		})

		It("should parse values > MaxUint8 to KO", func() {
			Expect(status.ParseInt64(256)).To(Equal(status.KO))
			Expect(status.ParseInt64(math.MaxInt64)).To(Equal(status.KO))
		})
	})

	Describe("ParseFloat64", func() {
		It("should parse 0.0 to KO", func() {
			Expect(status.ParseFloat64(0.0)).To(Equal(status.KO))
		})

		It("should parse 1.0 to Warn", func() {
			Expect(status.ParseFloat64(1.0)).To(Equal(status.Warn))
		})

		It("should parse 2.0 to OK", func() {
			Expect(status.ParseFloat64(2.0)).To(Equal(status.OK))
		})

		It("should floor fractional values", func() {
			Expect(status.ParseFloat64(1.9)).To(Equal(status.Warn))
			Expect(status.ParseFloat64(2.1)).To(Equal(status.OK))
			Expect(status.ParseFloat64(2.9)).To(Equal(status.OK))
		})

		It("should parse negative values to KO", func() {
			Expect(status.ParseFloat64(-1.0)).To(Equal(status.KO))
			Expect(status.ParseFloat64(-100.5)).To(Equal(status.KO))
		})

		It("should parse values > MaxUint8 to KO", func() {
			Expect(status.ParseFloat64(256.0)).To(Equal(status.KO))
			Expect(status.ParseFloat64(1000.5)).To(Equal(status.KO))
		})

		It("should handle edge cases", func() {
			Expect(status.ParseFloat64(0.1)).To(Equal(status.KO))
			Expect(status.ParseFloat64(0.9)).To(Equal(status.KO))
			Expect(status.ParseFloat64(3.0)).To(Equal(status.KO))
		})
	})
})
