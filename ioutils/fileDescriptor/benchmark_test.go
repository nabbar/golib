/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package fileDescriptor_test

import (
	"testing"

	"github.com/nabbar/golib/ioutils/fileDescriptor"
)

// Standard Go benchmarks for SystemFileDescriptor.
// Run with: go test -bench . -benchmem
//
// Expected results:
//   - Query: ~500 ns/op, 0 allocs/op
//   - Very consistent timing (low variance)
//   - Zero memory allocations

// BenchmarkQuery measures the performance of querying file descriptor limits.
// This is the most common operation and should be extremely fast.
func BenchmarkQuery(b *testing.B) {
	// Warmup
	fileDescriptor.SystemFileDescriptor(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileDescriptor.SystemFileDescriptor(0)
	}
}

// BenchmarkQueryParallel measures query performance under concurrent load.
// Should scale well as operations are independent and kernel-synchronized.
func BenchmarkQueryParallel(b *testing.B) {
	// Warmup
	fileDescriptor.SystemFileDescriptor(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fileDescriptor.SystemFileDescriptor(0)
		}
	})
}

// BenchmarkQueryWithErrorCheck measures query with full error handling.
// Should have negligible overhead compared to BenchmarkQuery.
func BenchmarkQueryWithErrorCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		current, max, err := fileDescriptor.SystemFileDescriptor(0)
		if err != nil {
			b.Fatal(err)
		}
		if current <= 0 || max < current {
			b.Fatalf("Invalid result: current=%d, max=%d", current, max)
		}
	}
}

// BenchmarkIncrease measures the performance of limit increase operations.
// Note: May fail due to permissions, which is acceptable.
func BenchmarkIncrease(b *testing.B) {
	initial, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		b.Skip("Cannot query initial limits")
	}

	target := initial + 10
	if target > max {
		b.Skip("Cannot test increase: already at maximum")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = fileDescriptor.SystemFileDescriptor(target)
		// Ignore errors (may be permission denied)
	}
}

// BenchmarkSequentialCalls measures overhead of sequential calls.
// Useful for comparing with concurrent access patterns.
func BenchmarkSequentialCalls(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Query twice in sequence
		fileDescriptor.SystemFileDescriptor(0)
		fileDescriptor.SystemFileDescriptor(0)
	}
}

// BenchmarkAlternatingOperations measures mixed query/increase pattern.
// Simulates a realistic usage pattern where checks alternate with increases.
func BenchmarkAlternatingOperations(b *testing.B) {
	initial, max, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		b.Skip("Cannot query initial limits")
	}

	target := initial + 5
	if target > max {
		target = initial
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			fileDescriptor.SystemFileDescriptor(0)
		} else {
			fileDescriptor.SystemFileDescriptor(target)
		}
	}
}

// BenchmarkValueExtraction measures the cost of extracting return values.
// Verifies that return value handling doesn't add significant overhead.
func BenchmarkValueExtraction(b *testing.B) {
	var current, max int
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		current, max, err = fileDescriptor.SystemFileDescriptor(0)
		// Use values to prevent optimization
		_ = current
		_ = max
		_ = err
	}
}

// BenchmarkCompareWithReset measures impact of resetting to same value.
// Useful for understanding no-op cost.
func BenchmarkCompareWithReset(b *testing.B) {
	current, _, err := fileDescriptor.SystemFileDescriptor(0)
	if err != nil {
		b.Skip("Cannot query limits")
	}

	b.Run("Query", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fileDescriptor.SystemFileDescriptor(0)
		}
	})

	b.Run("SetSameValue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fileDescriptor.SystemFileDescriptor(current)
		}
	})
}

// BenchmarkCacheMiss measures the cost when values might not be cached.
// Varies the requested value to test different code paths.
func BenchmarkCacheMiss(b *testing.B) {
	initial, _, _ := fileDescriptor.SystemFileDescriptor(0)

	values := []int{0, -1, initial, initial + 1, initial + 100}

	for _, val := range values {
		b.Run(b.Name(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fileDescriptor.SystemFileDescriptor(val)
			}
		})
	}
}
