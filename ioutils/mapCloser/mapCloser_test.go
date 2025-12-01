/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package mapCloser_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/nabbar/golib/ioutils/mapCloser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MapCloser", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
	})

	// Basic Operations tests core functionality: create, add, get, clean, close.
	// These are the fundamental operations that users will perform most often.
	Context("Basic Operations", func() {
		It("should create a new closer", func() {
			closer := New(ctx)
			defer closer.Close()

			Expect(closer).ToNot(BeNil())
			Expect(closer.Len()).To(Equal(0))
		})

		It("should add single closer", func() {
			closer := New(ctx)
			defer closer.Close()

			mock := newMockCloser()
			closer.Add(mock)

			Expect(closer.Len()).To(Equal(1))
		})

		It("should add multiple closers", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(newMockCloser(), newMockCloser(), newMockCloser())

			Expect(closer.Len()).To(Equal(3))
		})

		It("should get all added closers", func() {
			closer := New(ctx)
			defer closer.Close()

			mock1 := newMockCloser()
			mock2 := newMockCloser()
			closer.Add(mock1, mock2)

			closers := closer.Get()
			Expect(closers).To(HaveLen(2))
		})

		It("should clean all closers", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(newMockCloser(), newMockCloser())
			Expect(closer.Len()).To(Equal(2))

			closer.Clean()
			Expect(closer.Len()).To(Equal(0))
		})

		It("should close all closers", func() {
			closer := New(ctx)

			mock1 := newMockCloser()
			mock2 := newMockCloser()
			closer.Add(mock1, mock2)

			err := closer.Close()

			Expect(err).ToNot(HaveOccurred())
			Eventually(func() bool {
				return mock1.IsClosed() && mock2.IsClosed()
			}, "1s").Should(BeTrue())
		})
	})

	// Clone Operations tests the cloning functionality which allows creating
	// independent copies of a closer with the same state but separate storage.
	Context("Clone Operations", func() {
		It("should create independent copy", func() {
			closer1 := New(ctx)
			defer closer1.Close()

			closer1.Add(newMockCloser())

			closer2 := closer1.Clone()
			defer closer2.Close()

			Expect(closer2).ToNot(BeNil())
			Expect(closer2.Len()).To(Equal(1))
		})

		It("should not affect original when modified", func() {
			closer1 := New(ctx)
			defer closer1.Close()

			closer1.Add(newMockCloser())

			closer2 := closer1.Clone()
			defer closer2.Close()
			closer2.Add(newMockCloser())

			Expect(closer1.Len()).To(Equal(1))
			Expect(closer2.Len()).To(Equal(2))
		})
	})

	// Error Handling tests how the closer behaves when closers return errors,
	// including error aggregation and continued operation despite failures.
	Context("Error Handling", func() {
		It("should return error when closer fails", func() {
			closer := New(ctx)

			expectedErr := errors.New("close failed")
			closer.Add(newErrorCloser(expectedErr))

			err := closer.Close()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("close failed"))
		})

		It("should collect multiple errors", func() {
			closer := New(ctx)

			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			closer.Add(newErrorCloser(err1), newErrorCloser(err2))

			err := closer.Close()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error 1"))
			Expect(err.Error()).To(ContainSubstring("error 2"))
		})

		It("should continue closing even after errors", func() {
			closer := New(ctx)

			errorMock := newErrorCloser(errors.New("fail"))
			successMock := newMockCloser()
			closer.Add(errorMock, successMock)

			closer.Close()

			Eventually(func() bool {
				return errorMock.IsClosed() && successMock.IsClosed()
			}, "1s").Should(BeTrue())
		})

		It("should handle nil closers gracefully", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(nil, newMockCloser(), nil)

			Expect(closer.Len()).To(Equal(3))

			// Get should filter nils
			closers := closer.Get()
			Expect(closers).To(HaveLen(1))
		})
	})

	// Context Cancellation tests the automatic cleanup triggered by context
	// cancellation or timeout, which is the core feature of this package.
	Context("Context Cancellation", func() {
		It("should work with cancelled context", func() {
			localCtx, localCancel := context.WithCancel(context.Background())
			closer := New(localCtx)

			mock := newMockCloser()
			closer.Add(mock)

			// Manually close to test functionality
			err := closer.Close()
			Expect(err).ToNot(HaveOccurred())

			localCancel()

			Eventually(func() bool {
				return mock.IsClosed()
			}, "1s").Should(BeTrue())
		})

		It("should work with timeout context", func() {
			timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer timeoutCancel()

			closer := New(timeoutCtx)
			mock := newMockCloser()
			closer.Add(mock)

			// Manually close
			err := closer.Close()
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() bool {
				return mock.IsClosed()
			}, "1s").Should(BeTrue())
		})
	})

	// Concurrency tests verify thread-safety by calling methods concurrently
	// from multiple goroutines. All operations should be safe without data races.
	Context("Concurrency", func() {
		It("should handle concurrent adds", func() {
			closer := New(ctx)
			defer closer.Close()

			var wg sync.WaitGroup
			concurrency := 50

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					closer.Add(newMockCloser())
				}()
			}

			wg.Wait()
			Expect(closer.Len()).To(Equal(concurrency))
		})

		It("should handle concurrent reads", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(newMockCloser(), newMockCloser())

			var wg sync.WaitGroup
			concurrency := 50

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = closer.Get()
					_ = closer.Len()
				}()
			}

			wg.Wait()
		})

		It("should handle concurrent add and get", func() {
			closer := New(ctx)
			defer closer.Close()

			var wg sync.WaitGroup
			operations := 100

			for i := 0; i < operations/2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					closer.Add(newMockCloser())
				}()
			}

			for i := 0; i < operations/2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = closer.Get()
				}()
			}

			wg.Wait()
		})
	})

	// Edge Cases tests boundary conditions and unusual scenarios that users
	// might encounter, ensuring robust behavior in all situations.
	Context("Edge Cases", func() {
		It("should handle empty closer close", func() {
			closer := New(ctx)

			err := closer.Close()

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle many closers", func() {
			closer := New(ctx)
			defer closer.Close()

			for i := 0; i < 100; i++ {
				closer.Add(newMockCloser())
			}

			Expect(closer.Len()).To(Equal(100))
		})

		It("should work with io.Closer interface", func() {
			closer := New(ctx)
			defer closer.Close()

			var iocloser io.Closer = newMockCloser()
			closer.Add(iocloser)

			Expect(closer.Len()).To(Equal(1))
		})

		It("should handle clean and re-add", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(newMockCloser())
			closer.Clean()
			closer.Add(newMockCloser())

			Expect(closer.Len()).To(Equal(1))
		})

		It("should handle repeated clean", func() {
			closer := New(ctx)
			defer closer.Close()

			closer.Add(newMockCloser())
			closer.Clean()
			closer.Clean()

			Expect(closer.Len()).To(Equal(0))
		})

		It("should handle closed context operations", func() {
			localCtx, localCancel := context.WithCancel(context.Background())
			closer := New(localCtx)

			closer.Add(newMockCloser())

			// Close first
			err := closer.Close()
			Expect(err).ToNot(HaveOccurred())

			localCancel()

			// Operations after close should be safe
			closer.Add(newMockCloser())
			Expect(closer.Get()).To(BeEmpty())

			clone := closer.Clone()
			Expect(clone).To(BeNil())
		})

		It("should handle nil operations gracefully", func() {
			closer := New(ctx)
			defer closer.Close()

			// Add nil values
			closer.Add(nil)
			closer.Add(nil, nil)

			Expect(closer.Len()).To(Equal(3))
			Expect(closer.Get()).To(BeEmpty())
		})

		It("should handle Get with mixed closers", func() {
			closer := New(ctx)
			defer closer.Close()

			mock1 := newMockCloser()
			mock2 := newMockCloser()

			closer.Add(mock1, nil, mock2, nil)

			closers := closer.Get()
			Expect(closers).To(HaveLen(2))
		})

		It("should handle Len overflow protection", func() {
			closer := New(ctx)
			defer closer.Close()

			// Add many items to test counter
			for i := 0; i < 1000; i++ {
				closer.Add(newMockCloser())
			}

			Expect(closer.Len()).To(Equal(1000))
		})

		It("should return error on double close", func() {
			closer := New(ctx)

			err1 := closer.Close()
			Expect(err1).ToNot(HaveOccurred())

			err2 := closer.Close()
			Expect(err2).To(HaveOccurred())
		})
	})

	// Robustness tests verify that the closer handles various edge conditions
	// and maintains correct behavior even in unusual circumstances.
	Context("Robustness", func() {
		It("should handle repeated Add operations", func() {
			closer := New(ctx)
			defer closer.Close()

			mock1 := newMockCloser()
			
			// Add same closer multiple times
			closer.Add(mock1)
			closer.Add(mock1)
			closer.Add(mock1)

			Expect(closer.Len()).To(Equal(3))
			
			// Get should return it three times
			result := closer.Get()
			Expect(result).To(HaveLen(3))
		})

		It("should handle Clean after adding many closers", func() {
			closer := New(ctx)
			defer closer.Close()

			// Add many closers
			closers := createMockClosers(100)
			for _, c := range closers {
				closer.Add(c)
			}

			Expect(closer.Len()).To(Equal(100))

			// Clean should remove all
			closer.Clean()
			Expect(closer.Len()).To(Equal(0))
			Expect(closer.Get()).To(BeEmpty())

			// None should be closed
			Expect(testCloserCount(closers...)).To(Equal(0))
		})

		It("should handle Get after Close", func() {
			closer := New(ctx)
			
			mock := newMockCloser()
			closer.Add(mock)

			closer.Close()

			// Get should return empty after close
			result := closer.Get()
			Expect(result).To(BeEmpty())
		})

		It("should handle Add after Close", func() {
			closer := New(ctx)
			
			closer.Add(newMockCloser())
			closer.Close()

			// Add after close should be no-op
			initialLen := closer.Len()
			closer.Add(newMockCloser())
			
			// Length should not increase
			Expect(closer.Len()).To(Equal(initialLen))
		})

		It("should handle Clean after Close", func() {
			closer := New(ctx)
			
			closer.Add(newMockCloser())
			closer.Close()

			// Clean after close should be no-op
			Expect(func() { closer.Clean() }).ToNot(Panic())
		})
	})

	// Performance tests verify that the closer can handle large numbers of
	// closers and high concurrency without significant degradation.
	Context("Performance", func() {
		It("should handle large number of closers efficiently", func() {
			closer := New(ctx)

			mocks := make([]*mockCloser, 500)
			for i := 0; i < 500; i++ {
				mocks[i] = newMockCloser()
				closer.Add(mocks[i])
			}

			start := time.Now()
			err := closer.Close()
			duration := time.Since(start)

			Expect(err).ToNot(HaveOccurred())
			Expect(duration).To(BeNumerically("<", 2*time.Second))
		})

		It("should handle high concurrency", func() {
			closer := New(ctx)
			defer closer.Close()

			var wg sync.WaitGroup
			var count atomic.Int64

			for i := 0; i < 200; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					closer.Add(newMockCloser())
					count.Add(1)
				}()
			}

			wg.Wait()
			Expect(count.Load()).To(Equal(int64(200)))
		})
	})
})
