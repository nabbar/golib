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

package viper_test

import (
	libvpr "github.com/nabbar/golib/viper"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Handling", func() {
	Describe("Error Codes", func() {
		It("should have ErrorParamEmpty defined", func() {
			err := libvpr.ErrorParamEmpty.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorParamEmpty))
		})

		It("should have ErrorParamMissing defined", func() {
			err := libvpr.ErrorParamMissing.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorParamMissing))
		})

		It("should have ErrorHomePathNotFound defined", func() {
			err := libvpr.ErrorHomePathNotFound.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorHomePathNotFound))
		})

		It("should have ErrorBasePathNotFound defined", func() {
			err := libvpr.ErrorBasePathNotFound.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorBasePathNotFound))
		})

		It("should have ErrorRemoteProvider defined", func() {
			err := libvpr.ErrorRemoteProvider.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorRemoteProvider))
		})

		It("should have ErrorRemoteProviderSecure defined", func() {
			err := libvpr.ErrorRemoteProviderSecure.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorRemoteProviderSecure))
		})

		It("should have ErrorRemoteProviderRead defined", func() {
			err := libvpr.ErrorRemoteProviderRead.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorRemoteProviderRead))
		})

		It("should have ErrorRemoteProviderMarshall defined", func() {
			err := libvpr.ErrorRemoteProviderMarshall.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorRemoteProviderMarshall))
		})

		It("should have ErrorConfigRead defined", func() {
			err := libvpr.ErrorConfigRead.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorConfigRead))
		})

		It("should have ErrorConfigReadDefault defined", func() {
			err := libvpr.ErrorConfigReadDefault.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorConfigReadDefault))
		})

		It("should have ErrorConfigIsDefault defined", func() {
			err := libvpr.ErrorConfigIsDefault.Error(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.GetCode()).To(Equal(libvpr.ErrorConfigIsDefault))
		})
	})

	Describe("Error Messages", func() {
		It("should return correct message for ErrorParamMissing", func() {
			err := libvpr.ErrorParamMissing.Error(nil)
			Expect(err.Error()).To(ContainSubstring("parameter is missing"))
		})

		It("should return correct message for ErrorHomePathNotFound", func() {
			err := libvpr.ErrorHomePathNotFound.Error(nil)
			Expect(err.Error()).To(ContainSubstring("home path"))
		})

		It("should return correct message for ErrorBasePathNotFound", func() {
			err := libvpr.ErrorBasePathNotFound.Error(nil)
			Expect(err.Error()).To(ContainSubstring("base config path"))
		})

		It("should return correct message for ErrorRemoteProvider", func() {
			err := libvpr.ErrorRemoteProvider.Error(nil)
			Expect(err.Error()).To(ContainSubstring("remote provider"))
		})

		It("should return correct message for ErrorRemoteProviderSecure", func() {
			err := libvpr.ErrorRemoteProviderSecure.Error(nil)
			Expect(err.Error()).To(ContainSubstring("secure remote provider"))
		})

		It("should return correct message for ErrorRemoteProviderRead", func() {
			err := libvpr.ErrorRemoteProviderRead.Error(nil)
			Expect(err.Error()).To(ContainSubstring("read config from remote"))
		})

		It("should return correct message for ErrorRemoteProviderMarshall", func() {
			err := libvpr.ErrorRemoteProviderMarshall.Error(nil)
			Expect(err.Error()).To(ContainSubstring("marshall config"))
		})

		It("should return correct message for ErrorConfigRead", func() {
			err := libvpr.ErrorConfigRead.Error(nil)
			Expect(err.Error()).To(ContainSubstring("read config from file"))
		})

		It("should return correct message for ErrorConfigReadDefault", func() {
			err := libvpr.ErrorConfigReadDefault.Error(nil)
			Expect(err.Error()).To(ContainSubstring("default config"))
		})

		It("should return correct message for ErrorConfigIsDefault", func() {
			err := libvpr.ErrorConfigIsDefault.Error(nil)
			Expect(err.Error()).To(ContainSubstring("default config"))
		})
	})

	Describe("Error Chaining", func() {
		It("should chain errors correctly", func() {
			originalErr := libvpr.ErrorParamMissing.Error(nil)
			chainedErr := libvpr.ErrorConfigRead.Error(originalErr)

			Expect(chainedErr).ToNot(BeNil())
			Expect(chainedErr.GetCode()).To(Equal(libvpr.ErrorConfigRead))
			Expect(chainedErr.HasParent()).To(BeTrue())
		})

		It("should preserve error code through chain", func() {
			err := libvpr.ErrorRemoteProvider.Error(nil)
			Expect(err.GetCode()).To(Equal(libvpr.ErrorRemoteProvider))
			Expect(err.IsCode(libvpr.ErrorRemoteProvider)).To(BeTrue())
		})
	})

	Describe("Error Code Comparison", func() {
		It("should correctly identify error codes", func() {
			err := libvpr.ErrorParamMissing.Error(nil)
			Expect(err.IsCode(libvpr.ErrorParamMissing)).To(BeTrue())
			Expect(err.IsCode(libvpr.ErrorConfigRead)).To(BeFalse())
		})

		It("should handle HasCode correctly", func() {
			err := libvpr.ErrorRemoteProviderRead.Error(nil)
			Expect(err.HasCode(libvpr.ErrorRemoteProviderRead)).To(BeTrue())
			Expect(err.HasCode(libvpr.ErrorParamEmpty)).To(BeFalse())
		})
	})

	Describe("Error Interface Compliance", func() {
		It("should implement error interface", func() {
			err := libvpr.ErrorConfigRead.Error(nil)
			var _ error = err
			Expect(err.Error()).ToNot(BeEmpty())
		})

		It("should return non-empty error string", func() {
			errors := []error{
				libvpr.ErrorParamEmpty.Error(nil),
				libvpr.ErrorParamMissing.Error(nil),
				libvpr.ErrorHomePathNotFound.Error(nil),
				libvpr.ErrorBasePathNotFound.Error(nil),
				libvpr.ErrorRemoteProvider.Error(nil),
				libvpr.ErrorRemoteProviderSecure.Error(nil),
				libvpr.ErrorRemoteProviderRead.Error(nil),
				libvpr.ErrorRemoteProviderMarshall.Error(nil),
				libvpr.ErrorConfigRead.Error(nil),
				libvpr.ErrorConfigReadDefault.Error(nil),
				libvpr.ErrorConfigIsDefault.Error(nil),
			}

			for _, err := range errors {
				Expect(err.Error()).ToNot(BeEmpty())
			}
		})
	})

	Describe("Error Code Uniqueness", func() {
		It("should have unique error codes", func() {
			codes := []interface{}{
				libvpr.ErrorParamEmpty,
				libvpr.ErrorParamMissing,
				libvpr.ErrorHomePathNotFound,
				libvpr.ErrorBasePathNotFound,
				libvpr.ErrorRemoteProvider,
				libvpr.ErrorRemoteProviderSecure,
				libvpr.ErrorRemoteProviderRead,
				libvpr.ErrorRemoteProviderMarshall,
				libvpr.ErrorConfigRead,
				libvpr.ErrorConfigReadDefault,
				libvpr.ErrorConfigIsDefault,
			}

			seen := make(map[interface{}]bool)
			for _, code := range codes {
				Expect(seen[code]).To(BeFalse(), "Duplicate error code found")
				seen[code] = true
			}
		})
	})
})
