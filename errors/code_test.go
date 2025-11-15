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

var _ = Describe("CodeError", func() {
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

	Describe("CodeError type conversions", func() {
		It("Uint16 should return uint16 value", func() {
			code := TestErrorCode1
			Expect(code.Uint16()).To(Equal(uint16(TestErrorCode1)))
		})

		It("Int should return int value", func() {
			code := TestErrorCode1
			Expect(code.Int()).To(Equal(int(TestErrorCode1)))
		})

		It("String should return string value", func() {
			code := TestErrorCode1
			Expect(code.String()).To(Equal("1000"))
		})

		It("Message should return message", func() {
			code := TestErrorCode1
			Expect(code.Message()).To(Equal("test error 1"))
		})

		It("GetMessage (deprecated) should still work", func() {
			code := TestErrorCode1
			// GetMessage is deprecated but should return String()
			Expect(code.GetMessage()).To(Equal(code.String()))
		})
	})

	Describe("CodeError.Error()", func() {
		It("should create Error from CodeError", func() {
			err := TestErrorCode1.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(TestErrorCode1))
		})

		It("should handle error with nil parent", func() {
			err := TestErrorCode1.Error(nil)
			Expect(err.HasParent()).To(BeFalse())
		})

		It("should create error with Error method from unregistered code", func() {
			unregisteredCode := CodeError(55555)
			err := unregisteredCode.Error()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown error"))
		})
	})

	Describe("CodeError.Errorf()", func() {
		It("should create formatted error without placeholders", func() {
			err := TestErrorCode1.Errorf()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("test error 1"))
		})

		It("should create formatted error with placeholders", func() {
			// Register a code with placeholders
			testCode := CodeError(9998)
			RegisterIdFctMessage(testCode, func(code CodeError) string {
				if code == testCode {
					return "error with %s and %d"
				}
				return ""
			})

			err := testCode.Errorf("value", 42)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("value"))
			Expect(err.Error()).To(ContainSubstring("42"))
		})

		It("should handle extra arguments", func() {
			testCode := CodeError(9997)
			RegisterIdFctMessage(testCode, func(code CodeError) string {
				if code == testCode {
					return "error with %s"
				}
				return ""
			})

			// Pass more arguments than placeholders
			err := testCode.Errorf("first", "second", "third")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("first"))
			// Extra arguments should be ignored
		})

		It("should handle message without placeholders", func() {
			err := TestErrorCode1.Errorf("ignored", "arguments")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("test error 1"))
		})

		It("should handle multiple placeholders", func() {
			testCode := CodeError(9996)
			RegisterIdFctMessage(testCode, func(code CodeError) string {
				if code == testCode {
					return "error with %s, %d, and %v"
				}
				return ""
			})

			err := testCode.Errorf("string", 123, true)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("string"))
			Expect(err.Error()).To(ContainSubstring("123"))
			Expect(err.Error()).To(ContainSubstring("true"))
		})

		It("should handle fewer arguments than placeholders", func() {
			testCode := CodeError(9995)
			RegisterIdFctMessage(testCode, func(code CodeError) string {
				if code == testCode {
					return "error with %s and %d"
				}
				return ""
			})

			// Pass fewer arguments than placeholders
			err := testCode.Errorf("only-one")
			Expect(err).ToNot(BeNil())
			// Should still work, fmt.Sprintf will handle missing args
		})

		It("should create error with Errorf from unregistered code", func() {
			unregisteredCode := CodeError(55554)
			err := unregisteredCode.Errorf("arg1", "arg2")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown error"))
		})
	})

	Describe("CodeError.IfError()", func() {
		It("should return nil if no parent error", func() {
			err := TestErrorCode1.IfError(nil)
			Expect(err).To(BeNil())
		})

		It("should return error if parent error exists", func() {
			parentErr := TestErrorCode2.Error(nil)
			err := TestErrorCode1.IfError(parentErr)
			Expect(err).ToNot(BeNil())
			Expect(err.HasParent()).To(BeTrue())
		})
	})

	Describe("ParseCodeError", func() {
		It("should parse positive int to CodeError", func() {
			code := ParseCodeError(1000)
			Expect(code).To(Equal(CodeError(1000)))
		})

		It("should return UnknownError for negative int", func() {
			code := ParseCodeError(-1)
			Expect(code).To(Equal(UnknownError))
		})

		It("should cap at MaxUint16", func() {
			code := ParseCodeError(int64(70000))
			Expect(code.Uint16()).To(Equal(uint16(65535)))
		})
	})

	Describe("NewCodeError", func() {
		It("should create CodeError from uint16", func() {
			code := NewCodeError(1500)
			Expect(code).To(Equal(CodeError(1500)))
			Expect(code.Uint16()).To(Equal(uint16(1500)))
		})

		It("should create UnknownError from zero", func() {
			code := NewCodeError(0)
			Expect(code).To(Equal(UnknownError))
		})

		It("should create CodeError from max uint16", func() {
			code := NewCodeError(65535)
			Expect(code.Uint16()).To(Equal(uint16(65535)))
		})
	})

	Describe("Error registration", func() {
		It("ExistInMapMessage should check registration", func() {
			Expect(ExistInMapMessage(TestErrorCode1)).To(BeTrue())
			Expect(ExistInMapMessage(ParseCodeError(99999))).To(BeFalse())
		})

		It("RegisterIdFctMessage should register error messages", func() {
			newCode := CodeError(9999)
			RegisterIdFctMessage(newCode, func(code CodeError) string {
				if code == newCode {
					return "new error message"
				}
				return ""
			})

			Expect(ExistInMapMessage(newCode)).To(BeTrue())
			Expect(newCode.Message()).To(Equal("new error message"))
		})
	})

	Describe("GetCodePackages", func() {
		It("should return code packages map", func() {
			packages := GetCodePackages("github.com/nabbar/golib")
			Expect(packages).ToNot(BeNil())
			// Should have at least our test codes
			Expect(len(packages)).To(BeNumerically(">=", 1))
		})

		It("should handle empty rootPackage", func() {
			packages := GetCodePackages("")
			Expect(packages).ToNot(BeNil())
		})
	})

	Describe("Error constants and messages", func() {
		It("should have UnknownError constant", func() {
			Expect(UnknownError).To(Equal(CodeError(0)))
		})

		It("should have UnknownMessage constant", func() {
			Expect(UnknownMessage).To(Equal("unknown error"))
		})

		It("should have NullMessage constant", func() {
			Expect(NullMessage).To(Equal(""))
		})

		It("UnknownError Message should return UnknownMessage", func() {
			Expect(UnknownError.Message()).To(Equal(UnknownMessage))
		})

		It("should handle Message for unknown code", func() {
			unknownCode := ParseCodeError(99999)
			Expect(unknownCode.Message()).To(Equal("unknown error"))
		})

		It("should handle UnknownError code Message", func() {
			code := UnknownError
			Expect(code.Message()).To(Equal("unknown error"))
		})
	})
})
