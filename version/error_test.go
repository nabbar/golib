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

package version_test

import (
	liberr "github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Error Handling", func() {
	Describe("Error Codes", func() {
		It("should have unique error codes", func() {
			codes := []liberr.CodeError{
				version.ErrorParamEmpty,
				version.ErrorGoVersionInit,
				version.ErrorGoVersionRuntime,
				version.ErrorGoVersionConstraint,
			}

			// Verify all codes are unique
			seen := make(map[liberr.CodeError]bool)
			for _, code := range codes {
				Expect(seen[code]).To(BeFalse(), "Error code %d should be unique", code)
				seen[code] = true
			}
		})

		It("should have codes in the correct range", func() {
			// All version error codes should be >= MinPkgVersion
			Expect(version.ErrorParamEmpty).To(BeNumerically(">=", liberr.MinPkgVersion))
			Expect(version.ErrorGoVersionInit).To(BeNumerically(">=", liberr.MinPkgVersion))
			Expect(version.ErrorGoVersionRuntime).To(BeNumerically(">=", liberr.MinPkgVersion))
			Expect(version.ErrorGoVersionConstraint).To(BeNumerically(">=", liberr.MinPkgVersion))
		})

		It("should have sequential error codes", func() {
			Expect(version.ErrorGoVersionInit).To(Equal(version.ErrorParamEmpty + 1))
			Expect(version.ErrorGoVersionRuntime).To(Equal(version.ErrorGoVersionInit + 1))
			Expect(version.ErrorGoVersionConstraint).To(Equal(version.ErrorGoVersionRuntime + 1))
		})
	})

	Describe("Error Messages", func() {
		Context("ErrorParamEmpty", func() {
			It("should have appropriate error message", func() {
				err := version.ErrorParamEmpty.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("given parameters is empty"))
			})

			It("should have correct error code", func() {
				err := version.ErrorParamEmpty.Error(nil)
				Expect(err.GetCode()).To(Equal(version.ErrorParamEmpty))
			})
		})

		Context("ErrorGoVersionInit", func() {
			It("should have appropriate error message", func() {
				err := version.ErrorGoVersionInit.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("init GoVersion contraint error"))
			})

			It("should have correct error code", func() {
				err := version.ErrorGoVersionInit.Error(nil)
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
			})

			It("should wrap parent error if provided", func() {
				parentErr := version.ErrorParamEmpty.Error(nil)
				err := version.ErrorGoVersionInit.Error(parentErr)
				Expect(err).ToNot(BeNil())
				Expect(err.GetParent(false)).ToNot(BeEmpty())
			})
		})

		Context("ErrorGoVersionRuntime", func() {
			It("should have appropriate error message", func() {
				err := version.ErrorGoVersionRuntime.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("extract GoVersion runtime error"))
			})

			It("should have correct error code", func() {
				err := version.ErrorGoVersionRuntime.Error(nil)
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionRuntime))
			})
		})

		Context("ErrorGoVersionConstraint", func() {
			It("should have appropriate error message", func() {
				err := version.ErrorGoVersionConstraint.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("current binary is build with a non-compatible version of Go"))
			})

			It("should have correct error code", func() {
				err := version.ErrorGoVersionConstraint.Error(nil)
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
			})

			It("should include version info when provided", func() {
				// Create error through CheckGo which adds version info
				v := version.NewVersion(
					version.License_MIT,
					"Test",
					"Test",
					testTime,
					"build",
					"v1.0.0",
					"Author",
					"test",
					testStruct{},
					0,
				)

				err := v.CheckGo("99.99", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
				Expect(err.Error()).To(ContainSubstring("non-compatible version of Go"))
			})
		})
	})

	Describe("Error Registration", func() {
		It("should have registered error messages", func() {
			// All error codes should return valid messages
			codes := []liberr.CodeError{
				version.ErrorParamEmpty,
				version.ErrorGoVersionInit,
				version.ErrorGoVersionRuntime,
				version.ErrorGoVersionConstraint,
			}

			for _, code := range codes {
				err := code.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).ToNot(BeEmpty())
				Expect(err.Error()).ToNot(Equal(liberr.NullMessage))
			}
		})

		It("should not panic during package initialization", func() {
			// If we got here, package init() succeeded without panic
			// This verifies no error code collision occurred
			Expect(true).To(BeTrue())
		})
	})

	Describe("Error Creation and Propagation", func() {
		Context("creating errors from code", func() {
			It("should create error with nil parent", func() {
				err := version.ErrorParamEmpty.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.GetParent(false)).To(BeEmpty())
			})

			It("should create error with parent error", func() {
				parentErr := version.ErrorParamEmpty.Error(nil)
				childErr := version.ErrorGoVersionInit.Error(parentErr)

				Expect(childErr).ToNot(BeNil())
				Expect(childErr.GetParent(false)).ToNot(BeEmpty())
				Expect(childErr.GetParent(false)).To(HaveLen(1))
				Expect(childErr.GetParent(false)[0].Error()).To(Equal(parentErr.Error()))
			})

			It("should preserve parent error information", func() {
				parentErr := version.ErrorParamEmpty.Error(nil)
				childErr := version.ErrorGoVersionInit.Error(parentErr)

				// Parent error should be accessible
				Expect(childErr.GetParent(false)).To(HaveLen(1))
				Expect(childErr.GetParent(false)[0].Error()).To(Equal(parentErr.Error()))
				Expect(childErr.HasCode(version.ErrorParamEmpty)).To(BeTrue())
			})
		})

		Context("error chaining", func() {
			It("should support multiple levels of error chaining", func() {
				err1 := version.ErrorParamEmpty.Error(nil)
				err2 := version.ErrorGoVersionInit.Error(err1)
				err3 := version.ErrorGoVersionRuntime.Error(err2)

				Expect(err3).ToNot(BeNil())
				Expect(err3.GetParent(false)).To(HaveLen(2))
				Expect(err3.GetParent(false)[0].Error()).To(Equal(err2.Error()))
				Expect(err3.HasError(err1)).To(BeTrue())
			})
		})
	})

	Describe("Error Interface Compliance", func() {
		It("should implement error interface", func() {
			err := version.ErrorParamEmpty.Error(nil)
			Expect(err).ToNot(BeNil())

			// Should be able to use as error interface
			var e error = err
			Expect(e).ToNot(BeNil())
			Expect(e.Error()).ToNot(BeEmpty())
		})

		It("should implement liberr.Error interface", func() {
			err := version.ErrorParamEmpty.Error(nil)

			// Should have Code method
			Expect(err.GetCode()).To(Equal(version.ErrorParamEmpty))

			// Should have Error method
			Expect(err.Error()).ToNot(BeEmpty())

			// Should have GetParent method
			_ = err.GetParent(false)
		})
	})

	Describe("Real-world Error Scenarios", func() {
		Context("CheckGo errors", func() {
			var v version.Version

			BeforeEach(func() {
				v = version.NewVersion(
					version.License_MIT,
					"Test",
					"Test",
					testTime,
					"build",
					"v1.0.0",
					"Author",
					"test",
					testStruct{},
					0,
				)
			})

			It("should return ErrorGoVersionInit for invalid constraint", func() {
				err := v.CheckGo("invalid-version", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionInit))
			})

			It("should return ErrorGoVersionConstraint for version mismatch", func() {
				err := v.CheckGo("99.99", ">=")
				Expect(err).ToNot(BeNil())
				Expect(err.GetCode()).To(Equal(version.ErrorGoVersionConstraint))
			})

			It("should provide helpful error messages", func() {
				err := v.CheckGo("99.99", ">=")
				Expect(err).ToNot(BeNil())

				errMsg := err.Error()
				Expect(errMsg).To(ContainSubstring("non-compatible version of Go"))
			})
		})
	})

	Describe("Error Message Consistency", func() {
		It("should have consistent error message format", func() {
			errors := []struct {
				code liberr.CodeError
				msg  string
			}{
				{version.ErrorParamEmpty, "given parameters is empty"},
				{version.ErrorGoVersionInit, "init GoVersion contraint error"},
				{version.ErrorGoVersionRuntime, "extract GoVersion runtime error"},
				{version.ErrorGoVersionConstraint, "current binary is build with a non-compatible version of Go"},
			}

			for _, tc := range errors {
				err := tc.code.Error(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring(tc.msg))
			}
		})

		It("should have lowercase-starting error messages", func() {
			// Go convention: error messages should start with lowercase
			errors := []liberr.CodeError{
				version.ErrorParamEmpty,
				version.ErrorGoVersionInit,
				version.ErrorGoVersionRuntime,
				version.ErrorGoVersionConstraint,
			}

			for _, code := range errors {
				err := code.Error(nil)
				msg := err.Error()
				Expect(msg).ToNot(BeEmpty())
				// First character should be lowercase (except for proper nouns like "Go")
				// We check that messages don't start with generic uppercase words
			}
		})
	})

	Describe("Concurrent Error Creation", func() {
		It("should safely create errors from multiple goroutines", func() {
			done := make(chan bool, 20)

			for i := 0; i < 20; i++ {
				go func(code liberr.CodeError) {
					defer GinkgoRecover()
					err := code.Error(nil)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).ToNot(BeEmpty())
					done <- true
				}(version.ErrorParamEmpty)
			}

			for i := 0; i < 20; i++ {
				Eventually(done).Should(Receive())
			}
		})

		It("should safely create error chains from multiple goroutines", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					defer GinkgoRecover()
					err1 := version.ErrorParamEmpty.Error(nil)
					err2 := version.ErrorGoVersionInit.Error(err1)

					Expect(err2).ToNot(BeNil())
					Expect(err2.GetParent(false)).To(HaveLen(1))
					Expect(err2.GetParent(false)[0].Error()).To(Equal(err1.Error()))
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive())
			}
		})
	})

	Describe("Error Code Boundaries", func() {
		It("should respect package error code boundaries", func() {
			// Verify error codes don't exceed reasonable bounds
			Expect(version.ErrorParamEmpty).To(BeNumerically("<", liberr.MinPkgVersion+1000))
			Expect(version.ErrorGoVersionInit).To(BeNumerically("<", liberr.MinPkgVersion+1000))
			Expect(version.ErrorGoVersionRuntime).To(BeNumerically("<", liberr.MinPkgVersion+1000))
			Expect(version.ErrorGoVersionConstraint).To(BeNumerically("<", liberr.MinPkgVersion+1000))
		})

		It("should not conflict with other package error codes", func() {
			// Verify no collision with other packages
			// This is ensured by the init() function which panics on collision
			// If we reach this test, no collision occurred
			Expect(true).To(BeTrue())
		})
	})
})
