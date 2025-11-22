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
	"sync"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/term"
)

var _ = Describe("TTY Advanced Restore Operations", func() {
	Describe("Restore with actual terminal", func() {
		Context("when in terminal", func() {
			BeforeEach(func() {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					Skip("Skipping terminal-dependent test: not running in a terminal")
				}
			})

			It("should restore terminal state successfully", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle multiple consecutive restores", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				for i := 0; i < 5; i++ {
					err = saver.Restore()
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should restore with different file descriptors", func() {
				// Test with stdin
				saver1, err := tty.New(os.Stdin, false)
				Expect(err).ToNot(HaveOccurred())
				err = saver1.Restore()
				Expect(err).ToNot(HaveOccurred())

				// Test with stdout if it's a terminal
				if term.IsTerminal(int(os.Stdout.Fd())) {
					saver2, err := tty.New(os.Stdout, false)
					Expect(err).ToNot(HaveOccurred())
					err = saver2.Restore()
					Expect(err).ToNot(HaveOccurred())
				}
			})

			It("should be safe with concurrent restores", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						err := saver.Restore()
						Expect(err).ToNot(HaveOccurred())
					}()
				}
				wg.Wait()
			})
		})

		Context("when not in terminal", func() {
			It("should handle non-terminal gracefully", func() {
				// Create with a non-terminal input
				tempFile, err := os.CreateTemp("", "tty-restore-*.txt")
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					name := tempFile.Name()
					_ = tempFile.Close()
					_ = os.Remove(name)
				}()

				saver, err := tty.New(tempFile, false)
				Expect(err).ToNot(HaveOccurred())

				// Should not error even with non-terminal
				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Restore function wrapper", func() {
		Context("with various states", func() {
			It("should handle nil state", func() {
				Expect(func() {
					tty.Restore(nil)
				}).ToNot(Panic())
			})

			It("should restore valid state", func() {
				mock := newMockTTYSaver(false)
				tty.Restore(mock)
				Expect(mock.WasCalled()).To(BeTrue())
			})

			It("should handle failing restore", func() {
				mock := newMockTTYSaver(true)
				Expect(func() {
					tty.Restore(mock)
				}).ToNot(Panic())
				Expect(mock.WasCalled()).To(BeTrue())
			})
		})

		Context("in defer statements", func() {
			It("should work in defer", func() {
				mock := newMockTTYSaver(false)

				func() {
					defer tty.Restore(mock)
					// Do something
				}()

				Expect(mock.WasCalled()).To(BeTrue())
			})

			It("should handle panic with defer", func() {
				mock := newMockTTYSaver(false)

				Expect(func() {
					defer tty.Restore(mock)
					panic("test panic")
				}).To(Panic())

				Expect(mock.WasCalled()).To(BeTrue())
			})
		})
	})

	Describe("Restore edge cases", func() {
		Context("with closed file descriptors", func() {
			It("should handle closed fd gracefully", func() {
				tempFile, err := os.CreateTemp("", "tty-closed-*.txt")
				Expect(err).ToNot(HaveOccurred())
				name := tempFile.Name()
				defer os.Remove(name)

				saver, err := tty.New(tempFile, false)
				Expect(err).ToNot(HaveOccurred())

				// Close the file
				_ = tempFile.Close()

				// Should not panic
				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("concurrent restore and new", func() {
			BeforeEach(func() {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					Skip("Skipping terminal-dependent test: not running in a terminal")
				}
			})

			It("should handle concurrent operations", func() {
				var wg sync.WaitGroup
				done := make(chan bool, 20)

				// Create multiple savers and restore concurrently
				for i := 0; i < 10; i++ {
					wg.Add(2)

					// Create saver
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						saver, err := tty.New(nil, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(saver).ToNot(BeNil())
						done <- true
					}()

					// Restore saver
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						saver, err := tty.New(nil, false)
						Expect(err).ToNot(HaveOccurred())
						Expect(saver).ToNot(BeNil())
						err = saver.Restore()
						Expect(err).ToNot(HaveOccurred())
						done <- true
					}()
				}

				wg.Wait()
				close(done)

				count := 0
				for range done {
					count++
				}
				Expect(count).To(Equal(20))
			})
		})
	})

	Describe("Terminal state verification", func() {
		Context("IsTerminal checks", func() {
			It("should report correct terminal state", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				expected := term.IsTerminal(int(os.Stdin.Fd()))
				Expect(saver.IsTerminal()).To(Equal(expected))
			})

			It("should remain consistent after restore", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				before := saver.IsTerminal()
				err = saver.Restore()
				Expect(err).ToNot(HaveOccurred())
				after := saver.IsTerminal()

				Expect(before).To(Equal(after))
			})
		})
	})
})
