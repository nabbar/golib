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
 */

package router_test

import (
	librtr "github.com/nabbar/golib/router"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router/Errors", func() {
	Describe("Error Codes", func() {
		It("should have ErrorParamEmpty code", func() {
			err := librtr.ErrorParamEmpty.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
		})

		It("should have ErrorConfigValidator code", func() {
			err := librtr.ErrorConfigValidator.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid config, validation error"))
		})

		It("should have ErrorHeaderAuthMissing code", func() {
			err := librtr.ErrorHeaderAuthMissing.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("missing authorization header"))
		})

		It("should have ErrorHeaderAuthEmpty code", func() {
			err := librtr.ErrorHeaderAuthEmpty.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("authorization header is empty"))
		})

		It("should have ErrorHeaderAuthRequire code", func() {
			err := librtr.ErrorHeaderAuthRequire.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("authorization check failed"))
		})

		It("should have ErrorHeaderAuthForbidden code", func() {
			err := librtr.ErrorHeaderAuthForbidden.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("authorization check success but unauthorized client"))
		})

		It("should have ErrorHeaderAuth code", func() {
			err := librtr.ErrorHeaderAuth.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("authorization check return an invalid response code"))
		})
	})

	Describe("Error Chaining", func() {
		It("should chain errors correctly", func() {
			baseErr := librtr.ErrorParamEmpty.Error(nil)
			chainedErr := librtr.ErrorConfigValidator.Error(baseErr)

			Expect(chainedErr).ToNot(BeNil())
			Expect(chainedErr.Error()).To(ContainSubstring("invalid config"))
		})
	})
})
