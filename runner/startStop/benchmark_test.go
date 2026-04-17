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

package startStop_test

import (
	"context"
	"testing"

	librun "github.com/nabbar/golib/runner/startStop"
)

// noopFunc is a placeholder function that does nothing.
func noopFunc(ctx context.Context) error {
	return nil
}

// blockingFunc is a function that blocks until the context is cancelled.
func blockingFunc(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// BenchmarkStatus measures the overhead of IsRunning and Uptime calls.
func BenchmarkStatus(b *testing.B) {
	run := librun.New(blockingFunc, noopFunc)
	_ = run.Start(context.Background())
	defer func() { _ = run.Stop(context.Background()) }()

	b.Run("IsRunning", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = run.IsRunning()
		}
	})

	b.Run("Uptime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = run.Uptime()
		}
	})
}

// BenchmarkLifecycle measures the performance of Start and Stop transitions.
func BenchmarkLifecycle(b *testing.B) {
	ctx := context.Background()
	run := librun.New(noopFunc, noopFunc)

	b.Run("StartStop_Serial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = run.Start(ctx)
			_ = run.Stop(ctx)
		}
	})

	b.Run("Restart_Serial", func(b *testing.B) {
		_ = run.Start(ctx)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = run.Restart(ctx)
		}
		b.StopTimer()
		_ = run.Stop(ctx)
	})
}

// BenchmarkConcurrency measures performance under high concurrent status checking.
func BenchmarkConcurrency(b *testing.B) {
	run := librun.New(blockingFunc, noopFunc)
	_ = run.Start(context.Background())
	defer func() { _ = run.Stop(context.Background()) }()

	b.Run("Concurrent_Status", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = run.IsRunning()
				_ = run.Uptime()
			}
		})
	})
}

// BenchmarkErrorOperations measures overhead of error tracking.
func BenchmarkErrorOperations(b *testing.B) {
	run := librun.New(noopFunc, noopFunc)

	b.Run("ErrorsList_Empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = run.ErrorsList()
		}
	})
}
