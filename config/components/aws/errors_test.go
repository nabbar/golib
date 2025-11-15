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

package aws_test

import (
	"fmt"

	. "github.com/nabbar/golib/config/components/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liberr "github.com/nabbar/golib/errors"
)

var _ = Describe("Error Handling", func() {
	Describe("Error Codes", func() {
		It("should have ErrorParamEmpty defined", func() {
			err := ErrorParamEmpty.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("empty"))
		})

		It("should have ErrorParamInvalid defined", func() {
			err := ErrorParamInvalid.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid"))
		})

		It("should have ErrorComponentNotInitialized defined", func() {
			err := ErrorComponentNotInitialized.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("initialized"))
		})

		It("should have ErrorConfigInvalid defined", func() {
			err := ErrorConfigInvalid.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid"))
		})

		It("should have ErrorComponentStart defined", func() {
			err := ErrorComponentStart.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("start"))
		})

		It("should have ErrorComponentReload defined", func() {
			err := ErrorComponentReload.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("reload"))
		})

		It("should have ErrorDependencyLogDefault defined", func() {
			err := ErrorDependencyLogDefault.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("Logger"))
		})
	})

	Describe("Error Wrapping", func() {
		It("should wrap parent errors for ErrorParamEmpty", func() {
			parentErr := fmt.Errorf("parent error")
			err := ErrorParamEmpty.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorParamInvalid", func() {
			parentErr := fmt.Errorf("validation failed")
			err := ErrorParamInvalid.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorComponentNotInitialized", func() {
			parentErr := fmt.Errorf("context is nil")
			err := ErrorComponentNotInitialized.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorConfigInvalid", func() {
			parentErr := fmt.Errorf("missing required field")
			err := ErrorConfigInvalid.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorComponentStart", func() {
			parentErr := fmt.Errorf("connection failed")
			err := ErrorComponentStart.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorComponentReload", func() {
			parentErr := fmt.Errorf("config parse error")
			err := ErrorComponentReload.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should wrap parent errors for ErrorDependencyLogDefault", func() {
			parentErr := fmt.Errorf("logger not available")
			err := ErrorDependencyLogDefault.Error(parentErr)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("Error Messages", func() {
		It("should return descriptive message for ErrorParamEmpty", func() {
			err := ErrorParamEmpty.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("parameters"))
			Expect(msg).To(ContainSubstring("empty"))
		})

		It("should return descriptive message for ErrorParamInvalid", func() {
			err := ErrorParamInvalid.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("parameters"))
			Expect(msg).To(ContainSubstring("invalid"))
		})

		It("should return descriptive message for ErrorComponentNotInitialized", func() {
			err := ErrorComponentNotInitialized.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("component"))
			Expect(msg).To(ContainSubstring("initialized"))
		})

		It("should return descriptive message for ErrorConfigInvalid", func() {
			err := ErrorConfigInvalid.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("config"))
			Expect(msg).To(ContainSubstring("invalid"))
		})

		It("should return descriptive message for ErrorComponentStart", func() {
			err := ErrorComponentStart.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("start"))
			Expect(msg).To(ContainSubstring("component"))
		})

		It("should return descriptive message for ErrorComponentReload", func() {
			err := ErrorComponentReload.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("reload"))
			Expect(msg).To(ContainSubstring("component"))
		})

		It("should return descriptive message for ErrorDependencyLogDefault", func() {
			err := ErrorDependencyLogDefault.Error(nil)
			msg := err.Error()
			Expect(msg).To(ContainSubstring("Logger"))
		})
	})

	Describe("Error Code Uniqueness", func() {
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

			codeMap := make(map[liberr.CodeError]bool)
			for _, code := range codes {
				Expect(codeMap[code]).To(BeFalse(), "Duplicate error code: %v", code)
				codeMap[code] = true
			}
		})
	})

	Describe("Error Code Registration", func() {
		It("should not panic during initialization", func() {
			// This test ensures that the init() function doesn't panic
			// when registering error messages
			Expect(func() {
				_ = ErrorParamEmpty.Error(nil)
			}).NotTo(Panic())
		})

		It("should retrieve registered messages", func() {
			// All error codes should return non-null messages
			err := ErrorParamEmpty.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorParamInvalid.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorComponentNotInitialized.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorConfigInvalid.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorComponentStart.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorComponentReload.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())

			err = ErrorDependencyLogDefault.Error(nil)
			Expect(err.Error()).NotTo(BeEmpty())
		})
	})

	Describe("Error Chaining", func() {
		It("should preserve error chain for multiple wraps", func() {
			baseErr := fmt.Errorf("base error")
			err1 := ErrorParamInvalid.Error(baseErr)
			err2 := ErrorComponentStart.Error(err1)

			Expect(err2).NotTo(BeNil())
			Expect(err2.HasParent()).To(BeTrue())
		})

		It("should handle nil parent gracefully", func() {
			err := ErrorParamEmpty.Error(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.HasParent()).To(BeFalse())
		})
	})

	Describe("Error Context", func() {
		It("should provide context in error messages", func() {
			contextErr := fmt.Errorf("bucket name is empty")
			err := ErrorParamInvalid.Error(contextErr)

			msg := err.Error()
			Expect(msg).To(ContainSubstring("invalid"))
			// Parent error should be included
			Expect(err.HasParent()).To(BeTrue())
		})

		It("should handle complex error chains", func() {
			err1 := fmt.Errorf("level 1 error")
			err2 := ErrorConfigInvalid.Error(err1)
			err3 := ErrorComponentStart.Error(err2)

			Expect(err3).NotTo(BeNil())
			Expect(err3.HasParent()).To(BeTrue())
		})
	})
})

var _ = Describe("Error Edge Cases", func() {
	Context("with nil parent errors", func() {
		It("should not panic with nil parent", func() {
			Expect(func() {
				_ = ErrorParamEmpty.Error(nil)
				_ = ErrorParamInvalid.Error(nil)
				_ = ErrorComponentNotInitialized.Error(nil)
				_ = ErrorConfigInvalid.Error(nil)
				_ = ErrorComponentStart.Error(nil)
				_ = ErrorComponentReload.Error(nil)
				_ = ErrorDependencyLogDefault.Error(nil)
			}).NotTo(Panic())
		})
	})

	Context("with empty parent errors", func() {
		It("should handle empty error messages", func() {
			emptyErr := fmt.Errorf("")
			err := ErrorParamInvalid.Error(emptyErr)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).NotTo(BeEmpty())
		})
	})

	Context("with very long parent errors", func() {
		It("should handle long error messages", func() {
			longMsg := string(make([]byte, 1000))
			for range longMsg {
				longMsg = "x" + longMsg[1:]
			}
			longErr := fmt.Errorf("%s", longMsg)
			err := ErrorConfigInvalid.Error(longErr)
			Expect(err).NotTo(BeNil())
		})
	})
})

var _ = Describe("Error Usage Patterns", func() {
	It("should be usable in if-error checks", func() {
		// Simulate a function returning an error
		testFunc := func() error {
			return ErrorParamEmpty.Error(fmt.Errorf("test error"))
		}

		err := testFunc()
		Expect(err).To(HaveOccurred())
	})

	It("should be usable in error comparison", func() {
		err1 := ErrorParamEmpty.Error(nil)
		err2 := ErrorParamInvalid.Error(nil)

		// Different error codes should produce different errors
		Expect(err1.Error()).NotTo(Equal(err2.Error()))
	})

	It("should provide error codes for programmatic handling", func() {
		// Error codes should be comparable
		Expect(ErrorParamEmpty).NotTo(Equal(ErrorParamInvalid))
		Expect(ErrorComponentStart).NotTo(Equal(ErrorComponentReload))
	})

	It("should support error wrapping patterns", func() {
		originalErr := fmt.Errorf("original error")

		// Wrap with first error code
		wrapped1 := ErrorConfigInvalid.Error(originalErr)
		Expect(wrapped1).NotTo(BeNil())
		Expect(wrapped1.HasParent()).To(BeTrue())

		// Wrap again with different error code
		wrapped2 := ErrorComponentStart.Error(wrapped1)
		Expect(wrapped2).NotTo(BeNil())
		Expect(wrapped2.HasParent()).To(BeTrue())
	})
})
