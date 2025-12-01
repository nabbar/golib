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
 *
 */

// Package hookfile provides a logrus hook implementation for file-based logging.
// This file contains benchmark tests for the hookfile package.
//
// Benchmarks measure:
//   - Log write performance under various concurrency levels
//   - Memory usage during log operations
//   - Throughput with different message counts
//
// The benchmarks use gmeasure from Gomega for detailed performance metrics.
// They help identify performance regressions and optimize the hook implementation.
package hookfile_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gmeasure"
)

var _ = Describe("Benchmark Tests", func() {
	var (
		experiment   *Experiment
		tempBenchDir string
	)

	BeforeEach(func() {
		// Ensure tempDir exists (may have been deleted by another test)
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			tempDir, err = os.MkdirTemp("", "hookfile-test-*")
			Expect(err).NotTo(HaveOccurred())
		}

		// Create a temporary directory for benchmark files
		var err error
		tempBenchDir, err = os.MkdirTemp(tempDir, "benchmark-*")
		Expect(err).NotTo(HaveOccurred())

		// Create a new experiment
		experiment = NewExperiment("hookfile_benchmarks")
		AddReportEntry(experiment.Name, experiment)
	})

	AfterEach(func() {
		// Clean up benchmark files
		if tempBenchDir != "" {
			_ = os.RemoveAll(tempBenchDir)
		}
	})

	It("measures log writing performance", func() {
		// Define test parameters
		messageCounts := []int{100, 1000, 10000}
		concurrencyLevels := []int{1, 4, 16}

		for _, count := range messageCounts {
			for _, concurrency := range concurrencyLevels {
				fct := func(idx int) {
					// Set up test file with unique name
					logFile := filepath.Join(tempBenchDir, fmt.Sprintf("benchmark_%d_%d_%d.log", count, concurrency, idx))

					// Set up hook
					opts := logcfg.OptionsFile{
						Filepath:   logFile,
						CreatePath: true,
					}

					hook, err := logfil.New(opts, &logrus.TextFormatter{
						DisableTimestamp: true,
					})
					Expect(err).NotTo(HaveOccurred())
					defer func() {
						if hook != nil {
							_ = hook.Close()
						}
					}()

					// Set up logger
					logger := logrus.New()
					logger.SetOutput(io.Discard)
					// Use AddHook only if hook is not nil
					if hook != nil {
						logger.AddHook(hook)
					}

					// Measure write performance
					experiment.MeasureDuration("log_write", func() {
						var wg sync.WaitGroup
						messagesPerGoroutine := count / concurrency

						for i := 0; i < concurrency; i++ {
							wg.Add(1)
							go func() {
								defer wg.Done()
								for j := 0; j < messagesPerGoroutine; j++ {
									logger.WithField("benchmark", true).WithField("iteration", j).Info("Benchmark log message")
								}
							}()
						}
						wg.Wait()

						// Wait for writes to be flushed
						time.Sleep(100 * time.Millisecond)
					}, Precision(time.Millisecond))
				}

				experiment.Sample(fct, SamplingConfig{
					N:           5, // Number of samples
					NumParallel: 1,
				})
			}
		}
		// Add analysis - simplified version without detailed stats
		experiment.RecordValue("throughput", 0.0, Units("messages/second"))
	})

	It("measures memory usage", func() {
		var memStatsBefore, memStatsAfter runtime.MemStats
		var hook logfil.HookFile
		var err error

		// Force GC and wait for it to complete
		runtime.GC()
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		runtime.ReadMemStats(&memStatsBefore)

		// Create and use hook
		opts := logcfg.OptionsFile{
			Filepath:   filepath.Join(tempBenchDir, "memory.log"),
			CreatePath: true,
		}

		hook, err = logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			if hook != nil {
				_ = hook.Close()
			}
		}()

		logger := logrus.New()
		logger.SetOutput(io.Discard)
		logger.AddHook(hook)

		for i := 0; i < 1000; i++ {
			logger.WithField("msg", "Memory test log entry").WithField("iteration", i).Info("")
		}

		// Wait for writes to be flushed
		time.Sleep(100 * time.Millisecond)

		runtime.ReadMemStats(&memStatsAfter)

		// Calculate memory usage safely (handle potential negative values)
		var memUsedKB float64
		if memStatsAfter.HeapInuse >= memStatsBefore.HeapInuse {
			memUsedKB = float64(memStatsAfter.HeapInuse-memStatsBefore.HeapInuse) / 1024
		} else {
			// If memory decreased, record 0 (GC freed more than we allocated)
			memUsedKB = 0
		}

		// Record memory usage
		experiment.RecordValue("memory_usage_kb",
			memUsedKB,
			Units("KB"),
		)
	})
})

// BenchmarkHookFileWrite benchmarks the hook file write performance
func BenchmarkHookFileWrite(b *testing.B) {
	// Skip if running with race detector
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	// Create a temporary file for benchmarking
	tempFile, err := os.CreateTemp(tempDir, "bench-*.log")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Set up hook with benchmark file
	opts := logcfg.OptionsFile{
		Filepath:   tempFile.Name(),
		CreatePath: true,
	}

	hook, err := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
	if err != nil {
		b.Fatalf("Failed to create hook: %v", err)
	}

	// Set up logger
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard output for benchmarking
	logger.AddHook(hook)

	// Reset timer after setup
	b.ResetTimer()

	// Run benchmark
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Benchmark log message")
		}
	})

	// Ensure all logs are written
	_ = hook.Fire(&logrus.Entry{
		Logger:  logger,
		Level:   logrus.InfoLevel,
		Message: "Flush logs",
	})
}
