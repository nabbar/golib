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
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Model Operations", func() {
	Describe("Is Functions", func() {
		It("should return true for IsDays when duration is >= 1 day", func() {
			// TC-MD-015
			Expect(durbig.Days(1).IsDays()).To(BeTrue())
			Expect(durbig.Days(2).IsDays()).To(BeTrue())
		})

		It("should return false for IsDays when duration is < 1 day", func() {
			// TC-MD-016
			Expect(durbig.Hours(23).IsDays()).To(BeFalse())
			Expect(durbig.Minutes(1439).IsDays()).To(BeFalse())
		})

		It("should return true for IsHours when duration is >= 1 hour", func() {
			// TC-MD-017
			Expect(durbig.Hours(1).IsHours()).To(BeTrue())
			Expect(durbig.Hours(2).IsHours()).To(BeTrue())
		})

		It("should return false for IsHours when duration is < 1 hour", func() {
			// TC-MD-018
			Expect(durbig.Minutes(59).IsHours()).To(BeFalse())
			Expect(durbig.Seconds(3599).IsHours()).To(BeFalse())
		})

		It("should return true for IsMinutes when duration is >= 1 minute", func() {
			// TC-MD-019
			Expect(durbig.Minutes(1).IsMinutes()).To(BeTrue())
			Expect(durbig.Minutes(2).IsMinutes()).To(BeTrue())
		})

		It("should return false for IsMinutes when duration is < 1 minute", func() {
			// TC-MD-020
			Expect(durbig.Seconds(59).IsMinutes()).To(BeFalse())
		})

		It("should return true for IsSeconds when duration is >= 1 second", func() {
			// TC-MD-021
			Expect(durbig.Seconds(1).IsSeconds()).To(BeTrue())
			Expect(durbig.Seconds(2).IsSeconds()).To(BeTrue())
		})

		It("should return false for IsSeconds when duration is < 1 second", func() {
			// TC-MD-022
			// Duration is an integer of seconds, so this can only happen for 0 or negative
			Expect(durbig.Duration(0).IsSeconds()).To(BeFalse())
			Expect(durbig.Duration(-1).IsSeconds()).To(BeFalse())
		})
	})

	Describe("ViperDecoderHook", func() {
		var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

		BeforeEach(func() {
			hook = durbig.ViperDecoderHook()
		})

		It("should create valid decoder hook", func() {
			// TC-MD-001
			Expect(hook).ToNot(BeNil())
		})

		It("should decode string to Duration", func() {
			// TC-MD-002
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "5h30m"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeAssignableToTypeOf(durbig.Duration(0)))

			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should decode duration with days", func() {
			// TC-MD-003
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "2d12h"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			expected := durbig.Days(2) + durbig.Hours(12)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})

		It("should pass through non-string types", func() {
			// TC-MD-004
			fromType := reflect.TypeOf(0)
			toType := reflect.TypeOf(durbig.Duration(0))
			data := 12345

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when target is not Duration", func() {
			// TC-MD-005
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(0)
			data := "5h30m"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when data is not string type", func() {
			// TC-MD-006
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := 12345 // Not a string

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should return error for invalid duration string", func() {
			// TC-MD-007
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "invalid"

			_, err := hook(fromType, toType, data)

			Expect(err).To(HaveOccurred())
		})

		It("should handle zero duration", func() {
			// TC-MD-008
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "0s"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(durbig.Seconds(0)))
		})

		It("should handle negative duration", func() {
			// TC-MD-009
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "-5h"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d.Int64()).To(Equal(durbig.Hours(-5).Int64()))
		})

		It("should handle complex duration strings", func() {
			// TC-MD-010
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "5d23h15m13s"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			expected := durbig.Days(5) + durbig.Hours(23) + durbig.Minutes(15) + durbig.Seconds(13)
			Expect(d.Int64()).To(Equal(expected.Int64()))
		})

		It("should handle duration strings with spaces", func() {
			// TC-MD-011
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := " 5h 30m "

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should handle duration strings with quotes", func() {
			// TC-MD-012
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := `"5h30m"`

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(durbig.Hours(5) + durbig.Minutes(30)))
		})

		It("should handle all supported units", func() {
			// TC-MD-013
			units := map[string]durbig.Duration{
				"10s": durbig.Seconds(10),
				"5m":  durbig.Minutes(5),
				"2h":  durbig.Hours(2),
				"1d":  durbig.Days(1),
				"7d":  durbig.Days(7),
			}

			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))

			for input, expected := range units {
				result, err := hook(fromType, toType, input)
				Expect(err).ToNot(HaveOccurred(), "for input "+input)

				d, ok := result.(durbig.Duration)
				Expect(ok).To(BeTrue(), "for input "+input)
				Expect(d.Int64()).To(Equal(expected.Int64()), "for input "+input)
			}
		})

		It("should handle empty string as error", func() {
			// TC-MD-014
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := ""

			_, err := hook(fromType, toType, data)

			Expect(err).To(HaveOccurred())
		})

		It("should handle very large durations", func() {
			// TC-MD-023
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "10000d"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(durbig.Days(10000)))
		})

		It("should handle fractional duration strings", func() {
			// TC-MD-024
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "1.5h"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(durbig.Duration)
			Expect(ok).To(BeTrue())
			Expect(d.Int64()).To(Equal(int64(5400))) // 1.5 hours = 5400 seconds
		})
	})
})

func BenchmarkViperDecoderHook(b *testing.B) {
	// TC-BM-001
	hook := durbig.ViperDecoderHook()
	fromType := reflect.TypeOf("")
	toType := reflect.TypeOf(durbig.Duration(0))
	data := "5d23h15m13s"

	for i := 0; i < b.N; i++ {
		_, _ = hook(fromType, toType, data)
	}
}

func BenchmarkIsDays(b *testing.B) {
	// TC-BM-002
	d := durbig.Days(5)
	for i := 0; i < b.N; i++ {
		_ = d.IsDays()
	}
}

func BenchmarkIsHours(b *testing.B) {
	// TC-BM-003
	d := durbig.Hours(5)
	for i := 0; i < b.N; i++ {
		_ = d.IsHours()
	}
}

func BenchmarkIsMinutes(b *testing.B) {
	// TC-BM-004
	d := durbig.Minutes(5)
	for i := 0; i < b.N; i++ {
		_ = d.IsMinutes()
	}
}

func BenchmarkIsSeconds(b *testing.B) {
	// TC-BM-005
	d := durbig.Seconds(5)
	for i := 0; i < b.N; i++ {
		_ = d.IsSeconds()
	}
}
