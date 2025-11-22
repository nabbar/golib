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
	"time"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TTY Signal Handler", func() {
	Describe("SignalHandler()", func() {
		Context("with nil state", func() {
			It("should not panic when setting up handler", func() {
				Expect(func() {
					tty.SignalHandler(nil)
				}).ToNot(Panic())

				// Give goroutine time to start
				time.Sleep(10 * time.Millisecond)
			})
		})

		Context("with valid mock state", func() {
			It("should set up signal handler without error", func() {
				mock := newMockTTYSaver(false)

				Expect(func() {
					tty.SignalHandler(mock)
				}).ToNot(Panic())

				// Give goroutine time to start
				time.Sleep(10 * time.Millisecond)

				// Handler should be running (we can't easily test signal reception
				// without actually sending signals which would terminate the test)
			})

			It("should handle multiple signal handler registrations", func() {
				mock1 := newMockTTYSaver(false)
				mock2 := newMockTTYSaver(false)

				// Register multiple handlers (should not panic or deadlock)
				tty.SignalHandler(mock1)
				tty.SignalHandler(mock2)

				time.Sleep(10 * time.Millisecond)
			})
		})

		Context("goroutine lifecycle", func() {
			It("should start a goroutine for signal handling", func() {
				mock := newMockTTYSaver(false)

				// Set up handler
				tty.SignalHandler(mock)

				// Give goroutine time to start
				time.Sleep(10 * time.Millisecond)

				// We can't easily verify the goroutine is running,
				// but we can verify no panic occurred
			})
		})

		Context("concurrent signal handler setup", func() {
			It("should handle concurrent SignalHandler calls", func() {
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func() {
						mock := newMockTTYSaver(false)
						tty.SignalHandler(mock)
						done <- true
					}()
				}

				// Wait for all goroutines
				for i := 0; i < 10; i++ {
					select {
					case <-done:
					case <-time.After(1 * time.Second):
						Fail("timeout waiting for signal handler setup")
					}
				}
			})
		})
	})

	Describe("Signal Handler Behavior", func() {
		Context("defensive checks", func() {
			It("should handle nil state gracefully in handler", func() {
				// This tests that Restore(nil) is safe
				Expect(func() {
					tty.SignalHandler(nil)
					time.Sleep(10 * time.Millisecond)
				}).ToNot(Panic())
			})
		})

		Context("stress test", func() {
			It("should handle rapid handler registrations", func() {
				for i := 0; i < 100; i++ {
					mock := newMockTTYSaver(false)
					tty.SignalHandler(mock)
				}

				// Give time for goroutines to start
				time.Sleep(50 * time.Millisecond)

				// Should not panic or deadlock
			})
		})
	})
})
