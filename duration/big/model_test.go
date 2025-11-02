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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	durbig "github.com/nabbar/golib/duration/big"
)

var _ = Describe("Big Duration Model Operations", func() {
	Describe("ViperDecoderHook", func() {
		var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

		BeforeEach(func() {
			hook = durbig.ViperDecoderHook()
		})

		It("should create valid decoder hook", func() {
			Expect(hook).ToNot(BeNil())
		})

		It("should decode string to Duration", func() {
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
			fromType := reflect.TypeOf(0)
			toType := reflect.TypeOf(durbig.Duration(0))
			data := 12345

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when target is not Duration", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(0)
			data := "5h30m"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when data is not string type", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := 12345 // Not a string

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should return error for invalid duration string", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := "invalid"

			_, err := hook(fromType, toType, data)

			Expect(err).To(HaveOccurred())
		})

		It("should handle zero duration", func() {
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
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(durbig.Duration(0))
			data := ""

			_, err := hook(fromType, toType, data)

			Expect(err).To(HaveOccurred())
		})

		It("should handle very large durations", func() {
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
