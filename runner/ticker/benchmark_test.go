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

package ticker

import (
	"context"
	"sync"
	"testing"
	"time"

	libsrv "github.com/nabbar/golib/runner"
)

func BenchmarkLifecycleStartStop(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		t := New(time.Hour, func(ctx context.Context, tck libsrv.TickUpdate) error { return nil })
		_ = t.Start(ctx)
		_ = t.Stop(ctx)
	}
}

func BenchmarkLifecycleRestart(b *testing.B) {
	ctx := context.Background()
	t := New(time.Hour, func(ctx context.Context, tck libsrv.TickUpdate) error { return nil })
	_ = t.Start(ctx)
	defer t.Stop(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.Restart(ctx)
	}
}

func BenchmarkLifecycleConcurrentExchange(b *testing.B) {
	var mu sync.Mutex
	var current Ticker
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			next := New(time.Hour, func(ctx context.Context, tck libsrv.TickUpdate) error { return nil })
			_ = next.Start(ctx)

			mu.Lock()
			old := current
			current = next
			mu.Unlock()

			if old != nil {
				_ = old.Stop(ctx)
			}
		}
	})
	if current != nil {
		_ = current.Stop(ctx)
	}
}

func BenchmarkStatusIsRunning(b *testing.B) {
	t := New(time.Hour, func(ctx context.Context, tck libsrv.TickUpdate) error { return nil })
	_ = t.Start(context.Background())
	defer t.Stop(context.Background())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.IsRunning()
	}
}

func BenchmarkStatusUptime(b *testing.B) {
	t := New(time.Hour, func(ctx context.Context, tck libsrv.TickUpdate) error { return nil })
	_ = t.Start(context.Background())
	defer t.Stop(context.Background())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.Uptime()
	}
}

func BenchmarkStatusErrorsLast(b *testing.B) {
	t := New(time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
		return context.DeadlineExceeded
	})
	_ = t.Start(context.Background())
	time.Sleep(2 * time.Millisecond)
	defer t.Stop(context.Background())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.ErrorsLast()
	}
}

func BenchmarkContentionMixed(b *testing.B) {
	t := New(time.Millisecond, func(ctx context.Context, tck libsrv.TickUpdate) error {
		return nil
	})
	_ = t.Start(context.Background())
	defer t.Stop(context.Background())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = t.IsRunning()
			_ = t.Uptime()
			_ = t.ErrorsLast()
		}
	})
}
