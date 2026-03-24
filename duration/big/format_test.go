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
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Formatting Operations", func() {
	Describe("String", func() {
		It("should format zero duration", func() {
			// TC-FO-001
			d := durbig.Seconds(0)
			Expect(d.String()).To(Equal("0s"))
		})

		It("should format seconds only", func() {
			// TC-FO-002
			d := durbig.Seconds(45)
			Expect(d.String()).To(Equal("45s"))
		})

		It("should format minutes and seconds", func() {
			// TC-FO-003
			d := durbig.Minutes(5) + durbig.Seconds(30)
			Expect(d.String()).To(Equal("5m30s"))
		})

		It("should format hours, minutes, and seconds", func() {
			// TC-FO-004
			d := durbig.Hours(2) + durbig.Minutes(30) + durbig.Seconds(45)
			Expect(d.String()).To(Equal("2h30m45s"))
		})

		It("should format days", func() {
			// TC-FO-005
			d := durbig.Days(5)
			Expect(d.String()).To(Equal("5d"))
		})

		It("should format days with hours", func() {
			// TC-FO-006
			d := durbig.Days(5) + durbig.Hours(12)
			Expect(d.String()).To(Equal("5d12h"))
		})

		It("should format complete duration", func() {
			// TC-FO-007
			d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
			Expect(d.String()).To(Equal("5d23h15m13s"))
		})

		It("should format negative duration", func() {
			// TC-FO-008
			d := durbig.Seconds(-30)
			Expect(d.String()).To(ContainSubstring("-"))
		})

		It("should format negative days", func() {
			// TC-FO-009
			d := durbig.Days(-5)
			Expect(d.String()).To(ContainSubstring("-"))
		})

		It("should format very large duration", func() {
			// TC-FO-010
			d := durbig.Days(10000)
			Expect(d.String()).To(Equal("10000d"))
		})

		It("should format hours without days when less than 24h", func() {
			// TC-FO-036
			d := durbig.Hours(23)
			Expect(d.String()).To(Equal("23h"))
		})

		It("should omit zero components at start", func() {
			// TC-FO-037
			d := durbig.Minutes(5)
			Expect(d.String()).To(Equal("5m"))
		})
	})

	Describe("Time", func() {
		It("should convert to time.Duration for small values", func() {
			// TC-FO-013
			d := durbig.Hours(5) + durbig.Minutes(30)
			td, err := d.Time()
			Expect(err).ToNot(HaveOccurred())
			Expect(td).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should convert zero duration", func() {
			// TC-FO-014
			d := durbig.Seconds(0)
			td, err := d.Time()
			Expect(err).ToNot(HaveOccurred())
			Expect(td).To(Equal(time.Duration(0)))
		})

		It("should convert negative duration", func() {
			// TC-FO-015
			d := durbig.Seconds(-30)
			td, err := d.Time()
			Expect(err).ToNot(HaveOccurred())
			Expect(td).To(Equal(-30 * time.Second))
		})

		It("should return error for overflow", func() {
			// TC-FO-016
			// Create a duration larger than time.Duration can hold
			d := durbig.Days(1000000) // Very large number of days
			_, err := d.Time()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("overflow"))
		})

		It("should handle maximum safe time.Duration", func() {
			// TC-FO-017
			// math.MaxInt64 nanoseconds ~= 292 years
			maxSeconds := int64(math.MaxInt64 / int64(time.Second))
			d := durbig.Seconds(maxSeconds - 1)
			td, err := d.Time()
			Expect(err).ToNot(HaveOccurred())
			Expect(td).To(BeNumerically(">", 0))
		})

		It("should handle values just under overflow", func() {
			// TC-FO-018
			// Just below overflow threshold
			maxSeconds := int64(math.MaxInt64 / int64(time.Second))
			d := durbig.Seconds(maxSeconds - 100)
			td, err := d.Time()
			Expect(err).ToNot(HaveOccurred())
			Expect(td).To(BeNumerically(">", 0))
		})
	})

	Describe("Int64", func() {
		It("should return int64 value for positive duration", func() {
			// TC-FO-019
			d := durbig.Seconds(100)
			Expect(d.Int64()).To(Equal(int64(100)))
		})

		It("should return int64 value for negative duration", func() {
			// TC-FO-020
			d := durbig.Seconds(-100)
			Expect(d.Int64()).To(Equal(int64(-100)))
		})

		It("should return zero for zero duration", func() {
			// TC-FO-021
			d := durbig.Seconds(0)
			Expect(d.Int64()).To(Equal(int64(0)))
		})

		It("should handle large values", func() {
			// TC-FO-022
			d := durbig.Days(1000)
			expected := int64(1000 * 24 * 3600)
			Expect(d.Int64()).To(Equal(expected))
		})

		It("should handle MaxInt64", func() {
			// TC-FO-023
			d := durbig.Seconds(math.MaxInt64)
			Expect(d.Int64()).To(Equal(int64(math.MaxInt64)))
		})

		It("should handle MinInt64", func() {
			// TC-FO-024
			d := durbig.Seconds(math.MinInt64)
			Expect(d.Int64()).To(Equal(int64(math.MinInt64)))
		})
	})

	Describe("Uint64", func() {
		It("should return uint64 value for positive duration", func() {
			// TC-FO-025
			d := durbig.Seconds(100)
			Expect(d.Uint64()).To(Equal(uint64(100)))
		})

		It("should return zero for negative duration", func() {
			// TC-FO-026
			d := durbig.Seconds(-100)
			Expect(d.Uint64()).To(Equal(uint64(0)))
		})

		It("should return zero for zero duration", func() {
			// TC-FO-027
			d := durbig.Seconds(0)
			Expect(d.Uint64()).To(Equal(uint64(0)))
		})

		It("should handle large positive values", func() {
			// TC-FO-028
			d := durbig.Days(1000)
			expected := uint64(1000 * 24 * 3600)
			Expect(d.Uint64()).To(Equal(expected))
		})

		It("should return 0 for negative days", func() {
			// TC-FO-029
			d := durbig.Days(-5)
			Expect(d.Uint64()).To(Equal(uint64(0)))
		})
	})

	Describe("Float64", func() {
		It("should return float64 value for positive duration", func() {
			// TC-FO-030
			d := durbig.Seconds(100)
			Expect(d.Float64()).To(Equal(float64(100)))
		})

		It("should return float64 value for negative duration", func() {
			// TC-FO-031
			d := durbig.Seconds(-100)
			Expect(d.Float64()).To(Equal(float64(-100)))
		})

		It("should return zero for zero duration", func() {
			// TC-FO-032
			d := durbig.Seconds(0)
			Expect(d.Float64()).To(Equal(float64(0)))
		})

		It("should handle large values", func() {
			// TC-FO-033
			d := durbig.Days(1000)
			expected := float64(1000 * 24 * 3600)
			Expect(d.Float64()).To(Equal(expected))
		})

		It("should handle fractional conversion", func() {
			// TC-FO-034
			d := durbig.Seconds(5)
			Expect(d.Float64()).To(BeNumerically("~", 5.0))
		})

		It("should preserve precision for large numbers", func() {
			// TC-FO-035
			d := durbig.Days(365)
			expected := float64(365 * 24 * 3600)
			Expect(d.Float64()).To(Equal(expected))
		})
	})

	Describe("String Formatting Edge Cases", func() {
		It("should format duration with only minutes", func() {
			// TC-FO-038
			d := durbig.Minutes(45)
			Expect(d.String()).To(Equal("45m"))
		})

		It("should format duration with only hours", func() {
			// TC-FO-039
			d := durbig.Hours(12)
			Expect(d.String()).To(Equal("12h"))
		})

		It("should format 1 second", func() {
			// TC-FO-040
			d := durbig.Seconds(1)
			Expect(d.String()).To(Equal("1s"))
		})

		It("should format 1 minute", func() {
			// TC-FO-041
			d := durbig.Minutes(1)
			Expect(d.String()).To(Equal("1m"))
		})

		It("should format 1 hour", func() {
			// TC-FO-042
			d := durbig.Hours(1)
			Expect(d.String()).To(Equal("1h"))
		})

		It("should format 1 day", func() {
			// TC-FO-043
			d := durbig.Days(1)
			Expect(d.String()).To(Equal("1d"))
		})

		It("should format maximum duration", func() {
			// TC-FO-044
			d := durbig.Duration(math.MaxInt64)
			str := d.String()
			Expect(str).ToNot(BeEmpty())
			Expect(str).To(ContainSubstring("d"))
		})
	})

	Describe("Type Conversion Round-trip", func() {
		It("should maintain value through Int64 conversion", func() {
			// TC-FO-045
			original := int64(12345)
			d := durbig.Seconds(original)
			Expect(d.Int64()).To(Equal(original))
		})

		It("should maintain value through Float64 conversion", func() {
			// TC-FO-046
			original := float64(12345)
			d := durbig.ParseFloat64(original)
			Expect(d.Float64()).To(BeNumerically("~", original, 1))
		})

		It("should handle conversion chain", func() {
			// TC-FO-047
			original := durbig.Hours(5) + durbig.Minutes(30)
			f := original.Float64()
			d := durbig.ParseFloat64(f)
			Expect(d.Int64()).To(Equal(original.Int64()))
		})
	})

	Describe("Components Access", func() {
		It("should return correct Days component", func() {
			// TC-FO-048
			d := durbig.Days(5) + durbig.Hours(12)
			Expect(d.Days()).To(Equal(int64(5)))
		})

		It("should return correct Hours component (total hours)", func() {
			// TC-FO-049
			d := durbig.Days(1) + durbig.Hours(2)
			Expect(d.Hours()).To(Equal(int64(26)))
		})

		It("should return correct Minutes component (total minutes)", func() {
			// TC-FO-050
			d := durbig.Hours(1) + durbig.Minutes(30)
			Expect(d.Minutes()).To(Equal(int64(90)))
		})

		It("should return correct Seconds component (total seconds)", func() {
			// TC-FO-051
			d := durbig.Minutes(1) + durbig.Seconds(30)
			Expect(d.Seconds()).To(Equal(int64(90)))
		})
	})
})

func BenchmarkString(b *testing.B) {
	// TC-BM-007
	d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
	for i := 0; i < b.N; i++ {
		_ = d.String()
	}
}

func BenchmarkTime(b *testing.B) {
	// TC-BM-008
	d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
	for i := 0; i < b.N; i++ {
		_, _ = d.Time()
	}
}
