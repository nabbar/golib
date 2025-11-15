/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package console_test

import (
	. "github.com/nabbar/golib/console"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Handling", func() {
	Describe("Error Code Constants", func() {
		It("should define ErrorParamEmpty", func() {
			Expect(ErrorParamEmpty).To(BeNumerically(">", 0))
		})

		It("should define ErrorColorIOFprintf", func() {
			Expect(ErrorColorIOFprintf).To(BeNumerically(">", 0))
		})

		It("should define ErrorColorBufWrite", func() {
			Expect(ErrorColorBufWrite).To(BeNumerically(">", 0))
		})

		It("should define ErrorColorBufUndefined", func() {
			Expect(ErrorColorBufUndefined).To(BeNumerically(">", 0))
		})

		It("should have errors in ascending order", func() {
			Expect(ErrorParamEmpty).To(BeNumerically("<", ErrorColorIOFprintf))
			Expect(ErrorColorIOFprintf).To(BeNumerically("<", ErrorColorBufWrite))
			Expect(ErrorColorBufWrite).To(BeNumerically("<", ErrorColorBufUndefined))
		})

		It("should have unique error codes", func() {
			codes := []interface{}{
				ErrorParamEmpty,
				ErrorColorIOFprintf,
				ErrorColorBufWrite,
				ErrorColorBufUndefined,
			}

			// Check all codes are different
			for i := 0; i < len(codes); i++ {
				for j := i + 1; j < len(codes); j++ {
					Expect(codes[i]).ToNot(Equal(codes[j]))
				}
			}
		})
	})

	Describe("Error Messages", func() {
		It("should provide message for ErrorParamEmpty", func() {
			err := ErrorParamEmpty.Error()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
		})

		It("should provide message for ErrorColorIOFprintf", func() {
			err := ErrorColorIOFprintf.Error()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("cannot write on IO"))
		})

		It("should provide message for ErrorColorBufWrite", func() {
			err := ErrorColorBufWrite.Error()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("cannot write on buffer"))
		})

		It("should provide message for ErrorColorBufUndefined", func() {
			err := ErrorColorBufUndefined.Error()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("buffer is not defined"))
		})
	})

	Describe("Error Usage", func() {
		Context("ErrorColorBufUndefined", func() {
			It("should be returned when buffer is nil", func() {
				_, err := ColorPrint.BuffPrintf(nil, "test")
				Expect(err).ToNot(BeNil())
				Expect(err).To(MatchError(ContainSubstring("buffer is not defined")))
			})

			It("should have meaningful error message", func() {
				_, err := ColorPrint.BuffPrintf(nil, "test")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("buffer"))
				Expect(err.Error()).To(ContainSubstring("not defined"))
			})
		})

		Context("Error consistency", func() {
			It("should return same error for same condition", func() {
				_, err1 := ColorPrint.BuffPrintf(nil, "test1")
				_, err2 := ColorPrint.BuffPrintf(nil, "test2")

				Expect(err1).To(Equal(err2))
			})
		})
	})

	Describe("Error creation", func() {
		It("should create error from code", func() {
			err := ErrorParamEmpty.Error()
			Expect(err).ToNot(BeNil())
		})

		It("should create error with nil parent", func() {
			err := ErrorParamEmpty.Error(nil)
			Expect(err).ToNot(BeNil())
		})

		It("should create independent error instances", func() {
			err1 := ErrorParamEmpty.Error()
			err2 := ErrorParamEmpty.Error()

			// Should be equal in content but may be different instances
			Expect(err1.Error()).To(Equal(err2.Error()))
		})
	})
})
