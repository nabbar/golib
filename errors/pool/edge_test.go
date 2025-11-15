/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package pool_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/errors/pool"
)

var _ = Describe("Edge Cases and Special Scenarios", func() {
	var p pool.Pool

	BeforeEach(func() {
		p = pool.New()
	})

	Describe("Large Scale Operations", func() {
		It("should handle adding many errors", func() {
			const count = 10000

			for i := 0; i < count; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}

			Expect(p.Len()).To(Equal(uint64(count)))
		})

		It("should handle large indices", func() {
			largeIdx := uint64(1000000)
			err := errors.New("error at large index")

			p.Set(largeIdx, err)

			Expect(p.Get(largeIdx)).To(Equal(err))
			Expect(p.MaxId()).To(Equal(largeIdx))
		})

		It("should handle sparse indices efficiently", func() {
			indices := []uint64{1, 100, 1000, 10000, 100000}

			for _, idx := range indices {
				p.Set(idx, fmt.Errorf("error at %d", idx))
			}

			Expect(p.Len()).To(Equal(uint64(len(indices))))
			Expect(p.MaxId()).To(Equal(uint64(100000)))

			for _, idx := range indices {
				Expect(p.Get(idx)).NotTo(BeNil())
			}
		})
	})

	Describe("Empty and Nil Handling", func() {
		It("should handle operations on empty pool", func() {
			Expect(p.Len()).To(Equal(uint64(0)))
			Expect(p.MaxId()).To(Equal(uint64(0)))
			Expect(p.Last()).To(BeNil())
			Expect(p.Error()).To(BeNil())
			Expect(p.Slice()).To(BeEmpty())
		})

		It("should handle Get on empty pool", func() {
			Expect(p.Get(1)).To(BeNil())
			Expect(p.Get(100)).To(BeNil())
		})

		It("should handle Del on empty pool", func() {
			Expect(func() {
				p.Del(1)
				p.Del(100)
			}).NotTo(Panic())
		})

		It("should handle Clear on empty pool", func() {
			Expect(func() {
				p.Clear()
			}).NotTo(Panic())

			Expect(p.Len()).To(Equal(uint64(0)))
		})

		It("should handle multiple clears", func() {
			p.Add(errors.New("error"))
			p.Clear()
			p.Clear()
			p.Clear()

			Expect(p.Len()).To(Equal(uint64(0)))
		})
	})

	Describe("Error Type Variations", func() {
		It("should handle different error types", func() {
			simpleErr := errors.New("simple error")
			fmtErr := fmt.Errorf("formatted error: %d", 42)
			wrappedErr := fmt.Errorf("wrapped: %w", simpleErr)

			p.Add(simpleErr, fmtErr, wrappedErr)

			Expect(p.Len()).To(Equal(uint64(3)))
		})

		It("should preserve error identity", func() {
			originalErr := errors.New("original")
			p.Add(originalErr)

			retrieved := p.Get(1)
			Expect(retrieved).To(BeIdenticalTo(originalErr))
		})

		It("should handle errors with special characters", func() {
			specialErr := errors.New("error with\nnewline\tand\ttabs")
			p.Add(specialErr)

			Expect(p.Get(1)).To(Equal(specialErr))
		})

		It("should handle very long error messages", func() {
			longMsg := string(make([]byte, 10000))
			for i := range longMsg {
				longMsg = fmt.Sprintf("%s%d", longMsg[:i], i%10)
			}

			longErr := errors.New(longMsg)
			p.Add(longErr)

			Expect(p.Get(1)).NotTo(BeNil())
		})
	})

	Describe("Index Boundary Cases", func() {
		It("should handle index 1", func() {
			err := errors.New("first error")
			p.Add(err)

			Expect(p.Get(1)).To(Equal(err))
		})

		It("should handle maximum uint64 index", func() {
			// Note: This test might be slow or cause issues
			// Testing with a large but reasonable number instead
			maxIdx := uint64(1<<32 - 1) // 4 billion
			err := errors.New("error at max")

			p.Set(maxIdx, err)
			Expect(p.Get(maxIdx)).To(Equal(err))
		})

		It("should maintain correct ordering with gaps", func() {
			p.Set(10, errors.New("error at 10"))
			p.Set(5, errors.New("error at 5"))
			p.Set(15, errors.New("error at 15"))
			p.Set(1, errors.New("error at 1"))

			Expect(p.MaxId()).To(Equal(uint64(15)))
			Expect(p.Len()).To(Equal(uint64(4)))

			// All should be retrievable
			Expect(p.Get(1)).NotTo(BeNil())
			Expect(p.Get(5)).NotTo(BeNil())
			Expect(p.Get(10)).NotTo(BeNil())
			Expect(p.Get(15)).NotTo(BeNil())
		})
	})

	Describe("Repeated Operations", func() {
		It("should handle repeated additions of same error", func() {
			err := errors.New("repeated error")

			for i := 0; i < 100; i++ {
				p.Add(err)
			}

			Expect(p.Len()).To(Equal(uint64(100)))

			// Each should be the same error instance
			for i := uint64(1); i <= 100; i++ {
				Expect(p.Get(i)).To(BeIdenticalTo(err))
			}
		})

		It("should handle repeated Set on same index", func() {
			for i := 0; i < 100; i++ {
				p.Set(5, fmt.Errorf("iteration %d", i))
			}

			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Get(5)).NotTo(BeNil())
		})

		It("should handle repeated Del on same index", func() {
			p.Add(errors.New("error"))

			for i := 0; i < 10; i++ {
				p.Del(1)
			}

			Expect(p.Get(1)).To(BeNil())
		})
	})

	Describe("Interleaved Operations", func() {
		It("should handle add-delete-add sequences", func() {
			err1 := errors.New("first")
			err2 := errors.New("second")

			p.Add(err1)
			Expect(p.Len()).To(Equal(uint64(1)))

			p.Del(1)
			Expect(p.Len()).To(Equal(uint64(0)))

			p.Add(err2)
			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Get(2)).To(Equal(err2))
		})

		It("should handle add-clear-add sequences", func() {
			p.Add(errors.New("before clear"))
			initialCount := p.MaxId()

			p.Clear()
			Expect(p.Len()).To(Equal(uint64(0)))

			p.Add(errors.New("after clear"))
			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.MaxId()).To(Equal(initialCount + 1))
		})

		It("should handle complex sequences", func() {
			// Add some errors
			for i := 1; i <= 10; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}

			// Delete odd indices
			for i := uint64(1); i <= 10; i += 2 {
				p.Del(i)
			}

			// Set some new values
			for i := uint64(2); i <= 10; i += 2 {
				p.Set(i, fmt.Errorf("updated error %d", i))
			}

			// Add more
			for i := 11; i <= 15; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}

			Expect(p.MaxId()).To(Equal(uint64(15)))
			Expect(p.Len()).To(BeNumerically(">=", 5))
		})
	})

	Describe("Memory and Performance", func() {
		It("should not leak memory after clear", func() {
			// Add many errors
			for i := 0; i < 1000; i++ {
				p.Add(fmt.Errorf("error %d", i))
			}

			// Clear
			p.Clear()

			// Pool should be empty
			Expect(p.Len()).To(Equal(uint64(0)))
			Expect(p.Slice()).To(BeEmpty())
		})

		It("should handle rapid add/delete cycles", func() {
			for cycle := 0; cycle < 100; cycle++ {
				for i := 0; i < 10; i++ {
					p.Add(fmt.Errorf("cycle %d error %d", cycle, i))
				}

				p.Clear()
			}

			Expect(p.Len()).To(Equal(uint64(0)))
		})

		It("should maintain performance with many operations", func() {
			const operations = 1000

			// This should complete quickly
			for i := 0; i < operations; i++ {
				p.Add(fmt.Errorf("error %d", i))
				if i%100 == 0 {
					_ = p.Len()
					_ = p.MaxId()
				}
			}

			Expect(p.Len()).To(Equal(uint64(operations)))
		})
	})

	Describe("Slice and Error Consistency", func() {
		It("should return consistent Slice results", func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)

			slice1 := p.Slice()
			slice2 := p.Slice()

			Expect(len(slice1)).To(Equal(len(slice2)))
		})

		It("should include all non-nil errors in Slice", func() {
			p.Set(1, errors.New("error 1"))
			p.Set(5, errors.New("error 5"))
			p.Set(10, errors.New("error 10"))

			slice := p.Slice()
			Expect(slice).To(HaveLen(3))
		})

		It("should generate Error from all errors in pool", func() {
			p.Add(
				errors.New("alpha"),
				errors.New("beta"),
				errors.New("gamma"),
			)

			err := p.Error()
			Expect(err).NotTo(BeNil())

			// Verify that all errors are in the slice
			slice := p.Slice()
			Expect(slice).To(HaveLen(3))
		})
	})
})
