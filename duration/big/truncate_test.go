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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Truncate Operations", func() {
	Describe("Truncate with Custom Unit", func() {
		It("should truncate to minute", func() {
			d := durbig.Minutes(5) + durbig.Seconds(45)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(5).Int64()))
		})

		It("should truncate to hour", func() {
			d := durbig.Hours(3) + durbig.Minutes(45)
			truncated := d.Truncate(durbig.Hour)
			Expect(truncated.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should truncate to day", func() {
			d := durbig.Days(2) + durbig.Hours(12)
			truncated := d.Truncate(durbig.Day)
			Expect(truncated.Int64()).To(Equal(durbig.Days(2).Int64()))
		})

		It("should return unchanged when unit is negative", func() {
			d := durbig.Minutes(5) + durbig.Seconds(30)
			truncated := d.Truncate(durbig.Duration(-1))
			Expect(truncated).To(Equal(d))
		})

		It("should return unchanged when unit is zero", func() {
			d := durbig.Minutes(5) + durbig.Seconds(30)
			truncated := d.Truncate(0)
			Expect(truncated).To(Equal(d))
		})

		It("should handle zero duration", func() {
			d := durbig.Seconds(0)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			d := durbig.Minutes(-5) - durbig.Seconds(30)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(-5).Int64()))
		})

		It("should handle exact multiple", func() {
			d := durbig.Minutes(10)
			truncated := d.Truncate(durbig.Minute)
			Expect(truncated).To(Equal(d))
		})

		It("should truncate to custom unit", func() {
			d := durbig.Seconds(127)
			truncated := d.Truncate(durbig.Seconds(10))
			Expect(truncated.Int64()).To(Equal(int64(120)))
		})
	})

	Describe("Round with Custom Unit", func() {
		It("should round up when above halfway", func() {
			d := durbig.Seconds(55)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should round down when below halfway", func() {
			d := durbig.Seconds(25)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round away from zero at halfway", func() {
			d := durbig.Seconds(30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should handle negative duration rounding", func() {
			d := durbig.Seconds(-55)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(-durbig.Minute.Int64()))
		})

		It("should return unchanged when unit is negative", func() {
			d := durbig.Minutes(5) + durbig.Seconds(30)
			rounded := d.Round(durbig.Duration(-1))
			Expect(rounded).To(Equal(d))
		})

		It("should return unchanged when unit is zero", func() {
			d := durbig.Minutes(5) + durbig.Seconds(30)
			rounded := d.Round(0)
			Expect(rounded).To(Equal(d))
		})

		It("should handle zero duration", func() {
			d := durbig.Seconds(0)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round hours correctly", func() {
			d := durbig.Hours(2) + durbig.Minutes(45)
			rounded := d.Round(durbig.Hour)
			Expect(rounded.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should round days correctly", func() {
			d := durbig.Days(1) + durbig.Hours(18)
			rounded := d.Round(durbig.Day)
			Expect(rounded.Int64()).To(Equal(durbig.Days(2).Int64()))
		})

		It("should handle exact multiple", func() {
			d := durbig.Hours(5)
			rounded := d.Round(durbig.Hour)
			Expect(rounded).To(Equal(d))
		})
	})

	Describe("TruncateMinutes", func() {
		It("should truncate to minutes", func() {
			d := durbig.Minutes(5) + durbig.Seconds(45)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(5).Int64()))
		})

		It("should handle zero duration", func() {
			d := durbig.Seconds(0)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			d := durbig.Minutes(-5) - durbig.Seconds(30)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(durbig.Minutes(-5).Int64()))
		})

		It("should not change exact minutes", func() {
			d := durbig.Minutes(10)
			truncated := d.TruncateMinutes()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate duration with hours", func() {
			d := durbig.Hours(2) + durbig.Minutes(30) + durbig.Seconds(45)
			truncated := d.TruncateMinutes()
			expected := durbig.Hours(2) + durbig.Minutes(30)
			Expect(truncated.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("TruncateHours", func() {
		It("should truncate to hours", func() {
			d := durbig.Hours(3) + durbig.Minutes(45)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(durbig.Hours(3).Int64()))
		})

		It("should handle zero duration", func() {
			d := durbig.Seconds(0)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			d := durbig.Hours(-3) - durbig.Minutes(30)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(durbig.Hours(-3).Int64()))
		})

		It("should not change exact hours", func() {
			d := durbig.Hours(5)
			truncated := d.TruncateHours()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate duration with days", func() {
			d := durbig.Days(2) + durbig.Hours(15) + durbig.Minutes(30)
			truncated := d.TruncateHours()
			expected := durbig.Days(2) + durbig.Hours(15)
			Expect(truncated.Int64()).To(Equal(expected.Int64()))
		})
	})

	Describe("TruncateDays", func() {
		It("should truncate to days", func() {
			d := durbig.Days(5) + durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(5).Int64()))
		})

		It("should handle zero duration", func() {
			d := durbig.Seconds(0)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate negative duration", func() {
			d := durbig.Days(-5) - durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(-5).Int64()))
		})

		It("should not change exact days", func() {
			d := durbig.Days(7)
			truncated := d.TruncateDays()
			Expect(truncated).To(Equal(d))
		})

		It("should truncate partial days", func() {
			d := durbig.Days(10) + durbig.Hours(23) + durbig.Minutes(59) + durbig.Seconds(59)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(durbig.Days(10).Int64()))
		})

		It("should handle duration less than a day", func() {
			d := durbig.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("Truncate Edge Cases", func() {
		It("should handle very small truncation unit", func() {
			d := durbig.Seconds(123)
			truncated := d.Truncate(durbig.Seconds(1))
			Expect(truncated).To(Equal(d))
		})

		It("should handle very large truncation unit", func() {
			d := durbig.Days(100) + durbig.Hours(12)
			truncated := d.Truncate(durbig.Days(10))
			Expect(truncated.Int64()).To(Equal(durbig.Days(100).Int64()))
		})

		It("should truncate 59 seconds to 0", func() {
			d := durbig.Seconds(59)
			truncated := d.TruncateMinutes()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate 59 minutes to 0", func() {
			d := durbig.Minutes(59)
			truncated := d.TruncateHours()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})

		It("should truncate 23 hours to 0", func() {
			d := durbig.Hours(23)
			truncated := d.TruncateDays()
			Expect(truncated.Int64()).To(Equal(int64(0)))
		})
	})

	Describe("Round Edge Cases", func() {
		It("should round 30 seconds up", func() {
			d := durbig.Seconds(30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should round 29 seconds down", func() {
			d := durbig.Seconds(29)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(int64(0)))
		})

		It("should round 31 seconds up", func() {
			d := durbig.Seconds(31)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(durbig.Minute.Int64()))
		})

		It("should handle negative halfway rounding", func() {
			d := durbig.Seconds(-30)
			rounded := d.Round(durbig.Minute)
			Expect(rounded.Int64()).To(Equal(-durbig.Minute.Int64()))
		})
	})

	Describe("Truncate Chain Operations", func() {
		It("should allow chaining truncate operations", func() {
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
