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
)

var _ = Describe("TTY Basic Operations", func() {
	Describe("New()", func() {
		Context("with default stdin (nil reader)", func() {
			It("should use os.Stdin by default", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with explicit stdin", func() {
			It("should accept os.Stdin explicitly", func() {
				saver, err := tty.New(os.Stdin, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with signal handling enabled", func() {
			It("should create saver with signal handling", func() {
				saver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})

		Context("with signal handling disabled", func() {
			It("should create saver without signal handling", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())
			})
		})
	})

	Describe("TTYSaver Interface", func() {
		Context("with mock implementation", func() {
			It("should implement TTYSaver interface", func() {
				mock := newMockTTYSaver(false)

				// Verify it implements the interface
				var _ tty.TTYSaver = mock

				err := mock.Restore()
				Expect(err).ToNot(HaveOccurred())
				Expect(mock.WasCalled()).To(BeTrue())
			})

			It("should handle restore errors", func() {
				mock := newMockTTYSaver(true)

				err := mock.Restore()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrorMockRestore))
				Expect(mock.WasCalled()).To(BeTrue())
			})
		})
	})

	Describe("Error Constants", func() {
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
	})
})
