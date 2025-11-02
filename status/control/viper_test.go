/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package control_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/status/control"
)

var _ = Describe("Mode/ViperDecoderHook", func() {
	var hook func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error)

	BeforeEach(func() {
		hook = control.ViperDecoderHook()
	})

	Describe("Type checking", func() {
		It("should return data unchanged if target type is not Mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(0)
			data := "should"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should return data unchanged if source type is not string", func() {
			from := reflect.TypeOf(0)
			to := reflect.TypeOf(control.Mode(0))
			data := 123

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})

		It("should return data unchanged if data is not a string", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := 123

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})
	})

	Describe("Valid conversions", func() {
		It("should decode 'should' to Should mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "should"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Should))
		})

		It("should decode 'Should' to Should mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "Should"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Should))
		})

		It("should decode 'must' to Must mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "must"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Must))
		})

		It("should decode 'Must' to Must mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "Must"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Must))
		})

		It("should decode 'anyof' to AnyOf mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "anyof"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.AnyOf))
		})

		It("should decode 'AnyOf' to AnyOf mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "AnyOf"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.AnyOf))
		})

		It("should decode 'quorum' to Quorum mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "quorum"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Quorum))
		})

		It("should decode 'Quorum' to Quorum mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "Quorum"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Quorum))
		})

		It("should decode invalid string to Ignore mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "invalid"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Ignore))
		})

		It("should decode empty string to Ignore mode", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := ""

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Ignore))
		})
	})

	Describe("Case insensitivity", func() {
		It("should handle SHOULD", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "SHOULD"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Should))
		})

		It("should handle MUST", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "MUST"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Must))
		})

		It("should handle ANYOF", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "ANYOF"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.AnyOf))
		})

		It("should handle QUORUM", func() {
			from := reflect.TypeOf("")
			to := reflect.TypeOf(control.Mode(0))
			data := "QUORUM"

			result, err := hook(from, to, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(control.Quorum))
		})
	})
})
