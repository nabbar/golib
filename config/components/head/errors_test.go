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

package head_test

import (
	"fmt"

	. "github.com/nabbar/golib/config/components/head"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	liberr "github.com/nabbar/golib/errors"
)

// Error handling tests verify error codes and messages
var _ = Describe("Error Handling", func() {
	Describe("Error Constants", func() {
		Context("error code uniqueness", func() {
			It("should have unique error codes", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
					ErrorReloadPoolServer,
					ErrorReloadTLSDefault,
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
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
					ErrorReloadPoolServer,
					ErrorReloadTLSDefault,
				}

				for _, code := range codes {
					Expect(code).To(BeNumerically(">", 0))
				}
			})
		})

		Context("error messages", func() {
			It("should have message for ErrorParamEmpty", func() {
				err := ErrorParamEmpty.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("empty"))
			})

			It("should have message for ErrorParamInvalid", func() {
				err := ErrorParamInvalid.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("invalid"))
			})

			It("should have message for ErrorComponentNotInitialized", func() {
				err := ErrorComponentNotInitialized.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("initialized"))
			})

			It("should have message for ErrorConfigInvalid", func() {
				err := ErrorConfigInvalid.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("invalid"))
			})

			It("should have message for ErrorReloadPoolServer", func() {
				err := ErrorReloadPoolServer.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("pool"))
			})

			It("should have message for ErrorReloadTLSDefault", func() {
				err := ErrorReloadTLSDefault.Error(nil)
				Expect(err).NotTo(BeNil())
				msg := err.Error()
				Expect(msg).NotTo(BeEmpty())
				Expect(msg).To(ContainSubstring("TLS"))
			})
		})
	})

	Describe("Error Creation", func() {
		Context("with nil parent error", func() {
			It("should create error without parent", func() {
				err := ErrorParamEmpty.Error(nil)
				Expect(err).NotTo(BeNil())
				Expect(err.HasParent()).To(BeFalse())
			})

			It("should have correct error code", func() {
				err := ErrorParamInvalid.Error(nil)
				Expect(err).NotTo(BeNil())
				Expect(err.Code()).To(Equal(uint16(ErrorParamInvalid)))
			})
		})

		Context("with parent error", func() {
			It("should wrap parent error", func() {
				parentErr := fmt.Errorf("parent error")
				err := ErrorParamInvalid.Error(parentErr)
				Expect(err).NotTo(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should include parent error message", func() {
				parentErr := fmt.Errorf("specific error details")
				err := ErrorComponentNotInitialized.Error(parentErr)
				Expect(err).NotTo(BeNil())
				Expect(err.HasParent()).To(BeTrue())
				// Parent errors are accessible via GetParent()
				parents := err.GetParent(false)
				Expect(parents).To(HaveLen(1))
			})

			It("should maintain error chain", func() {
				parentErr := fmt.Errorf("root cause")
				err1 := ErrorParamInvalid.Error(parentErr)
				err2 := ErrorComponentNotInitialized.Error(err1)

				Expect(err2).NotTo(BeNil())
				Expect(err2.HasParent()).To(BeTrue())
			})
		})
	})

	Describe("Error Usage", func() {
		Context("in component operations", func() {
			It("should panic for SetDependencies on nil component", func() {
				Expect(func() {
					var cpt CptHead
					_ = cpt.SetDependencies([]string{"dep1"})
				}).To(Panic())
			})
		})

		Context("error comparison", func() {
			It("should match error codes", func() {
				err := ErrorParamEmpty.Error(nil)
				Expect(err.Code()).To(Equal(uint16(ErrorParamEmpty)))
			})

			It("should distinguish between different errors", func() {
				err1 := ErrorParamEmpty.Error(nil)
				err2 := ErrorParamInvalid.Error(nil)

				Expect(err1.Code()).NotTo(Equal(err2.Code()))
			})
		})
	})

	Describe("Error Messages Quality", func() {
		Context("message content", func() {
			It("should have descriptive messages", func() {
				errors := map[liberr.CodeError]string{
					ErrorParamEmpty:              "empty",
					ErrorParamInvalid:            "invalid",
					ErrorComponentNotInitialized: "initialized",
					ErrorConfigInvalid:           "config",
					ErrorReloadPoolServer:        "pool",
					ErrorReloadTLSDefault:        "TLS",
				}

				for code, expectedWord := range errors {
					err := code.Error(nil)
					msg := err.Error()
					Expect(msg).To(ContainSubstring(expectedWord),
						fmt.Sprintf("Error code %d should contain '%s' in message", code, expectedWord))
				}
			})

			It("should not return null message", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
					ErrorReloadPoolServer,
					ErrorReloadTLSDefault,
				}

				for _, code := range codes {
					err := code.Error(nil)
					msg := err.Error()
					Expect(msg).NotTo(Equal(liberr.NullMessage))
				}
			})
		})

		Context("message formatting", func() {
			It("should be human-readable", func() {
				err := ErrorComponentNotInitialized.Error(nil)
				msg := err.Error()

				// Should be lowercase (standard Go error convention)
				Expect(msg).NotTo(MatchRegexp("^[A-Z]"))
			})

			It("should not have trailing punctuation", func() {
				codes := []liberr.CodeError{
					ErrorParamEmpty,
					ErrorParamInvalid,
					ErrorComponentNotInitialized,
					ErrorConfigInvalid,
				}

				for _, code := range codes {
					err := code.Error(nil)
					msg := err.Error()
					Expect(msg).NotTo(HaveSuffix("."))
					Expect(msg).NotTo(HaveSuffix("!"))
				}
			})
		})
	})

	Describe("Error Integration", func() {
		Context("with other packages", func() {
			It("should integrate with liberr package", func() {
				err := ErrorParamEmpty.Error(nil)

				// Should implement error interface
				var _ error = err

				// Should have liberr.Error methods
				Expect(err.Code()).To(BeNumerically(">", 0))
				Expect(err.Error()).NotTo(BeEmpty())
			})

			It("should support error wrapping", func() {
				original := fmt.Errorf("original error")
				wrapped := ErrorParamInvalid.Error(original)

				Expect(wrapped.HasParent()).To(BeTrue())
				parents := wrapped.GetParent(false)
				Expect(parents).To(HaveLen(1))
			})
		})

		Context("error propagation", func() {
			It("should propagate through error chain", func() {
				err1 := fmt.Errorf("level 1")
				err2 := ErrorParamInvalid.Error(err1)
				err3 := ErrorComponentNotInitialized.Error(err2)

				Expect(err3.HasParent()).To(BeTrue())
				// Error chain is maintained
				parents := err3.GetParent(true) // recursive
				Expect(len(parents)).To(BeNumerically(">=", 1))
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("error creation", func() {
			It("should handle multiple creations of same error", func() {
				err1 := ErrorParamEmpty.Error(nil)
				err2 := ErrorParamEmpty.Error(nil)

				Expect(err1.Code()).To(Equal(err2.Code()))
				Expect(err1.Error()).To(Equal(err2.Error()))
			})

			It("should handle concurrent error creation", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						err := ErrorParamInvalid.Error(nil)
						Expect(err).NotTo(BeNil())
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					Eventually(done).Should(Receive())
				}
			})
		})

		Context("error with various parent types", func() {
			It("should handle standard error as parent", func() {
				parent := fmt.Errorf("standard error")
				err := ErrorParamInvalid.Error(parent)
				Expect(err).NotTo(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})

			It("should handle liberr.Error as parent", func() {
				parent := ErrorParamEmpty.Error(nil)
				err := ErrorParamInvalid.Error(parent)
				Expect(err).NotTo(BeNil())
				Expect(err.HasParent()).To(BeTrue())
			})
		})
	})
})
