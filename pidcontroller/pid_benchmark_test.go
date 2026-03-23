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

package pidcontroller_test

import (
	"context"
	"testing"
	"time"

	"github.com/nabbar/golib/pidcontroller"
)

// BenchmarkPIDRange measures the performance of the Range function for a standard transition.
func BenchmarkPIDRange(b *testing.B) {
	// Setup standard PID parameters
	pid := pidcontroller.New(0.5, 0.1, 0.05)
	min, max := 0.0, 1000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pid.Range(min, max)
	}
}

// BenchmarkPIDRangeCtx measures the performance of the RangeCtx function with context handling.
func BenchmarkPIDRangeCtx(b *testing.B) {
	// Setup standard PID parameters
	pid := pidcontroller.New(0.5, 0.1, 0.05)
	min, max := 0.0, 1000.0

	// Pre-allocate context for minimal overhead in loop, though cancellation requires new context
	// Using a long timeout effectively tests the loop performance + ctx check overhead
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a fresh context for each iteration if cancellation logic was part of the bench
		// But here we want to benchmark the calculation loop
		_ = pid.RangeCtx(ctx, min, max)
	}
}

// BenchmarkPIDRangeSmallSteps measures performance when steps are small (high iteration count).
func BenchmarkPIDRangeSmallSteps(b *testing.B) {
	// Low P gain means small increments
	pid := pidcontroller.New(0.01, 0.0, 0.0)
	min, max := 0.0, 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pid.Range(min, max)
	}
}

// BenchmarkPIDRangeLargeRange measures performance over a very large numerical range.
func BenchmarkPIDRangeLargeRange(b *testing.B) {
	pid := pidcontroller.New(0.5, 0.0, 0.0)
	min, max := 0.0, 1e6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pid.Range(min, max)
	}
}

// BenchmarkPIDCreation measures the overhead of creating a new PID controller.
func BenchmarkPIDCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pidcontroller.New(0.5, 0.1, 0.1)
	}
}

// BenchmarkPIDRangeCtxTimeout measures performance when context timeout occurs frequently.
// This benchmark involves context creation overhead.
func BenchmarkPIDRangeCtxTimeout(b *testing.B) {
	pid := pidcontroller.New(0.0001, 0.0, 0.0) // Very slow progression
	min, max := 0.0, 1000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Short timeout to force early exit
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
		_ = pid.RangeCtx(ctx, min, max)
		cancel()
	}
}
