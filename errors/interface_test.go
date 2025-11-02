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

	. "github.com/nabbar/golib/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interface Functions", func() {
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

	Describe("Is", func() {
		It("should detect Error interface", func() {
			err := TestErrorCode1.Error(nil)
			Expect(Is(err)).To(BeTrue())
		})

		It("should return false for standard errors", func() {
			stdErr := errors.New("standard error")
			Expect(Is(stdErr)).To(BeFalse())
		})

		It("should return false for nil", func() {
			Expect(Is(nil)).To(BeFalse())
		})
	})

	Describe("Get", func() {
		It("should get Error interface from error", func() {
			err := TestErrorCode1.Error(nil)
			result := Get(err)
			Expect(result).ToNot(BeNil())
			Expect(result.IsCode(TestErrorCode1)).To(BeTrue())
		})

		It("should return nil for standard errors", func() {
			stdErr := errors.New("standard error")
			result := Get(stdErr)
			Expect(result).To(BeNil())
		})

		It("should return nil for nil error", func() {
			result := Get(nil)
			Expect(result).To(BeNil())
		})
	})

	Describe("Has", func() {
		It("should detect code in error chain", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			Expect(Has(err, TestErrorCode2)).To(BeTrue())
		})

		It("should return false for non-existent code", func() {
			err := TestErrorCode1.Error(nil)
			Expect(Has(err, TestErrorCode2)).To(BeFalse())
		})

		It("should return false for nil error", func() {
			Expect(Has(nil, TestErrorCode1)).To(BeFalse())
		})

		It("should detect own code", func() {
			err := TestErrorCode1.Error(nil)
			Expect(Has(err, TestErrorCode1)).To(BeTrue())
		})
	})

	Describe("ContainsString", func() {
		It("should find string in error message", func() {
			err := TestErrorCode1.Error(nil)
			Expect(ContainsString(err, "test error")).To(BeTrue())
		})

		It("should return false for non-existent string", func() {
			err := TestErrorCode1.Error(nil)
			Expect(ContainsString(err, "not found")).To(BeFalse())
		})

		It("should return false for nil error", func() {
			Expect(ContainsString(nil, "test")).To(BeFalse())
		})

		It("should search in standard errors", func() {
			stdErr := errors.New("standard error message")
			Expect(ContainsString(stdErr, "standard")).To(BeTrue())
		})
	})

	Describe("IsCode", func() {
		It("should detect code in error", func() {
			err := TestErrorCode1.Error(nil)
			Expect(IsCode(err, TestErrorCode1)).To(BeTrue())
		})

		It("should return false for different code", func() {
			err := TestErrorCode1.Error(nil)
			Expect(IsCode(err, TestErrorCode2)).To(BeFalse())
		})

		It("should return false for nil error", func() {
			Expect(IsCode(nil, TestErrorCode1)).To(BeFalse())
		})

		It("should return false for standard errors", func() {
			stdErr := errors.New("standard error")
			Expect(IsCode(stdErr, TestErrorCode1)).To(BeFalse())
		})
	})

	Describe("Interface compatibility", func() {
		It("should verify Return interface is defined", func() {
			// This is a compile-time check that Return interface exists
			r := NewDefaultReturn()
			var _ Return = r
			Expect(r).ToNot(BeNil())
		})

		It("should verify ReturnGin interface is defined", func() {
			// This is a compile-time check that ReturnGin interface exists
			r := NewDefaultReturn()
			var _ ReturnGin = r
			Expect(r).ToNot(BeNil())
		})

		It("Error interface should be implemented", func() {
			err := TestErrorCode1.Error(nil)
			var _ error = err
			var _ Error = err
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("nil error handling", func() {
		It("should handle nil errors in various functions", func() {
			Expect(Get(nil)).To(BeNil())
			Expect(Has(nil, TestErrorCode1)).To(BeFalse())
			Expect(ContainsString(nil, "test")).To(BeFalse())
			Expect(IsCode(nil, TestErrorCode1)).To(BeFalse())
		})

		It("should handle AddOrNew with both nil errors", func() {
			result := AddOrNew(nil, nil)
			Expect(result).To(BeNil())
		})
	})
})
