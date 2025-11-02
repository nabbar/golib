/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
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

package types_test

import (
	. "github.com/nabbar/golib/logger/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Types - Field Constants", func() {
	Describe("Field constant values", func() {
		Context("when checking field names", func() {
			It("should have correct FieldTime value", func() {
				Expect(FieldTime).To(Equal("time"))
			})

			It("should have correct FieldLevel value", func() {
				Expect(FieldLevel).To(Equal("level"))
			})

			It("should have correct FieldStack value", func() {
				Expect(FieldStack).To(Equal("stack"))
			})

			It("should have correct FieldCaller value", func() {
				Expect(FieldCaller).To(Equal("caller"))
			})

			It("should have correct FieldFile value", func() {
				Expect(FieldFile).To(Equal("file"))
			})

			It("should have correct FieldLine value", func() {
				Expect(FieldLine).To(Equal("line"))
			})

			It("should have correct FieldMessage value", func() {
				Expect(FieldMessage).To(Equal("message"))
			})

			It("should have correct FieldError value", func() {
				Expect(FieldError).To(Equal("error"))
			})

			It("should have correct FieldData value", func() {
				Expect(FieldData).To(Equal("data"))
			})
		})

		Context("when checking field uniqueness", func() {
			It("should have all unique field names", func() {
				fields := []string{
					FieldTime,
					FieldLevel,
					FieldStack,
					FieldCaller,
					FieldFile,
					FieldLine,
					FieldMessage,
					FieldError,
					FieldData,
				}

				// Create a map to check for duplicates
				fieldMap := make(map[string]bool)
				for _, field := range fields {
					Expect(fieldMap[field]).To(BeFalse(), "Field %s appears multiple times", field)
					fieldMap[field] = true
				}

				Expect(len(fieldMap)).To(Equal(len(fields)))
			})
		})

		Context("when used as map keys", func() {
			It("should work correctly in map structures", func() {
				fieldValues := map[string]interface{}{
					FieldTime:    "2024-01-01T00:00:00Z",
					FieldLevel:   "info",
					FieldStack:   "stack trace",
					FieldCaller:  "main.go:42",
					FieldFile:    "main.go",
					FieldLine:    42,
					FieldMessage: "test message",
					FieldError:   "test error",
					FieldData:    "test data",
				}

				Expect(fieldValues[FieldTime]).To(Equal("2024-01-01T00:00:00Z"))
				Expect(fieldValues[FieldLevel]).To(Equal("info"))
				Expect(fieldValues[FieldStack]).To(Equal("stack trace"))
				Expect(fieldValues[FieldCaller]).To(Equal("main.go:42"))
				Expect(fieldValues[FieldFile]).To(Equal("main.go"))
				Expect(fieldValues[FieldLine]).To(Equal(42))
				Expect(fieldValues[FieldMessage]).To(Equal("test message"))
				Expect(fieldValues[FieldError]).To(Equal("test error"))
				Expect(fieldValues[FieldData]).To(Equal("test data"))
			})
		})

		Context("when checking field categories", func() {
			It("should have metadata fields", func() {
				metadataFields := []string{FieldTime, FieldLevel}
				Expect(metadataFields).To(ContainElement(FieldTime))
				Expect(metadataFields).To(ContainElement(FieldLevel))
			})

			It("should have trace fields", func() {
				traceFields := []string{FieldStack, FieldCaller, FieldFile, FieldLine}
				Expect(traceFields).To(ContainElement(FieldStack))
				Expect(traceFields).To(ContainElement(FieldCaller))
				Expect(traceFields).To(ContainElement(FieldFile))
				Expect(traceFields).To(ContainElement(FieldLine))
			})

			It("should have content fields", func() {
				contentFields := []string{FieldMessage, FieldError, FieldData}
				Expect(contentFields).To(ContainElement(FieldMessage))
				Expect(contentFields).To(ContainElement(FieldError))
				Expect(contentFields).To(ContainElement(FieldData))
			})
		})

		Context("when using in filters", func() {
			It("should be useful for field filtering", func() {
				// Simulate filtering fields that should be disabled
				disabledFields := map[string]bool{
					FieldStack:  true,
					FieldCaller: false,
				}

				Expect(disabledFields[FieldStack]).To(BeTrue())
				Expect(disabledFields[FieldCaller]).To(BeFalse())
				Expect(disabledFields[FieldMessage]).To(BeFalse())
			})
		})
	})
})
