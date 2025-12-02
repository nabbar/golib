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
// This file contains integration tests for the hookfile package, including:
//   - Log rotation detection and handling
//   - Log level filtering
//   - File creation and directory handling
//   - Formatter integration
//
// These tests verify that the hook correctly interacts with the file system
// and handles external log rotation scenarios.
package hookfile_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration Tests", func() {
	var (
		tempIntDir  string
		testLogFile string
	)

	BeforeEach(func() {
		// Close all hooks before cleanup
		logfil.ResetOpenFiles()

		// Ensure tempDir exists (may have been deleted by another test)
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			tempDir, err = os.MkdirTemp("", "hookfile-test-*")
			Expect(err).NotTo(HaveOccurred())
		}

		var err error
		tempIntDir, err = os.MkdirTemp(tempDir, "integration-test-*")
		Expect(err).NotTo(HaveOccurred())

		testLogFile = filepath.Join(tempIntDir, "test.log")
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
		// Clean up test log file after each test
		if tempIntDir != "" {
			_ = os.RemoveAll(tempIntDir)
		}
	})

	It("should handle log rotation", func() {
		// This test verifies that the hook correctly detects and handles external log rotation.
		// External rotation occurs when tools like logrotate rename/move the log file.
		// The hook should detect this and create a new file at the original path.

		// Create initial log file with CreatePath enabled for rotation support
		opts := logcfg.OptionsFile{
			Filepath:   testLogFile,
			CreatePath: true, // Required for rotation detection and file recreation
		}

		hook, err := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			if hook != nil {
				_ = hook.Close()
			}
		}()

		// Write first log entry to establish the initial file
		logger := logrus.New()
		firstEntry := logrus.NewEntry(logger)
		firstEntry.Level = logrus.InfoLevel
		firstEntry.Message = "ignored value" // Message field is ignored, data is used
		firstEntry.Data = logrus.Fields{
			"test": "first",
			"msg":  "First log entry",
		}

		err = hook.Fire(firstEntry)
		Expect(err).NotTo(HaveOccurred())

		// Wait for sync cycle to complete (sync timer is 1 second)
		// This ensures the first entry is flushed to disk
		time.Sleep(2 * time.Second) // Wait for sync timer (1s) + margin

		// Simulate external log rotation by renaming the file
		// This is what logrotate or similar tools do
		rotatedFile := testLogFile + ".1"
		Expect(os.Rename(testLogFile, rotatedFile)).To(Succeed())

		// Wait for the next sync cycle to detect the rotation
		// The SyncFct will notice the file at testLogFile is different/missing
		time.Sleep(2 * time.Second) // Wait for next sync cycle

		// Write second log entry - should go to newly created file after rotation detection
		secondEntry := logrus.NewEntry(logger)
		secondEntry.Level = logrus.InfoLevel
		secondEntry.Message = "ignored value"
		secondEntry.Data = logrus.Fields{
			"test": "second",
			"msg":  "Second log entry",
		}

		err = hook.Fire(secondEntry)
		Expect(err).NotTo(HaveOccurred())

		// Wait for write to complete
		time.Sleep(500 * time.Millisecond)

		// Third log entry to ensure new file is working
		thirdEntry := logrus.NewEntry(logger)
		thirdEntry.Level = logrus.InfoLevel
		thirdEntry.Message = "ignored value"
		thirdEntry.Data = logrus.Fields{
			"test": "third",
			"msg":  "Third log entry",
		}

		err = hook.Fire(thirdEntry)
		Expect(err).NotTo(HaveOccurred())

		// Ensure all writes are complete
		time.Sleep(500 * time.Millisecond)

		// Close the hook to flush any remaining logs
		Expect(hook.Close()).To(Succeed())
		hook = nil

		// Verify rotated file contains first entry
		content, err := os.ReadFile(rotatedFile)
		Expect(err).NotTo(HaveOccurred())
		contentStr := string(content)
		Expect(contentStr).To(ContainSubstring("test=first"), "Rotated file should contain first log entry")
		Expect(contentStr).To(ContainSubstring("fields.msg=\"First log entry\""), "Rotated file should contain first log message")

		// Verify new file contains second and third entries
		content, err = os.ReadFile(testLogFile)
		Expect(err).NotTo(HaveOccurred())
		contentStr = string(content)
		Expect(contentStr).To(ContainSubstring("test=second"), "New file should contain second log entry")
		Expect(contentStr).To(ContainSubstring("fields.msg=\"Second log entry\""), "New file should contain second log message")
		Expect(contentStr).To(ContainSubstring("test=third"), "New file should contain third log entry")
		Expect(contentStr).To(ContainSubstring("fields.msg=\"Third log entry\""), "New file should contain third log message")
	})

	It("should handle multiple hooks", func() {
		// Create multiple hooks writing to different files
		const numHooks = 3
		files := make([]string, numHooks)
		hooks := make([]logfil.HookFile, numHooks)

		// Setup all hooks
		for i := 0; i < numHooks; i++ {
			files[i] = filepath.Join(tempIntDir, "nested", "dir", fmt.Sprintf("hook_%d.log", i))

			opts := logcfg.OptionsFile{
				Filepath:   files[i],
				CreatePath: true,
			}

			var err error
			hooks[i], err = logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
			Expect(err).NotTo(HaveOccurred())
		}

		// Create test entry
		testEntry := logrus.NewEntry(logrus.New())
		testEntry.Level = logrus.InfoLevel
		testEntry.Message = "ignored value"
		testEntry.Data = logrus.Fields{
			"test":  true,
			"multi": "hook",
			"msg":   "Test log message",
		}

		// Send the entry to all hooks
		for _, h := range hooks {
			err := h.Fire(testEntry)
			Expect(err).NotTo(HaveOccurred())
		}

		// Ensure all writes are complete
		time.Sleep(250 * time.Millisecond)

		// Close all hooks and verify files
		for i, h := range hooks {
			// Close the hook to flush any remaining logs
			err := h.Close()
			Expect(err).NotTo(HaveOccurred())

			// Verify the log file was created and contains the message
			content, err := os.ReadFile(files[i])
			Expect(err).NotTo(HaveOccurred(), "Failed to read log file for hook %d: %s", i, files[i])
			contentStr := string(content)
			Expect(contentStr).To(ContainSubstring("test=true"),
				"File %s should contain test field", files[i])
			Expect(contentStr).To(ContainSubstring("multi=hook"),
				"File %s should contain multi field", files[i])
			Expect(contentStr).To(ContainSubstring("msg=\"Test log message\""),
				"File %s should contain message field", files[i])
		}
	})

	It("should handle log levels correctly", func() {
		// This test verifies that the hook respects the configured log levels
		// and only writes entries matching those levels to the file.

		opts := logcfg.OptionsFile{
			Filepath:   testLogFile,
			CreatePath: true,
			LogLevel:   []string{"error", "warning"}, // Only error and warning levels
		}

		hook, err := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
		Expect(err).NotTo(HaveOccurred())

		// Capture output for verification
		var buf bytes.Buffer
		logger := logrus.New()
		logger.SetOutput(&buf)
		logger.AddHook(hook)

		// Log at different levels with proper data fields
		// IMPORTANT: The Message field is ignored by the formatter.
		// All log data must be in the Data field, including the "msg" key.

		debugEntry := logrus.NewEntry(logger)
		debugEntry.Level = logrus.DebugLevel
		debugEntry.Message = "ignored value"
		debugEntry.Data = logrus.Fields{"msg": "Debug message"}
		hook.Fire(debugEntry) // Should be filtered out (not in configured levels)

		infoEntry := logrus.NewEntry(logger)
		infoEntry.Level = logrus.InfoLevel
		infoEntry.Message = "ignored value"
		infoEntry.Data = logrus.Fields{"msg": "Info message"}
		hook.Fire(infoEntry) // Should be filtered out (not in configured levels)

		warnEntry := logrus.NewEntry(logger)
		warnEntry.Level = logrus.WarnLevel
		warnEntry.Message = "ignored value"
		warnEntry.Data = logrus.Fields{"msg": "Warning message"}
		hook.Fire(warnEntry) // Should be logged (warning is configured)

		errorEntry := logrus.NewEntry(logger)
		errorEntry.Level = logrus.ErrorLevel
		errorEntry.Message = "ignored value"
		errorEntry.Data = logrus.Fields{"msg": "Error message"}
		hook.Fire(errorEntry) // Should be logged

		// Give the hook time to process
		time.Sleep(250 * time.Millisecond)

		// Close the hook to ensure all logs are written
		err = hook.Close()
		Expect(err).NotTo(HaveOccurred())

		// Verify only error and warning messages are logged
		content, err := os.ReadFile(testLogFile)
		Expect(err).NotTo(HaveOccurred())
		contentStr := string(content)

		Expect(contentStr).NotTo(ContainSubstring("fields.msg=\"Debug message\""), "Debug message should not be logged")
		Expect(contentStr).NotTo(ContainSubstring("fields.msg=\"Info message\""), "Info message should not be logged")
		Expect(contentStr).To(ContainSubstring("level=warning"), "Warning level should be in log")
		Expect(contentStr).To(ContainSubstring("fields.msg=\"Warning message\""), "Warning message should be logged")
		Expect(contentStr).To(ContainSubstring("level=error"), "Error level should be in log")
		Expect(contentStr).To(ContainSubstring("fields.msg=\"Error message\""), "Error message should be logged")
	})

	It("should handle concurrent log writes to the same file", func() {
		const (
			numGoroutines = 10
			numLogs       = 100
		)

		opts := logcfg.OptionsFile{
			Filepath:   testLogFile,
			CreatePath: true,
		}

		hook, err := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
		Expect(err).NotTo(HaveOccurred())

		logger := logrus.New()
		logger.SetOutput(io.Discard)
		logger.AddHook(hook)

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numLogs; j++ {
					entry := logrus.NewEntry(logger)
					entry.Level = logrus.InfoLevel
					entry.Message = "ignored value"
					entry.Data = logrus.Fields{
						"goroutine": id,
						"iteration": j,
						"msg":       fmt.Sprintf("Log entry %d", j),
					}
					hook.Fire(entry)
				}
			}(i)
		}

		wg.Wait()

		// Give the hook time to process all logs
		time.Sleep(500 * time.Millisecond)

		// Close the hook to ensure all logs are written
		err = hook.Close()
		Expect(err).NotTo(HaveOccurred())

		// Verify all logs were written
		content, err := os.ReadFile(testLogFile)
		Expect(err).NotTo(HaveOccurred())

		// Verify we have the expected number of log entries
		contentStr := string(content)
		totalExpected := numGoroutines * numLogs
		actualCount := strings.Count(contentStr, "msg=\"Log entry")
		Expect(actualCount).To(Equal(totalExpected),
			"Expected %d log entries with 'msg=\"Log entry', got %d. Content: %s",
			totalExpected, actualCount, contentStr)

		// Verify goroutine and iteration fields are present
		for i := 0; i < numGoroutines; i++ {
			Expect(contentStr).To(ContainSubstring(fmt.Sprintf("goroutine=%d", i)),
				"Logs should contain goroutine ID %d", i)
		}

		// Verify all iterations are present for each goroutine
		for i := 0; i < numGoroutines; i++ {
			for j := 0; j < numLogs; j++ {
				expected := fmt.Sprintf("goroutine=%d iteration=%d", i, j)
				Expect(contentStr).To(ContainSubstring(expected),
					"Logs should contain %s", expected)
			}
		}
	})
})
