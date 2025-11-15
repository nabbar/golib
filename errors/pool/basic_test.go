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

var _ = Describe("Basic Operations", func() {
	var p pool.Pool

	BeforeEach(func() {
		p = pool.New()
	})

	Describe("Pool Creation", func() {
		It("should create a new pool", func() {
			Expect(p).NotTo(BeNil())
		})

		It("should start with length 0", func() {
			Expect(p.Len()).To(Equal(uint64(0)))
		})

		It("should have MaxId 0 when empty", func() {
			Expect(p.MaxId()).To(Equal(uint64(0)))
		})

		It("should return nil for Last() when empty", func() {
			Expect(p.Last()).To(BeNil())
		})

		It("should return nil for Error() when empty", func() {
			Expect(p.Error()).To(BeNil())
		})

		It("should return empty slice when empty", func() {
			Expect(p.Slice()).To(BeEmpty())
		})
	})

	Describe("Add Operation", func() {
		It("should add a single error", func() {
			err := errors.New("test error")
			p.Add(err)

			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Last()).To(Equal(err))
		})

		It("should add multiple errors", func() {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			err3 := errors.New("error 3")

			p.Add(err1, err2, err3)

			Expect(p.Len()).To(Equal(uint64(3)))
		})

		It("should ignore nil errors", func() {
			err1 := errors.New("error 1")
			p.Add(err1, nil, nil)

			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Last()).To(Equal(err1))
		})

		It("should handle empty Add call", func() {
			p.Add()
			Expect(p.Len()).To(Equal(uint64(0)))
		})

		It("should handle Add with only nil errors", func() {
			p.Add(nil, nil, nil)
			Expect(p.Len()).To(Equal(uint64(0)))
		})

		It("should maintain order of additions", func() {
			err1 := errors.New("first")
			err2 := errors.New("second")
			err3 := errors.New("third")

			p.Add(err1)
			p.Add(err2)
			p.Add(err3)

			Expect(p.Get(1)).To(Equal(err1))
			Expect(p.Get(2)).To(Equal(err2))
			Expect(p.Get(3)).To(Equal(err3))
		})
	})

	Describe("Get Operation", func() {
		BeforeEach(func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)
		})

		It("should get error by index", func() {
			err := p.Get(1)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("error 1"))
		})

		It("should return nil for non-existent index", func() {
			err := p.Get(999)
			Expect(err).To(BeNil())
		})

		It("should return nil for index 0", func() {
			err := p.Get(0)
			Expect(err).To(BeNil())
		})

		It("should get all errors correctly", func() {
			for i := uint64(1); i <= 3; i++ {
				err := p.Get(i)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal(fmt.Sprintf("error %d", i)))
			}
		})
	})

	Describe("Set Operation", func() {
		It("should set error at specific index", func() {
			newErr := errors.New("new error")
			p.Set(5, newErr)

			Expect(p.Get(5)).To(Equal(newErr))
		})

		It("should overwrite existing error", func() {
			p.Add(errors.New("original"))
			newErr := errors.New("replacement")

			p.Set(1, newErr)

			Expect(p.Get(1)).To(Equal(newErr))
		})

		It("should ignore nil errors", func() {
			p.Add(errors.New("original"))
			p.Set(1, nil)

			// Should still have the original error
			Expect(p.Get(1)).NotTo(BeNil())
		})

		It("should allow setting at any index", func() {
			p.Set(100, errors.New("error at 100"))
			Expect(p.Get(100)).NotTo(BeNil())
			Expect(p.MaxId()).To(Equal(uint64(100)))
		})
	})

	Describe("Delete Operation", func() {
		BeforeEach(func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)
		})

		It("should delete error by index", func() {
			p.Del(2)

			Expect(p.Get(2)).To(BeNil())
			Expect(p.Len()).To(Equal(uint64(2)))
		})

		It("should not affect other errors", func() {
			p.Del(2)

			Expect(p.Get(1)).NotTo(BeNil())
			Expect(p.Get(3)).NotTo(BeNil())
		})

		It("should handle deletion of non-existent index", func() {
			p.Del(999)
			Expect(p.Len()).To(Equal(uint64(3)))
		})

		It("should handle deletion of index 0", func() {
			p.Del(0)
			Expect(p.Len()).To(Equal(uint64(3)))
		})
	})

	Describe("Clear Operation", func() {
		BeforeEach(func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)
		})

		It("should clear all errors", func() {
			p.Clear()

			Expect(p.Len()).To(Equal(uint64(0)))
			Expect(p.Slice()).To(BeEmpty())
		})

		It("should reset MaxId after clear", func() {
			p.Clear()
			Expect(p.MaxId()).To(Equal(uint64(0)))
		})

		It("should allow adding after clear", func() {
			p.Clear()
			newErr := errors.New("new error")
			p.Add(newErr)

			Expect(p.Len()).To(Equal(uint64(1)))
			Expect(p.Last()).To(Equal(newErr))
		})
	})

	Describe("Len Operation", func() {
		It("should return correct length", func() {
			Expect(p.Len()).To(Equal(uint64(0)))

			p.Add(errors.New("error 1"))
			Expect(p.Len()).To(Equal(uint64(1)))

			p.Add(errors.New("error 2"))
			Expect(p.Len()).To(Equal(uint64(2)))
		})

		It("should decrease after deletion", func() {
			p.Add(errors.New("error 1"), errors.New("error 2"))
			Expect(p.Len()).To(Equal(uint64(2)))

			p.Del(1)
			Expect(p.Len()).To(Equal(uint64(1)))
		})
	})

	Describe("MaxId Operation", func() {
		It("should return highest index", func() {
			p.Add(errors.New("error 1"))
			Expect(p.MaxId()).To(Equal(uint64(1)))

			p.Add(errors.New("error 2"))
			Expect(p.MaxId()).To(Equal(uint64(2)))
		})

		It("should handle sparse indices", func() {
			p.Set(5, errors.New("error at 5"))
			p.Set(10, errors.New("error at 10"))
			p.Set(3, errors.New("error at 3"))

			Expect(p.MaxId()).To(Equal(uint64(10)))
		})

		It("should update after deletion of max", func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)

			p.Del(3)
			Expect(p.MaxId()).To(Equal(uint64(2)))
		})
	})

	Describe("Last Operation", func() {
		It("should return last error", func() {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")

			p.Add(err1)
			Expect(p.Last()).To(Equal(err1))

			p.Add(err2)
			Expect(p.Last()).To(Equal(err2))
		})

		It("should return nil when empty", func() {
			Expect(p.Last()).To(BeNil())
		})

		It("should return error at MaxId", func() {
			p.Set(5, errors.New("error at 5"))
			p.Set(10, errors.New("error at 10"))

			Expect(p.Last()).To(Equal(p.Get(10)))
		})
	})

	Describe("Slice Operation", func() {
		It("should return all errors as slice", func() {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			err3 := errors.New("error 3")

			p.Add(err1, err2, err3)

			slice := p.Slice()
			Expect(slice).To(HaveLen(3))
			Expect(slice).To(ContainElement(err1))
			Expect(slice).To(ContainElement(err2))
			Expect(slice).To(ContainElement(err3))
		})

		It("should return empty slice when no errors", func() {
			Expect(p.Slice()).To(BeEmpty())
		})

		It("should not include deleted errors", func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
				errors.New("error 3"),
			)

			p.Del(2)

			slice := p.Slice()
			Expect(slice).To(HaveLen(2))
		})
	})

	Describe("Error Operation", func() {
		It("should return nil when no errors", func() {
			Expect(p.Error()).To(BeNil())
		})

		It("should return combined error when errors exist", func() {
			p.Add(
				errors.New("error 1"),
				errors.New("error 2"),
			)

			err := p.Error()
			Expect(err).NotTo(BeNil())
		})

		It("should create error from error slice", func() {
			p.Add(
				errors.New("first error"),
				errors.New("second error"),
			)

			err := p.Error()
			Expect(err).NotTo(BeNil())

			// Verify the error is created from the slice
			slice := p.Slice()
			Expect(slice).To(HaveLen(2))
		})
	})
})
