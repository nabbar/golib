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
	"runtime"
	"sync"
	"time"

	iotagg "github.com/nabbar/golib/ioutils/aggregator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega/gmeasure"
)

// Performance benchmarks for Aggregator operations.
// These benchmarks measure the performance and resource characteristics
// of write aggregation, lifecycle management, and monitoring metrics.
//
// Benchmarks are organized following patterns from ioutils/multi:
//   - Aggregated experiments grouping related variations
//   - Systematic variations (buffer sizes, goroutine counts, data sizes)
//   - Real-world scenario testing
//   - Statistical analysis with gmeasure
//
// Run with: go test -v to see performance reports.
var _ = Describe("[TC-BC] Aggregator Performance Benchmarks", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(testCtx, time.Minute)
	})

	AfterEach(func() {
		if cancel != nil {
			cancel()
		}
		runtime.GC()
	})

	Describe("TC-BC-001: Write operations", func() {
		It("TC-BC-002: should benchmark Write with varying buffer sizes and data sizes", func() {
			experiment := NewExperiment("Write operations")
			AddReportEntry(experiment.Name, experiment)

			// Small data (10 bytes)
			smallData := []byte("test data")

			// Medium data (1KB)
			mediumData := make([]byte, 1024)

			// Large data (10KB)
			largeData := make([]byte, 10240)

			writer := newTestWriter()
			cfg := iotagg.Config{
				FctWriter: writer.Write,
			}

			cfg.BufWriter = 10
			aggsml, _ := iotagg.New(ctx, cfg)
			defer aggsml.Close()

			cfg.BufWriter = 100
			aggmed, _ := iotagg.New(ctx, cfg)
			defer aggmed.Close()

			cfg.BufWriter = 1000
			agglrg, _ := iotagg.New(ctx, cfg)
			defer agglrg.Close()

			aggsml.Start(ctx)
			aggmed.Start(ctx)
			agglrg.Start(ctx)

			experiment.SampleDuration("Small buffer (10), small data", func(idx int) {
				for i := 0; i < 100; i++ {
					aggsml.Write(smallData)
				}
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Medium buffer (100), small data", func(idx int) {
				for i := 0; i < 100; i++ {
					aggmed.Write(smallData)
				}
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Large buffer (1000), small data", func(idx int) {
				for i := 0; i < 100; i++ {
					agglrg.Write(smallData)
				}
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Large buffer (1000), 1KB data", func(idx int) {
				for i := 0; i < 100; i++ {
					agglrg.Write(mediumData)
				}
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("Large buffer (1000), 10KB data", func(idx int) {
				for i := 0; i < 50; i++ {
					agglrg.Write(largeData)
				}
			}, SamplingConfig{N: 50, Duration: 0})
		})
	})

	Describe("TC-BC-003: Concurrent operations", func() {
		It("TC-BC-004: should benchmark concurrent writes with varying goroutine counts", func() {
			experiment := NewExperiment("Concurrent write operations")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}
			agg, _ := iotagg.New(ctx, cfg)
			defer agg.Close()

			agg.Start(ctx)

			data := []byte("test data")
			writesPerGoroutine := 50

			experiment.SampleDuration("1 goroutine", func(idx int) {
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < writesPerGoroutine; j++ {
						agg.Write(data)
					}
				}()
				wg.Wait()
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("5 goroutines", func(idx int) {
				var wg sync.WaitGroup
				for i := 0; i < 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := 0; j < writesPerGoroutine; j++ {
							agg.Write(data)
						}
					}()
				}
				wg.Wait()
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("10 goroutines", func(idx int) {
				var wg sync.WaitGroup
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := 0; j < writesPerGoroutine; j++ {
							agg.Write(data)
						}
					}()
				}
				wg.Wait()
			}, SamplingConfig{N: 100, Duration: 0})

			experiment.SampleDuration("20 goroutines", func(idx int) {
				var wg sync.WaitGroup
				for i := 0; i < 20; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for j := 0; j < writesPerGoroutine; j++ {
							agg.Write(data)
						}
					}()
				}
				wg.Wait()
			}, SamplingConfig{N: 50, Duration: 0})
		})
	})

	Describe("TC-BC-007: Monitoring metrics", func() {
		It("TC-BC-008: should benchmark metrics read operations", func() {
			experiment := NewExperiment("Monitoring metrics")
			AddReportEntry(experiment.Name, experiment)

			writer := newTestWriter()
			cfg := iotagg.Config{
				BufWriter: 1000,
				FctWriter: writer.Write,
			}

			agg, _ := iotagg.New(ctx, cfg)
			agg.Start(ctx)
			defer agg.Close()

			// Write some data to populate metrics
			for i := 0; i < 100; i++ {
				agg.Write([]byte("test data"))
			}

			experiment.SampleDuration("NbWaiting", func(idx int) {
				_ = agg.NbWaiting()
			}, SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("NbProcessing", func(idx int) {
				_ = agg.NbProcessing()
			}, SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("SizeWaiting", func(idx int) {
				_ = agg.SizeWaiting()
			}, SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("SizeProcessing", func(idx int) {
				_ = agg.SizeProcessing()
			}, SamplingConfig{N: 10000, Duration: 0})

			experiment.SampleDuration("All 4 metrics", func(idx int) {
				_ = agg.NbWaiting()
				_ = agg.NbProcessing()
				_ = agg.SizeWaiting()
				_ = agg.SizeProcessing()
			}, SamplingConfig{N: 10000, Duration: 0})
		})
	})
})
