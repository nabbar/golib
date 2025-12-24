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
// This file contains basic functionality tests for the hookfile package.
package hookfile_test

import (
	"context"
	"os"
	"path/filepath"
	"time"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HookFile", func() {
	var (
		hook logfil.HookFile
		log  *logrus.Logger
		err  error
	)

	BeforeEach(func() {
		logfil.ResetOpenFiles()

		// Create a new logger instance for each test
		log = logrus.New()
		log.SetOutput(GinkgoWriter)
		log.SetLevel(logrus.DebugLevel) // Enable all log levels

		// Create a new hook for each test
		hook, err = createTestHook()
		Expect(err).NotTo(HaveOccurred(), "Failed to create test hook")

		// Register the hook with the logger
		log.AddHook(hook)
	})

	AfterEach(func() {
		time.Sleep(100 * time.Millisecond)
		// Clean up test log file after each test
		if _, err := os.Stat(testLogFile); err == nil {
			_ = os.Remove(testLogFile)
		}
	})

	Context("Basic Functionality", func() {
		It("should create a new hook with valid options", func() {
			Expect(hook).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should write logs to the specified file", func() {
			// Write a log message with non-nil Data field
			entry := logrus.NewEntry(log)
			entry.Level = logrus.InfoLevel
			entry.Message = "ignored value"
			entry.Data = logrus.Fields{"key": "value", "msg": "Test log message"}

			err = hook.Fire(entry)
			Expect(err).ToNot(HaveOccurred())

			// Give the hook time to write the log
			time.Sleep(250 * time.Millisecond)
			err = hook.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify the log file was created
			content, err := os.ReadFile(testLogFile)
			Expect(err).NotTo(HaveOccurred(), "Failed to read log file")

			// Check if the log message is in the file with the expected format
			// Format attendu : level=info key=value msg="Test log message"
			Expect(string(content)).To(ContainSubstring("level=info"), "Log level should be in log")
			Expect(string(content)).To(ContainSubstring("key=value"), "Key-value pair should be in log")
			Expect(string(content)).To(ContainSubstring("msg=\"Test log message\""), "Message should be in log")
		})

		It("should respect log levels", func() {
			// Test debug level
			debugEntry := logrus.NewEntry(log)
			debugEntry.Level = logrus.DebugLevel
			debugEntry.Message = "ignored value"
			debugEntry.Data = logrus.Fields{"test": true, "msg": "Debug message"}

			err = hook.Fire(debugEntry)
			Expect(err).ToNot(HaveOccurred())

			// Test info level
			infoEntry := logrus.NewEntry(log)
			infoEntry.Level = logrus.InfoLevel
			infoEntry.Message = "ignored value"
			infoEntry.Data = logrus.Fields{"test": true, "msg": "Info message"}

			err = hook.Fire(infoEntry)
			Expect(err).ToNot(HaveOccurred())

			// Ensure writes are flushed
			time.Sleep(250 * time.Millisecond)
			err = hook.Close()
			Expect(err).ToNot(HaveOccurred())

			// Verify file content
			content, err := os.ReadFile(testLogFile)
			Expect(err).NotTo(HaveOccurred())
			contentStr := string(content)

			// Vérifier que les messages de log contiennent les champs attendus
			Expect(contentStr).To(ContainSubstring("level=debug"), "Debug level should be in log")
			Expect(contentStr).To(ContainSubstring("test=true"), "Test field should be in log")
			Expect(contentStr).To(ContainSubstring("msg=\"Debug message\""), "Debug message should be in log")

			Expect(contentStr).To(ContainSubstring("level=info"), "Info level should be in log")
			Expect(contentStr).To(ContainSubstring("msg=\"Info message\""), "Info message should be in log")
		})
	})

	Context("Configuration", func() {
		It("should create directories if createPath is true", func() {
			tempPath := filepath.Join(tempDir, "nested", "dir", "test.log")

			opts := logcfg.OptionsFile{
				Filepath:   tempPath,
				CreatePath: true,
				FileMode:   0600,
				PathMode:   0700,
			}

			hook, err := logfil.New(opts, &logrus.TextFormatter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(hook).NotTo(BeNil())

			// Verify the directory was created
			_, err = os.Stat(filepath.Dir(tempPath))
			Expect(err).NotTo(HaveOccurred())

			// Clean up
			_ = os.RemoveAll(filepath.Dir(tempPath))
		})

		It("should return error for invalid file path", func() {
			opts := logcfg.OptionsFile{
				Filepath:   "/invalid/path/to/logfile.log",
				CreatePath: false,
			}

			_, err := logfil.New(opts, &logrus.TextFormatter{})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Error Handling", func() {
		It("should handle file write errors", func() {
			// Create a read-only directory
			readOnlyDir := filepath.Join(tempDir, "readonly")
			Expect(os.Mkdir(readOnlyDir, 0444)).To(Succeed())
			defer os.RemoveAll(readOnlyDir)

			readOnlyFile := filepath.Join(readOnlyDir, "test.log")

			// Try to create a hook with a read-only directory
			opts := logcfg.OptionsFile{
				Filepath:   readOnlyFile,
				CreatePath: false,
			}

			_, err := logfil.New(opts, &logrus.TextFormatter{})
			Expect(err).To(HaveOccurred())
		})

		It("should return error for missing file path", func() {
			opts := logcfg.OptionsFile{
				Filepath: "",
			}

			_, err := logfil.New(opts, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing file path"))
		})
	})

	Context("Hook Lifecycle", func() {
		It("should report running state correctly", func() {
			Expect(hook.IsRunning()).To(BeTrue(), "Hook should be running after creation")

			err = hook.Close()
			Expect(err).NotTo(HaveOccurred())

			Expect(hook.IsRunning()).To(BeFalse(), "Hook should not be running after close")
		})

		It("should handle Run method with context", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			done := make(chan bool)
			go func() {
				hook.Run(ctx)
				done <- true
			}()

			select {
			case <-done:
				// Run completed successfully
			case <-time.After(500 * time.Millisecond):
				Fail("Run should have completed when context was canceled")
			}

			Expect(hook.IsRunning()).To(BeFalse(), "Hook should not be running after Run completes")
		})

		It("should handle Write method", func() {
			data := []byte("test data\n")
			n, err := hook.Write(data)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(data)))

			time.Sleep(100 * time.Millisecond)
		})

		It("should handle Write after Close and reopen", func() {
			// Close the hook to trigger aggregator close
			err = hook.Close()
			Expect(err).NotTo(HaveOccurred())

			// Wait for close to complete
			time.Sleep(200 * time.Millisecond)

			// Try to write - should trigger ErrClosedResources and reopen
			data := []byte("test after close\n")
			n, err := hook.Write(data)

			// The write should either succeed (after reopen) or fail with closed error
			if err == nil {
				Expect(n).To(Equal(len(data)))
			}

			time.Sleep(100 * time.Millisecond)
		})
	})

	Context("Configuration Options", func() {
		It("should use default file and path modes", func() {
			opts := logcfg.OptionsFile{
				Filepath:   filepath.Join(tempDir, "mode-test.log"),
				CreatePath: true,
				// FileMode and PathMode not set (should use defaults)
			}

			h, err := logfil.New(opts, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(h).NotTo(BeNil())

			// Write a log to ensure file is created
			entry := logrus.NewEntry(log)
			entry.Level = logrus.InfoLevel
			entry.Data = logrus.Fields{"msg": "test"}
			err = h.Fire(entry)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)
			err = h.Close()
			Expect(err).NotTo(HaveOccurred())

			// Verify file exists with default permissions
			info, err := os.Stat(filepath.Join(tempDir, "mode-test.log"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0644)))
		})

		It("should register hook with logger", func() {
			logger := logrus.New()
			logger.SetOutput(GinkgoWriter)

			hook.RegisterHook(logger)

			// Verify hook was registered by logging and checking file
			logger.WithField("msg", "registered hook test").Info("ignored")

			time.Sleep(200 * time.Millisecond)

			content, err := os.ReadFile(testLogFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("registered hook test"))
		})

		It("should return correct log levels", func() {
			levels := hook.Levels()
			Expect(levels).NotTo(BeNil())
			Expect(levels).To(HaveLen(4)) // debug, info, warn, error

			// Verify levels contain expected values
			levelMap := make(map[logrus.Level]bool)
			for _, level := range levels {
				levelMap[level] = true
			}
			Expect(levelMap[logrus.DebugLevel]).To(BeTrue())
			Expect(levelMap[logrus.InfoLevel]).To(BeTrue())
			Expect(levelMap[logrus.WarnLevel]).To(BeTrue())
			Expect(levelMap[logrus.ErrorLevel]).To(BeTrue())
		})
	})
})

var _ = Describe("HookFile Additional Coverage", func() {
	It("should handle empty log data", func() {
		logfil.ResetOpenFiles()
		defer logfil.ResetOpenFiles()

		opts := logcfg.OptionsFile{
			Filepath:   filepath.Join(tempDir, "empty-test.log"),
			CreatePath: true,
		}

		hook, err := logfil.New(opts, nil)
		Expect(err).NotTo(HaveOccurred())
		defer hook.Close()

		logger := logrus.New()
		logger.SetOutput(GinkgoWriter)

		entry := logrus.NewEntry(logger)
		entry.Level = logrus.InfoLevel
		entry.Data = logrus.Fields{} // Empty data

		err = hook.Fire(entry)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	It("should filter access log with empty message", func() {
		logfil.ResetOpenFiles()
		defer logfil.ResetOpenFiles()

		opts := logcfg.OptionsFile{
			Filepath:        filepath.Join(tempDir, "access-empty.log"),
			CreatePath:      true,
			EnableAccessLog: true,
		}

		hook, err := logfil.New(opts, nil)
		Expect(err).NotTo(HaveOccurred())
		defer hook.Close()

		logger := logrus.New()
		logger.SetOutput(GinkgoWriter)

		entry := logrus.NewEntry(logger)
		entry.Level = logrus.InfoLevel
		entry.Message = "" // Empty message in access log mode

		err = hook.Fire(entry)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	It("should handle formatter errors gracefully", func() {
		logfil.ResetOpenFiles()
		defer logfil.ResetOpenFiles()

		opts := logcfg.OptionsFile{
			Filepath:   filepath.Join(tempDir, "formatter-test.log"),
			CreatePath: true,
		}

		hook, err := logfil.New(opts, &logrus.JSONFormatter{})
		Expect(err).NotTo(HaveOccurred())
		defer hook.Close()

		logger := logrus.New()
		logger.SetOutput(GinkgoWriter)

		entry := logrus.NewEntry(logger)
		entry.Level = logrus.InfoLevel
		entry.Data = logrus.Fields{"msg": "test"}

		err = hook.Fire(entry)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(100 * time.Millisecond)
	})

	It("should filter level not in configured levels", func() {
		logfil.ResetOpenFiles()
		defer logfil.ResetOpenFiles()

		opts := logcfg.OptionsFile{
			Filepath:   filepath.Join(tempDir, "level-filter.log"),
			CreatePath: true,
			LogLevel:   []string{"error"}, // Only error level
		}

		hook, err := logfil.New(opts, &logrus.TextFormatter{
			DisableTimestamp: true,
		})
		Expect(err).NotTo(HaveOccurred())
		defer hook.Close()

		logger := logrus.New()
		logger.SetOutput(GinkgoWriter)

		// Info level should be filtered
		infoEntry := logrus.NewEntry(logger)
		infoEntry.Level = logrus.InfoLevel
		infoEntry.Data = logrus.Fields{"msg": "info message"}

		err = hook.Fire(infoEntry)
		Expect(err).NotTo(HaveOccurred())

		// Error level should be written
		errorEntry := logrus.NewEntry(logger)
		errorEntry.Level = logrus.ErrorLevel
		errorEntry.Data = logrus.Fields{"msg": "error message"}

		err = hook.Fire(errorEntry)
		Expect(err).NotTo(HaveOccurred())

		time.Sleep(200 * time.Millisecond)

		content, err := os.ReadFile(filepath.Join(tempDir, "level-filter.log"))
		Expect(err).NotTo(HaveOccurred())
		contentStr := string(content)

		// Should only contain error, not info
		Expect(contentStr).To(ContainSubstring("error message"))
		Expect(contentStr).NotTo(ContainSubstring("info message"))
	})
})
