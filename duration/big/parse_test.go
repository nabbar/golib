/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package big_test

import (
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Parsing Operations", func() {
	Describe("Parse", func() {
		It("should parse simple seconds", func() {
			d, err := durbig.Parse("30s")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(30)))
		})

		It("should parse simple minutes", func() {
			d, err := durbig.Parse("5m")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5 * 60)))
		})

		It("should parse simple hours", func() {
			d, err := durbig.Parse("2h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(2 * 3600)))
		})

		It("should parse simple days", func() {
			d, err := durbig.Parse("3d")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(3 * 24 * 3600)))
		})

		It("should parse complex duration", func() {
			d, err := durbig.Parse("5d23h15m13s")
			Expect(err).ToNot(HaveOccurred())
			expected := int64(5*24*3600 + 23*3600 + 15*60 + 13)
			Expect(d.Int64()).To(Equal(expected))
		})

		It("should parse fractional seconds", func() {
			d, err := durbig.Parse("1.5s")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(1))) // Truncates to seconds
		})

		It("should parse fractional minutes", func() {
			d, err := durbig.Parse("2.5m")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(150))) // 2.5 * 60 = 150
		})

		It("should parse fractional hours", func() {
			d, err := durbig.Parse("1.5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5400))) // 1.5 * 3600 = 5400
		})

		It("should parse negative duration", func() {
			d, err := durbig.Parse("-5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(-5 * 3600)))
		})

		It("should parse zero duration", func() {
			d, err := durbig.Parse("0s")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(0)))
		})

		It("should parse single zero", func() {
			d, err := durbig.Parse("0")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(0)))
		})

		It("should return error for empty string", func() {
			_, err := durbig.Parse("")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for invalid format", func() {
			_, err := durbig.Parse("invalid")
			Expect(err).To(HaveOccurred())
		})

		It("should handle plus sign prefix", func() {
			d, err := durbig.Parse("+5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5 * 3600)))
		})

		It("should handle duration with spaces", func() {
			d, err := durbig.Parse("5h 30m")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5*3600 + 30*60)))
		})

		It("should handle quoted duration strings", func() {
			d, err := durbig.Parse("\"5h30m\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5*3600 + 30*60)))
		})

		It("should return error for just sign", func() {
			_, err := durbig.Parse("-")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for just plus sign", func() {
			_, err := durbig.Parse("+")
			Expect(err).To(HaveOccurred())
		})

		It("should handle very large duration", func() {
			d, err := durbig.Parse("100000d")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(100000 * 24 * 3600)))
		})

		It("should handle multiple components", func() {
			d, err := durbig.Parse("1d2h3m4s")
			Expect(err).ToNot(HaveOccurred())
			expected := int64(1*24*3600 + 2*3600 + 3*60 + 4)
			Expect(d.Int64()).To(Equal(expected))
		})
	})

	Describe("ParseByte", func() {
		It("should parse valid byte array", func() {
			d, err := durbig.ParseByte([]byte("3h45m"))
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(3*3600 + 45*60)))
		})

		It("should return error for invalid byte array", func() {
			_, err := durbig.ParseByte([]byte("invalid"))
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty byte array", func() {
			_, err := durbig.ParseByte([]byte(""))
			Expect(err).To(HaveOccurred())
		})

		It("should parse duration with days", func() {
			d, err := durbig.ParseByte([]byte("7d"))
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(7 * 24 * 3600)))
		})
	})

	Describe("Constructor Functions", func() {
		Describe("Seconds", func() {
			It("should create duration from positive seconds", func() {
				d := durbig.Seconds(30)
				Expect(d.Int64()).To(Equal(int64(30)))
			})

			It("should create duration from negative seconds", func() {
				d := durbig.Seconds(-30)
				Expect(d.Int64()).To(Equal(int64(-30)))
			})

			It("should create zero duration", func() {
				d := durbig.Seconds(0)
				Expect(d.Int64()).To(Equal(int64(0)))
			})

			It("should handle large values", func() {
				d := durbig.Seconds(1000000)
				Expect(d.Int64()).To(Equal(int64(1000000)))
			})
		})

		Describe("Minutes", func() {
			It("should create duration from minutes", func() {
				d := durbig.Minutes(5)
				Expect(d.Int64()).To(Equal(int64(5 * 60)))
			})

			It("should create duration from negative minutes", func() {
				d := durbig.Minutes(-5)
				Expect(d.Int64()).To(Equal(int64(-5 * 60)))
			})

			It("should create zero duration", func() {
				d := durbig.Minutes(0)
				Expect(d.Int64()).To(Equal(int64(0)))
			})
		})

		Describe("Hours", func() {
			It("should create duration from hours", func() {
				d := durbig.Hours(3)
				Expect(d.Int64()).To(Equal(int64(3 * 3600)))
			})

			It("should create duration from negative hours", func() {
				d := durbig.Hours(-3)
				Expect(d.Int64()).To(Equal(int64(-3 * 3600)))
			})

			It("should create zero duration", func() {
				d := durbig.Hours(0)
				Expect(d.Int64()).To(Equal(int64(0)))
			})
		})

		Describe("Days", func() {
			It("should create duration from days", func() {
				d := durbig.Days(7)
				Expect(d.Int64()).To(Equal(int64(7 * 24 * 3600)))
			})

			It("should create duration from negative days", func() {
				d := durbig.Days(-7)
				Expect(d.Int64()).To(Equal(int64(-7 * 24 * 3600)))
			})

			It("should create zero duration", func() {
				d := durbig.Days(0)
				Expect(d.Int64()).To(Equal(int64(0)))
			})

			It("should handle large number of days", func() {
				d := durbig.Days(365)
				Expect(d.Int64()).To(Equal(int64(365 * 24 * 3600)))
			})
		})
	})

	Describe("ParseDuration", func() {
		It("should parse time.Duration", func() {
			td := 5*time.Hour + 30*time.Minute
			d := durbig.ParseDuration(td)
			Expect(d.Int64()).To(Equal(int64(5*3600 + 30*60)))
		})

		It("should parse zero time.Duration", func() {
			td := time.Duration(0)
			d := durbig.ParseDuration(td)
			Expect(d.Int64()).To(Equal(int64(0)))
		})

		It("should parse negative time.Duration", func() {
			td := -5 * time.Hour
			d := durbig.ParseDuration(td)
			Expect(d.Int64()).To(Equal(int64(-5 * 3600)))
		})

		It("should handle subsecond precision", func() {
			td := 500 * time.Millisecond
			d := durbig.ParseDuration(td)
			Expect(d.Int64()).To(Equal(int64(0))) // Truncates to seconds
		})

		It("should handle very small durations", func() {
			td := 100 * time.Nanosecond
			d := durbig.ParseDuration(td)
			Expect(d.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("ParseFloat64", func() {
		It("should parse positive float", func() {
			d := durbig.ParseFloat64(30.5)
			Expect(d.Int64()).To(Equal(int64(31))) // Rounds to nearest
		})

		It("should parse zero", func() {
			d := durbig.ParseFloat64(0)
			Expect(d.Int64()).To(Equal(int64(0)))
		})

		It("should parse negative float", func() {
			d := durbig.ParseFloat64(-30.5)
			Expect(d.Int64()).To(Equal(int64(-31))) // Rounds to nearest even (banker's rounding)
		})

		It("should handle very large positive values", func() {
			d := durbig.ParseFloat64(math.MaxFloat64)
			Expect(d.Int64()).To(Equal(int64(math.MaxInt64)))
		})

		It("should handle very large negative values", func() {
			d := durbig.ParseFloat64(-math.MaxFloat64)
			Expect(d.Int64()).To(Equal(int64(-math.MaxInt64)))
		})

		It("should handle fractional seconds", func() {
			d := durbig.ParseFloat64(5.9)
			Expect(d.Int64()).To(Equal(int64(6))) // Rounds to nearest
		})

		It("should handle values just below MaxInt64", func() {
			// Float64 loses precision at very large values
			d := durbig.ParseFloat64(1e18) // 1 quintillion
			Expect(d.Int64()).To(BeNumerically(">", int64(0)))
		})

		It("should handle values just above MinInt64", func() {
			d := durbig.ParseFloat64(float64(math.MinInt64 + 100))
			Expect(d.Int64()).To(BeNumerically("~", int64(math.MinInt64+100), 100))
		})
	})

	Describe("Edge Cases and Boundaries", func() {
		It("should handle MaxInt64 seconds", func() {
			d := durbig.Seconds(math.MaxInt64)
			Expect(d.Int64()).To(Equal(int64(math.MaxInt64)))
		})

		It("should handle MinInt64 seconds", func() {
			d := durbig.Seconds(math.MinInt64)
			Expect(d.Int64()).To(Equal(int64(math.MinInt64)))
		})

		It("should parse duration with all components", func() {
			d, err := durbig.Parse("365d23h59m59s")
			Expect(err).ToNot(HaveOccurred())
			expected := int64(365*24*3600 + 23*3600 + 59*60 + 59)
			Expect(d.Int64()).To(Equal(expected))
		})

		It("should handle duration string with multiple spaces", func() {
			d, err := durbig.Parse("  5h  30m  ")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Int64()).To(Equal(int64(5*3600 + 30*60)))
		})
	})
})
