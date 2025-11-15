/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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
	"context"

	. "github.com/nabbar/golib/httpserver/pool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool", func() {
	Describe("Pool Creation", func() {
		It("should create empty pool", func() {
			pool := New(nil, nil)

			Expect(pool).ToNot(BeNil())
			Expect(pool.Len()).To(Equal(0))
		})

		It("should create pool with context", func() {
			pool := New(context.Background(), nil)
			Expect(pool).ToNot(BeNil())
		})
	})

	Describe("Pool Management", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)
		})

		It("should have zero length when empty", func() {
			Expect(pool.Len()).To(Equal(0))
		})

		It("should clean pool", func() {
			pool.Clean()
			Expect(pool.Len()).To(Equal(0))
		})
	})

	Describe("Pool Filter Operations", func() {
		var pool Pool

		BeforeEach(func() {
			pool = New(nil, nil)
		})

		It("should check if server exists", func() {
			exists := pool.Has("127.0.0.1:8080")
			Expect(exists).To(BeFalse())
		})

		It("should get monitor names", func() {
			names := pool.MonitorNames()
			Expect(names).ToNot(BeNil())
			Expect(len(names)).To(Equal(0))
		})
	})

	Describe("Pool Clone", func() {
		It("should clone pool", func() {
			original := New(nil, nil)
			ctx := context.Background()

			cloned := original.Clone(ctx)

			Expect(cloned).ToNot(BeNil())
			Expect(cloned).ToNot(Equal(original))
		})

		It("should clone empty pool", func() {
			original := New(nil, nil)

			cloned := original.Clone(context.Background())

			Expect(cloned.Len()).To(Equal(0))
		})
	})
})
