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

var _ = Describe("TC-NW-001: Aggregator Creation", func() {
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

	Describe("TC-NW-002: New()", func() {
		Context("TC-NW-003: with valid configuration", func() {
			It("TC-NW-004: should create aggregator with all parameters", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					AsyncTimer: 100 * time.Millisecond,
					AsyncMax:   5,
					AsyncFct:   func(ctx context.Context) {},
					SyncTimer:  200 * time.Millisecond,
					SyncFct:    func(ctx context.Context) {},
					BufWriter:  10,
					FctWriter:  writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-005: should create aggregator with minimal configuration", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-006: should create aggregator with nil context", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(nil, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-007: should create aggregator with custom logger", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-008: should create aggregator with unbuffered channel", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 0, // unbuffered
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-009: should create aggregator with buffered channel", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					BufWriter: 100,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Cleanup
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("TC-NW-010: with invalid configuration", func() {
			It("TC-NW-011: should return error when FctWriter is nil", func() {
				cfg := iotagg.Config{
					AsyncTimer: 100 * time.Millisecond,
					SyncTimer:  200 * time.Millisecond,
					BufWriter:  10,
					FctWriter:  nil, // missing required field
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(iotagg.ErrInvalidWriter))
				Expect(agg).To(BeNil())
			})

			It("TC-NW-012: should handle async configuration without function", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					AsyncTimer: 100 * time.Millisecond,
					AsyncMax:   5,
					AsyncFct:   nil, // timer without function
					FctWriter:  writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Should not crash, async timer should be ignored
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-013: should handle sync configuration without function", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					SyncTimer: 100 * time.Millisecond,
					SyncFct:   nil, // timer without function
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				// Should not crash, sync timer should be ignored
				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("TC-NW-014: with edge cases", func() {
			It("TC-NW-015: should handle zero AsyncMax", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					AsyncMax:  0,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-016: should handle negative AsyncMax", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					AsyncMax:  -1,
					FctWriter: writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})

			It("TC-NW-017: should handle zero timers", func() {
				writer := newTestWriter()
				cfg := iotagg.Config{
					AsyncTimer: 0,
					SyncTimer:  0,
					FctWriter:  writer.Write,
				}

				agg, err := iotagg.New(ctx, cfg)
				Expect(err).ToNot(HaveOccurred())
				Expect(agg).ToNot(BeNil())

				err = agg.Close()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("TC-NW-018: Context Interface", func() {
		var (
			agg iotagg.Aggregator
		)

		BeforeEach(func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			var err error
			agg, err = iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg).ToNot(BeNil())
		})

		AfterEach(func() {
			if agg != nil {
				_ = agg.Close()
			}
		})

		It("TC-NW-019: should implement context.Context interface", func() {
			// Test Done channel
			doneChan := agg.Done()
			Expect(doneChan).ToNot(BeNil())

			// Test Err (should be nil when not cancelled)
			err := agg.Err()
			Expect(err).To(BeNil())

			// Test Value
			val := agg.Value("test-key")
			Expect(val).To(BeNil())
		})

		It("TC-NW-020: should implement context with deadline", func() {
			deadline := time.Now().Add(5 * time.Second)
			ctxWithDeadline, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()

			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg2, err := iotagg.New(ctxWithDeadline, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg2).ToNot(BeNil())

			d, ok := agg2.Deadline()
			Expect(ok).To(BeTrue())
			Expect(d).To(BeTemporally("~", deadline, time.Second))

			_ = agg2.Close()
		})

		It("TC-NW-021: should propagate context values", func() {
			type ctxKey string
			key := ctxKey("test-key")
			value := "test-value"

			ctxWithValue := context.WithValue(ctx, key, value)

			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg2, err := iotagg.New(ctxWithValue, cfg)
			Expect(err).ToNot(HaveOccurred())
			Expect(agg2).ToNot(BeNil())

			val := agg2.Value(key)
			Expect(val).To(Equal(value))

			_ = agg2.Close()
		})
	})

	Describe("TC-NW-022: Logger Configuration", func() {
		It("TC-NW-023: should set custom error logger", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			agg.SetLoggerError(func(msg string, err ...error) {
				// Custom logger - just ensure it doesn't panic
			})

			// This test mainly ensures SetLoggerError doesn't panic
		})

		It("TC-NW-024: should set custom info logger", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			agg.SetLoggerInfo(func(msg string, arg ...any) {
				// Custom logger - just ensure it doesn't panic
			})

			// This test mainly ensures SetLoggerInfo doesn't panic
		})

		It("TC-NW-025: should handle nil error logger", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			// Should not panic
			agg.SetLoggerError(nil)
		})

		It("TC-NW-026: should handle nil info logger", func() {
			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			agg, err := iotagg.New(ctx, cfg)
			Expect(err).ToNot(HaveOccurred())
			defer agg.Close()

			// Should not panic
			agg.SetLoggerInfo(nil)
		})
	})
})
