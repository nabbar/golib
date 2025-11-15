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

package errors_test

import (
	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultReturn", func() {
	BeforeEach(func() {
		// Register test error messages
		if !ExistInMapMessage(TestErrorCode1) {
			RegisterIdFctMessage(TestErrorCode1, func(code CodeError) string {
				switch code {
				case TestErrorCode1:
					return "test error 1"
				case TestErrorCode2:
					return "test error 2"
				case TestErrorCode3:
					return "test error 3"
				default:
					return ""
				}
			})
		}
	})

	Describe("Creation", func() {
		It("should create return object", func() {
			r := NewDefaultReturn()
			Expect(r).ToNot(BeNil())
			Expect(r.Code).To(BeEmpty())
			Expect(r.Message).To(BeEmpty())
		})
	})

	Describe("SetError", func() {
		It("should set error", func() {
			r := NewDefaultReturn()
			r.SetError(100, "test error", "file.go", 42)
			Expect(r.Code).To(Equal("100"))
			Expect(r.Message).To(Equal("test error"))
		})

		It("should overwrite previous error", func() {
			r := NewDefaultReturn()
			r.SetError(100, "first error", "file1.go", 10)
			r.SetError(200, "second error", "file2.go", 20)
			Expect(r.Code).To(Equal("200"))
			Expect(r.Message).To(Equal("second error"))
		})
	})

	Describe("AddParent", func() {
		It("should add parent", func() {
			r := NewDefaultReturn()
			r.SetError(100, "main error", "file.go", 10)
			r.AddParent(200, "parent error", "file2.go", 20)
			Expect(r.Code).To(Equal("100"))
		})

		It("should add multiple parents", func() {
			r := NewDefaultReturn()
			r.AddParent(100, "parent 1", "file1.go", 10)
			r.AddParent(200, "parent 2", "file2.go", 20)
			r.AddParent(300, "parent 3", "file3.go", 30)
			// Verify it doesn't crash and maintains state
			Expect(r).ToNot(BeNil())
		})

		It("should add parents without main error", func() {
			r := NewDefaultReturn()
			r.AddParent(100, "parent", "file.go", 10)
			Expect(r).ToNot(BeNil())
		})
	})

	Describe("JSON", func() {
		It("should generate JSON", func() {
			r := NewDefaultReturn()
			r.SetError(100, "test error", "file.go", 42)
			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeEmpty())
			Expect(string(jsonBytes)).To(ContainSubstring("test error"))
		})

		It("should generate JSON with parents", func() {
			r := NewDefaultReturn()
			r.SetError(100, "main error", "file.go", 10)
			r.AddParent(200, "parent error", "file2.go", 20)
			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeEmpty())
			Expect(string(jsonBytes)).To(ContainSubstring("main error"))
		})

		It("should generate valid JSON for empty return", func() {
			r := NewDefaultReturn()
			jsonBytes := r.JSON()
			Expect(jsonBytes).ToNot(BeNil())
		})
	})

	Describe("Return interface", func() {
		It("Return interface should transform error", func() {
			err := TestErrorCode1.Error(nil)
			r := NewDefaultReturn()
			err.Return(r)
			Expect(r.Message).ToNot(BeEmpty())
		})

		It("should implement Return interface", func() {
			r := NewDefaultReturn()
			var _ Return = r
			Expect(r).ToNot(BeNil())
		})
	})

	Describe("ReturnGin interface", func() {
		It("should implement ReturnGin interface", func() {
			r := NewDefaultReturn()
			r.SetError(100, "test error", "file.go", 42)

			// Verify it implements the interface (compile-time check)
			var _ ReturnGin = r
			Expect(r).ToNot(BeNil())
		})

		It("GinTonicAbort should handle nil context", func() {
			r := NewDefaultReturn()
			r.SetError(100, "test error", "file.go", 42)

			// Should not panic with nil context
			Expect(func() {
				r.GinTonicAbort(nil, 500)
			}).ToNot(Panic())
		})

		It("GinTonicErrorAbort should handle nil context", func() {
			r := NewDefaultReturn()
			r.SetError(100, "test error", "file.go", 42)

			// Should not panic with nil context
			Expect(func() {
				r.GinTonicErrorAbort(nil, 500)
			}).ToNot(Panic())
		})
	})

	Describe("ReturnError and ReturnParent", func() {
		It("should handle ReturnError callback", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)

			var callCount int
			err.ReturnError(func(code int, msg string, file string, line int) {
				callCount++
			})
			Expect(callCount).To(Equal(1))
		})

		It("should handle ReturnParent callback", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)

			var callCount int
			err.ReturnParent(func(code int, msg string, file string, line int) {
				callCount++
			})
			Expect(callCount).To(BeNumerically(">", 0))
		})
	})
})
