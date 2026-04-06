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
// This file contains concurrency tests for the hookfile package.
//
// These tests verify that:
//   - Multiple goroutines can safely write to the same log file
//   - The hook handles concurrent writes without data corruption
//   - File handles are properly managed under concurrent access
//   - Hook creation and destruction work correctly in concurrent scenarios
//
// All tests use the race detector to ensure thread-safety.
package hookfile_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concurrency Tests", func() {
	var (
		hook logfil.HookFile
		log  *logrus.Logger
		err  error
	)

	BeforeEach(func() {
		// Close all hooks before cleanup
		logfil.ResetOpenFiles()

		log = logrus.New()
		log.SetOutput(GinkgoWriter)

		hook, err = createTestHook()
		Expect(err).NotTo(HaveOccurred())

		hook.RegisterHook(log)
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
		// Clean up test log file after each test
		if _, err := os.Stat(testLogFile); err == nil {
			_ = os.Remove(testLogFile)
		}
	})

	It("should handle concurrent log writes", func() {
		// This test verifies thread-safety by having multiple goroutines
		// write to the same log file simultaneously. It ensures:
		// 1. No data corruption occurs
		// 2. All log entries are written
		// 3. No race conditions exist (verified with -race flag)

		const (
			numGoroutines = 10  // Number of concurrent writers
			numLogs       = 100 // Logs per goroutine (total: 1000 entries)
		)

		var (
			wg      sync.WaitGroup
			errChan = make(chan error, numGoroutines*numLogs) // Buffered to avoid blocking
		)

		// Launch concurrent writers
		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				// Each goroutine writes numLogs entries
				for j := 0; j < numLogs; j++ {
					entry := logrus.NewEntry(log)
					entry.Level = logrus.InfoLevel
					entry.Message = "ignored value" // Not used by formatter
					entry.Data = logrus.Fields{
						"goroutine": id, // Track which goroutine wrote this
						"iteration": j,  // Track iteration number
						"test":      true,
						"msg":       "Log entry", // Actual message in Data field
					}

					// Fire the entry and collect any errors
					if err := hook.Fire(entry); err != nil {
						errChan <- fmt.Errorf("goroutine %d, iteration %d: %v", id, j, err)
					}
				}
			}(i)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Check for errors
		close(errChan)
		for err := range errChan {
			Expect(err).NotTo(HaveOccurred())
		}

		// Ensure all writes are flushed (longer with race detector)
		time.Sleep(2000 * time.Millisecond)

		// Close the hook to flush any remaining logs
		Expect(hook.Close()).To(Succeed())

		// Verify all logs were written
		content, err := os.ReadFile(testLogFile)
		Expect(err).NotTo(HaveOccurred())
		contentStr := string(content)

		// Vérifier que nous avons le bon nombre d'entrées de log
		expectedEntries := numGoroutines * numLogs
		// Le format attendu est : level=info goroutine=X iteration=Y test=true fields.msg="Log entry"
		// Nous allons compter le nombre de lignes qui contiennent "fields.msg=\"Log entry\""
		actualEntries := strings.Count(contentStr, "fields.msg=\"Log entry\"")
		Expect(actualEntries).To(Equal(expectedEntries),
			"Expected %d log entries, got %d",
			expectedEntries, actualEntries)
	})
})

func BenchmarkConcurrentLogWrites(b *testing.B) {
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
	hook, err := logfil.New(getOptions(tempFile.Name()), &logrus.TextFormatter{DisableTimestamp: true})
	if err != nil {
		b.Fatalf("Failed to create hook: %v", err)
	}

	// Set up logger
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard output for benchmarking
	hook.RegisterHook(logger)

	// Run benchmark with multiple goroutines
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
