/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package aggregator_test

import (
	"context"
	"sync"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TC-CC-001: Concurrency and Race Conditions", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(testCtx)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	Describe("TC-CC-002: Concurrent Writes", func() {
		It("TC-CC-003: should handle concurrent writes safely", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Concurrent writes
			numGoroutines := 50
			writesPerGoroutine := 20
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < writesPerGoroutine; j++ {
						data := []byte{byte(id), byte(j)}
						n, err := agg.Write(data)
						if err == nil {
							Expect(n).To(Equal(len(data)))
						}
						// Allow ErrClosedResources if test is finishing
					}
				}(i)
			}

			wg.Wait()

			// Wait for processing
			Eventually(func() int32 {
				return writer.GetCallCount()
			}, 5*time.Second, 100*time.Millisecond).Should(BeNumerically(">", 0))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify we got many writes
			Expect(writer.GetCallCount()).To(BeNumerically(">", 100))
		})

		It("TC-CC-004: should handle concurrent writes with small buffer", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 5, // Small buffer to force blocking
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 10
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					data := []byte{byte(id)}
					n, err := agg.Write(data)
					if err == nil {
						Expect(n).To(Equal(len(data)))
					}
				}(i)
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-CC-005: Concurrent Start/Stop", func() {

		It("TC-CC-006: should handle concurrent Stop calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 10
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					err := agg.Stop(ctx)
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()

			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())
		})

		It("TC-CC-007: should handle concurrent Restart calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 10
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					err := agg.Restart(ctx)
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-CC-008: Concurrent Reads (Status)", func() {
		It("TC-CC-009: should handle concurrent IsRunning calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 50
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 100; j++ {
						_ = agg.IsRunning()
					}
				}()
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-CC-010: should handle concurrent Uptime calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 50
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 100; j++ {
						_ = agg.Uptime()
					}
				}()
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-CC-011: should handle concurrent ErrorsList calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 50
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 100; j++ {
						_ = agg.ErrorsList()
						_ = agg.ErrorsLast()
					}
				}()
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-CC-012: Mixed Operations", func() {
		It("TC-CC-013: should handle concurrent writes and status reads", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 100,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 50
			var wg sync.WaitGroup

			// Writers
			for i := 0; i < numGoroutines/2; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 10; j++ {
						data := []byte{byte(id), byte(j)}
						_, _ = agg.Write(data)
						time.Sleep(time.Millisecond)
					}
				}(i)
			}

			// Status readers
			for i := 0; i < numGoroutines/2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 100; j++ {
						_ = agg.IsRunning()
						_ = agg.Uptime()
						time.Sleep(time.Millisecond)
					}
				}()
			}

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-CC-014: should handle writes during start/stop transitions", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 100,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			// Writers
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 20; j++ {
						data := []byte{byte(id), byte(j)}
						_, _ = agg.Write(data)
						time.Sleep(5 * time.Millisecond)
					}
				}(i)
			}

			// Controller
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				for i := 0; i < 5; i++ {
					_ = agg.Start(ctx)
					time.Sleep(20 * time.Millisecond)
					_ = agg.Stop(ctx)
					time.Sleep(20 * time.Millisecond)
				}
			}()

			wg.Wait()

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-CC-015: should handle concurrent Close calls", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numGoroutines := 10
			var wg sync.WaitGroup

			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer GinkgoRecover()

					err := agg.Close()
					Expect(err).ToNot(HaveOccurred())
				}()
			}

			wg.Wait()
		})
	})

	Describe("TC-CC-016: Stress Tests", func() {
		It("TC-CC-017: should handle high volume of writes", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			numWrites := 10000
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				for i := 0; i < numWrites; i++ {
					data := []byte{byte(i % 256)}
					_, _ = agg.Write(data)
				}
			}()

			wg.Wait()

			// Give time to process
			time.Sleep(2 * time.Second)

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			// Should have processed many writes
			Expect(writer.GetCallCount()).To(BeNumerically(">", 100))
		})
		It("TC-CC-018: should handle context cancellation under load", func() {
			localCtx, localCancel := context.WithCancel(ctx)

			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 100,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(localCtx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(localCtx)
			Expect(err).ToNot(HaveOccurred())

			var wg sync.WaitGroup

			// Start writers
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					defer GinkgoRecover()

					for j := 0; j < 100; j++ {
						data := []byte{byte(id), byte(j)}
						_, _ = agg.Write(data)
					}
				}(i)
			}

			// Cancel context after a bit
			time.Sleep(50 * time.Millisecond)
			localCancel()

			wg.Wait()

			// Aggregator should stop
			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())
		})
	})
})
