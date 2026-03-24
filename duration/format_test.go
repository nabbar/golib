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

var _ = Describe("Duration Formatting", func() {
	Describe("String", func() {
		It("should format duration with days", func() {
			// TC-FT-001
			d := libdur.Days(5) + libdur.Hours(23) + libdur.Minutes(15) + libdur.Seconds(13)
			Expect(d.String()).To(Equal("5d23h15m13s"))
		})

		It("should format duration without days", func() {
			// TC-FT-002
			d := libdur.Hours(23) + libdur.Minutes(15) + libdur.Seconds(13)
			Expect(d.String()).To(Equal("23h15m13s"))
		})

		It("should format simple durations", func() {
			// TC-FT-003
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
			// TC-FT-004
			d := libdur.Seconds(0)
			Expect(d.String()).To(Equal("0s"))
		})

		It("should format negative duration", func() {
			// TC-FT-005
			d := libdur.Seconds(-30)
			str := d.String()
			Expect(str).To(ContainSubstring("-"))
		})

		It("should format milliseconds", func() {
			// TC-FT-006
			d := libdur.ParseDuration(500 * time.Millisecond)
			Expect(d.String()).To(Equal("500ms"))
		})

		It("should format microseconds", func() {
			// TC-FT-007
			d := libdur.ParseDuration(250 * time.Microsecond)
			Expect(d.String()).To(Equal("250µs"))
		})

		It("should format nanoseconds", func() {
			// TC-FT-008
			d := libdur.ParseDuration(100 * time.Nanosecond)
			Expect(d.String()).To(Equal("100ns"))
		})
	})

	Describe("Time", func() {
		It("should convert to time.Duration", func() {
			// TC-FT-009
			d := libdur.Hours(5) + libdur.Minutes(30)
			td := d.Time()
			Expect(td).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should handle zero", func() {
			// TC-FT-010
			d := libdur.Seconds(0)
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle negative", func() {
			// TC-FT-011
			d := libdur.Seconds(-10)
			Expect(d.Time()).To(Equal(-10 * time.Second))
		})
	})

	Describe("Days", func() {
		It("should calculate days correctly", func() {
			// TC-FT-012
			d := libdur.Days(7)
			Expect(d.Days()).To(Equal(int64(7)))
		})

		It("should handle fractional days", func() {
			// TC-FT-013
			d := libdur.Hours(36) // 1.5 days
			Expect(d.Days()).To(Equal(int64(1)))
		})

		It("should handle zero", func() {
			// TC-FT-014
			d := libdur.Hours(0)
			Expect(d.Days()).To(Equal(int64(0)))
		})

		It("should handle less than a day", func() {
			// TC-FT-015
			d := libdur.Hours(12)
			Expect(d.Days()).To(Equal(int64(0)))
		})

		It("should handle negative duration", func() {
			// TC-FT-016
			d := libdur.Days(-5)
			Expect(d.Days()).To(Equal(int64(-5)))
		})

		It("should handle very large durations", func() {
			// TC-FT-017
			// Use a large but not overflow-inducing value
			d := libdur.Days(100000)
			days := d.Days()
			Expect(days).To(Equal(int64(100000)))
		})
	})

	Describe("Float64", func() {
		It("should convert to float64", func() {
			// TC-FT-018
			d := libdur.Seconds(5)
			f := d.Float64()
			Expect(f).To(Equal(float64(5 * time.Second)))
		})

		It("should handle zero", func() {
			// TC-FT-019
			d := libdur.Seconds(0)
			Expect(d.Float64()).To(Equal(float64(0)))
		})

		It("should handle negative", func() {
			// TC-FT-020
			d := libdur.Seconds(-5)
			Expect(d.Float64()).To(Equal(float64(-5 * time.Second)))
		})

		It("should preserve precision", func() {
			// TC-FT-021
			d := libdur.ParseDuration(1500 * time.Millisecond)
			f := d.Float64()
			Expect(f).To(Equal(float64(1500 * time.Millisecond)))
		})
	})

	Describe("Hours", func() {
		It("should calculate hours correctly", func() {
			// TC-FT-022
			d := libdur.Hours(2)
			Expect(d.Hours()).To(Equal(int64(2)))
		})
		It("should calculate hours correctly with days", func() {
			// TC-FT-023
			d := libdur.Days(1)
			Expect(d.Hours()).To(Equal(int64(24)))
		})
	})

	Describe("Minutes", func() {
		It("should calculate minutes correctly", func() {
			// TC-FT-024
			d := libdur.Minutes(2)
			Expect(d.Minutes()).To(Equal(int64(2)))
		})
		It("should calculate minutes correctly with hours", func() {
			// TC-FT-025
			d := libdur.Hours(1)
			Expect(d.Minutes()).To(Equal(int64(60)))
		})
	})

	Describe("Seconds", func() {
		It("should calculate seconds correctly", func() {
			// TC-FT-026
			d := libdur.Seconds(2)
			Expect(d.Seconds()).To(Equal(int64(2)))
		})
	})

	Describe("Milliseconds", func() {
		It("should calculate milliseconds correctly", func() {
			// TC-FT-027
			d := libdur.ParseDuration(2 * time.Millisecond)
			Expect(d.Milliseconds()).To(Equal(int64(2)))
		})
	})

	Describe("Microseconds", func() {
		It("should calculate microseconds correctly", func() {
			// TC-FT-028
			d := libdur.ParseDuration(2 * time.Microsecond)
			Expect(d.Microseconds()).To(Equal(int64(2)))
		})
	})

	Describe("Nanoseconds", func() {
		It("should calculate nanoseconds correctly", func() {
			// TC-FT-029
			d := libdur.ParseDuration(2 * time.Nanosecond)
			Expect(d.Nanoseconds()).To(Equal(int64(2)))
		})
	})

	Describe("Uint64", func() {
		It("should convert to uint64", func() {
			// TC-FT-030
			d := libdur.Seconds(2)
			Expect(d.Uint64()).To(Equal(uint64(2 * time.Second)))
		})
		It("should convert to uint64 with negative value", func() {
			// TC-FT-031
			d := libdur.Seconds(-2)
			Expect(d.Uint64()).To(Equal(uint64(2 * time.Second)))
		})
	})

	Describe("Int64", func() {
		It("should convert to int64", func() {
			// TC-FT-032
			d := libdur.Seconds(2)
			Expect(d.Int64()).To(Equal(int64(2 * time.Second)))
		})
	})

	Describe("Duration", func() {
		It("should return time.Duration", func() {
			// TC-FT-033
			d := libdur.Seconds(2)
			Expect(d.Duration()).To(Equal(2 * time.Second))
		})
	})
})
