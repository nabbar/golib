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
	"syscall"
	"time"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/term"
)

var _ = Describe("TTY Signal Handling", func() {
	Describe("Signal() method", func() {
		Context("with terminal support", func() {
			BeforeEach(func() {
				if !term.IsTerminal(int(os.Stdin.Fd())) {
					Skip("Skipping terminal-dependent test: not running in a terminal")
				}
			})

			It("should handle signal correctly with signal enabled", func() {
				saver, err := tty.New(nil, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				// Start signal handler in background
				done := make(chan error, 1)
				go func() {
					done <- saver.Signal()
				}()

				// Give the signal handler time to set up
				time.Sleep(50 * time.Millisecond)

				// Send SIGTERM to ourselves
				proc, err := os.FindProcess(os.Getpid())
				Expect(err).ToNot(HaveOccurred())

				err = proc.Signal(syscall.SIGTERM)
				Expect(err).ToNot(HaveOccurred())

				// Wait for signal to be processed
				select {
				case err := <-done:
					Expect(err).ToNot(HaveOccurred())
				case <-time.After(2 * time.Second):
					Fail("timeout waiting for signal handler")
				}
			})
		})

		Context("with signal disabled", func() {
			It("should return immediately when signal handling is disabled", func() {
				saver, err := tty.New(nil, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(saver).ToNot(BeNil())

				// Signal should return immediately (no-op)
				err = saver.Signal()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with mock", func() {
			It("should call Signal on mock", func() {
				mock := newMockTTYSaver(false)
				Expect(mock.SignalWasCalled()).To(BeFalse())

				err := mock.Signal()
				Expect(err).ToNot(HaveOccurred())
				Expect(mock.SignalWasCalled()).To(BeTrue())
			})

			It("should handle non-terminal mock", func() {
				mock := newMockTTYSaverWithTerminal(false, false)
				err := mock.Signal()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("SignalHandler with different signal types", func() {
		Context("signal handling scenarios", func() {
			It("should set up handler without blocking", func() {
				mock := newMockTTYSaver(false)

				start := time.Now()
				tty.SignalHandler(mock)
				elapsed := time.Since(start)

				// Should return immediately, not block
				Expect(elapsed).To(BeNumerically("<", 100*time.Millisecond))
			})

			It("should handle nil state gracefully", func() {
				Expect(func() {
					tty.SignalHandler(nil)
				}).ToNot(Panic())
			})

			It("should create goroutine for signal handling", func() {
				mock := newMockTTYSaver(false)

				// Set up handler
				tty.SignalHandler(mock)

				// Give time for goroutine to start
				time.Sleep(10 * time.Millisecond)

				// Should not panic or deadlock
			})
		})
	})

	Describe("Signal edge cases", func() {
		Context("concurrent signal setup", func() {
			It("should handle multiple concurrent signal setups", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						defer GinkgoRecover()
						mock := newMockTTYSaver(false)
						tty.SignalHandler(mock)
						done <- true
					}()
				}

				// Wait for all to complete
				for i := 0; i < 10; i++ {
					select {
					case <-done:
					case <-time.After(2 * time.Second):
						Fail("timeout waiting for concurrent signal handlers")
					}
				}
			})
		})

		Context("signal with errors", func() {
			It("should handle mock with error", func() {
				mock := newMockTTYSaver(true)
				tty.SignalHandler(mock)

				time.Sleep(10 * time.Millisecond)
			})
		})
	})
})
