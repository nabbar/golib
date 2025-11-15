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

package duration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

var _ = Describe("Duration Truncate Operations", func() {
	Describe("TruncateMicroseconds", func() {
		It("should truncate to microseconds", func() {
			// 1500ns + 500ns = 2000ns = 2µs, so truncated value should be 2µs
			d := libdur.ParseDuration(1500*time.Nanosecond + 500*time.Nanosecond)
			truncated := d.TruncateMicroseconds()
			Expect(truncated.Time()).To(Equal(2 * time.Microsecond))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateMicroseconds()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact microseconds", func() {
			d := libdur.ParseDuration(5 * time.Microsecond)
			truncated := d.TruncateMicroseconds()
			Expect(truncated.Time()).To(Equal(5 * time.Microsecond))
		})
	})

	Describe("TruncateMilliseconds", func() {
		It("should truncate to milliseconds", func() {
			d := libdur.ParseDuration(1500*time.Microsecond + 250*time.Microsecond)
			truncated := d.TruncateMilliseconds()
			Expect(truncated.Time()).To(Equal(1 * time.Millisecond))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateMilliseconds()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact milliseconds", func() {
			d := libdur.ParseDuration(500 * time.Millisecond)
			truncated := d.TruncateMilliseconds()
			Expect(truncated.Time()).To(Equal(500 * time.Millisecond))
		})
	})

	Describe("TruncateSeconds", func() {
		It("should truncate to seconds", func() {
			d := libdur.ParseDuration(5*time.Second + 750*time.Millisecond)
			truncated := d.TruncateSeconds()
			Expect(truncated.Time()).To(Equal(5 * time.Second))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateSeconds()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact seconds", func() {
			d := libdur.Seconds(30)
			truncated := d.TruncateSeconds()
			Expect(truncated.Time()).To(Equal(30 * time.Second))
		})

		It("should handle fractional seconds", func() {
			d := libdur.ParseDuration(3*time.Second + 999*time.Millisecond)
			truncated := d.TruncateSeconds()
			Expect(truncated.Time()).To(Equal(3 * time.Second))
		})
	})

	Describe("TruncateMinutes", func() {
		It("should truncate to minutes", func() {
			d := libdur.ParseDuration(5*time.Minute + 45*time.Second)
			truncated := d.TruncateMinutes()
			Expect(truncated.Time()).To(Equal(5 * time.Minute))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateMinutes()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact minutes", func() {
			d := libdur.Minutes(10)
			truncated := d.TruncateMinutes()
			Expect(truncated.Time()).To(Equal(10 * time.Minute))
		})

		It("should handle less than a minute", func() {
			d := libdur.Seconds(30)
			truncated := d.TruncateMinutes()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})
	})

	Describe("TruncateHours", func() {
		It("should truncate to hours", func() {
			d := libdur.ParseDuration(3*time.Hour + 45*time.Minute)
			truncated := d.TruncateHours()
			Expect(truncated.Time()).To(Equal(3 * time.Hour))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateHours()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact hours", func() {
			d := libdur.Hours(24)
			truncated := d.TruncateHours()
			Expect(truncated.Time()).To(Equal(24 * time.Hour))
		})

		It("should handle less than an hour", func() {
			d := libdur.Minutes(30)
			truncated := d.TruncateHours()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})
	})

	Describe("TruncateDays", func() {
		It("should truncate to days", func() {
			d := libdur.Days(5) + libdur.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(5 * 24 * time.Hour))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle exact days", func() {
			d := libdur.Days(7)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(7 * 24 * time.Hour))
		})

		It("should handle less than a day", func() {
			d := libdur.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle complex duration", func() {
			d := libdur.Days(2) + libdur.Hours(18) + libdur.Minutes(45) + libdur.Seconds(30)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(2 * 24 * time.Hour))
		})
	})

	Describe("Truncate Chain", func() {
		It("should allow chaining truncate operations", func() {
			d := libdur.ParseDuration(5*time.Hour + 45*time.Minute + 30*time.Second + 500*time.Millisecond)

			// Truncate to minutes
			m := d.TruncateMinutes()
			Expect(m.Time()).To(Equal(5*time.Hour + 45*time.Minute))

			// Then truncate to hours
			h := d.TruncateHours()
			Expect(h.Time()).To(Equal(5 * time.Hour))
		})
	})

	Describe("Truncate with Negative Durations", func() {
		It("should truncate negative seconds", func() {
			d := libdur.Seconds(-35)
			truncated := d.TruncateSeconds()
			Expect(truncated.Time()).To(Equal(-35 * time.Second))
		})

		It("should truncate negative minutes", func() {
			// -5m30s = -5.5 minutes, truncated should be -6 minutes (floor rounds toward negative infinity)
			d := libdur.Minutes(-5) - libdur.Seconds(30)
			truncated := d.TruncateMinutes()
			Expect(truncated.Time()).To(Equal(-6 * time.Minute))
		})

		It("should truncate negative hours", func() {
			// -3h30m = -3.5 hours, truncated should be -4 hours (floor rounds toward negative infinity)
			d := libdur.Hours(-3) - libdur.Minutes(30)
			truncated := d.TruncateHours()
			Expect(truncated.Time()).To(Equal(-4 * time.Hour))
		})

		It("should truncate negative days", func() {
			// -2d12h = -2.5 days, truncated should be -3 days (floor rounds toward negative infinity)
			d := libdur.Days(-2) - libdur.Hours(12)
			truncated := d.TruncateDays()
			Expect(truncated.Time()).To(Equal(-3 * 24 * time.Hour))
		})
	})

	Describe("Truncate Edge Cases", func() {
		It("should handle very large milliseconds", func() {
			d := libdur.ParseDuration(999999 * time.Millisecond)
			truncated := d.TruncateMilliseconds()
			Expect(truncated.Time()).To(Equal(999999 * time.Millisecond))
		})

		It("should handle fractional nanoseconds", func() {
			d := libdur.ParseDuration(1234 * time.Nanosecond)
			truncated := d.TruncateMicroseconds()
			Expect(truncated.Time()).To(Equal(1 * time.Microsecond))
		})

		It("should truncate mixed precision correctly", func() {
			d := libdur.ParseDuration(1*time.Hour + 23*time.Minute + 45*time.Second + 678*time.Millisecond + 912*time.Microsecond + 345*time.Nanosecond)

			// TruncateMilliseconds removes microseconds and nanoseconds
			ms := d.TruncateMilliseconds()
			expectedMs := 1*time.Hour + 23*time.Minute + 45*time.Second + 678*time.Millisecond
			Expect(ms.Time()).To(Equal(expectedMs))
		})
	})
})
