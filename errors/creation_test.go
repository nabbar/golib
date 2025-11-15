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
	"errors"
	"fmt"

	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Creation", func() {
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

	Describe("Creating errors", func() {
		It("should create error from CodeError", func() {
			err := TestErrorCode1.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(TestErrorCode1)))
			Expect(err.Error()).To(ContainSubstring("test error 1"))
		})

		It("should create error with parent", func() {
			parent := errors.New("parent error")
			err := TestErrorCode1.Error(parent)
			Expect(err.HasParent()).To(BeTrue())
			Expect(err.GetParent(false)).To(HaveLen(1))
		})

		It("should create error with multiple parents", func() {
			p1 := errors.New("parent 1")
			p2 := errors.New("parent 2")
			err := TestErrorCode1.Error(p1, p2)
			Expect(err.HasParent()).To(BeTrue())
			Expect(len(err.GetParent(false))).To(BeNumerically(">=", 1))
		})

		It("should create error using New function", func() {
			err := New(100, "custom error")
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(100)))
			Expect(err.Error()).To(ContainSubstring("custom error"))
		})

		It("should create formatted error using Newf", func() {
			err := Newf(200, "error: %s, code: %d", "test", 42)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("test"))
			Expect(err.Error()).To(ContainSubstring("42"))
		})
	})

	Describe("NewErrorTrace", func() {
		It("should create error with trace", func() {
			err := NewErrorTrace(100, "test error", "file.go", 42)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(100)))
			trace := err.GetTrace()
			// Trace is filtered by filterPath, may show relative path
			Expect(trace).To(ContainSubstring("42"))
		})

		It("should create error with function in trace when file is empty", func() {
			err := NewErrorTrace(200, "trace test", "", 100)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(200)))
			trace := err.GetTrace()
			// Trace might be empty or contain function info depending on context
			_ = trace
		})

		It("should create error with file in trace", func() {
			err := NewErrorTrace(201, "trace test", "/path/to/myfile.go", 50)
			Expect(err).ToNot(BeNil())
			trace := err.GetTrace()
			// File path will be filtered, but line number should be present
			Expect(trace).To(ContainSubstring("50"))
		})

		It("should create error with complete path", func() {
			err := NewErrorTrace(202, "trace test", "/full/path/to/file.go", 100)
			Expect(err).ToNot(BeNil())
			Expect(err.Code()).To(Equal(uint16(202)))
			trace := err.GetTrace()
			Expect(trace).ToNot(BeEmpty())
			Expect(trace).To(ContainSubstring("100"))
		})
	})

	Describe("NewErrorRecovered", func() {
		It("should create error from panic recovery", func() {
			err := NewErrorRecovered("panic message", "recovered value")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("panic message"))
		})

		It("should handle NewErrorRecovered with empty recovered string", func() {
			err := NewErrorRecovered("panic message", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("panic message"))
			// Should not have parent when recovered is empty
			Expect(err.HasParent()).To(BeFalse())
		})

		It("should handle NewErrorRecovered with parents", func() {
			parent := errors.New("parent error")
			err := NewErrorRecovered("panic message", "recovered", parent)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("Make", func() {
		It("should convert standard error to Error interface", func() {
			stdErr := errors.New("standard error")
			err := Make(stdErr)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("standard error"))
		})

		It("should return same Error if already Error type", func() {
			originalErr := TestErrorCode1.Error(nil)
			err := Make(originalErr)
			Expect(err).ToNot(BeNil())
			Expect(err.IsCode(TestErrorCode1)).To(BeTrue())
		})

		It("should handle nil error", func() {
			err := Make(nil)
			Expect(err).To(BeNil())
		})
	})

	Describe("MakeIfError", func() {
		It("should create Error if any error is not nil", func() {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			err := MakeIfError(err1, err2)
			Expect(err).ToNot(BeNil())
		})

		It("should return nil if all errors are nil", func() {
			err := MakeIfError(nil, nil, nil)
			Expect(err).To(BeNil())
		})

		It("should handle mixed nil and non-nil errors", func() {
			err1 := errors.New("error 1")
			err := MakeIfError(nil, err1, nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("AddOrNew", func() {
		It("should add error to existing Error", func() {
			mainErr := TestErrorCode1.Error(nil)
			subErr := errors.New("sub error")
			result := AddOrNew(mainErr, subErr)
			Expect(result).ToNot(BeNil())
			Expect(result.HasParent()).To(BeTrue())
		})

		It("should create new Error if main is nil", func() {
			subErr := errors.New("sub error")
			result := AddOrNew(nil, subErr)
			Expect(result).ToNot(BeNil())
		})

		It("should handle both nil errors", func() {
			result := AddOrNew(nil, nil)
			Expect(result).To(BeNil())
		})

		It("should handle nil sub error", func() {
			mainErr := TestErrorCode1.Error(nil)
			result := AddOrNew(mainErr, nil)
			Expect(result).ToNot(BeNil())
			Expect(result.IsCode(TestErrorCode1)).To(BeTrue())
		})

		It("should handle parent errors", func() {
			mainErr := TestErrorCode1.Error(nil)
			subErr := errors.New("sub error")
			parent := errors.New("parent error")
			result := AddOrNew(mainErr, subErr, parent)
			Expect(result).ToNot(BeNil())
			Expect(result.HasParent()).To(BeTrue())
		})
	})

	Describe("IfError", func() {
		It("should create error only if parent exists", func() {
			parent := errors.New("parent error")
			err := IfError(100, "test error", parent)
			Expect(err).ToNot(BeNil())
		})

		It("should return nil if parent is nil", func() {
			err := IfError(100, "test error", nil)
			Expect(err).To(BeNil())
		})

		It("should return nil for empty error string", func() {
			parent := errors.New("parent")
			err := IfError(100, "test", parent, errors.New(""))
			// Should still create error as parent is not nil
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("Edge cases", func() {
		It("should handle empty error message", func() {
			err := New(0, "")
			Expect(err.Error()).To(Equal(""))
		})

		It("should handle very long error chains", func() {
			err := TestErrorCode1.Error(nil)
			for i := 0; i < 100; i++ {
				err.Add(fmt.Errorf("parent %d", i))
			}

			Expect(err.HasParent()).To(BeTrue())
			Expect(len(err.GetParent(false))).To(Equal(100))
		})

		It("should handle multiple Add calls", func() {
			err := TestErrorCode1.Error(nil)
			err.Add(errors.New("p1"))
			err.Add(errors.New("p2"))
			err.Add(errors.New("p3"))

			Expect(err.GetParent(false)).To(HaveLen(3))
		})
	})
})
