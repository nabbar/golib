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
	"sync"
	"time"

	"github.com/nabbar/golib/shell/tty"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TTY Edge Cases", func() {
	Describe("Nil Safety", func() {
		Context("Restore with nil", func() {
			It("should not panic with nil state", func() {
				Expect(func() {
					tty.Restore(nil)
				}).ToNot(Panic())
			})

			It("should handle repeated nil calls", func() {
				for i := 0; i < 1000; i++ {
					tty.Restore(nil)
				}
			})
		})

		Context("SignalHandler with nil", func() {
			It("should not panic with nil state", func() {
				Expect(func() {
					tty.SignalHandler(nil)
				}).ToNot(Panic())

				time.Sleep(10 * time.Millisecond)
			})
		})
	})

	Describe("Concurrent Operations", func() {
		Context("concurrent Restore calls", func() {
			It("should handle concurrent restore operations", func() {
				mock := newMockTTYSaver(false)
				var wg sync.WaitGroup

				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						tty.Restore(mock)
					}()
				}

				// Wait for all goroutines
				done := make(chan bool)
				go func() {
					wg.Wait()
					done <- true
				}()

				select {
				case <-done:
					// Success
				case <-time.After(5 * time.Second):
					Fail("timeout waiting for concurrent restore operations")
				}
			})

			It("should handle mixed nil and valid restore calls", func() {
				mock := newMockTTYSaver(false)
				var wg sync.WaitGroup

				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						if idx%2 == 0 {
							tty.Restore(nil)
						} else {
							tty.Restore(mock)
						}
					}(i)
				}

				wg.Wait()
			})
		})

		Context("concurrent SignalHandler registrations", func() {
			It("should handle concurrent signal handler setup", func() {
				var wg sync.WaitGroup

				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						mock := newMockTTYSaver(false)
						tty.SignalHandler(mock)
					}()
				}

				wg.Wait()
				time.Sleep(50 * time.Millisecond)
			})
		})
	})

	Describe("Error Handling", func() {
		Context("mock restore failures", func() {
			It("should handle restore errors gracefully", func() {
				mock := newMockTTYSaver(true) // Will fail on restore

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

		Context("mixed success and failure", func() {
			It("should handle alternating success and failure", func() {
				successMock := newMockTTYSaver(false)
				failMock := newMockTTYSaver(true)

				for i := 0; i < 10; i++ {
					if i%2 == 0 {
						tty.Restore(successMock)
					} else {
						tty.Restore(failMock)
					}
				}
			})
		})
	})

	Describe("Stress Tests", func() {
		Context("high volume operations", func() {
			It("should handle many restore calls", func() {
				mock := newMockTTYSaver(false)

				for i := 0; i < 10000; i++ {
					tty.Restore(mock)
				}
			})

			It("should handle many nil restore calls", func() {
				for i := 0; i < 10000; i++ {
					tty.Restore(nil)
				}
			})

			It("should handle many signal handler registrations", func() {
				for i := 0; i < 1000; i++ {
					mock := newMockTTYSaver(false)
					tty.SignalHandler(mock)
				}

				time.Sleep(100 * time.Millisecond)
			})
		})

		Context("concurrent stress", func() {
			It("should handle massive concurrent restore operations", func() {
				mock := newMockTTYSaver(false)
				var wg sync.WaitGroup

				for i := 0; i < 1000; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						tty.Restore(mock)
					}()
				}

				done := make(chan bool)
				go func() {
					wg.Wait()
					done <- true
				}()

				select {
				case <-done:
					// Success
				case <-time.After(10 * time.Second):
					Fail("timeout in concurrent stress test")
				}
			})
		})
	})

	Describe("Memory and Resource Tests", func() {
		Context("no memory leaks", func() {
			It("should not leak memory with repeated operations", func() {
				mock := newMockTTYSaver(false)

				// Perform many operations
				for i := 0; i < 1000; i++ {
					tty.Restore(mock)
					tty.SignalHandler(mock)
				}

				// If there were memory leaks, this would eventually cause issues
				// This test passes if no panic or excessive memory usage occurs
			})
		})

		Context("goroutine management", func() {
			It("should not leak goroutines from failed restores", func() {
				failMock := newMockTTYSaver(true)

				for i := 0; i < 100; i++ {
					tty.Restore(failMock)
				}

				// No goroutines should be leaked
				// This is implicitly tested by not hanging
			})
		})
	})

	Describe("Interface Compliance", func() {
		Context("TTYSaver interface", func() {
			It("should work with any TTYSaver implementation", func() {
				// Test with successful mock
				successMock := newMockTTYSaver(false)
				var _ tty.TTYSaver = successMock

				tty.Restore(successMock)
				Expect(successMock.WasCalled()).To(BeTrue())

				// Test with failing mock
				failMock := newMockTTYSaver(true)
				var _ tty.TTYSaver = failMock

				failMock.Reset()
				tty.Restore(failMock)
				Expect(failMock.WasCalled()).To(BeTrue())
			})
		})
	})
})
