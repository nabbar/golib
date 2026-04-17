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
// This file contains standard Go benchmarks for the hookfile package.
package hookfile_test

import (
	"io"
	"path/filepath"
	"strconv"
	"testing"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"
)

// setupBenchmarkLogger creates a logger with a file hook for benchmarking.
func setupBenchmarkLogger(b *testing.B, opts logcfg.OptionsFile) (*logrus.Logger, logfil.HookFile) {
	b.Helper()

	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		b.Fatalf("Failed to create hook: %v", err)
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	hook.RegisterHook(logger)

	return logger, hook
}

func BenchmarkNewHook(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_write")
	defer logfil.ResetOpenFiles()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hook, err := logfil.New(getOptions(logFile+"-"+strconv.Itoa(i)+".log"), &logrus.TextFormatter{
			DisableTimestamp: true,
		})
		if err != nil {
			b.Fatal(err)
		}
		_ = hook.Close()
	}
}

// BenchmarkHookFileWrite benchmarks basic sequential logging performance.
// It measures the overhead of the hook and the file writing process.
func BenchmarkHookFileWrite(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_write.log")

	logger, hook := setupBenchmarkLogger(b, getOptions(logFile))
	defer func() {
		_ = hook.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// IMPORTANT: In normal mode, the Message parameter is ignored if Data is empty.
		// We use WithField to ensure the hook has something to process and write.
		logger.WithField("msg", "Benchmark log message").Info("")
	}
}

// BenchmarkHookFileWriteParallel benchmarks concurrent logging performance.
// This tests the thread-safety and contention of the aggregator.
func BenchmarkHookFileWriteParallel(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_parallel.log")

	logger, hook := setupBenchmarkLogger(b, getOptions(logFile))
	defer func() {
		_ = hook.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.WithField("msg", "Benchmark log message").Info("")
		}
	})
}

// BenchmarkHookFileAccessLog benchmarks performance in access log mode.
// This mode bypasses the logrus formatter and writes the message directly.
func BenchmarkHookFileAccessLog(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_access.log")

	opts := getOptions(logFile)
	opts.EnableAccessLog = true

	logger, hook := setupBenchmarkLogger(b, opts)
	defer func() {
		_ = hook.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// In access log mode, the Message parameter IS used directly.
		logger.Info("Access log message")
	}
}

// BenchmarkHookFileWithFields benchmarks logging with multiple fields.
// This measures the overhead of field processing and formatting.
func BenchmarkHookFileWithFields(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_fields.log")

	logger, hook := setupBenchmarkLogger(b, getOptions(logFile))
	defer func() {
		_ = hook.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.WithFields(logrus.Fields{
			"msg":      "Benchmark log message",
			"user":     "john_doe",
			"id":       12345,
			"remote":   "192.168.1.1",
			"status":   200,
			"duration": "15ms",
			"method":   "GET",
			"path":     "/api/v1/resource",
		}).Info("")
	}
}

// BenchmarkHookFileFiltered benchmarks the overhead when logs are filtered by level.
// This should be very fast as it returns early in the hook's Fire method.
func BenchmarkHookFileFiltered(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_filtered.log")

	opts := getOptions(logFile)
	opts.LogLevel = []string{"error"} // Only allow error logs

	logger, hook := setupBenchmarkLogger(b, opts)
	defer func() {
		_ = hook.Close()
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Info level is filtered out by the hook
		logger.WithField("msg", "Benchmark log message").Info("")
	}
}

// BenchmarkHookFileTrace benchmarks performance with trace info enabled.
func BenchmarkHookFileTrace(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "bench_trace.log")

	opts := getOptions(logFile)
	opts.EnableTrace = true

	logger, hook := setupBenchmarkLogger(b, opts)
	defer func() {
		_ = hook.Close()
	}()

	// Enable caller reporting in logrus to provide data for the trace fields
	logger.SetReportCaller(true)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.WithField("msg", "Benchmark log message").Info("")
	}
}

// BenchmarkLogrusDiscard benchmarks logrus overhead with io.Discard for comparison.
func BenchmarkLogrusDiscard(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.WithField("msg", "Benchmark log message").Info("")
	}
}
