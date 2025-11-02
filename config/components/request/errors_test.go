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

package request_test

import (
	"fmt"

	. "github.com/nabbar/golib/config/components/request"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liberr "github.com/nabbar/golib/errors"
)

var _ = Describe("Error Handling", func() {
	Describe("Error Constants", func() {
		Context("error code uniqueness", func() {
			It("should have unique error codes", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
					ErrorDependencyTLSDefault,
					ErrorComponentStart,
					ErrorComponentReload,
				}

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
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
					ErrorDependencyTLSDefault,
					ErrorComponentStart,
					ErrorComponentReload,
				}

				for _, code := range codes {
					Expect(code).To(BeNumerically(">", 0))
				}
			})
		})

		Context("error messages", func() {
			It("should return correct error messages", func() {
				testCases := []struct {
					code     liberr.CodeError
					expected string
				}{
					{ErrorParamEmpty, "parameters is empty"},
					{ErrorParamInvalid, "parameters is invalid"},
					{ErrorComponentNotInitialized, "not be correctly initialized"},
					{ErrorConfigInvalid, "invalid config"},
					{ErrorDependencyTLSDefault, "cannot retrieve TLS"},
					{ErrorComponentStart, "cannot start component"},
					{ErrorComponentReload, "cannot restart component"},
				}

				for _, tc := range testCases {
					err := tc.code.Error(nil)
					Expect(err.Error()).To(ContainSubstring(tc.expected))
				}
			})
		})

		Context("error wrapping", func() {
			It("should allow wrapping an error", func() {
				innerErr := fmt.Errorf("inner error")
				wrappedErr := ErrorComponentStart.Error(innerErr)
				Expect(wrappedErr).To(HaveOccurred())

				par := wrappedErr.GetParent(false)
				for _, e := range par {
					Expect(e.Error()).To(ContainSubstring("inner error"))
				}
			})

			It("should allow unwrapping", func() {
				innerErr := fmt.Errorf("inner error")
				wrappedErr := ErrorComponentStart.Error(innerErr)
				unwrapped := wrappedErr.Unwrap()
				Expect(unwrapped[len(unwrapped)-1].Error()).To(Equal(innerErr.Error()))
			})
		})
	})
})
