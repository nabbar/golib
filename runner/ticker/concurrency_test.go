/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package ticker_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libsrv "github.com/nabbar/golib/runner"
	. "github.com/nabbar/golib/runner/ticker"
)

var _ = Describe("Concurrency", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Concurrent Start/Stop", func() {
		It("should handle multiple concurrent starts and stops", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			var wg sync.WaitGroup
			numOps := 100

			for i := 0; i < numOps; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if i%2 == 0 {
						_ = tick.Start(ctx)
					} else {
						_ = tick.Stop(ctx)
					}
				}()
			}

			wg.Wait()

			// Final state should be stopped or running depending on the last operation
			// But it should not panic or deadlock
			_ = tick.Stop(ctx) // Ensure it's stopped for the next test
			Expect(tick.IsRunning()).To(BeFalse())
		})

		It("should allow concurrent calls to IsRunning and Uptime", func() {
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numCalls := 1000

			for i := 0; i < numCalls; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = tick.IsRunning()
					_ = tick.Uptime()
				}()
			}

			wg.Wait()
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Concurrent Ticker Function Execution", func() {
		It("should not block on ticker function if it takes time", func() {
			processingTime := 50 * time.Millisecond
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				time.Sleep(processingTime)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond) // Allow multiple ticks to occur

			// The ticker should still be running, and not blocked indefinitely
			Expect(tick.IsRunning()).To(BeTrue())

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent errors from ticker function", func() {
			errorCount := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				errorCount.Add(1)
				return context.Canceled
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(errorCount.Load()).To(BeNumerically(">", uint32(1)))
			Expect(tick.ErrorsList()).To(ContainElement(context.Canceled))
		})
	})

	Describe("Context Cancellation in Concurrent Scenarios", func() {
		It("should stop all operations when Stop is called", func() {
			counter := int32(0)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numGoroutines := 10

			// Multiple goroutines checking status
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					for {
						if !tick.IsRunning() {
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
				}(i)
			}

			time.Sleep(100 * time.Millisecond)

			// Call Stop to stop the ticker
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for all goroutines to finish
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Success
			case <-time.After(2 * time.Second):
				Fail("Goroutines did not finish after ticker stopped")
			}
		})

		It("should handle concurrent operations during Stop", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			// Start operations
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tick.IsRunning()
					_ = tick.Uptime()
				}(i)
			}

			time.Sleep(50 * time.Millisecond)

			// Call Stop while operations are running
			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			wg.Wait()

			// Should be stopped
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Stress Test", func() {
		It("should handle high-frequency ticks under concurrent load", func() {
			counter := new(atomic.Uint32)
			tick := New(1*time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numReaders := 10

			for i := 0; i < numReaders; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < 100; j++ {
						_ = tick.IsRunning()
						_ = tick.Uptime()
						_ = tick.ErrorsLast()
						_ = tick.ErrorsList()
						time.Sleep(100 * time.Microsecond)
					}
				}()
			}

			time.Sleep(200 * time.Millisecond) // Let it tick for a while

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			wg.Wait()

			Expect(counter.Load()).To(BeNumerically(">", uint32(100)))
			Expect(tick.IsRunning()).To(BeFalse())
		})
	})
})
