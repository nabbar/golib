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

package atomic_test

import (
	"context"
	"sync/atomic"
	"testing"

	libatm "github.com/nabbar/golib/atomic"
)

// BenchmarkCast measures the overhead of the Cast function with different types.
func BenchmarkCast(b *testing.B) {
	b.Run("Int", func(b *testing.B) {
		var val any = 42
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = libatm.Cast[int](val)
		}
	})

	b.Run("String", func(b *testing.B) {
		var val any = "test"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = libatm.Cast[string](val)
		}
	})

	b.Run("Interface", func(b *testing.B) {
		type testInterface interface {
			Foo()
		}
		type testImpl struct{}

		var val any = testImpl{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = libatm.Cast[testInterface](val)
		}
	})
}

// BenchmarkIsEmpty measures the overhead of the IsEmpty function.
func BenchmarkIsEmpty(b *testing.B) {
	var val any = 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = libatm.IsEmpty[int](val)
	}
}

// BenchmarkValueOperations measures the performance of Value[T] operations.
func BenchmarkValueOperations(b *testing.B) {
	b.Run("Load_NoDefault", func(b *testing.B) {
		v := libatm.NewValue[int]()
		v.Store(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})

	b.Run("Load_WithDefault", func(b *testing.B) {
		v := libatm.NewValueDefault[int](10, 20)
		v.Store(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})

	b.Run("Store_NoDefault", func(b *testing.B) {
		v := libatm.NewValue[int]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.Store(i)
		}
	})

	b.Run("Store_WithDefault", func(b *testing.B) {
		v := libatm.NewValueDefault[int](10, 20)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.Store(i)
		}
	})
}

// BenchmarkContextInterface measures performance on context-like interfaces.
func BenchmarkContextInterface(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v := libatm.NewValue[context.Context]()
	v.Store(ctx)

	b.Run("Load_Context_NoDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := v.Load()
			_ = c.Err()
		}
	})

	vDef := libatm.NewValueDefault[context.Context](context.Background(), context.Background())
	vDef.Store(ctx)

	b.Run("Load_Context_WithDefault", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := vDef.Load()
			_ = c.Err()
		}
	})
}

// BenchmarkNativeComparison comparison with native sync/atomic.Value and type assertion.
func BenchmarkNativeComparison(b *testing.B) {
	b.Run("NativeAtomicValueLoad", func(b *testing.B) {
		var v atomic.Value
		v.Store(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			val := v.Load()
			if val != nil {
				_ = val.(int)
			}
		}
	})

	b.Run("NativeAtomicInt64Load", func(b *testing.B) {
		var v atomic.Int64
		v.Store(42)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})
}
