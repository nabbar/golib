/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package size_test

import (
	"reflect"

	. "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ViperDecoderHook", func() {
	var hook func(reflect.Type, reflect.Type, interface{}) (interface{}, error)

	BeforeEach(func() {
		hook = ViperDecoderHook()
	})

	Describe("String to Size conversion", func() {
		It("should decode valid size string", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"100MB",
			)
			Expect(err).ToNot(HaveOccurred())

			size, ok := result.(Size)
			Expect(ok).To(BeTrue())
			Expect(size).To(BeNumerically("~", 100*SizeMega, float64(100*SizeMega)*0.01))
		})

		It("should decode various size formats", func() {
			tests := map[string]Size{
				"1KB":   SizeKilo,
				"5MB":   5 * SizeMega,
				"10GB":  10 * SizeGiga,
				"2TB":   2 * SizeTera,
				"1.5MB": Size(1.5 * float64(SizeMega)),
			}

			for input, expected := range tests {
				result, err := hook(
					reflect.TypeOf(""),
					reflect.TypeOf(Size(0)),
					input,
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(BeNumerically("~", expected, float64(expected)*0.05))
			}
		})

		It("should return error for invalid size string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"invalid",
			)
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"",
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Integer to Size conversion", func() {
		Context("Int types", func() {
			It("should decode int", func() {
				result, err := hook(
					reflect.TypeOf(int(0)),
					reflect.TypeOf(Size(0)),
					int(1024),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024)))
			})

			It("should decode int8", func() {
				result, err := hook(
					reflect.TypeOf(int8(0)),
					reflect.TypeOf(Size(0)),
					int8(100),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(100)))
			})

			It("should decode int16", func() {
				result, err := hook(
					reflect.TypeOf(int16(0)),
					reflect.TypeOf(Size(0)),
					int16(1024),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024)))
			})

			It("should decode int32", func() {
				result, err := hook(
					reflect.TypeOf(int32(0)),
					reflect.TypeOf(Size(0)),
					int32(5120),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(5120)))
			})

			It("should decode int64", func() {
				result, err := hook(
					reflect.TypeOf(int64(0)),
					reflect.TypeOf(Size(0)),
					int64(10240),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(10240)))
			})

			It("should handle negative int values", func() {
				result, err := hook(
					reflect.TypeOf(int(0)),
					reflect.TypeOf(Size(0)),
					int(-1024),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024))) // Absolute value
			})
		})

		Context("Uint types", func() {
			It("should decode uint", func() {
				result, err := hook(
					reflect.TypeOf(uint(0)),
					reflect.TypeOf(Size(0)),
					uint(1024),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024)))
			})

			It("should decode uint8", func() {
				result, err := hook(
					reflect.TypeOf(uint8(0)),
					reflect.TypeOf(Size(0)),
					uint8(100),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(100)))
			})

			It("should decode uint16", func() {
				result, err := hook(
					reflect.TypeOf(uint16(0)),
					reflect.TypeOf(Size(0)),
					uint16(1024),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024)))
			})

			It("should decode uint32", func() {
				result, err := hook(
					reflect.TypeOf(uint32(0)),
					reflect.TypeOf(Size(0)),
					uint32(5120),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(5120)))
			})

			It("should decode uint64", func() {
				result, err := hook(
					reflect.TypeOf(uint64(0)),
					reflect.TypeOf(Size(0)),
					uint64(10240),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(10240)))
			})
		})

		Context("Float types", func() {
			It("should decode float32", func() {
				result, err := hook(
					reflect.TypeOf(float32(0)),
					reflect.TypeOf(Size(0)),
					float32(1024.5),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(1024)))
			})

			It("should decode float64", func() {
				result, err := hook(
					reflect.TypeOf(float64(0)),
					reflect.TypeOf(Size(0)),
					float64(5120.7),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(Equal(Size(5120)))
			})

			It("should handle negative float values", func() {
				result, err := hook(
					reflect.TypeOf(float64(0)),
					reflect.TypeOf(Size(0)),
					float64(-1024.5),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				// math.Floor(-1024.5) = -1025, then abs = 1025
				Expect(size).To(Equal(Size(1025)))
			})
		})
	})

	Describe("Byte slice to Size conversion", func() {
		It("should decode byte slice", func() {
			result, err := hook(
				reflect.TypeOf([]byte{}),
				reflect.TypeOf(Size(0)),
				[]byte("10MB"),
			)
			Expect(err).ToNot(HaveOccurred())

			size, ok := result.(Size)
			Expect(ok).To(BeTrue())
			Expect(size).To(BeNumerically("~", 10*SizeMega, float64(10*SizeMega)*0.01))
		})

		It("should handle various byte slice formats", func() {
			tests := map[string]Size{
				"1KB": SizeKilo,
				"5MB": 5 * SizeMega,
				"2GB": 2 * SizeGiga,
			}

			for input, expected := range tests {
				result, err := hook(
					reflect.TypeOf([]byte{}),
					reflect.TypeOf(Size(0)),
					[]byte(input),
				)
				Expect(err).ToNot(HaveOccurred())

				size, ok := result.(Size)
				Expect(ok).To(BeTrue())
				Expect(size).To(BeNumerically("~", expected, float64(expected)*0.01))
			}
		})

		It("should return error for invalid byte slice", func() {
			_, err := hook(
				reflect.TypeOf([]byte{}),
				reflect.TypeOf(Size(0)),
				[]byte("invalid"),
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Pass-through behavior", func() {
		It("should pass through non-Size target types", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(123),
				"test",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("test"))
		})

		It("should pass through non-matching source types", func() {
			result, err := hook(
				reflect.TypeOf(true),
				reflect.TypeOf(Size(0)),
				true,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("should pass through when both types don't match", func() {
			result, err := hook(
				reflect.TypeOf(true),
				reflect.TypeOf(""),
				true,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(true))
		})

		It("should pass through complex types", func() {
			type customType struct {
				Value int
			}

			input := customType{Value: 42}
			result, err := hook(
				reflect.TypeOf(customType{}),
				reflect.TypeOf(Size(0)),
				input,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(input))
		})
	})

	Describe("Edge cases", func() {
		It("should handle zero values", func() {
			result, err := hook(
				reflect.TypeOf(int(0)),
				reflect.TypeOf(Size(0)),
				int(0),
			)
			Expect(err).ToNot(HaveOccurred())

			size, ok := result.(Size)
			Expect(ok).To(BeTrue())
			Expect(size).To(Equal(SizeNul))
		})

		It("should handle nil data appropriately", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(123),
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should handle type mismatches", func() {
			// Data doesn't match the from type
			result, err := hook(
				reflect.TypeOf(int(0)),
				reflect.TypeOf(Size(0)),
				"not an int",
			)
			Expect(err).ToNot(HaveOccurred())
			// Should pass through since type assertion fails
			Expect(result).To(Equal("not an int"))
		})
	})

	Describe("Error handling", func() {
		It("should return error for unparseable string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"not a size",
			)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for empty string", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"",
			)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for string without unit", func() {
			_, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"123",
			)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Type consistency", func() {
		It("should always return Size type for valid conversions", func() {
			inputs := []interface{}{
				int(1024),
				int64(1024),
				uint64(1024),
				float64(1024),
				"1KB",
				[]byte("1KB"),
			}

			for _, input := range inputs {
				result, err := hook(
					reflect.TypeOf(input),
					reflect.TypeOf(Size(0)),
					input,
				)
				Expect(err).ToNot(HaveOccurred())

				_, ok := result.(Size)
				Expect(ok).To(BeTrue(), "Result should be Size type for input: %v", input)
			}
		})

		It("should maintain type safety", func() {
			result, err := hook(
				reflect.TypeOf(""),
				reflect.TypeOf(Size(0)),
				"10MB",
			)
			Expect(err).ToNot(HaveOccurred())

			// Should be able to use as Size
			size, ok := result.(Size)
			Expect(ok).To(BeTrue())
			Expect(size.Uint64()).To(BeNumerically(">", 0))
		})
	})

	Describe("Performance", func() {
		It("should handle repeated conversions", func() {
			for i := 0; i < 1000; i++ {
				_, err := hook(
					reflect.TypeOf(""),
					reflect.TypeOf(Size(0)),
					"10MB",
				)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("should handle various type conversions efficiently", func() {
			for i := 0; i < 100; i++ {
				_, _ = hook(reflect.TypeOf(int(0)), reflect.TypeOf(Size(0)), int(1024))
				_, _ = hook(reflect.TypeOf(int64(0)), reflect.TypeOf(Size(0)), int64(1024))
				_, _ = hook(reflect.TypeOf(uint64(0)), reflect.TypeOf(Size(0)), uint64(1024))
				_, _ = hook(reflect.TypeOf(float64(0)), reflect.TypeOf(Size(0)), float64(1024))
				_, _ = hook(reflect.TypeOf(""), reflect.TypeOf(Size(0)), "1KB")
			}
		})
	})
})
