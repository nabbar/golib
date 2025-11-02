/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package http_test

import (
	"context"
	"net/http"
	"testing"

	httpComp "github.com/nabbar/golib/config/components/http"
	htpool "github.com/nabbar/golib/httpserver/pool"
)

// BenchmarkNew benchmarks component creation
func BenchmarkNew(b *testing.B) {
	ctx := context.Background()

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpComp.New(ctx, httpComp.DefaultTlsKey, handler)
	}
}

// BenchmarkSetTLSKey benchmarks TLS key updates
func BenchmarkSetTLSKey(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cpt.SetTLSKey("test-key")
	}
}

// BenchmarkSetHandler benchmarks handler updates
func BenchmarkSetHandler(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{
			"test": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cpt.SetHandler(handler)
	}
}

// BenchmarkGetPool benchmarks pool retrieval
func BenchmarkGetPool(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.GetPool()
	}
}

// BenchmarkSetPool benchmarks pool updates
func BenchmarkSetPool(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	handler := func() map[string]http.Handler {
		return map[string]http.Handler{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool := htpool.New(ctx, handler)
		cpt.SetPool(pool)
	}
}

// BenchmarkDependencies benchmarks dependency retrieval
func BenchmarkDependencies(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.Dependencies()
	}
}

// BenchmarkSetDependencies benchmarks dependency updates
func BenchmarkSetDependencies(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	// Initialize with minimal context
	cpt.Init("test", ctx, nil, nil, nil, nil)

	deps := []string{"dep1", "dep2", "dep3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.SetDependencies(deps)
	}
}

// BenchmarkDefaultConfig benchmarks default configuration generation
func BenchmarkDefaultConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpComp.DefaultConfig("")
	}
}

// BenchmarkDefaultConfigIndented benchmarks indented configuration generation
func BenchmarkDefaultConfigIndented(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = httpComp.DefaultConfig("  ")
	}
}

// BenchmarkIsStarted benchmarks started state check
func BenchmarkIsStarted(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.IsStarted()
	}
}

// BenchmarkIsRunning benchmarks running state check
func BenchmarkIsRunning(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.IsRunning()
	}
}

// BenchmarkType benchmarks type retrieval
func BenchmarkType(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpt.Type()
	}
}

// BenchmarkConcurrentGetPool benchmarks concurrent pool access
func BenchmarkConcurrentGetPool(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cpt.GetPool()
		}
	})
}

// BenchmarkConcurrentSetTLSKey benchmarks concurrent TLS key updates
func BenchmarkConcurrentSetTLSKey(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cpt.SetTLSKey("test-key")
		}
	})
}

// BenchmarkConcurrentDependencies benchmarks concurrent dependency access
func BenchmarkConcurrentDependencies(b *testing.B) {
	ctx := context.Background()
	cpt := httpComp.New(ctx, httpComp.DefaultTlsKey, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cpt.Dependencies()
		}
	})
}
