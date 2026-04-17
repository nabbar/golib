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

package aggregator_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"
)

// Performance benchmarks for Aggregator operations.
// These benchmarks measure the performance and resource characteristics
// of write aggregation, lifecycle management, and monitoring metrics.

// benchWriter is a no-op writer for benchmarks to avoid memory exhaustion
// and mutex contention that would occur with testWriter.
func benchWriter(p []byte) (int, error) {
	return len(p), nil
}

// fillData populates a byte slice with dummy data to ensure benchmarks
// reflect processing of non-empty content.
func fillData(p []byte) {
	for i := range p {
		p[i] = byte(i % 256)
	}
}

// BenchmarkWriteThroughput measures bytes per second throughput.
func BenchmarkWriteThroughput(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data := make([]byte, 4096) // 4KB chunks
	fillData(data)

	cfg := iotagg.Config{
		BufWriter:  1000,
		BufMaxSize: 4096, // Matching data size
		FctWriter:  benchWriter,
	}

	agg, _ := iotagg.New(ctx, cfg)
	_ = agg.Start(ctx)
	defer agg.Close()

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agg.Write(data)
	}
	b.StopTimer()
}

// BenchmarkWrite measures performance of Write with varying buffer and data sizes.
func BenchmarkWrite(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	smallData := []byte("test data")
	mediumData := make([]byte, 1024)
	fillData(mediumData)
	largeData := make([]byte, 10240)
	fillData(largeData)

	benchmarks := []struct {
		name       string
		bufSize    int
		bufMaxSize int
		data       []byte
		numWrites  int
	}{
		{"SmallBuf_SmallData", 10, 64, smallData, 100},
		{"MedBuf_SmallData", 100, 64, smallData, 100},
		{"LrgBuf_SmallData", 1000, 64, smallData, 100},
		{"LrgBuf_1KBData", 1000, 1024, mediumData, 100},
		{"LrgBuf_10KBData", 1000, 10240, largeData, 50},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cfg := iotagg.Config{
				BufWriter:  bm.bufSize,
				BufMaxSize: bm.bufMaxSize,
				FctWriter:  benchWriter,
			}
			agg, _ := iotagg.New(ctx, cfg)
			_ = agg.Start(ctx)
			defer agg.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < bm.numWrites; j++ {
					_, _ = agg.Write(bm.data)
				}
			}
			b.StopTimer()
			runtime.GC()
		})
	}
}

// BenchmarkConcurrentWrite measures performance of concurrent writes with varying goroutine counts.
func BenchmarkConcurrentWrite(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data := []byte("test data")
	cfg := iotagg.Config{
		BufWriter:  1000,
		BufMaxSize: 64, // Matching small data size
		FctWriter:  benchWriter,
	}

	writesPerGoroutine := 50

	goroutines := []int{1, 5, 10, 20}

	for _, g := range goroutines {
		b.Run(fmt.Sprintf("%d_Goroutines", g), func(b *testing.B) {
			agg, _ := iotagg.New(ctx, cfg)
			_ = agg.Start(ctx)
			defer agg.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(g)
				for j := 0; j < g; j++ {
					go func() {
						defer wg.Done()
						for k := 0; k < writesPerGoroutine; k++ {
							_, _ = agg.Write(data)
						}
					}()
				}
				wg.Wait()
			}
			b.StopTimer()
			runtime.GC()
		})
	}
}

// BenchmarkCallbacks measures overhead of Sync and Async callbacks on Write performance.
func BenchmarkCallbacks(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data := []byte("test data")
	cfg := iotagg.Config{
		AsyncTimer: 10 * time.Millisecond,
		AsyncMax:   -1,
		AsyncFct:   func(ctx context.Context) {},
		SyncTimer:  10 * time.Millisecond,
		SyncFct:    func(ctx context.Context) {},
		BufWriter:  1000,
		BufMaxSize: 64,
		FctWriter:  benchWriter,
	}

	agg, _ := iotagg.New(ctx, cfg)
	_ = agg.Start(ctx)
	defer agg.Close()

	b.Run("WithActiveCallbacks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agg.Write(data)
		}
	})
}

// BenchmarkMetrics measures performance of monitoring metrics read operations.
func BenchmarkMetrics(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := iotagg.Config{
		BufWriter:  1000,
		BufMaxSize: 64,
		FctWriter:  benchWriter,
	}

	agg, _ := iotagg.New(ctx, cfg)
	_ = agg.Start(ctx)
	defer agg.Close()

	// Write some data to populate metrics
	for i := 0; i < 100; i++ {
		_, _ = agg.Write([]byte("test data"))
	}

	b.Run("NbWaiting", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = agg.NbWaiting()
		}
	})

	b.Run("NbProcessing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = agg.NbProcessing()
		}
	})

	b.Run("SizeWaiting", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = agg.SizeWaiting()
		}
	})

	b.Run("SizeProcessing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = agg.SizeProcessing()
		}
	})

	b.Run("AllMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = agg.NbWaiting()
			_ = agg.NbProcessing()
			_ = agg.SizeWaiting()
			_ = agg.SizeProcessing()
		}
	})

	b.Run("WriteWithMetrics", func(b *testing.B) {
		data := []byte("test data for metrics")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = agg.Write(data)
			_ = agg.NbWaiting()
			_ = agg.NbProcessing()
			_ = agg.SizeWaiting()
			_ = agg.SizeProcessing()
		}
	})
}
