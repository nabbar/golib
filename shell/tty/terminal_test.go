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
	"os"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/term"
)

var _ = Describe("TTY Terminal Detection", func() {
	// Helper to check if we're in a terminal
	isInTerminal := func() bool {
		return term.IsTerminal(int(os.Stdin.Fd()))
	}

	Describe("IsTerminal()", func() {
		Context("when running in an actual terminal", func() {
			BeforeEach(func() {
				if !isInTerminal() {
					Skip("Skipping terminal-dependent test: not running in a terminal")
				}
			})

			It("should detect terminal correctly", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				Expect(saver.IsTerminal()).To(BeTrue())
			})

			It("should work with explicit stdin", func() {
				saver, err := tty.New(os.Stdin, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
				Expect(saver.IsTerminal()).To(BeTrue())
			})
		})

		Context("when not in a terminal", func() {
			It("should return false for non-terminal", func() {
				// The mock is not a real terminal
				mock := newMockTTYSaverWithTerminal(false, false)
				Expect(mock.IsTerminal()).To(BeFalse())
			})

			It("should return true for mocked terminal", func() {
				mock := newMockTTYSaverWithTerminal(false, true)
				Expect(mock.IsTerminal()).To(BeTrue())
			})
		})

		Context("with nil saver", func() {
			It("should handle nil saver gracefully", func() {
				// Create a saver with non-terminal input
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				// IsTerminal should not panic
				Expect(func() {
					_ = saver.IsTerminal()
				}).ToNot(Panic())
			})
		})
	})

	Describe("Terminal environment detection", func() {
		It("should correctly identify terminal environment", func() {
			fd := int(os.Stdin.Fd())
			expected := term.IsTerminal(fd)

			saver, err := tty.New(nil, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(saver).ToNot(BeNil())

			// Should match the system's terminal detection
			Expect(saver.IsTerminal()).To(Equal(expected))
		})

		Context("with explicit file descriptors", func() {
			It("should detect os.Stdin terminal state", func() {
				expected := term.IsTerminal(int(os.Stdin.Fd()))

				saver, err := tty.New(os.Stdin, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver.IsTerminal()).To(Equal(expected))
			})

			It("should detect os.Stdout terminal state", func() {
				expected := term.IsTerminal(int(os.Stdout.Fd()))

				saver, err := tty.New(os.Stdout, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver.IsTerminal()).To(Equal(expected))
			})

			It("should detect os.Stderr terminal state", func() {
				expected := term.IsTerminal(int(os.Stderr.Fd()))

				saver, err := tty.New(os.Stderr, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver.IsTerminal()).To(Equal(expected))
			})
		})
	})

	Describe("Terminal-dependent operations", func() {
		Context("when in terminal", func() {
			BeforeEach(func() {
				if !isInTerminal() {
					Skip("Skipping terminal-dependent test: not running in a terminal")
				}
			})

			It("should allow Restore on terminal", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver.IsTerminal()).To(BeTrue())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle multiple restores", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver.IsTerminal()).To(BeTrue())

				// Multiple restores should not error
				for i := 0; i < 5; i++ {
					err = saver.Restore()
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})

		Context("when not in terminal", func() {
			It("should handle restore gracefully on non-terminal", func() {
				// In non-terminal environment, restore should not error
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
