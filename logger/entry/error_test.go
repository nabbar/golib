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

package entry_test

import (
	"errors"
	"fmt"

	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Operations", func() {
	Describe("ErrorClean", func() {
		Context("on valid entry", func() {
			It("should clean empty error slice", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorClean()

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should clean error slice with errors", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.ErrorAdd(false, errors.New("error1"), errors.New("error2"))
				result := e.ErrorClean()

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should allow multiple cleanups", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.ErrorAdd(false, errors.New("error1"))
				e = e.ErrorClean()
				e = e.ErrorAdd(false, errors.New("error2"))
				e = e.ErrorClean()

				Expect(e).ToNot(BeNil())
			})
		})
	})

	Describe("ErrorSet", func() {
		Context("with valid error slice", func() {
			It("should set single error", func() {
				e := logent.New(loglvl.InfoLevel)
				errs := []error{errors.New("test error")}
				result := e.ErrorSet(errs)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should set multiple errors", func() {
				e := logent.New(loglvl.InfoLevel)
				errs := []error{
					errors.New("error1"),
					errors.New("error2"),
					errors.New("error3"),
				}
				result := e.ErrorSet(errs)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should accept slice with nil errors", func() {
				e := logent.New(loglvl.InfoLevel)
				errs := []error{
					errors.New("error1"),
					nil,
					errors.New("error2"),
				}
				result := e.ErrorSet(errs)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should replace existing errors", func() {
				e := logent.New(loglvl.InfoLevel)
				errs1 := []error{errors.New("error1")}
				errs2 := []error{errors.New("error2")}

				e = e.ErrorSet(errs1)
				e = e.ErrorSet(errs2)

				Expect(e).ToNot(BeNil())
			})
		})

		Context("with empty or nil slice", func() {
			It("should handle empty slice", func() {
				e := logent.New(loglvl.InfoLevel)
				errs := []error{}
				result := e.ErrorSet(errs)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle nil slice", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorSet(nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})
	})

	Describe("ErrorAdd", func() {
		Context("with cleanNil false", func() {
			It("should add single error", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false, errors.New("test error"))

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should add multiple errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false,
					errors.New("error1"),
					errors.New("error2"),
					errors.New("error3"),
				)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should include nil errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false,
					errors.New("error1"),
					nil,
					errors.New("error2"),
				)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should accumulate errors from multiple calls", func() {
				e := logent.New(loglvl.InfoLevel)
				e = e.ErrorAdd(false, errors.New("error1"))
				e = e.ErrorAdd(false, errors.New("error2"))
				e = e.ErrorAdd(false, errors.New("error3"))

				Expect(e).ToNot(BeNil())
			})

			It("should handle no errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("with cleanNil true", func() {
			It("should skip nil errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(true,
					errors.New("error1"),
					nil,
					errors.New("error2"),
					nil,
				)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should add non-nil errors only", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(true,
					errors.New("error1"),
					errors.New("error2"),
				)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle all nil errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(true, nil, nil, nil)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("with wrapped errors", func() {
			It("should handle fmt.Errorf wrapped errors", func() {
				e := logent.New(loglvl.InfoLevel)
				baseErr := errors.New("base error")
				wrappedErr := fmt.Errorf("wrapped: %w", baseErr)
				result := e.ErrorAdd(false, wrappedErr)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})

			It("should handle multiple wrapped errors", func() {
				e := logent.New(loglvl.InfoLevel)
				err1 := errors.New("error 1")
				err2 := fmt.Errorf("wrapped: %w", err1)
				err3 := fmt.Errorf("double wrapped: %w", err2)
				result := e.ErrorAdd(false, err3)

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})

		Context("on uninitialized error slice", func() {
			It("should initialize and add errors", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false, errors.New("first error"))

				Expect(result).ToNot(BeNil())
				Expect(result).To(Equal(e))
			})
		})
	})

	Describe("Error Method Chaining", func() {
		Context("with valid entry", func() {
			It("should chain error operations", func() {
				result := logent.New(loglvl.InfoLevel).
					ErrorAdd(false, errors.New("error1")).
					ErrorAdd(false, errors.New("error2")).
					ErrorClean().
					ErrorAdd(false, errors.New("error3"))

				Expect(result).ToNot(BeNil())
			})

			It("should chain with ErrorSet", func() {
				errs := []error{errors.New("error1"), errors.New("error2")}
				result := logent.New(loglvl.InfoLevel).
					ErrorSet(errs).
					ErrorAdd(false, errors.New("error3"))

				Expect(result).ToNot(BeNil())
			})

			It("should maintain entry reference", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false, errors.New("error1")).
					ErrorAdd(false, errors.New("error2"))

				Expect(result).To(Equal(e))
			})
		})
	})

	Describe("Error Integration", func() {
		Context("with complete entry setup", func() {
			It("should work with other entry methods", func() {
				e := logent.New(loglvl.ErrorLevel).
					SetLevel(loglvl.WarnLevel).
					ErrorAdd(false, errors.New("warning occurred")).
					DataSet("additional context")

				Expect(e).ToNot(BeNil())
			})

			It("should handle mixed operations", func() {
				errs := []error{errors.New("error1")}
				e := logent.New(loglvl.InfoLevel).
					ErrorSet(errs).
					ErrorAdd(false, errors.New("error2")).
					ErrorClean().
					ErrorAdd(true, errors.New("error3"), nil)

				Expect(e).ToNot(BeNil())
			})
		})
	})

	Describe("Error Edge Cases", func() {
		Context("with various error types", func() {
			It("should handle different error implementations", func() {
				customErr := &customError{msg: "custom error"}
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false,
					errors.New("standard error"),
					customErr,
					fmt.Errorf("formatted error"),
				)

				Expect(result).ToNot(BeNil())
			})

			It("should handle error with special characters", func() {
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false,
					errors.New("error with \n newline"),
					errors.New("error with \t tab"),
					errors.New("error with \"quotes\""),
				)

				Expect(result).ToNot(BeNil())
			})

			It("should handle very long error messages", func() {
				longMsg := string(make([]byte, 10000))
				e := logent.New(loglvl.InfoLevel)
				result := e.ErrorAdd(false, errors.New(longMsg))

				Expect(result).ToNot(BeNil())
			})
		})
	})
})

// customError is a custom error type for testing
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}
