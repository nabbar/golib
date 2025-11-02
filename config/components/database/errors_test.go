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

package database_test

import (
	"errors"

	. "github.com/nabbar/golib/config/components/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liberr "github.com/nabbar/golib/errors"
)

// Error code tests verify error definitions and message handling
var _ = Describe("Error Codes", func() {
	Describe("Error Code Definitions", func() {
		It("should have defined error codes", func() {
			// Verify error codes are defined (non-zero)
			Expect(ErrorParamEmpty).NotTo(BeZero())
			Expect(ErrorParamInvalid).NotTo(BeZero())
			Expect(ErrorComponentNotInitialized).NotTo(BeZero())
			Expect(ErrorConfigInvalid).NotTo(BeZero())
			Expect(ErrorComponentStart).NotTo(BeZero())
			Expect(ErrorComponentReload).NotTo(BeZero())
			Expect(ErrorDependencyLogDefault).NotTo(BeZero())
		})

		It("should have unique error codes", func() {
			codes := []liberr.CodeError{
				ErrorParamEmpty,
				ErrorParamInvalid,
				ErrorComponentNotInitialized,
				ErrorConfigInvalid,
				ErrorComponentStart,
				ErrorComponentReload,
				ErrorDependencyLogDefault,
			}

			// Check uniqueness
			seen := make(map[liberr.CodeError]bool)
			for _, code := range codes {
				Expect(seen[code]).To(BeFalse(), "Duplicate error code found: %d", code)
				seen[code] = true
			}
		})

		It("should have sequential error codes", func() {
			// Error codes should be sequential
			Expect(ErrorParamInvalid).To(Equal(ErrorParamEmpty + 1))
			Expect(ErrorComponentNotInitialized).To(Equal(ErrorParamInvalid + 1))
			Expect(ErrorConfigInvalid).To(Equal(ErrorComponentNotInitialized + 1))
			Expect(ErrorComponentStart).To(Equal(ErrorConfigInvalid + 1))
			Expect(ErrorComponentReload).To(Equal(ErrorComponentStart + 1))
			Expect(ErrorDependencyLogDefault).To(Equal(ErrorComponentReload + 1))
		})
	})

	Describe("Error Messages", func() {
		It("should return correct message for ErrorParamEmpty", func() {
			err := ErrorParamEmpty.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("empty"))
			Expect(msg).To(ContainSubstring("parameter"))
		})

		It("should return correct message for ErrorParamInvalid", func() {
			err := ErrorParamInvalid.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("invalid"))
			Expect(msg).To(ContainSubstring("parameter"))
		})

		It("should return correct message for ErrorComponentNotInitialized", func() {
			err := ErrorComponentNotInitialized.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("not"))
			Expect(msg).To(ContainSubstring("initialized"))
		})

		It("should return correct message for ErrorConfigInvalid", func() {
			err := ErrorConfigInvalid.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("invalid"))
			Expect(msg).To(ContainSubstring("config"))
		})

		It("should return correct message for ErrorComponentStart", func() {
			err := ErrorComponentStart.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("database"))
			Expect(msg).To(ContainSubstring("connection"))
		})

		It("should return correct message for ErrorComponentReload", func() {
			err := ErrorComponentReload.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("database"))
			Expect(msg).To(ContainSubstring("connection"))
		})

		It("should return correct message for ErrorDependencyLogDefault", func() {
			err := ErrorDependencyLogDefault.Error(nil)
			Expect(err).NotTo(BeNil())
			msg := err.Error()
			Expect(msg).To(ContainSubstring("Logger"))
		})
	})

	Describe("Error With Parent", func() {
		It("should include parent error for ErrorParamEmpty", func() {
			parent := errors.New("parent error")
			err := ErrorParamEmpty.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorParamInvalid", func() {
			parent := errors.New("validation failed")
			err := ErrorParamInvalid.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorComponentNotInitialized", func() {
			parent := errors.New("nil component")
			err := ErrorComponentNotInitialized.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorConfigInvalid", func() {
			parent := errors.New("missing DSN")
			err := ErrorConfigInvalid.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorComponentStart", func() {
			parent := errors.New("connection refused")
			err := ErrorComponentStart.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorComponentReload", func() {
			parent := errors.New("cannot reconnect")
			err := ErrorComponentReload.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should include parent error for ErrorDependencyLogDefault", func() {
			parent := errors.New("logger not found")
			err := ErrorDependencyLogDefault.Error(parent)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("Error Without Parent", func() {
		It("should work without parent error", func() {
			codes := []liberr.CodeError{
				ErrorParamEmpty,
				ErrorParamInvalid,
				ErrorComponentNotInitialized,
				ErrorConfigInvalid,
				ErrorComponentStart,
				ErrorComponentReload,
				ErrorDependencyLogDefault,
			}

			for _, code := range codes {
				err := code.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
			}
		})
	})
})

// Error behavior tests
var _ = Describe("Error Behavior", func() {
	Context("when creating errors", func() {
		It("should create independent error instances", func() {
			err1 := ErrorParamEmpty.Error(nil)
			err2 := ErrorParamEmpty.Error(nil)

			Expect(err1).NotTo(BeNil())
			Expect(err2).NotTo(BeNil())
			// They should be different instances
			Expect(err1).NotTo(BeIdenticalTo(err2))
		})

		It("should preserve parent error information", func() {
			parent := errors.New("original error")
			err := ErrorParamEmpty.Error(parent)

			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
			// Error message should contain information about the error
			msg := err.Error()
			Expect(msg).NotTo(BeEmpty())
		})

		It("should handle nil parent gracefully", func() {
			err := ErrorParamEmpty.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeFalse())
		})
	})

	Context("with error chains", func() {
		It("should support error chaining", func() {
			parent := errors.New("root cause")
			err1 := ErrorConfigInvalid.Error(parent)
			err2 := ErrorComponentStart.Error(err1)

			Expect(err2).NotTo(BeNil())
			Expect(err2.HasParent()).To(BeTrue())
		})
	})
})

// Error message formatting tests
var _ = Describe("Error Message Formatting", func() {
	It("should have consistent message format", func() {
		codes := []liberr.CodeError{
			ErrorParamEmpty,
			ErrorParamInvalid,
			ErrorComponentNotInitialized,
			ErrorConfigInvalid,
			ErrorComponentStart,
			ErrorComponentReload,
			ErrorDependencyLogDefault,
		}

		for _, code := range codes {
			err := code.Error(nil)
			msg := err.Error()
			// Messages should be lowercase and descriptive
			Expect(msg).NotTo(BeEmpty())
			Expect(len(msg)).To(BeNumerically(">", 10))
		}
	})

	It("should have descriptive messages", func() {
		// Each message should give enough context
		testCases := []struct {
			code     liberr.CodeError
			keywords []string
		}{
			{ErrorParamEmpty, []string{"parameter", "empty"}},
			{ErrorParamInvalid, []string{"parameter", "invalid"}},
			{ErrorComponentNotInitialized, []string{"component", "initialized"}},
			{ErrorConfigInvalid, []string{"config", "invalid"}},
			{ErrorComponentStart, []string{"database", "connection"}},
			{ErrorComponentReload, []string{"database", "connection"}},
			{ErrorDependencyLogDefault, []string{"Logger"}},
		}

		for _, tc := range testCases {
			err := tc.code.Error(nil)
			msg := err.Error()
			for _, keyword := range tc.keywords {
				Expect(msg).To(ContainSubstring(keyword),
					"Error message should contain '%s'", keyword)
			}
		}
	})
})
