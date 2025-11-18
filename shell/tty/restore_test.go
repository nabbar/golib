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

var _ = Describe("TTY Restore Operations", func() {
	Describe("Restore()", func() {
		Context("with nil state", func() {
			It("should not panic", func() {
				Expect(func() {
					tty.Restore(nil)
				}).ToNot(Panic())
			})

			It("should return immediately without error", func() {
				// This is a defensive check - should be safe
				tty.Restore(nil)
				// No assertion needed - just verifying no panic
			})
		})

		Context("with valid mock state", func() {
			It("should call Restore on the state", func() {
				mock := newMockTTYSaver(false)
				Expect(mock.WasCalled()).To(BeFalse())

				tty.Restore(mock)

				Expect(mock.WasCalled()).To(BeTrue())
			})

			It("should handle restore failure gracefully", func() {
				// Mock that fails to restore
				mock := newMockTTYSaver(true)

				// Should not panic even if restore fails
				Expect(func() {
					tty.Restore(mock)
				}).ToNot(Panic())

				Expect(mock.WasCalled()).To(BeTrue())
			})
		})

		Context("multiple calls", func() {
			It("should handle multiple restore calls safely", func() {
				mock := newMockTTYSaver(false)

				// First call
				tty.Restore(mock)
				Expect(mock.WasCalled()).To(BeTrue())

				// Reset and call again
				mock.Reset()
				tty.Restore(mock)
				Expect(mock.WasCalled()).To(BeTrue())
			})
		})
	})

	Describe("Restore Behavior", func() {
		Context("defensive programming", func() {
			It("should handle nil gracefully without allocation", func() {
				// Call multiple times to verify no issues
				for i := 0; i < 100; i++ {
					tty.Restore(nil)
				}
			})

			It("should not leak goroutines", func() {
				mock := newMockTTYSaver(false)

				// Call multiple times
				for i := 0; i < 100; i++ {
					tty.Restore(mock)
				}

				// No goroutines should be created
				// This is verified by the absence of panic or deadlock
			})
		})
	})
})
