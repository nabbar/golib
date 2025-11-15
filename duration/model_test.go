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
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libdur "github.com/nabbar/golib/duration"
)

var _ = Describe("Duration Model Operations", func() {
	Describe("ViperDecoderHook", func() {
		var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

		BeforeEach(func() {
			hook = libdur.ViperDecoderHook()
		})

		It("should create valid decoder hook", func() {
			Expect(hook).ToNot(BeNil())
		})

		It("should decode string to Duration", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "5h30m"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeAssignableToTypeOf(libdur.Duration(0)))

			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(libdur.Hours(5) + libdur.Minutes(30)))
		})

		It("should decode duration with days", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "2d12h"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d.Days()).To(Equal(int64(2)))
		})

		It("should pass through non-string types", func() {
			fromType := reflect.TypeOf(0)
			toType := reflect.TypeOf(libdur.Duration(0))
			data := 12345

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when target is not Duration", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(time.Duration(0))
			data := "5h30m"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should pass through when data is not string type", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := 12345 // Not a string

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should return error for invalid duration string", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "invalid"

			_, err := hook(fromType, toType, data)

			Expect(err).To(HaveOccurred())
		})

		It("should handle zero duration", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "0s"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(libdur.Seconds(0)))
		})

		It("should handle negative duration", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "-5h"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d.Time()).To(Equal(-5 * time.Hour))
		})

		It("should handle complex duration strings", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := "5d23h15m13s"

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			expected := 5*24*time.Hour + 23*time.Hour + 15*time.Minute + 13*time.Second
			Expect(d.Time()).To(Equal(expected))
		})

		It("should handle duration strings with spaces", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := " 5h 30m "

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(libdur.Hours(5) + libdur.Minutes(30)))
		})

		It("should handle duration strings with quotes", func() {
			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))
			data := `"5h30m"`

			result, err := hook(fromType, toType, data)

			Expect(err).ToNot(HaveOccurred())
			d, ok := result.(libdur.Duration)
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(libdur.Hours(5) + libdur.Minutes(30)))
		})

		It("should handle all supported units", func() {
			units := map[string]time.Duration{
				"100ns": 100 * time.Nanosecond,
				"100us": 100 * time.Microsecond,
				"100ms": 100 * time.Millisecond,
				"10s":   10 * time.Second,
				"5m":    5 * time.Minute,
				"2h":    2 * time.Hour,
				"1d":    24 * time.Hour,
			}

			fromType := reflect.TypeOf("")
			toType := reflect.TypeOf(libdur.Duration(0))

			for input, expected := range units {
				result, err := hook(fromType, toType, input)
				Expect(err).ToNot(HaveOccurred(), "for input "+input)

				d, ok := result.(libdur.Duration)
				Expect(ok).To(BeTrue(), "for input "+input)
				Expect(d.Time()).To(Equal(expected), "for input "+input)
			}
		})
	})
})
