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

		It("should handle Set as append if index out of bounds", func() {
			err := errors.New("error at large index")
			p.Set(1000000, err)

			Expect(p.Get(0)).To(Equal(err))
			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.MaxId()).To(Equal(uint64(1)))
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
			Expect(p.Get(0)).To(BeNil())
			Expect(p.Get(99)).To(BeNil())
		})

		It("should handle Del on empty pool", func() {
			Expect(func() {
				p.Del(0)
				p.Del(99)
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

			retrieved := p.Get(0)
			Expect(retrieved).To(BeIdenticalTo(originalErr))
		})

		It("should handle errors with special characters", func() {
			specialErr := errors.New("error with\nnewline\tand\ttabs")
			p.Add(specialErr)

			Expect(p.Get(0)).To(Equal(specialErr))
		})
	})

	Describe("Index Boundary Cases", func() {
		It("should handle index 0", func() {
			err := errors.New("first error")
			p.Add(err)

			Expect(p.Get(0)).To(Equal(err))
		})

		It("should maintain correct ordering after deletions", func() {
			p.Add(errors.New("e0"), errors.New("e1"), errors.New("e2"))
			p.Del(1)

			Expect(p.Len()).To(Equal(uint64(2)))
			Expect(p.Get(0).Error()).To(Equal("e0"))
			Expect(p.Get(1).Error()).To(Equal("e2"))
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
			for i := uint64(0); i < 100; i++ {
				Expect(p.Get(i)).To(BeIdenticalTo(err))
			}
		})

		It("should handle repeated Set on same index", func() {
			p.Add(errors.New("e0"))
			for i := 0; i < 10; i++ {
				p.Set(0, fmt.Errorf("iteration %d", i))
			}

			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Get(0).Error()).To(ContainSubstring("iteration 9"))
		})
	})

	Describe("Interleaved Operations", func() {
		It("should handle add-delete-add sequences", func() {
			err1 := errors.New("first")
			err2 := errors.New("second")

			p.Add(err1)
			Expect(p.Len()).To(Equal(uint64(1)))

			p.Del(0)
			Expect(p.Len()).To(Equal(uint64(0)))

			p.Add(err2)
			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Get(0)).To(Equal(err2))
		})

		It("should handle add-clear-add sequences", func() {
			p.Add(errors.New("before clear"))
			p.Clear()
			Expect(p.Len()).To(Equal(uint64(0)))

			p.Add(errors.New("after clear"))
			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.MaxId()).To(Equal(uint64(1)))
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
	})
})
