/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package ticker_test

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/runner/ticker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// concurrency_test.go validates thread-safety and concurrent operations of the ticker package.
//
// Test Coverage:
//   - Concurrent Start/Stop/Restart calls on a single ticker
//   - Concurrent status checks (IsRunning, Uptime)
//   - Concurrent error collection reads
//   - Multiple independent tickers running simultaneously
//   - Mixed concurrent operations
//   - Context cancellation during concurrent operations
//   - Stress tests with high-frequency operations
//
// Testing Strategy:
// These tests use sync.WaitGroup to coordinate multiple goroutines and atomic operations
// for counters. They deliberately create race conditions to verify the ticker's internal
// synchronization is correct.
//
// Run with race detector:
//
//	CGO_ENABLED=1 go test -race ./...
//	CGO_ENABLED=1 ginkgo -v --race --repeat=10 .
//
// Important Notes:
//   - These tests are designed to expose race conditions if present
//   - The race detector may slow down execution significantly
//   - Timing assertions use generous margins to account for concurrent load
//   - Tests introduce deliberate delays to ensure goroutines overlap in time
var _ = Describe("Concurrency", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("Concurrent Start/Stop", func() {
		It("should handle concurrent Start calls", func() {
			counter := int32(0)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			var wg sync.WaitGroup
			numGoroutines := 5

			// Start concurrently with small delays to avoid excessive lock contention
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					time.Sleep(time.Duration(idx*5) * time.Millisecond)
					_ = tick.Start(ctx)
				}(i)
			}

			wg.Wait()

			// Should be running
			Expect(tick.IsRunning()).To(BeTrue())

			// Wait for some ticks
			time.Sleep(150 * time.Millisecond)
			Expect(atomic.LoadInt32(&counter)).To(BeNumerically(">=", int32(1)))

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent Stop calls", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			var wg sync.WaitGroup
			numGoroutines := 10

			// Stop concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tick.Stop(ctx)
				}(i)
			}

			wg.Wait()

			// Should be stopped
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})

		It("should handle concurrent Start and Stop calls", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			var wg sync.WaitGroup
			numOps := 5

			// Mix Start and Stop calls
			for i := 0; i < numOps; i++ {
				wg.Add(1)
				if i%2 == 0 {
					go func(idx int) {
						defer wg.Done()
						time.Sleep(time.Duration(idx*5) * time.Millisecond)
						_ = tick.Start(ctx)
					}(i)
				} else {
					go func(idx int) {
						defer wg.Done()
						time.Sleep(time.Duration(idx*5) * time.Millisecond)
						_ = tick.Stop(ctx)
					}(i)
				}
			}

			wg.Wait()
			time.Sleep(300 * time.Millisecond)

			// Cleanup
			_ = tick.Stop(ctx)
			Eventually(tick.IsRunning, 300*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})

		It("should handle concurrent Restart calls", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			var wg sync.WaitGroup
			numGoroutines := 5

			// Restart concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = tick.Restart(ctx)
				}()
			}

			wg.Wait()

			time.Sleep(10 * time.Millisecond)
			// Should be running
			Expect(tick.IsRunning()).To(BeTrue())

			time.Sleep(10 * time.Millisecond)
			Expect(counter.Load()).To(BeNumerically(">=", int32(1)))

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Concurrent Status Checks", func() {
		It("should handle concurrent IsRunning calls", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numGoroutines := 100
			results := make([]bool, numGoroutines)

			// Check IsRunning concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = tick.IsRunning()
				}(i)
			}

			wg.Wait()

			// All should see it as running
			for _, result := range results {
				Expect(result).To(BeTrue())
			}

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent Uptime calls", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			var wg sync.WaitGroup
			numGoroutines := 100
			uptimes := make([]time.Duration, numGoroutines)

			// Check Uptime concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					uptimes[idx] = tick.Uptime()
				}(i)
			}

			wg.Wait()

			// All should have reasonable uptime
			for _, uptime := range uptimes {
				Expect(uptime).To(BeNumerically(">", 0))
			}

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Concurrent Error Access", func() {
		It("should handle concurrent ErrorsLast calls", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			var wg sync.WaitGroup
			numGoroutines := 100

			// Access ErrorsLast concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tick.ErrorsLast()
				}(i)
			}

			wg.Wait()

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent ErrorsList calls", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			var wg sync.WaitGroup
			numGoroutines := 100

			// Access ErrorsList concurrently
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tick.ErrorsList()
				}(i)
			}

			wg.Wait()

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle concurrent error writes and reads", func() {
			counter := int32(0)
			tick := New(20*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numReaders := 50

			// Read errors concurrently while ticker is writing
			for i := 0; i < numReaders; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					for j := 0; j < 10; j++ {
						_ = tick.ErrorsLast()
						_ = tick.ErrorsList()
						time.Sleep(5 * time.Millisecond)
					}
				}(i)
			}

			wg.Wait()

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Concurrent Multiple Tickers", func() {
		It("should handle multiple independent tickers concurrently", func() {
			numTickers := 10
			counters := make([]int32, numTickers)
			tickers := make([]Ticker, numTickers)

			// Create and start multiple tickers
			for i := 0; i < numTickers; i++ {
				idx := i
				tickers[i] = New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
					atomic.AddInt32(&counters[idx], 1)
					return nil
				})

				err := tickers[i].Start(ctx)
				Expect(err).ToNot(HaveOccurred())
			}

			// Let them run (200ms / 50ms = ~4 ticks expected per ticker)
			time.Sleep(200 * time.Millisecond)

			// All should be running and have executed at least twice
			// Use conservative threshold to account for concurrent load
			for i := 0; i < numTickers; i++ {
				Expect(tickers[i].IsRunning()).To(BeTrue())
				Expect(atomic.LoadInt32(&counters[i])).To(BeNumerically(">=", int32(2)))
			}

			// Stop all concurrently
			var wg sync.WaitGroup
			for i := 0; i < numTickers; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tickers[idx].Stop(ctx)
				}(i)
			}
			wg.Wait()

			// All should be stopped
			for i := 0; i < numTickers; i++ {
				Eventually(tickers[i].IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
			}
		})

		It("should handle concurrent operations on multiple tickers", func() {
			numTickers := 5
			tickers := make([]Ticker, numTickers)

			for i := 0; i < numTickers; i++ {
				tickers[i] = New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
					return nil
				})
			}

			var wg sync.WaitGroup

			// Perform various operations concurrently on all tickers
			for i := 0; i < numTickers; i++ {
				wg.Add(4)

				go func(idx int) {
					defer wg.Done()
					_ = tickers[idx].Start(ctx)
				}(i)

				go func(idx int) {
					defer wg.Done()
					time.Sleep(50 * time.Millisecond)
					_ = tickers[idx].IsRunning()
				}(i)

				go func(idx int) {
					defer wg.Done()
					time.Sleep(100 * time.Millisecond)
					_ = tickers[idx].Uptime()
				}(i)

				go func(idx int) {
					defer wg.Done()
					time.Sleep(150 * time.Millisecond)
					_ = tickers[idx].Stop(ctx)
				}(i)
			}

			wg.Wait()

			// Cleanup
			for i := 0; i < numTickers; i++ {
				_ = tickers[i].Stop(ctx)
			}
		})
	})

	Describe("Race Conditions", func() {
		It("should not have race conditions in Start", func() {
			counter := int32(0)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			var wg sync.WaitGroup
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					time.Sleep(time.Duration(idx) * time.Millisecond)
					_ = tick.Start(ctx)
				}(i)
			}

			wg.Wait()
			time.Sleep(100 * time.Millisecond)

			err := tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not have race conditions in Stop", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(50 * time.Millisecond)

			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_ = tick.Stop(ctx)
				}(i)
			}

			wg.Wait()
		})

		It("should not have race conditions with mixed operations", func() {
			counter := int32(0)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			var wg sync.WaitGroup
			numOps := 24

			// Mix all operations
			for i := 0; i < numOps; i++ {
				wg.Add(1)

				op := i % 6
				go func(operation, idx int) {
					defer wg.Done()
					time.Sleep(time.Duration(idx/6) * time.Millisecond)

					switch operation {
					case 0:
						_ = tick.Start(ctx)
					case 1:
						_ = tick.Stop(ctx)
					case 2:
						_ = tick.Restart(ctx)
					case 3:
						_ = tick.IsRunning()
					case 4:
						_ = tick.Uptime()
					case 5:
						_ = tick.ErrorsLast()
					}
				}(op, i)
			}

			wg.Wait()

			// Cleanup
			_ = tick.Stop(ctx)
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Context Cancellation in Concurrent Scenarios", func() {
		It("should stop all operations when context is cancelled", func() {
			counter := int32(0)
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				atomic.AddInt32(&counter, 1)
				return nil
			})

			cancelCtx, cancelFunc := context.WithCancel(context.Background())
			err := tick.Start(cancelCtx)
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

			// Cancel context
			cancelFunc()

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
				Fail("Goroutines did not finish after context cancellation")
			}
		})

		It("should handle concurrent operations during context cancellation", func() {
			tick := New(50*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				return nil
			})

			cancelCtx, cancelFunc := context.WithCancel(context.Background())
			err := tick.Start(cancelCtx)
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

			// Cancel context while operations are running
			cancelFunc()

			wg.Wait()

			// Should stop eventually
			Eventually(tick.IsRunning, 500*time.Millisecond, 10*time.Millisecond).Should(BeFalse())
		})
	})

	Describe("Stress Test", func() {
		It("should handle high-frequency ticks under concurrent load", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			err := tick.Start(ctx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup
			numReaders := 20

			// Concurrent readers performing many status checks
			// This creates significant concurrent load on the ticker
			for i := 0; i < numReaders; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					for j := 0; j < 50; j++ {
						_ = tick.IsRunning()
						_ = tick.Uptime()
						_ = tick.ErrorsLast()
						time.Sleep(2 * time.Millisecond)
					}
				}(i)
			}

			wg.Wait()

			// Should have at least a few ticks during the concurrent load
			// Conservative threshold to account for race detector overhead
			Expect(counter.Load()).To(BeNumerically(">=", uint32(3)))

			err = tick.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle rapid start/stop cycles", func() {
			counter := new(atomic.Uint32)
			tick := New(10*time.Millisecond, func(ctx context.Context, tck *time.Ticker) error {
				counter.Add(1)
				return nil
			})

			for i := 0; i < 10; i++ {
				err := tick.Start(ctx)
				Expect(err).ToNot(HaveOccurred())

				time.Sleep(15 * time.Millisecond)

				err = tick.Stop(ctx)
				Expect(err).ToNot(HaveOccurred())
			}

			// Should have completed multiple cycles
			Expect(counter.Load()).To(BeNumerically(">=", int32(5)))
		})
	})
})
