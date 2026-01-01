/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package nobar_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libbar "github.com/nabbar/golib/semaphore/nobar"
)

var _ = Describe("Race Detection Tests", func() {
	It("should not have race conditions with concurrent Inc calls", func() {
		sem := createTestSemaphoreWithProgress(globalCtx, 50)
		bar := libbar.New(sem, 10000, false)

		var wg sync.WaitGroup
		const goroutines = 100
		const incrementsPerGoroutine = 100

		// Launch many concurrent goroutines calling Inc
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < incrementsPerGoroutine; j++ {
					bar.Inc(1)
					// Small delay to increase chance of overlapping calls
					time.Sleep(time.Microsecond)
				}
			}()
		}

		wg.Wait()
		bar.DeferMain()

		// Verify final state
		Expect(bar.Total()).To(Equal(int64(0)))
	})

	It("should not have race conditions with mixed Inc/Dec calls", func() {
		sem := createTestSemaphoreWithProgress(globalCtx, 50)
		bar := libbar.New(sem, 5000, false)

		var wg sync.WaitGroup

		// Incrementers
		for i := 0; i < 25; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					bar.Inc64(10)
					time.Sleep(time.Microsecond)
				}
			}()
		}

		// Decrementers
		for i := 0; i < 25; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					bar.Dec64(10)
					time.Sleep(time.Microsecond)
				}
			}()
		}

		wg.Wait()
		bar.DeferMain()

		// Total should remain unchanged
		Expect(bar.Total()).To(Equal(int64(0)))
	})

	It("should not have race conditions with concurrent reads and writes", func() {
		sem := createTestSemaphoreWithProgress(globalCtx, 20)
		bar := libbar.New(sem, 1000, false)

		var wg sync.WaitGroup

		// Writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					bar.Inc(1)
					time.Sleep(time.Microsecond)
				}
			}()
		}

		// Readers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					_ = bar.Current()
					_ = bar.Total()
					_ = bar.Completed()
					time.Sleep(time.Microsecond)
				}
			}()
		}

		wg.Wait()
		bar.DeferMain()

		Expect(bar.Total()).To(Equal(int64(0)))
	})
})
