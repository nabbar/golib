/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package types_test

import (
	"time"

	. "github.com/nabbar/golib/httpserver/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Field Types and Constants", func() {
	Describe("FieldType Constants", func() {
		It("should define FieldName constant", func() {
			Expect(FieldName).To(BeNumerically(">=", 0))
		})

		It("should define FieldBind constant", func() {
			Expect(FieldBind).To(BeNumerically(">", FieldName))
		})

		It("should define FieldExpose constant", func() {
			Expect(FieldExpose).To(BeNumerically(">", FieldBind))
		})

		It("should have unique values for each field type", func() {
			Expect(FieldName).ToNot(Equal(FieldBind))
			Expect(FieldName).ToNot(Equal(FieldExpose))
			Expect(FieldBind).ToNot(Equal(FieldExpose))
		})

		It("should be usable in switch statements", func() {
			testField := FieldName
			var result string

			switch testField {
			case FieldName:
				result = "name"
			case FieldBind:
				result = "bind"
			case FieldExpose:
				result = "expose"
			default:
				result = "unknown"
			}

			Expect(result).To(Equal("name"))
		})

		It("should handle all field types in switch", func() {
			fields := []FieldType{FieldName, FieldBind, FieldExpose}
			results := []string{}

			for _, field := range fields {
				switch field {
				case FieldName:
					results = append(results, "name")
				case FieldBind:
					results = append(results, "bind")
				case FieldExpose:
					results = append(results, "expose")
				}
			}

			Expect(results).To(Equal([]string{"name", "bind", "expose"}))
		})
	})

	Describe("HandlerDefault Constant", func() {
		It("should define default handler name", func() {
			Expect(HandlerDefault).To(Equal("default"))
		})

		It("should be usable as map key", func() {
			handlers := map[string]bool{
				HandlerDefault: true,
			}

			Expect(handlers).To(HaveKey(HandlerDefault))
			Expect(handlers[HandlerDefault]).To(BeTrue())
		})
	})

	Describe("Timeout Constants", func() {
		It("should define TimeoutWaitingPortFreeing", func() {
			Expect(TimeoutWaitingPortFreeing).To(Equal(250 * time.Microsecond))
		})

		It("should define TimeoutWaitingStop", func() {
			Expect(TimeoutWaitingStop).To(Equal(5 * time.Second))
		})

		It("should have reasonable timeout values", func() {
			Expect(TimeoutWaitingPortFreeing).To(BeNumerically(">", 0))
			Expect(TimeoutWaitingStop).To(BeNumerically(">", TimeoutWaitingPortFreeing))
		})

		It("should be usable with time operations", func() {
			start := time.Now()
			time.Sleep(TimeoutWaitingPortFreeing)
			elapsed := time.Since(start)

			Expect(elapsed).To(BeNumerically(">=", TimeoutWaitingPortFreeing))
		})
	})

	Describe("BadHandlerName Constant", func() {
		It("should define bad handler name", func() {
			Expect(BadHandlerName).To(Equal("no handler"))
		})

		It("should be different from HandlerDefault", func() {
			Expect(BadHandlerName).ToNot(Equal(HandlerDefault))
		})

		It("should be usable in comparisons", func() {
			handlerName := "no handler"
			Expect(handlerName).To(Equal(BadHandlerName))
		})
	})

	Describe("FieldType as Custom Type", func() {
		It("should allow variable declaration", func() {
			var field FieldType
			field = FieldName

			Expect(field).To(Equal(FieldName))
		})

		It("should allow comparison", func() {
			field1 := FieldName
			field2 := FieldName
			field3 := FieldBind

			Expect(field1 == field2).To(BeTrue())
			Expect(field1 == field3).To(BeFalse())
		})

		It("should be usable in maps", func() {
			fieldMap := map[FieldType]string{
				FieldName:   "name field",
				FieldBind:   "bind field",
				FieldExpose: "expose field",
			}

			Expect(fieldMap[FieldName]).To(Equal("name field"))
			Expect(fieldMap[FieldBind]).To(Equal("bind field"))
			Expect(fieldMap[FieldExpose]).To(Equal("expose field"))
		})

		It("should be usable in slices", func() {
			fields := []FieldType{FieldName, FieldBind, FieldExpose}

			Expect(fields).To(HaveLen(3))
			Expect(fields[0]).To(Equal(FieldName))
			Expect(fields[1]).To(Equal(FieldBind))
			Expect(fields[2]).To(Equal(FieldExpose))
		})

		It("should support type assertion", func() {
			var field interface{} = FieldName

			ft, ok := field.(FieldType)
			Expect(ok).To(BeTrue())
			Expect(ft).To(Equal(FieldName))
		})
	})

	Describe("Constants Integration", func() {
		It("should use constants together", func() {
			// Simulating usage in filtering
			filterBy := FieldName
			defaultHandler := HandlerDefault
			badHandler := BadHandlerName

			Expect(filterBy).To(Equal(FieldName))
			Expect(defaultHandler).ToNot(Equal(badHandler))
		})

		It("should use timeouts in context", func() {
			portTimeout := TimeoutWaitingPortFreeing
			stopTimeout := TimeoutWaitingStop

			Expect(stopTimeout).To(BeNumerically(">", portTimeout))
		})
	})
})
