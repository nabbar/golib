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
 */

package tty_test

import (
	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TTY Error Handling", func() {
	Describe("Error constants", func() {
		It("should have ErrorNotTTY defined", func() {
			Expect(tty.ErrorNotTTY).ToNot(BeNil())
			Expect(tty.ErrorNotTTY.Error()).To(ContainSubstring("not a terminal"))
		})

		It("should have ErrorTTYFailed defined", func() {
			Expect(tty.ErrorTTYFailed).ToNot(BeNil())
			Expect(tty.ErrorTTYFailed.Error()).To(ContainSubstring("failed to get terminal state"))
		})

		It("should have ErrorDevTTYFail defined", func() {
			Expect(tty.ErrorDevTTYFail).ToNot(BeNil())
			Expect(tty.ErrorDevTTYFail.Error()).To(ContainSubstring("failed to open /dev/tty"))
		})

		It("should have distinct error messages", func() {
			err1 := tty.ErrorNotTTY.Error()
			err2 := tty.ErrorTTYFailed.Error()
			err3 := tty.ErrorDevTTYFail.Error()

			Expect(err1).ToNot(Equal(err2))
			Expect(err2).ToNot(Equal(err3))
			Expect(err1).ToNot(Equal(err3))
		})
	})

	Describe("Error handling in New()", func() {
		Context("with various error scenarios", func() {
			It("should not error with nil input", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should not error with signal enabled", func() {
				saver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})

			It("should not error with signal disabled", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})
	})

	Describe("Error handling in Restore()", func() {
		Context("with mock errors", func() {
			It("should handle restore failure gracefully", func() {
				mock := newMockTTYSaver(true)

				// Should not panic
				Expect(func() {
					tty.Restore(mock)
				}).ToNot(Panic())

				Expect(mock.WasCalled()).To(BeTrue())
			})

			It("should handle multiple consecutive failures", func() {
				mock := newMockTTYSaver(true)

				for i := 0; i < 10; i++ {
					mock.Reset()
					tty.Restore(mock)
					Expect(mock.WasCalled()).To(BeTrue())
				}
			})
		})

		Context("with nil values", func() {
			It("should not panic with nil", func() {
				Expect(func() {
					tty.Restore(nil)
				}).ToNot(Panic())
			})

			It("should handle repeated nil calls", func() {
				for i := 0; i < 100; i++ {
					tty.Restore(nil)
				}
			})
		})
	})

	Describe("Error handling in SignalHandler()", func() {
		Context("with various states", func() {
			It("should not panic with nil", func() {
				Expect(func() {
					tty.SignalHandler(nil)
				}).ToNot(Panic())
			})

			It("should not panic with valid mock", func() {
				mock := newMockTTYSaver(false)
				Expect(func() {
					tty.SignalHandler(mock)
				}).ToNot(Panic())
			})

			It("should not panic with failing mock", func() {
				mock := newMockTTYSaver(true)
				Expect(func() {
					tty.SignalHandler(mock)
				}).ToNot(Panic())
			})
		})
	})

	Describe("Defensive programming checks", func() {
		Context("nil safety", func() {
			It("should handle nil in all public functions", func() {
				// New with nil is valid
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				// Restore with nil is safe
				Expect(func() {
					tty.Restore(nil)
				}).ToNot(Panic())

				// SignalHandler with nil is safe
				Expect(func() {
					tty.SignalHandler(nil)
				}).ToNot(Panic())
			})

			It("should handle nil saver methods", func() {
				// Mock with nil state behavior
				mock := newMockTTYSaverWithTerminal(false, false)

				err := mock.Restore()
				Expect(err).ToNot(HaveOccurred())

				err = mock.Signal()
				Expect(err).ToNot(HaveOccurred())

				isTerminal := mock.IsTerminal()
				Expect(isTerminal).To(BeFalse())
			})
		})

		Context("invalid file descriptors", func() {
			It("should handle invalid fd gracefully", func() {
				// Non-terminal input should not cause panic
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Concurrent error handling", func() {
		Context("with concurrent failures", func() {
			It("should handle concurrent restore failures", func() {
				mock := newMockTTYSaver(true)
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						tty.Restore(mock)
						done <- true
					}()
				}

				for i := 0; i < 10; i++ {
					<-done
				}
			})

			It("should handle mixed success and failure", func() {
				successMock := newMockTTYSaver(false)
				failMock := newMockTTYSaver(true)
				done := make(chan bool, 20)

				for i := 0; i < 10; i++ {
					go func(idx int) {
						defer GinkgoRecover()
						if idx%2 == 0 {
							tty.Restore(successMock)
						} else {
							tty.Restore(failMock)
						}
						done <- true
					}(i)
				}

				for i := 0; i < 10; i++ {
					<-done
				}
			})
		})
	})

	Describe("Error recovery", func() {
		Context("after errors", func() {
			It("should continue working after restore failure", func() {
				failMock := newMockTTYSaver(true)

				// First call fails
				tty.Restore(failMock)
				Expect(failMock.WasCalled()).To(BeTrue())

				// Reset and try again
				failMock.Reset()
				tty.Restore(failMock)
				Expect(failMock.WasCalled()).To(BeTrue())
			})

			It("should work with different mocks after failure", func() {
				failMock := newMockTTYSaver(true)
				successMock := newMockTTYSaver(false)

				// Fail
				tty.Restore(failMock)
				Expect(failMock.WasCalled()).To(BeTrue())

				// Succeed
				tty.Restore(successMock)
				Expect(successMock.WasCalled()).To(BeTrue())
			})
		})
	})
})
