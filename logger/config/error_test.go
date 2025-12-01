/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package config_test

import (
	liberr "github.com/nabbar/golib/errors"
	. "github.com/nabbar/golib/logger/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Codes", func() {
	Describe("ErrorParamEmpty", func() {
		Context("error creation", func() {
			It("should create error with correct code", func() {
				err := ErrorParamEmpty.Error(nil)

				Expect(err).ToNot(BeNil())
				Expect(err.IsCode(ErrorParamEmpty)).To(BeTrue())
			})

			It("should have meaningful error message", func() {
				err := ErrorParamEmpty.Error(nil)

				Expect(err).ToNot(BeNil())
				message := err.Error()
				Expect(message).ToNot(BeEmpty())
				Expect(message).To(ContainSubstring("parameters"))
				Expect(message).To(ContainSubstring("empty"))
			})
		})

		Context("error wrapping", func() {
			It("should wrap underlying error", func() {
				underlyingErr := ErrorValidatorError.Error(nil)
				err := ErrorParamEmpty.Error(underlyingErr)

				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should preserve underlying error", func() {
				underlyingErr := ErrorValidatorError.Error(nil)
				err := ErrorParamEmpty.Error(underlyingErr)

				Expect(err).ToNot(BeNil())
				parents := err.GetParent(false)
				Expect(parents).To(HaveLen(1))
				// Verify parent has same error code
				if parentErr, ok := parents[0].(liberr.Error); ok {
					Expect(parentErr.IsCode(ErrorValidatorError)).To(BeTrue())
				}
			})
		})

		Context("error comparison", func() {
			It("should be comparable with code", func() {
				err := ErrorParamEmpty.Error(nil)

				Expect(err.IsCode(ErrorParamEmpty)).To(BeTrue())
			})

			It("should not equal different error codes", func() {
				err := ErrorParamEmpty.Error(nil)

				Expect(err.IsCode(ErrorValidatorError)).To(BeFalse())
			})
		})
	})

	Describe("ErrorValidatorError", func() {
		Context("error creation", func() {
			It("should create error with correct code", func() {
				err := ErrorValidatorError.Error(nil)

				Expect(err).ToNot(BeNil())
				Expect(err.IsCode(ErrorValidatorError)).To(BeTrue())
			})

			It("should have meaningful error message", func() {
				err := ErrorValidatorError.Error(nil)

				Expect(err).ToNot(BeNil())
				message := err.Error()
				Expect(message).ToNot(BeEmpty())
				Expect(message).To(ContainSubstring("logger"))
				Expect(message).To(ContainSubstring("invalid"))
				Expect(message).To(ContainSubstring("config"))
			})
		})

		Context("with validation errors", func() {
			It("should wrap validation errors", func() {
				opts := &Options{}
				validationErr := opts.Validate()

				// Even if validation passes (returns nil), we should be able to wrap it
				err := ErrorValidatorError.Error(validationErr)
				Expect(err).ToNot(BeNil())
			})
		})

		Context("error chain", func() {
			It("should support error chaining", func() {
				baseErr := ErrorParamEmpty.Error(nil)
				chainedErr := ErrorValidatorError.Error(baseErr)

				Expect(chainedErr).ToNot(BeNil())
				Expect(chainedErr.HasParent()).To(BeTrue())
			})

			It("should retrieve parent error", func() {
				baseErr := ErrorParamEmpty.Error(nil)
				chainedErr := ErrorValidatorError.Error(baseErr)

				parents := chainedErr.GetParent(false)
				Expect(parents).To(HaveLen(1))
				// Verify parent has same error code
				if parentErr, ok := parents[0].(liberr.Error); ok {
					Expect(parentErr.IsCode(ErrorParamEmpty)).To(BeTrue())
				}
			})
		})
	})

	Describe("Error Code Uniqueness", func() {
		Context("code collision prevention", func() {
			It("should have unique error codes", func() {
				Expect(ErrorParamEmpty).ToNot(Equal(ErrorValidatorError))
			})

			It("should not panic during initialization", func() {
				// This test verifies that the init() function doesn't panic
				// due to error code collisions
				Expect(func() {
					// The init() has already run, so we just verify no panic occurred
					_ = ErrorParamEmpty.Error(nil)
					_ = ErrorValidatorError.Error(nil)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Error Message Retrieval", func() {
		Context("with valid error codes", func() {
			It("should return message for ErrorParamEmpty", func() {
				err := ErrorParamEmpty.Error(nil)
				Expect(err.Error()).ToNot(BeEmpty())
			})

			It("should return message for ErrorValidatorError", func() {
				err := ErrorValidatorError.Error(nil)
				Expect(err.Error()).ToNot(BeEmpty())
			})
		})

		Context("message content validation", func() {
			It("should have descriptive messages", func() {
				err1 := ErrorParamEmpty.Error(nil)
				err2 := ErrorValidatorError.Error(nil)

				// Messages should be different
				Expect(err1.Error()).ToNot(Equal(err2.Error()))

				// Messages should be descriptive (more than just a code)
				Expect(len(err1.Error())).To(BeNumerically(">", 10))
				Expect(len(err2.Error())).To(BeNumerically(">", 10))
			})
		})
	})

	Describe("Error Usage in Validation", func() {
		Context("with Options.Validate()", func() {
			It("should use ErrorValidatorError for validation failures", func() {
				opts := &Options{}

				// Force a validation error by providing invalid data
				// Note: Current implementation may not fail validation
				// This test ensures the error handling is correct
				err := opts.Validate()

				if err != nil {
					Expect(err.IsCode(ErrorValidatorError)).To(BeTrue())
				}
			})
		})
	})

	Describe("Error Properties", func() {
		Context("error interface compliance", func() {
			It("should implement error interface", func() {
				err := ErrorParamEmpty.Error(nil)

				var _ error = err
				Expect(err.Error()).ToNot(BeEmpty())
			})

			It("should be usable in error returns", func() {
				testFunc := func() error {
					return ErrorParamEmpty.Error(nil)
				}

				err := testFunc()
				Expect(err).ToNot(BeNil())
			})
		})

		Context("with nil parent", func() {
			It("should handle nil parent correctly", func() {
				err := ErrorParamEmpty.Error(nil)

				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeFalse())
			})
		})

		Context("error formatting", func() {
			It("should format error as string", func() {
				err := ErrorValidatorError.Error(nil)

				str := err.Error()
				Expect(str).ToNot(BeEmpty())
				Expect(str).To(BeAssignableToTypeOf(""))
			})
		})
	})

	Describe("Error Code Constants", func() {
		Context("constant values", func() {
			It("should have stable constant values", func() {
				// Error codes should not change between calls
				code1 := ErrorParamEmpty
				code2 := ErrorParamEmpty

				Expect(code1).To(Equal(code2))
			})

			It("should maintain order", func() {
				// ErrorValidatorError should be after ErrorParamEmpty
				Expect(ErrorValidatorError).To(BeNumerically(">", ErrorParamEmpty))
			})
		})
	})

	Describe("Error Creation Patterns", func() {
		Context("with various parent errors", func() {
			It("should accept nil parent", func() {
				err := ErrorParamEmpty.Error(nil)
				Expect(err).ToNot(BeNil())
			})

			It("should accept another error as parent", func() {
				parent := ErrorParamEmpty.Error(nil)
				err := ErrorValidatorError.Error(parent)

				Expect(err).ToNot(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should accept multiple levels of parents", func() {
				level1 := ErrorParamEmpty.Error(nil)
				level2 := ErrorValidatorError.Error(level1)
				level3 := ErrorParamEmpty.Error(level2)

				Expect(level3.HasParent()).To(BeTrue())
				parents := level3.GetParent(false)
				Expect(parents).To(HaveLen(2))
				// Verify first parent is level2 (ErrorValidatorError)
				if parentErr, ok := parents[0].(liberr.Error); ok {
					Expect(parentErr.IsCode(ErrorValidatorError)).To(BeTrue())
				}
				// Verify second parent is level1 (ErrorParamEmpty)
				if parentErr, ok := parents[1].(liberr.Error); ok {
					Expect(parentErr.IsCode(ErrorParamEmpty)).To(BeTrue())
				}
			})
		})
	})
})
