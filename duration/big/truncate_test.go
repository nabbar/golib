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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Truncate Operations", func() {
	Describe("Truncate with Custom Unit", func() {
		It("should truncate to minute", func() {
			// TC-TR-001
			d := durbig.Minutes(5) + durbig.Seconds(45)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(5).Int64()))
		})

		It("should truncate to hour", func() {
			// TC-TR-002
			d := durbig.Hours(3) + durbig.Minutes(45)
			truncated := d.Truncate(durbig.Hour)
			Expect(truncated.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should truncate to day", func() {
			// TC-TR-003
			d := durbig.Days(2) + durbig.Hours(12)
			truncated := d.Truncate(durbig.Day)
			Expect(truncated.Int64()).To(Equal(durbig.Days(2).Int64()))
		})

		It("should return unchanged when unit is negative", func() {
			// TC-TR-004
			d := durbig.Minutes(5) + durbig.Seconds(30)
			truncated := d.Truncate(durbig.Duration(-1))
			Expect(truncated).To(Equal(d))
		})

		It("should return unchanged when unit is zero", func() {
			// TC-TR-005
			d := durbig.Minutes(5) + durbig.Seconds(30)
			truncated := d.Truncate(0)
			Expect(truncated).To(Equal(d))
		})

		It("should handle zero duration", func() {
			// TC-TR-006
			d := durbig.Seconds(0)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			// TC-TR-007
			d := durbig.Minutes(-5) - durbig.Seconds(30)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(-5).Int64()))
		})

		It("should handle exact multiple", func() {
			// TC-TR-008
			d := durbig.Minutes(10)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated).To(Equal(d))
		})

		It("should truncate to custom unit", func() {
			// TC-TR-009
			d := durbig.Seconds(127)
			truncated := d.Truncate(durbig.Seconds(10))
			Expect(truncated.Int64()).To(Equal(int64(120)))
		})
	})

	Describe("Round with Custom Unit", func() {
		It("should round up when above halfway", func() {
			// TC-TR-010
			d := durbig.Seconds(55)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should round down when below halfway", func() {
			// TC-TR-011
			d := durbig.Seconds(25)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round away from zero at halfway", func() {
			// TC-TR-012
			d := durbig.Seconds(30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should handle negative duration rounding", func() {
			// TC-TR-013
			d := durbig.Seconds(-55)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(-durbig.Minute.Int64()))
		})

		It("should return unchanged when unit is negative", func() {
			// TC-TR-014
			d := durbig.Minutes(5) + durbig.Seconds(30)
			rounded := d.Round(durbig.Duration(-1))
			Expect(rounded).To(Equal(d))
		})

		It("should return unchanged when unit is zero", func() {
			// TC-TR-015
			d := durbig.Minutes(5) + durbig.Seconds(30)
			rounded := d.Round(0)
			Expect(rounded).To(Equal(d))
		})

		It("should handle zero duration", func() {
			// TC-TR-016
			d := durbig.Seconds(0)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round hours correctly", func() {
			// TC-TR-017
			d := durbig.Hours(2) + durbig.Minutes(45)
			rounded := d.Round(durbig.Hour)
			Expect(rounded.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should round days correctly", func() {
			// TC-TR-018
			d := durbig.Days(1) + durbig.Hours(18)
			rounded := d.Round(durbig.Day)
			Expect(rounded.Int64()).To(Equal(durbig.Days(2).Int64()))
		})

		It("should handle exact multiple", func() {
			// TC-TR-019
			d := durbig.Hours(5)
			rounded := d.Round(durbig.Hour)
			Expect(rounded).To(Equal(d))
		})
	})

	Describe("TruncateMinutes", func() {
		It("should truncate to minutes", func() {
			// TC-TR-020
			d := durbig.Minutes(5) + durbig.Seconds(45)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(5).Int64()))
		})

		It("should handle zero duration", func() {
			// TC-TR-021
			d := durbig.Seconds(0)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			// TC-TR-022
			d := durbig.Minutes(-5) - durbig.Seconds(30)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(-5).Int64()))
		})

		It("should not change exact minutes", func() {
			// TC-TR-023
			d := durbig.Minutes(10)
			truncated := d.TruncateMinutes()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate duration with hours", func() {
			// TC-TR-024
			d := durbig.Hours(2) + durbig.Minutes(30) + durbig.Seconds(45)
			truncated := d.TruncateMinutes()
			expected := durbig.Hours(2) + durbig.Minutes(30)
			Expect(truncated.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("TruncateHours", func() {
		It("should truncate to hours", func() {
			// TC-TR-025
			d := durbig.Hours(3) + durbig.Minutes(45)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should handle zero duration", func() {
			// TC-TR-026
			d := durbig.Seconds(0)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			// TC-TR-027
			d := durbig.Hours(-3) - durbig.Minutes(30)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(durbig.Hours(-3).Int64()))
		})

		It("should not change exact hours", func() {
			// TC-TR-028
			d := durbig.Hours(5)
			truncated := d.TruncateHours()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate duration with days", func() {
			// TC-TR-029
			d := durbig.Days(2) + durbig.Hours(15) + durbig.Minutes(30)
			truncated := d.TruncateHours()
			expected := durbig.Days(2) + durbig.Hours(15)
			Expect(truncated.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("TruncateDays", func() {
		It("should truncate to days", func() {
			// TC-TR-030
			d := durbig.Days(5) + durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(5).Int64()))
		})

		It("should handle zero duration", func() {
			// TC-TR-031
			d := durbig.Seconds(0)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			// TC-TR-032
			d := durbig.Days(-5) - durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(-5).Int64()))
		})

		It("should not change exact days", func() {
			// TC-TR-033
			d := durbig.Days(7)
			truncated := d.TruncateDays()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate partial days", func() {
			// TC-TR-034
			d := durbig.Days(10) + durbig.Hours(23) + durbig.Minutes(59) + durbig.Seconds(59)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(10).Int64()))
		})

		It("should handle duration less than a day", func() {
			// TC-TR-035
			d := durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("Truncate Edge Cases", func() {
		It("should handle very small truncation unit", func() {
			// TC-TR-036
			d := durbig.Seconds(123)
			truncated := d.Truncate(durbig.Seconds(1))
			Expect(truncated).To(Equal(d))
		})

		It("should handle very large truncation unit", func() {
			// TC-TR-037
			d := durbig.Days(100) + durbig.Hours(12)
			truncated := d.Truncate(durbig.Days(10))
			Expect(truncated.Int64()).To(Equal(durbig.Days(100).Int64()))
		})

		It("should truncate 59 seconds to 0", func() {
			// TC-TR-038
			d := durbig.Seconds(59)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate 59 minutes to 0", func() {
			// TC-TR-039
			d := durbig.Minutes(59)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate 23 hours to 0", func() {
			// TC-TR-040
			d := durbig.Hours(23)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("Round Edge Cases", func() {
		It("should round 30 seconds up", func() {
			// TC-TR-041
			d := durbig.Seconds(30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should round 29 seconds down", func() {
			// TC-TR-042
			d := durbig.Seconds(29)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round 31 seconds up", func() {
			// TC-TR-043
			d := durbig.Seconds(31)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should handle negative halfway rounding", func() {
			// TC-TR-044
			d := durbig.Seconds(-30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(-durbig.Minute.Int64()))
		})
	})

	Describe("Truncate Chain Operations", func() {
		It("should allow chaining truncate operations", func() {
			// TC-TR-045
			d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(45) + durbig.Seconds(30)

			// Truncate to minutes
			m := d.TruncateMinutes()
			expected := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(45)
			Expect(m.Int64()).To(Equal(expected.Int64()))

			// Then truncate to hours
			h := d.TruncateHours()
			expected = durbig.Days(5) + durbig.Hours(23)
			Expect(h.Int64()).To(Equal(expected.Int64()))

			// Then truncate to days
			days := d.TruncateDays()
			Expect(days.Int64()).To(Equal(durbig.Days(5).Int64()))
		})
	})
})

func BenchmarkTruncate(b *testing.B) {
	// TC-BM-009
	d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
	unit := durbig.Hours(2)
	for i := 0; i < b.N; i++ {
		_ = d.Truncate(unit)
	}
}

func BenchmarkRound(b *testing.B) {
	// TC-BM-010
	d := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
	unit := durbig.Hours(2)
	for i := 0; i < b.N; i++ {
		_ = d.Round(unit)
	}
}
