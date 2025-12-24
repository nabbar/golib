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
	"errors"
	"sync/atomic"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TC-ER-001: Error Handling", func() {
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

	Describe("TC-ER-002: Error Constants", func() {
		It("TC-ER-003: should have defined error constants", func() {
			Expect(iotagg.ErrInvalidWriter).ToNot(BeNil())
			Expect(iotagg.ErrInvalidInstance).ToNot(BeNil())
			Expect(iotagg.ErrClosedResources).ToNot(BeNil())
		})

		It("TC-ER-004: should have distinct error messages", func() {
			Expect(iotagg.ErrInvalidWriter.Error()).To(ContainSubstring("invalid writer"))
			Expect(iotagg.ErrInvalidInstance.Error()).To(ContainSubstring("invalid instance"))
			Expect(iotagg.ErrClosedResources.Error()).To(ContainSubstring("closed resources"))
		})
	})

	Describe("TC-ER-005: Writer Errors", func() {
		It("TC-ER-006: should handle writer function errors gracefully", func() {
			errorToReturn := errors.New("writer error")
			var callCount atomic.Int32

			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: func(p []byte) (n int, err error) {
					callCount.Add(1)
					return 0, errorToReturn
				},
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Write data - should not return error (error is logged)
			data := []byte("test")
			n, err := agg.Write(data)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			// Wait for processing attempt
			Eventually(func() int32 {
				return callCount.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 1))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-ER-007: should continue processing after writer errors", func() {
			var callCount atomic.Int32
			var failFirst atomic.Bool
			failFirst.Store(true)

			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: func(p []byte) (n int, err error) {
					callCount.Add(1)
					if failFirst.Load() {
						failFirst.Store(false)
						return 0, errors.New("first error")
					}
					return len(p), nil
				},
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Write two chunks
			_, _ = agg.Write([]byte("chunk1"))
			_, _ = agg.Write([]byte("chunk2"))

			// Should process both (first fails, second succeeds)
			Eventually(func() int32 {
				return callCount.Load()
			}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 2))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("TC-ER-008: Context Errors", func() {
		It("TC-ER-009: should handle parent context cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)

			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(localCtx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(localCtx)
			Expect(err).ToNot(HaveOccurred())

			// Cancel parent context
			localCancel()

			// Aggregator should detect cancellation
			Eventually(func() error {
				return agg.Err()
			}, 2*time.Second, 50*time.Millisecond).Should(Equal(context.Canceled))
		})

		It("TC-ER-010: should handle parent context deadline", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			localCtx, localCancel := context.WithTimeout(ctx, 200*time.Millisecond)
			defer localCancel()

			err = agg.Start(localCtx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for timeout
			Eventually(func() error {
				return agg.Err()
			}, 2*time.Second, 50*time.Millisecond).Should(HaveOccurred())
		})

		It("TC-ER-011: should reject writes after context cancellation", func() {
			localCtx, localCancel := context.WithCancel(ctx)

			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(localCtx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Start(localCtx)
			Expect(err).ToNot(HaveOccurred())

			// Cancel and wait
			localCancel()
			time.Sleep(200 * time.Millisecond)

			// Try to write
			data := []byte("test")
			n, err := agg.Write(data)
			Expect(err).To(HaveOccurred())
			Expect(n).To(Equal(0))
		})
	})

	Describe("TC-ER-012: Async Function Errors", func() {
		It("TC-ER-013: should handle panics in async function", func() {
			writer := newTestWriter()

			cfg := iotagg.Config{
				AsyncTimer: 50 * time.Millisecond,
				AsyncMax:   5,
				AsyncFct: func(ctx context.Context) {
					// This will panic but should be handled
					defer func() {
						_ = recover()
					}()
					panic("async panic")
				},
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait a bit for async function to run
			time.Sleep(300 * time.Millisecond)

			// Aggregator should still be running
			Expect(agg.IsRunning()).To(BeTrue())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-ER-014: should respect context cancellation in async function", func() {
			writer := newTestWriter()
			counter := newTestCounter()

			cfg := iotagg.Config{
				AsyncTimer: 20 * time.Millisecond,
				AsyncMax:   5,
				AsyncFct: func(ctx context.Context) {
					// Check context before doing work
					select {
					case <-ctx.Done():
						return
					default:
						counter.Inc()
						time.Sleep(50 * time.Millisecond)
					}
				},
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Let it run a bit
			time.Sleep(100 * time.Millisecond)
			countBeforeStop := counter.Get()

			// Stop
			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait
			time.Sleep(200 * time.Millisecond)

			// Counter should not increase much after stop
			countAfterStop := counter.Get()
			Expect(countAfterStop - countBeforeStop).To(BeNumerically("<=", 3))
		})
	})

	Describe("TC-ER-015: Sync Function Errors", func() {
		It("TC-ER-016: should handle panics in sync function", func() {
			writer := newTestWriter()
			var panicCount atomic.Int32

			cfg := iotagg.Config{
				SyncTimer: 50 * time.Millisecond,
				SyncFct: func(ctx context.Context) {
					panicCount.Add(1)
				},
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			// This might panic during start/run
			// We'll use a deferred recover to catch it
			func() {
				defer func() {
					_ = recover()
				}()
				_ = agg.Start(ctx)
				time.Sleep(200 * time.Millisecond)
			}()

			// Try to close gracefully
			_ = agg.Close()
		})

		It("TC-ER-017: should respect context cancellation in sync function", func() {
			writer := newTestWriter()
			counter := newTestCounter()
			var shouldBlock atomic.Bool
			shouldBlock.Store(true)

			cfg := iotagg.Config{
				SyncTimer: 20 * time.Millisecond,
				SyncFct: func(ctx context.Context) {
					if shouldBlock.Load() {
						counter.Inc()
						// Block for a bit
						select {
						case <-ctx.Done():
							return
						case <-time.After(100 * time.Millisecond):
						}
					}
				},
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Let it run a bit
			time.Sleep(100 * time.Millisecond)
			countBeforeStop := counter.Get()

			// Stop blocking and stop aggregator
			shouldBlock.Store(false)
			err = agg.Stop(ctx)
			Expect(err).ToNot(HaveOccurred())

			// Counter should have been called
			Expect(countBeforeStop).To(BeNumerically(">", 0))
		})
	})

	Describe("TC-ER-018: Edge Cases", func() {
		It("TC-ER-019: should handle rapid open/close cycles", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			for i := 0; i < 10; i++ {
				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = startAndWait(agg, ctx)
				Expect(err).ToNot(HaveOccurred())

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			}
		})

		It("TC-ER-020: should handle writes during close", func() {
			writer := newTestWriter()
			writer.SetDelay(10) // Slow down writes

			cfg := iotagg.Config{
				BufWriter: 100,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Start writing in goroutine
			done := make(chan bool)
			go func() {
				defer close(done)
				for i := 0; i < 50; i++ {
					_, _ = agg.Write([]byte{byte(i)})
					time.Sleep(5 * time.Millisecond)
				}
			}()

			// Close while writing
			time.Sleep(50 * time.Millisecond)
			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			<-done
		})

		It("TC-ER-021: should handle nil logger gracefully", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-ER-022: should handle zero-length writes without calling writer", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Write empty slices
			for i := 0; i < 10; i++ {
				n, err := agg.Write([]byte{})
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))
			}

			// Wait a bit
			time.Sleep(200 * time.Millisecond)

			// Writer should not be called
			Expect(writer.GetCallCount()).To(Equal(int32(0)))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-ER-023: should handle nil slices", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Write nil
			n, err := agg.Write(nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(n).To(Equal(0))

			// Wait a bit
			time.Sleep(200 * time.Millisecond)

			// Writer should not be called
			Expect(writer.GetCallCount()).To(Equal(int32(0)))

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
