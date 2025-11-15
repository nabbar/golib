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
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF ANY KIND, EXPRESS OR
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

var _ = Describe("Duration Formatting", func() {
	Describe("String", func() {
		It("should format duration with days", func() {
			d := libdur.Days(5) + libdur.Hours(23) + libdur.Minutes(15) + libdur.Seconds(13)
			Expect(d.String()).To(Equal("5d23h15m13s"))
		})

		It("should format duration without days", func() {
			d := libdur.Hours(23) + libdur.Minutes(15) + libdur.Seconds(13)
			Expect(d.String()).To(Equal("23h15m13s"))
		})

		It("should format simple durations", func() {
			tests := []struct {
				duration libdur.Duration
				expected string
			}{
				{libdur.Seconds(30), "30s"},
				{libdur.Minutes(5), "5m0s"},
				{libdur.Hours(2), "2h0m0s"},
				{libdur.Days(1), "1d"},
			}

			for _, tt := range tests {
				Expect(tt.duration.String()).To(Equal(tt.expected))
			}
		})

		It("should format zero duration", func() {
			d := libdur.Seconds(0)
			Expect(d.String()).To(Equal("0s"))
		})

		It("should format negative duration", func() {
			d := libdur.Seconds(-30)
			str := d.String()
			Expect(str).To(ContainSubstring("-"))
		})

		It("should format milliseconds", func() {
			d := libdur.ParseDuration(500 * time.Millisecond)
			Expect(d.String()).To(Equal("500ms"))
		})

		It("should format microseconds", func() {
			d := libdur.ParseDuration(250 * time.Microsecond)
			Expect(d.String()).To(Equal("250µs"))
		})

		It("should format nanoseconds", func() {
			d := libdur.ParseDuration(100 * time.Nanosecond)
			Expect(d.String()).To(Equal("100ns"))
		})
	})

	Describe("Time", func() {
		It("should convert to time.Duration", func() {
			d := libdur.Hours(5) + libdur.Minutes(30)
			td := d.Time()
			Expect(td).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle negative", func() {
			d := libdur.Seconds(-10)
			Expect(d.Time()).To(Equal(-10 * time.Second))
		})
	})

	Describe("Days", func() {
		It("should calculate days correctly", func() {
			d := libdur.Days(7)
			Expect(d.Days()).To(Equal(int64(7)))
		})

		It("should handle fractional days", func() {
			d := libdur.Hours(36) // 1.5 days
			Expect(d.Days()).To(Equal(int64(1)))
		})

		It("should handle zero", func() {
			d := libdur.Hours(0)
			Expect(d.Days()).To(Equal(int64(0)))
		})

		It("should handle less than a day", func() {
			d := libdur.Hours(12)
			Expect(d.Days()).To(Equal(int64(0)))
		})

		It("should handle negative duration", func() {
			d := libdur.Days(-5)
			Expect(d.Days()).To(Equal(int64(-5)))
		})

		It("should handle very large durations", func() {
			// Use a large but not overflow-inducing value
			d := libdur.Days(100000)
			days := d.Days()
			Expect(days).To(Equal(int64(100000)))
		})
	})

	Describe("Float64", func() {
		It("should convert to float64", func() {
			d := libdur.Seconds(5)
			f := d.Float64()
			Expect(f).To(Equal(float64(5 * time.Second)))
		})

		It("should handle zero", func() {
			d := libdur.Seconds(0)
			Expect(d.Float64()).To(Equal(float64(0)))
		})

		It("should handle negative", func() {
			d := libdur.Seconds(-5)
			Expect(d.Float64()).To(Equal(float64(-5 * time.Second)))
		})

		It("should preserve precision", func() {
			d := libdur.ParseDuration(1500 * time.Millisecond)
			f := d.Float64()
			Expect(f).To(Equal(float64(1500 * time.Millisecond)))
		})
	})
})
