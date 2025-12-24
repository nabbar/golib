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
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TC-WR-001: Writer Operations", func() {
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

	Describe("TC-WR-002: Write()", func() {
		Context("TC-WR-003: when aggregator is running", func() {
			It("TC-WR-004: should write data successfully", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 10,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				err = startAndWait(agg, ctx)
				Expect(err).ToNot(HaveOccurred())

				// Write some data
				data := []byte("test data")
				n, err := agg.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(data)))

				// Wait for data to be processed
				Eventually(func() int32 {
					return writer.GetCallCount()
				}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 1))

				// Verify data
				written := writer.GetData()
				Expect(written).To(HaveLen(1))
				Expect(written[0]).To(Equal(data))

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-WR-005: should write multiple data chunks", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 100,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = startAndWait(agg, ctx)
				Expect(err).ToNot(HaveOccurred())

				// Write multiple chunks
				chunks := [][]byte{
					[]byte("chunk 1"),
					[]byte("chunk 2"),
					[]byte("chunk 3"),
				}

				for _, chunk := range chunks {
					n, err := agg.Write(chunk)
					Expect(err).ToNot(HaveOccurred())
					Expect(n).To(Equal(len(chunk)))
				}

				// Wait for all data to be processed
				Eventually(func() int32 {
					return writer.GetCallCount()
				}, 2*time.Second, 50*time.Millisecond).Should(BeNumerically(">=", 3))

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-WR-006: should handle empty writes", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 10,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = startAndWait(agg, ctx)
				Expect(err).ToNot(HaveOccurred())

				// Write empty data
				data := []byte{}
				n, err := agg.Write(data)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(0))

				// Empty data should not be written
				time.Sleep(100 * time.Millisecond)
				Expect(writer.GetCallCount()).To(Equal(int32(0)))

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("TC-WR-007: when aggregator is not running", func() {
			It("TC-WR-008: should reject writes before start", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 10,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				// Write before starting should fail
				data := []byte("test data")
				n, err := agg.Write(data)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Or(Equal(iotagg.ErrInvalidInstance), Equal(iotagg.ErrClosedResources)))
				Expect(n).To(Equal(0))

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("TC-WR-009: when aggregator is closed", func() {
			It("TC-WR-010: should return error on write after close", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 10,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())

				err = startAndWait(agg, ctx)
				Expect(err).ToNot(HaveOccurred())

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())

				// Wait a bit for close to complete
				time.Sleep(100 * time.Millisecond)

				// Try to write after close
				data := []byte("test data")
				n, err := agg.Write(data)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(iotagg.ErrClosedResources))
				Expect(n).To(Equal(0))
			})
		})

		Context("TC-WR-011: when context is cancelled", func() {
			It("TC-WR-012: should return error on write after context cancel", func() {
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

				// Cancel context
				localCancel()

				// Wait for aggregator to stop
				Eventually(func() bool {
					return agg.IsRunning()
				}, 2*time.Second, 50*time.Millisecond).Should(BeFalse())

				// Wait for context to propagate
				Eventually(func() error {
					return agg.Err()
				}, 2*time.Second, 50*time.Millisecond).Should(Equal(context.Canceled))

				// Try to write after cancel
				// Note: Write may succeed if context propagation is slow,
				// but eventually it should fail
				data := []byte("test data")
				Eventually(func() error {
					_, err := agg.Write(data)
					return err
				}, 2*time.Second, 100*time.Millisecond).Should(HaveOccurred())
			})
		})
	})

	Describe("TC-WR-013: Close()", func() {
		It("TC-WR-014: should close successfully when running", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-015: should close successfully when not running", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-016: should be idempotent", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 10,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Close multiple times
			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())
		})

		It("TC-WR-017: should process pending writes before closing", func() {
			writer := newTestWriter()
			writer.SetDelay(10) // Add delay to writer
			cfg := iotagg.Config{
				BufWriter: 100,
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())

			err = startAndWait(agg, ctx)
			Expect(err).ToNot(HaveOccurred())

			// Wait for aggregator to be running
			Eventually(func() bool {
				return agg.IsRunning()
			}, 2*time.Second, 50*time.Millisecond).Should(BeTrue())

			// Write multiple chunks quickly
			numChunks := 10
			for i := 0; i < numChunks; i++ {
				_, _ = agg.Write([]byte("data")) // Ignore errors as close may happen
			}

			// Wait a tiny bit for at least one write to start processing
			time.Sleep(50 * time.Millisecond)

			// Close
			err = agg.Close()
			Expect(err).ToNot(HaveOccurred())

			// Some data should have been processed
			// (maybe not all if buffer was full, but at least some)
			Expect(writer.GetCallCount()).To(BeNumerically(">", 0))
		})
	})
})
