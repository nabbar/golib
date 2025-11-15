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

package log_test

import (
	"fmt"

	. "github.com/nabbar/golib/config/components/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liberr "github.com/nabbar/golib/errors"
)

// Error handling tests verify error codes and messages for the Log component.
// These tests ensure proper error definition, uniqueness, and message quality.
var _ = Describe("Error Handling", func() {
	Describe("Error Constants", func() {
		Context("error code uniqueness", func() {
			It("should have unique error codes", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorConfigInvalid,
					ErrorComponentNotInitialized,
					ErrorStartLog,
					ErrorReloadLog,
				}

				// Check uniqueness
				seen := make(map[liberr.CodeError]bool)
				for _, code := range codes {
					Expect(seen[code]).To(BeFalse(), fmt.Sprintf("Duplicate error code: %d", code))
					seen[code] = true
				}
			})

			It("should have valid error codes", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorConfigInvalid,
					ErrorComponentNotInitialized,
					ErrorStartLog,
					ErrorReloadLog,
				}

				for _, code := range codes {
					Expect(code).To(BeNumerically(">", 0))
				}
			})
		})

		Context("error messages", func() {
			It("should return correct error messages for defined codes", func() {
				testCases := []struct {
					code     liberr.CodeError
					expected string
				}{
					{ErrorParamEmpty, "parameters is empty"},
					{ErrorParamInvalid, "parameters is invalid"},
					{ErrorConfigInvalid, "invalid config"},
					{ErrorComponentNotInitialized, "not be correctly initialized"},
					{ErrorStartLog, "cannot start Logger"},
					{ErrorReloadLog, "cannot update Logger"},
				}

				for _, tc := range testCases {
					err := tc.code.Error()
					Expect(err.Error()).To(ContainSubstring(tc.expected))
				}
			})
		})

		Context("error wrapping and unwrapping", func() {
			It("should allow wrapping an error", func() {
				innerErr := fmt.Errorf("inner error")
				wrappedErr := ErrorStartLog.Error(innerErr)
				Expect(wrappedErr).To(HaveOccurred())
				Expect(wrappedErr.Error()).To(ContainSubstring("cannot start Logger"))

				par := wrappedErr.GetParent(false)
				for _, e := range par {
					Expect(e.Error()).To(ContainSubstring("inner error"))
				}
			})

			It("should allow unwrapping an error", func() {
				innerErr := fmt.Errorf("inner error")
				wrappedErr := ErrorStartLog.Error(innerErr)
				unwrapped := wrappedErr.Unwrap()
				Expect(unwrapped[len(unwrapped)-1].Error()).To(Equal(innerErr.Error()))
			})

			It("should return nil when unwrapping an error without cause", func() {
				err := ErrorStartLog.IfError(nil)
				Expect(err).To(BeNil())
			})

			It("should check error type using Is", func() {
				err := ErrorStartLog.Error()
				Expect(liberr.Is(err)).To(BeTrue())
			})

			It("should check error type using As", func() {
				err := ErrorStartLog.Error()
				Expect(liberr.Has(err, ErrorStartLog)).To(BeTrue())
				Expect(err.Code()).To(Equal(ErrorStartLog.Uint16()))
			})
		})

		Context("error code behavior", func() {
			It("should return the correct code for a CodeError", func() {
				err := ErrorParamEmpty.Error()
				Expect(err.Code()).To(Equal(ErrorParamEmpty.Uint16()))
			})

			It("should return 0 for a non-CodeError", func() {
				err := fmt.Errorf("simple error")
				Expect(liberr.Get(err)).To(BeNil())
			})

			It("should return the wrapped code", func() {
				innerErr := ErrorParamInvalid.Error()
				wrappedErr := ErrorStartLog.Error(innerErr)
				Expect(wrappedErr.Code()).To(Equal(ErrorStartLog.Uint16()))
			})
		})

		Context("concurrent error creation", func() {
			It("should handle concurrent error creation without race conditions", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						_ = ErrorParamEmpty.Error()
						_ = ErrorParamInvalid.Error()
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})
})
