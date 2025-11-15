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

var _ = Describe("Error Mode", func() {
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

	AfterEach(func() {
		// Reset to default mode after each test
		SetModeReturnError(ModeDefault)
	})

	Describe("Mode management", func() {
		It("should set and get mode", func() {
			SetModeReturnError(ModeReturnCodeError)
			Expect(GetModeReturnError()).To(Equal(ModeReturnCodeError))
		})

		It("should have default mode initially", func() {
			mode := GetModeReturnError()
			Expect(mode).ToNot(BeNil())
		})
	})

	Describe("Error Mode All Modes", func() {
		It("should handle ModeDefault", func() {
			SetModeReturnError(ModeDefault)
			err := TestErrorCode1.Error(nil)
			result := err.Error()
			Expect(result).ToNot(BeEmpty())
		})

		It("should handle ModeReturnCode", func() {
			SetModeReturnError(ModeReturnCode)
			err := TestErrorCode1.Error(nil)
			result := err.Error()
			Expect(result).To(ContainSubstring("1000"))
		})

		It("should handle ModeReturnCodeFull", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			SetModeReturnError(ModeReturnCodeFull)
			result := err.Error()
			Expect(result).ToNot(BeEmpty())
		})

		It("should handle ModeReturnCodeError", func() {
			SetModeReturnError(ModeReturnCodeError)
			err := TestErrorCode1.Error(nil)
			result := err.Error()
			Expect(result).To(ContainSubstring("test error 1"))
		})

		It("should handle ModeReturnCodeErrorFull", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			SetModeReturnError(ModeReturnCodeErrorFull)
			result := err.Error()
			Expect(result).To(ContainSubstring("test error 1"))
			Expect(result).To(ContainSubstring("test error 2"))
		})

		It("should handle ModeReturnCodeErrorTrace", func() {
			err := TestErrorCode1.Error(nil)
			SetModeReturnError(ModeReturnCodeErrorTrace)
			result := err.Error()
			Expect(result).ToNot(BeEmpty())
		})

		It("should handle ModeReturnCodeErrorTraceFull", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			SetModeReturnError(ModeReturnCodeErrorTraceFull)
			result := err.Error()
			Expect(result).ToNot(BeEmpty())
		})

		It("should handle ModeReturnStringError", func() {
			SetModeReturnError(ModeReturnStringError)
			err := TestErrorCode1.Error(nil)
			result := err.Error()
			Expect(result).To(ContainSubstring("test error 1"))
		})

		It("should handle ModeReturnStringErrorFull", func() {
			parent := errors.New("parent error")
			err := TestErrorCode1.Error(parent)
			SetModeReturnError(ModeReturnStringErrorFull)
			result := err.Error()
			Expect(result).To(ContainSubstring("test error 1"))
			Expect(result).To(ContainSubstring("parent error"))
		})
	})

	Describe("ErrorMode String representation", func() {
		It("should return string for all modes", func() {
			modes := []ErrorMode{
				ModeDefault,
				ModeReturnCode,
				ModeReturnCodeFull,
				ModeReturnCodeError,
				ModeReturnCodeErrorFull,
				ModeReturnCodeErrorTrace,
				ModeReturnCodeErrorTraceFull,
				ModeReturnStringError,
				ModeReturnStringErrorFull,
			}
			for _, mode := range modes {
				str := mode.String()
				Expect(str).ToNot(BeEmpty())
			}
		})

		It("should handle unknown mode", func() {
			unknownMode := ErrorMode(99)
			str := unknownMode.String()
			Expect(str).To(Equal("default"))
		})

		It("should return descriptive strings", func() {
			Expect(ModeReturnCode.String()).ToNot(Equal(ModeReturnCodeError.String()))
		})
	})

	Describe("Output formats", func() {
		It("should format Code output", func() {
			err := TestErrorCode1.Error(nil)
			code := err.Code()
			Expect(code).To(Equal(uint16(1000)))
		})

		It("should format CodeSlice output", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			codes := err.CodeSlice()
			Expect(codes).To(ContainElement(uint16(1000)))
		})

		It("should format StringError output", func() {
			err := TestErrorCode1.Error(nil)
			str := err.StringError()
			Expect(str).To(ContainSubstring("test error 1"))
		})

		It("should format StringErrorSlice output", func() {
			parent := errors.New("parent")
			err := TestErrorCode1.Error(parent)
			strs := err.StringErrorSlice()
			Expect(len(strs)).To(BeNumerically(">", 0))
		})

		It("should format CodeError output", func() {
			err := TestErrorCode1.Error(nil)
			output := err.CodeError("")
			Expect(output).To(ContainSubstring("1000"))
			Expect(output).To(ContainSubstring("test error 1"))
		})

		It("should format CodeErrorSlice output", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			outputs := err.CodeErrorSlice("")
			Expect(len(outputs)).To(BeNumerically(">", 0))
		})

		It("should format CodeErrorTrace output", func() {
			err := TestErrorCode1.Error(nil)
			output := err.CodeErrorTrace("")
			Expect(output).ToNot(BeEmpty())
		})

		It("should format CodeErrorTraceSlice output", func() {
			parent := TestErrorCode2.Error(nil)
			err := TestErrorCode1.Error(parent)
			outputs := err.CodeErrorTraceSlice("")
			Expect(len(outputs)).To(BeNumerically(">", 0))
		})
	})
})
