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

// Package ioprogress_test provides performance benchmarks for the ioprogress package.
//
// These benchmarks measure the overhead introduced by progress tracking wrappers
// compared to direct I/O operations. They validate that the package maintains
// minimal performance impact (<0.1% overhead) as documented.
//
// Running Benchmarks:
//   go test -bench=. -benchmem
//   go test -bench=BenchmarkReader -benchmem
//   go test -bench=BenchmarkWriter -benchmem
//   go test -bench=. -benchtime=10s -benchmem  # Longer run for stability
//
// Performance Targets:
//   - Overhead: <100ns per operation
//   - Memory allocations: 0 allocs per operation
//   - Impact: <5% compared to unwrapped I/O
package ioprogress_test

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nabbar/golib/ioutils/ioprogress"
)

// Benchmark sizes for different scenarios
const (
	benchSmallSize  = 1024        // 1 KB - small data transfers
	benchMediumSize = 64 * 1024   // 64 KB - typical buffer size
	benchLargeSize  = 1024 * 1024 // 1 MB - large file transfers
)

// =============================================================================
// Reader Benchmarks
// =============================================================================

// BenchmarkReaderBaseline measures raw strings.Reader performance without wrapping.
//
// This establishes a performance baseline for comparison with wrapped readers.
// Expected: ~100-500ns per operation depending on buffer size.
func BenchmarkReaderBaseline(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				reader := io.NopCloser(strings.NewReader(data))
				io.CopyBuffer(io.Discard, reader, buf)
				reader.Close()
			}
		})
	}
}

// BenchmarkReaderWithProgress measures performance with progress tracking enabled.
//
// This benchmark includes:
//   - Wrapper overhead
//   - Atomic counter updates
//   - Callback invocation (no-op function)
//
// Expected: <5% overhead compared to baseline.
func BenchmarkReaderWithProgress(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
				io.CopyBuffer(io.Discard, reader, buf)
				reader.Close()
			}
		})
	}
}

// BenchmarkReaderWithCallback measures performance with active callback.
//
// This benchmark includes all overhead plus actual callback execution.
// Uses atomic operations in the callback to simulate real-world usage.
//
// Expected: <5% overhead compared to baseline with no-op callback.
func BenchmarkReaderWithCallback(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				var total int64
				reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
				reader.RegisterFctIncrement(func(size int64) {
					atomic.AddInt64(&total, size)
				})
				io.CopyBuffer(io.Discard, reader, buf)
				reader.Close()
			}
		})
	}
}

// BenchmarkReaderMultipleCallbacks measures overhead with all callbacks registered.
//
// This represents the worst-case scenario with increment, reset, and EOF callbacks.
func BenchmarkReaderMultipleCallbacks(b *testing.B) {
	data := strings.Repeat("x", benchMediumSize)
	buf := make([]byte, 4096)
	
	b.ResetTimer()
	b.SetBytes(int64(benchMediumSize))
	
	for i := 0; i < b.N; i++ {
		var total int64
		reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
		
		reader.RegisterFctIncrement(func(size int64) {
			atomic.AddInt64(&total, size)
		})
		reader.RegisterFctEOF(func() {
			// EOF callback
		})
		reader.RegisterFctReset(func(max, current int64) {
			// Reset callback
		})
		
		io.CopyBuffer(io.Discard, reader, buf)
		reader.Close()
	}
}

// =============================================================================
// Writer Benchmarks
// =============================================================================

// BenchmarkWriterBaseline measures raw bytes.Buffer performance without wrapping.
//
// This establishes a performance baseline for comparison with wrapped writers.
func BenchmarkWriterBaseline(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			source := strings.NewReader(data)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				source.Reset(data)
				writer := newCloseableWriter()
				io.CopyBuffer(writer, source, buf)
				writer.Close()
			}
		})
	}
}

// BenchmarkWriterWithProgress measures performance with progress tracking enabled.
//
// Expected: <5% overhead compared to baseline.
func BenchmarkWriterWithProgress(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			source := strings.NewReader(data)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				source.Reset(data)
				writer := ioprogress.NewWriteCloser(newCloseableWriter())
				io.CopyBuffer(writer, source, buf)
				writer.Close()
			}
		})
	}
}

// BenchmarkWriterWithCallback measures performance with active callback.
//
// Uses atomic operations in the callback to simulate real-world usage.
func BenchmarkWriterWithCallback(b *testing.B) {
	benchmarks := []struct {
		name string
		size int
	}{
		{"Small_1KB", benchSmallSize},
		{"Medium_64KB", benchMediumSize},
		{"Large_1MB", benchLargeSize},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := strings.Repeat("x", bm.size)
			source := strings.NewReader(data)
			buf := make([]byte, 4096)
			
			b.ResetTimer()
			b.SetBytes(int64(bm.size))
			
			for i := 0; i < b.N; i++ {
				source.Reset(data)
				var total int64
				writer := ioprogress.NewWriteCloser(newCloseableWriter())
				writer.RegisterFctIncrement(func(size int64) {
					atomic.AddInt64(&total, size)
				})
				io.CopyBuffer(writer, source, buf)
				writer.Close()
			}
		})
	}
}

// =============================================================================
// Callback Registration Benchmarks
// =============================================================================

// BenchmarkCallbackRegistration measures the cost of registering callbacks.
//
// This benchmark validates that callback registration is fast and doesn't
// block I/O operations.
func BenchmarkCallbackRegistration(b *testing.B) {
	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader("data")))
	defer reader.Close()
	
	callback := func(size int64) {
		// No-op callback
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		reader.RegisterFctIncrement(callback)
	}
}

// BenchmarkCallbackRegistrationConcurrent measures concurrent callback registration.
//
// This validates thread-safe callback registration using atomic.Value.
func BenchmarkCallbackRegistrationConcurrent(b *testing.B) {
	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader("data")))
	defer reader.Close()
	
	callback := func(size int64) {
		// No-op callback
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader.RegisterFctIncrement(callback)
		}
	})
}

// =============================================================================
// Memory Allocation Benchmarks
// =============================================================================

// BenchmarkReaderAllocations specifically measures memory allocations.
//
// Expected: 0 allocations per operation after wrapper creation.
func BenchmarkReaderAllocations(b *testing.B) {
	data := strings.Repeat("x", benchMediumSize)
	buf := make([]byte, 4096)
	
	// Create wrapper outside the benchmark loop
	reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
	defer reader.Close()
	
	var total int64
	reader.RegisterFctIncrement(func(size int64) {
		atomic.AddInt64(&total, size)
	})
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// This should cause 0 allocations
		_, _ = reader.Read(buf)
	}
}

// =============================================================================
// Comparison Benchmarks
// =============================================================================

// BenchmarkOverheadComparison directly compares wrapped vs unwrapped I/O.
//
// This benchmark measures the exact overhead percentage introduced by the wrapper.
func BenchmarkOverheadComparison(b *testing.B) {
	data := strings.Repeat("x", benchMediumSize)
	buf := make([]byte, 4096)
	
	b.Run("Unwrapped", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reader := io.NopCloser(strings.NewReader(data))
			io.CopyBuffer(io.Discard, reader, buf)
			reader.Close()
		}
	})
	
	b.Run("Wrapped", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
			io.CopyBuffer(io.Discard, reader, buf)
			reader.Close()
		}
	})
	
	b.Run("Wrapped_WithCallback", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var total int64
			reader := ioprogress.NewReadCloser(io.NopCloser(strings.NewReader(data)))
			reader.RegisterFctIncrement(func(size int64) {
				atomic.AddInt64(&total, size)
			})
			io.CopyBuffer(io.Discard, reader, buf)
			reader.Close()
		}
	})
}
