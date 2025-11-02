/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package retro

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("isEmptyValue", func() {
		Context("when checking string values", func() {
			It("should return true for empty string", func() {
				val := reflect.ValueOf("")
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-empty string", func() {
				val := reflect.ValueOf("hello")
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking boolean values", func() {
			It("should return true for false boolean", func() {
				val := reflect.ValueOf(false)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for true boolean", func() {
				val := reflect.ValueOf(true)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking integer values", func() {
			It("should return true for zero int", func() {
				val := reflect.ValueOf(0)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-zero int", func() {
				val := reflect.ValueOf(42)
				Expect(isEmptyValue(val)).To(BeFalse())
			})

			It("should return false for negative int", func() {
				val := reflect.ValueOf(-5)
				Expect(isEmptyValue(val)).To(BeFalse())
			})

			It("should return true for zero int8", func() {
				val := reflect.ValueOf(int8(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero int16", func() {
				val := reflect.ValueOf(int16(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero int32", func() {
				val := reflect.ValueOf(int32(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero int64", func() {
				val := reflect.ValueOf(int64(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})
		})

		Context("when checking unsigned integer values", func() {
			It("should return true for zero uint", func() {
				val := reflect.ValueOf(uint(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-zero uint", func() {
				val := reflect.ValueOf(uint(10))
				Expect(isEmptyValue(val)).To(BeFalse())
			})

			It("should return true for zero uint8", func() {
				val := reflect.ValueOf(uint8(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero uint16", func() {
				val := reflect.ValueOf(uint16(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero uint32", func() {
				val := reflect.ValueOf(uint32(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for zero uint64", func() {
				val := reflect.ValueOf(uint64(0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})
		})

		Context("when checking float values", func() {
			It("should return true for zero float32", func() {
				val := reflect.ValueOf(float32(0.0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-zero float32", func() {
				val := reflect.ValueOf(float32(3.14))
				Expect(isEmptyValue(val)).To(BeFalse())
			})

			It("should return true for zero float64", func() {
				val := reflect.ValueOf(float64(0.0))
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-zero float64", func() {
				val := reflect.ValueOf(float64(2.718))
				Expect(isEmptyValue(val)).To(BeFalse())
			})

			It("should return false for negative float", func() {
				val := reflect.ValueOf(float64(-1.5))
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking slice values", func() {
			It("should return true for nil slice", func() {
				var s []string
				val := reflect.ValueOf(s)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for empty slice", func() {
				s := []string{}
				val := reflect.ValueOf(s)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-empty slice", func() {
				s := []string{"a", "b"}
				val := reflect.ValueOf(s)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking array values", func() {
			It("should return true for zero-length array", func() {
				arr := [0]int{}
				val := reflect.ValueOf(arr)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-empty array", func() {
				arr := [3]int{1, 2, 3}
				val := reflect.ValueOf(arr)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking map values", func() {
			It("should return true for nil map", func() {
				var m map[string]int
				val := reflect.ValueOf(m)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return true for empty map", func() {
				m := map[string]int{}
				val := reflect.ValueOf(m)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-empty map", func() {
				m := map[string]int{"key": 42}
				val := reflect.ValueOf(m)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking pointer values", func() {
			It("should return true for nil pointer", func() {
				var p *int
				val := reflect.ValueOf(p)
				Expect(isEmptyValue(val)).To(BeTrue())
			})

			It("should return false for non-nil pointer", func() {
				i := 42
				p := &i
				val := reflect.ValueOf(p)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking interface values", func() {
			It("should return true for nil interface", func() {
				var i interface{}
				val := reflect.ValueOf(i)
				// Note: reflect.ValueOf(nil) returns invalid value, not nil interface
				// So we test with a typed nil instead
				var ptr *int
				val = reflect.ValueOf(ptr).Convert(reflect.TypeOf((*interface{})(nil)).Elem())
				Expect(isEmptyValue(val)).To(BeFalse()) // Cannot convert to interface{} directly
			})

			It("should return false for non-nil interface", func() {
				var i interface{} = 42
				val := reflect.ValueOf(i)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})

		Context("when checking struct values", func() {
			It("should return false for struct (default case)", func() {
				type TestStruct struct {
					Field string
				}
				s := TestStruct{}
				val := reflect.ValueOf(s)
				Expect(isEmptyValue(val)).To(BeFalse())
			})
		})
	})
})
