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

package tlsmode_test

import (
	"reflect"

	. "github.com/nabbar/golib/mail/smtp/tlsmode"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TLS Mode Viper Decoder Hook", func() {
	var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

	BeforeEach(func() {
		hook = ViperDecoderHook()
	})

	Describe("Hook Initialization", func() {
		It("should create a valid hook function", func() {
			Expect(hook).ToNot(BeNil())
		})
	})

	Describe("Int Type Conversions", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should decode int values", func() {
			result, err := hook(reflect.TypeOf(int(0)), toType, int(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode int8 values", func() {
			result, err := hook(reflect.TypeOf(int8(0)), toType, int8(2))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})

		It("should decode int16 values", func() {
			result, err := hook(reflect.TypeOf(int16(0)), toType, int16(0))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSNone))
		})

		It("should decode int32 values", func() {
			result, err := hook(reflect.TypeOf(int32(0)), toType, int32(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode int64 values", func() {
			result, err := hook(reflect.TypeOf(int64(0)), toType, int64(2))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})
	})

	Describe("Uint Type Conversions", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should decode uint values", func() {
			result, err := hook(reflect.TypeOf(uint(0)), toType, uint(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode uint8 values", func() {
			result, err := hook(reflect.TypeOf(uint8(0)), toType, uint8(2))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})

		It("should decode uint16 values", func() {
			result, err := hook(reflect.TypeOf(uint16(0)), toType, uint16(0))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSNone))
		})

		It("should decode uint32 values", func() {
			result, err := hook(reflect.TypeOf(uint32(0)), toType, uint32(1))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode uint64 values", func() {
			result, err := hook(reflect.TypeOf(uint64(0)), toType, uint64(2))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})
	})

	Describe("Float Type Conversions", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should decode float32 values", func() {
			result, err := hook(reflect.TypeOf(float32(0)), toType, float32(1.0))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode float64 values", func() {
			result, err := hook(reflect.TypeOf(float64(0)), toType, float64(2.0))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})

		It("should floor fractional float values", func() {
			result, err := hook(reflect.TypeOf(float64(0)), toType, float64(1.9))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})
	})

	Describe("String Type Conversions", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should decode string values", func() {
			result, err := hook(reflect.TypeOf(""), toType, "starttls")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})

		It("should decode 'tls' string", func() {
			result, err := hook(reflect.TypeOf(""), toType, "tls")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStrictTLS))
		})

		It("should decode empty string", func() {
			result, err := hook(reflect.TypeOf(""), toType, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSNone))
		})
	})

	Describe("Byte Slice Conversions", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should decode byte slices", func() {
			result, err := hook(reflect.TypeOf([]byte{}), toType, []byte("starttls"))
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(TLSStartTLS))
		})
	})

	Describe("Non-Target Type Handling", func() {
		It("should pass through data for non-TLSMode target types", func() {
			otherType := reflect.TypeOf("")
			data := "test"
			result, err := hook(reflect.TypeOf(""), otherType, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})
	})

	Describe("Unsupported Type Handling", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should pass through unsupported types without error", func() {
			data := true
			result, err := hook(reflect.TypeOf(true), toType, data)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(data))
		})
	})

	Describe("All Type Roundtrip", func() {
		var toType reflect.Type

		BeforeEach(func() {
			toType = reflect.TypeOf(TLSMode(0))
		})

		It("should correctly decode all valid mode values across types", func() {
			testCases := []struct {
				fromValue interface{}
				expected  TLSMode
			}{
				{int(0), TLSNone},
				{int(1), TLSStartTLS},
				{int(2), TLSStrictTLS},
				{uint(0), TLSNone},
				{uint(1), TLSStartTLS},
				{uint(2), TLSStrictTLS},
				{float64(0.0), TLSNone},
				{float64(1.0), TLSStartTLS},
				{float64(2.0), TLSStrictTLS},
				{"", TLSNone},
				{"starttls", TLSStartTLS},
				{"tls", TLSStrictTLS},
				{[]byte("starttls"), TLSStartTLS},
				{[]byte("tls"), TLSStrictTLS},
			}

			for _, tc := range testCases {
				fromType := reflect.TypeOf(tc.fromValue)
				result, err := hook(fromType, toType, tc.fromValue)
				Expect(err).ToNot(HaveOccurred(), "Failed for value: %v", tc.fromValue)
				Expect(result).To(Equal(tc.expected), "Expected %v for input %v", tc.expected, tc.fromValue)
			}
		})
	})
})
