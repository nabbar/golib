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
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

var _ = Describe("Duration Parsing", func() {
	Describe("Parse", func() {
		It("should parse valid duration string", func() {
			// TC-PR-001
			d, err := libdur.Parse("5h30m")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should parse duration with days", func() {
			// TC-PR-002
			d, err := libdur.Parse("2d12h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(60 * time.Hour))
		})

		It("should parse negative duration", func() {
			// TC-PR-003
			d, err := libdur.Parse("-5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(-5 * time.Hour))
		})

		It("should parse zero duration", func() {
			// TC-PR-004
			d, err := libdur.Parse("0")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})

		It("should parse fractional duration", func() {
			// TC-PR-005
			d, err := libdur.Parse("1.5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(90 * time.Minute))
		})

		It("should parse all time units", func() {
			// TC-PR-006
			tests := []struct {
				input    string
				expected time.Duration
			}{
				{"100ns", 100 * time.Nanosecond},
				{"100us", 100 * time.Microsecond},
				{"100µs", 100 * time.Microsecond},
				{"100μs", 100 * time.Microsecond},
				{"100ms", 100 * time.Millisecond},
				{"10s", 10 * time.Second},
				{"5m", 5 * time.Minute},
				{"2h", 2 * time.Hour},
				{"1d", 24 * time.Hour},
			}

			for _, tt := range tests {
				d, err := libdur.Parse(tt.input)
				Expect(err).ToNot(HaveOccurred(), "for input "+tt.input)
				Expect(d.Time()).To(Equal(tt.expected), "for input "+tt.input)
			}
		})

		It("should parse complex duration", func() {
			// TC-PR-007
			d, err := libdur.Parse("5d23h15m13s")
			Expect(err).ToNot(HaveOccurred())
			expected := 5*24*time.Hour + 23*time.Hour + 15*time.Minute + 13*time.Second
			Expect(d.Time()).To(Equal(expected))
		})

		It("should handle quoted strings", func() {
			// TC-PR-008
			d, err := libdur.Parse("\"5h30m\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should handle strings with spaces", func() {
			// TC-PR-009
			d, err := libdur.Parse(" 5h 30m ")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(5*time.Hour + 30*time.Minute))
		})

		It("should return error for invalid format", func() {
			// TC-PR-010
			_, err := libdur.Parse("invalid")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for unknown unit", func() {
			// TC-PR-011
			_, err := libdur.Parse("5x")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing unit", func() {
			// TC-PR-012
			_, err := libdur.Parse("5")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for empty string", func() {
			// TC-PR-013
			_, err := libdur.Parse("")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for overflow", func() {
			// TC-PR-014
			_, err := libdur.Parse("99999999999999999999h")
			Expect(err).To(HaveOccurred())
		})

		It("should handle single zero", func() {
			// TC-PR-015
			d, err := libdur.Parse("0")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle plus sign prefix", func() {
			// TC-PR-016
			d, err := libdur.Parse("+5h")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(5 * time.Hour))
		})

		It("should return error for just sign", func() {
			// TC-PR-017
			_, err := libdur.Parse("-")
			Expect(err).To(HaveOccurred())
		})

		It("should return error for just plus sign", func() {
			// TC-PR-018
			_, err := libdur.Parse("+")
			Expect(err).To(HaveOccurred())
		})

		It("should handle fractional microseconds", func() {
			// TC-PR-019
			d, err := libdur.Parse("1.5µs")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(BeNumerically(">", time.Microsecond))
		})

		It("should return error for double unit", func() {
			// TC-PR-020
			_, err := libdur.Parse("5hh")
			Expect(err).To(HaveOccurred())
		})

		It("should handle very small fractional values", func() {
			// TC-PR-021
			d, err := libdur.Parse("0.001ms")
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(time.Microsecond))
		})

		It("should return error for dot without digits", func() {
			// TC-PR-022
			_, err := libdur.Parse(".s")
			Expect(err).To(HaveOccurred())
		})

		It("should handle multiple components", func() {
			// TC-PR-023
			d, err := libdur.Parse("1d2h3m4s5ms")
			Expect(err).ToNot(HaveOccurred())
			expected := 24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second + 5*time.Millisecond
			Expect(d.Time()).To(Equal(expected))
		})
	})

	Describe("ParseByte", func() {
		It("should parse valid byte array", func() {
			// TC-PR-024
			d, err := libdur.ParseByte([]byte("3h45m"))
			Expect(err).ToNot(HaveOccurred())
			Expect(d.Time()).To(Equal(3*time.Hour + 45*time.Minute))
		})

		It("should return error for invalid byte array", func() {
			// TC-PR-025
			_, err := libdur.ParseByte([]byte("invalid"))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Helper Functions", func() {
		It("should create duration from seconds", func() {
			// TC-PR-026
			d := libdur.Seconds(120)
			Expect(d.Time()).To(Equal(2 * time.Minute))
		})

		It("should create duration from minutes", func() {
			// TC-PR-027
			d := libdur.Minutes(90)
			Expect(d.Time()).To(Equal(90 * time.Minute))
		})

		It("should create duration from hours", func() {
			// TC-PR-028
			d := libdur.Hours(48)
			Expect(d.Time()).To(Equal(48 * time.Hour))
		})

		It("should create duration from days", func() {
			// TC-PR-029
			d := libdur.Days(7)
			Expect(d.Time()).To(Equal(7 * 24 * time.Hour))
		})

		It("should handle negative values", func() {
			// TC-PR-030
			d := libdur.Seconds(-30)
			Expect(d.Time()).To(Equal(-30 * time.Second))
		})

		It("should handle zero", func() {
			// TC-PR-031
			d := libdur.Seconds(0)
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})
	})

	Describe("ParseDuration", func() {
		It("should convert time.Duration", func() {
			// TC-PR-032
			td := 5*time.Hour + 30*time.Minute
			d := libdur.ParseDuration(td)
			Expect(d.Time()).To(Equal(td))
		})

		It("should handle negative duration", func() {
			// TC-PR-033
			td := -2 * time.Hour
			d := libdur.ParseDuration(td)
			Expect(d.Time()).To(Equal(td))
		})
	})

	Describe("ParseFloat64", func() {
		It("should convert positive float", func() {
			// TC-PR-034
			d := libdur.ParseFloat64(1000000000)
			Expect(d.Time()).To(Equal(time.Second))
		})

		It("should convert negative float", func() {
			// TC-PR-035
			d := libdur.ParseFloat64(-1000000000)
			Expect(d.Time()).To(Equal(-time.Second))
		})

		It("should handle zero", func() {
			// TC-PR-036
			d := libdur.ParseFloat64(0)
			Expect(d.Time()).To(Equal(time.Duration(0)))
		})

		It("should handle very large values (overflow protection)", func() {
			// TC-PR-037
			d := libdur.ParseFloat64(math.MaxFloat64)
			Expect(d.Time()).To(Equal(time.Duration(math.MaxInt64)))
		})

		It("should handle very small values (underflow protection)", func() {
			// TC-PR-038
			d := libdur.ParseFloat64(-math.MaxFloat64)
			Expect(d.Time()).To(Equal(time.Duration(-math.MaxInt64)))
		})

		It("should round values", func() {
			// TC-PR-039
			d := libdur.ParseFloat64(1500000000.7)
			expected := time.Duration(math.Round(1500000000.7))
			Expect(d.Time()).To(Equal(expected))
		})
	})
})
