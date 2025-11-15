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

var _ = Describe("Pattern Management", func() {
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

	Describe("Pattern configuration", func() {
		It("should set and get default pattern", func() {
			originalPattern := GetDefaultPattern()

			SetDefaultPattern("custom pattern: %s")
			Expect(GetDefaultPattern()).To(Equal("custom pattern: %s"))

			// Restore original pattern
			SetDefaultPattern(originalPattern)
		})

		It("should set and get default pattern trace", func() {
			originalPattern := GetDefaultPatternTrace()

			SetDefaultPatternTrace("custom trace: %s %s:%d")
			Expect(GetDefaultPatternTrace()).To(Equal("custom trace: %s %s:%d"))

			// Restore original pattern
			SetDefaultPatternTrace(originalPattern)
		})

		It("should have non-empty default patterns", func() {
			pattern := GetDefaultPattern()
			Expect(pattern).ToNot(BeEmpty())

			tracePattern := GetDefaultPatternTrace()
			Expect(tracePattern).ToNot(BeEmpty())
		})
	})

	Describe("Pattern usage", func() {
		It("should apply pattern to error output", func() {
			SetDefaultPattern("[ERROR] %s")
			err := TestErrorCode1.Error(nil)
			output := err.Error()
			Expect(output).ToNot(BeEmpty())

			// Reset pattern
			SetDefaultPattern("")
		})

		It("should apply trace pattern to error output", func() {
			originalPattern := GetDefaultPatternTrace()
			SetDefaultPatternTrace("TRACE: %s at %s:%d")

			err := TestErrorCode1.Error(nil)
			trace := err.GetTrace()
			Expect(trace).ToNot(BeEmpty())

			// Restore pattern
			SetDefaultPatternTrace(originalPattern)
		})
	})
})
