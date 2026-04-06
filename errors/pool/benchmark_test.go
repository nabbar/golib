/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package pool_test

import (
	"errors"
	"fmt"
	"testing"

	errpol "github.com/nabbar/golib/errors/pool"
)

func BenchmarkPoolOperations(b *testing.B) {
	testErr := errors.New("test error")

	b.Run("Add", func(b *testing.B) {
		p := errpol.New()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Add(testErr)
		}
	})

	sizes := []int{0, 1, 10, 100}
	for _, size := range sizes {
		p := errpol.New()
		for i := 0; i < size; i++ {
			p.Add(fmt.Errorf("error %d", i))
		}

		b.Run(fmt.Sprintf("Len/Size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = p.Len()
			}
		})

		b.Run(fmt.Sprintf("Slice/Size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = p.Slice()
			}
		})

		b.Run(fmt.Sprintf("Error/Size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = p.Error()
			}
		})
	}
}

func BenchmarkSparsePool(b *testing.B) {
	p := errpol.New()
	// Create a sparse pool with high sequence but few errors
	for i := uint64(1); i <= 1000; i++ {
		p.Set(i, errors.New("err"))
		if i%10 != 0 {
			p.Del(i)
		}
	}
	// Now only 100 errors remain, but sequence is at 1000

	b.Run("Len", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = p.Len()
		}
	})

	b.Run("Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = p.Slice()
		}
	})
}
