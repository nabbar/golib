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

package shell_test

import (
	"github.com/nabbar/golib/shell"
	"github.com/nabbar/golib/shell/command"
	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shell Constructor", func() {
	Describe("New()", func() {
		Context("with nil TTYSaver", func() {
			It("should create a shell with nil TTYSaver", func() {
				sh := shell.New(nil)
				Expect(sh).ToNot(BeNil())
			})

			It("should return a functional Shell", func() {
				sh := shell.New(nil)
				Expect(sh).ToNot(BeNil())

				// Verify it implements Shell interface by using its methods
				_, found := sh.Get("nonexistent")
				Expect(found).To(BeFalse())
			})

			It("should have working methods with nil TTYSaver", func() {
				sh := shell.New(nil)

				// Should not panic
				Expect(func() {
					_, _ = sh.Get("test")
					_ = sh.Desc("test")
					sh.Walk(func(name string, item command.Command) bool {
						return true
					})
				}).ToNot(Panic())
			})
		})

		Context("with valid TTYSaver", func() {
			var ttySaver tty.TTYSaver

			BeforeEach(func() {
				var err error
				ttySaver, err = tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create a shell with TTYSaver", func() {
				sh := shell.New(ttySaver)
				Expect(sh).ToNot(BeNil())
			})

			It("should accept TTYSaver with signal handling disabled", func() {
				ts, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ts)
				Expect(sh).ToNot(BeNil())
			})

			It("should accept TTYSaver with signal handling enabled", func() {
				ts, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())

				sh := shell.New(ts)
				Expect(sh).ToNot(BeNil())
			})
		})

		Context("concurrent shell creation", func() {
			It("should handle concurrent New() calls", func() {
				done := make(chan shell.Shell, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						sh := shell.New(nil)
						Expect(sh).ToNot(BeNil())
						done <- sh
					}()
				}

				for i := 0; i < 10; i++ {
					sh := <-done
					Expect(sh).ToNot(BeNil())
				}
			})

			It("should handle concurrent New() with TTYSavers", func() {
				done := make(chan shell.Shell, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						ts, err := tty.New(nil, false)
						Expect(err).ToNot(HaveOccurred())

						sh := shell.New(ts)
						Expect(sh).ToNot(BeNil())
						done <- sh
					}()
				}

				for i := 0; i < 10; i++ {
					sh := <-done
					Expect(sh).ToNot(BeNil())
				}
			})
		})

		Context("memory efficiency", func() {
			It("should create lightweight shells", func() {
				shells := make([]shell.Shell, 100)

				for i := 0; i < 100; i++ {
					shells[i] = shell.New(nil)
				}

				// All shells should be valid
				for _, sh := range shells {
					Expect(sh).ToNot(BeNil())
				}
			})
		})
	})
})
