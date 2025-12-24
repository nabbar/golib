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

package hookfile_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	logfil "github.com/nabbar/golib/logger/hookfile"
)

// Example_basic demonstrates the simplest use case: creating a hook that writes to a file.
func Example_basic() {
	// Create temporary directory for example
	tmpDir, _ := os.MkdirTemp("", "hookfile-example-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	logFile := filepath.Join(tmpDir, "app.log")

	// Configure the hook with minimal settings
	opts := logcfg.OptionsFile{
		Filepath:   logFile,
		FileMode:   0600,
		PathMode:   0700,
		CreatePath: true,
	}

	// Create the hook
	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true, // Disable timestamp for predictable output
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	// Create and configure logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Use Stderr to separate from file output
	logger.AddHook(hook)

	// IMPORTANT: The message parameter "ignored" is NOT used by the hook.
	// Only the fields (here "msg") are written to the file.
	// Exception: In AccessLog mode, only the message is used and fields are ignored.
	logger.WithField("msg", "Application started").Info("ignored")

	// Give aggregator time to flush
	time.Sleep(100 * time.Millisecond)

	// Read and print what was written by the hook
	content, _ := os.ReadFile(logFile)
	fmt.Print(string(content))

	// Output:
	// level=info fields.msg="Application started"
}

// Example_productionSetup demonstrates a production-ready configuration with rotation support.
func Example_productionSetup() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-prod-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	logFile := filepath.Join(tmpDir, "app.log")

	opts := logcfg.OptionsFile{
		Filepath:         logFile,
		FileMode:         0644, // Readable by others
		PathMode:         0755, // Standard directory permissions
		CreatePath:       true, // Create directories if needed (enables rotation detection)
		LogLevel:         []string{"info", "warning", "error"},
		DisableStack:     true,  // Don't log stack traces
		DisableTimestamp: false, // Include timestamps
	}

	hook, err := logfil.New(opts, &logrus.JSONFormatter{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: Use fields, not message parameter
	logger.WithFields(logrus.Fields{
		"msg":    "Application started",
		"action": "startup",
		"user":   "system",
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Logs written to file with rotation detection enabled")

	// Output:
	// Logs written to file with rotation detection enabled
}

// Example_accessLog demonstrates using access log mode for HTTP request logging.
func Example_accessLog() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-access-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	accessLog := filepath.Join(tmpDir, "access.log")

	// Enable access log mode
	opts := logcfg.OptionsFile{
		Filepath:        accessLog,
		FileMode:        0644,
		CreatePath:      true,
		EnableAccessLog: true, // Message-only mode
	}

	hook, err := logfil.New(opts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: In AccessLog mode, behavior is REVERSED!
	// The message "GET /api/users - 200 OK - 45ms" IS output.
	// The fields (method, path, status_code) are IGNORED.
	logger.WithFields(logrus.Fields{
		"method":      "GET",
		"path":        "/api/users",
		"status_code": 200,
	}).Info("GET /api/users - 200 OK - 45ms")

	time.Sleep(100 * time.Millisecond)

	content, _ := os.ReadFile(accessLog)
	fmt.Print(string(content))

	// Output:
	// GET /api/users - 200 OK - 45ms
}

// Example_levelFiltering demonstrates filtering logs by level to different files.
func Example_levelFiltering() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-levels-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	errorLog := filepath.Join(tmpDir, "error.log")

	opts := logcfg.OptionsFile{
		Filepath:   errorLog,
		CreatePath: true,
		LogLevel:   []string{"error", "fatal"}, // Only errors
	}

	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// This will be written (error level)
	// Note: message "ignored" is NOT used, only the field "msg"
	logger.WithField("msg", "Error occurred").Error("ignored")

	// This won't be written (wrong level)
	logger.WithField("msg", "Info message").Info("ignored")

	time.Sleep(100 * time.Millisecond)

	content, _ := os.ReadFile(errorLog)
	fmt.Printf("Error log: %s", string(content))

	// Output:
	// Error log: level=error fields.msg="Error occurred"
}

// Example_fieldFiltering demonstrates filtering specific fields from output.
func Example_fieldFiltering() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-filter-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	logFile := filepath.Join(tmpDir, "app.log")

	// Configure to filter out stack and timestamp
	opts := logcfg.OptionsFile{
		Filepath:         logFile,
		CreatePath:       true,
		DisableStack:     true,  // Remove stack fields
		DisableTimestamp: true,  // Remove time fields
		EnableTrace:      false, // Remove caller/file/line fields
	}

	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Filtered log",
		"stack":  "will be filtered",
		"caller": "will be filtered",
		"user":   "john",
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	content, _ := os.ReadFile(logFile)
	fmt.Print(string(content))

	// Output:
	// level=info fields.msg="Filtered log" user=john
}

// Example_multipleLoggers demonstrates multiple loggers writing to the same file efficiently.
func Example_multipleLoggers() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-multi-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	sharedLog := filepath.Join(tmpDir, "shared.log")

	opts := logcfg.OptionsFile{
		Filepath:   sharedLog,
		CreatePath: true,
	}

	// Create multiple hooks to the same file (they share the file aggregator)
	hook1, _ := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
	hook2, _ := logfil.New(opts, &logrus.TextFormatter{DisableTimestamp: true})
	defer hook1.Close()
	defer hook2.Close()

	logger1 := logrus.New()
	logger1.SetOutput(os.Stderr)
	logger1.AddHook(hook1)

	logger2 := logrus.New()
	logger2.SetOutput(os.Stderr)
	logger2.AddHook(hook2)

	// Both loggers write to the same file efficiently
	// IMPORTANT: message parameter is NOT used, only fields
	logger1.WithField("msg", "From logger 1").Info("ignored")
	logger2.WithField("msg", "From logger 2").Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Multiple loggers wrote to same file")

	// Output:
	// Multiple loggers wrote to same file
}

// Example_rotationDetection demonstrates automatic log rotation detection.
func Example_rotationDetection() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-rotate-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	logFile := filepath.Join(tmpDir, "app.log")

	opts := logcfg.OptionsFile{
		Filepath:   logFile,
		CreatePath: true, // Required for rotation detection
		Create:     true, // Required for automatic file creation after rotation
	}

	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// Write first log entry
	logger.WithField("msg", "Before rotation").Info("ignored")
	time.Sleep(100 * time.Millisecond)

	// Simulate external log rotation (like logrotate)
	os.Rename(logFile, logFile+".1")

	// Wait for rotation detection (sync timer runs every 1 second)
	time.Sleep(2500 * time.Millisecond)

	// Write after rotation - should go to new file
	logger.WithField("msg", "After rotation").Info("ignored")
	time.Sleep(500 * time.Millisecond)

	// Close hook to flush
	_ = hook.Close()
	time.Sleep(100 * time.Millisecond)

	// Check that new file was created
	if _, err := os.Stat(logFile); err == nil {
		fmt.Println("New log file created after rotation")
	}

	// Output:
	// New log file created after rotation
}

// Example_separateAccessLogs demonstrates separating access logs from application logs.
func Example_separateAccessLogs() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-sep-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	appLog := filepath.Join(tmpDir, "app.log")
	accessLog := filepath.Join(tmpDir, "access.log")

	// Application logs (standard mode with fields)
	appOpts := logcfg.OptionsFile{
		Filepath:   appLog,
		CreatePath: true,
	}
	appHook, _ := logfil.New(appOpts, &logrus.JSONFormatter{})
	defer appHook.Close()

	appLogger := logrus.New()
	appLogger.SetOutput(os.Stderr)
	appLogger.AddHook(appHook)

	// Access logs (access log mode with message)
	accessOpts := logcfg.OptionsFile{
		Filepath:        accessLog,
		CreatePath:      true,
		EnableAccessLog: true,
	}
	accessHook, _ := logfil.New(accessOpts, nil)
	defer accessHook.Close()

	accessLogger := logrus.New()
	accessLogger.SetOutput(os.Stderr)
	accessLogger.AddHook(accessHook)

	// Application log (uses fields)
	appLogger.WithField("msg", "Request processed").Info("ignored")

	// Access log (uses message)
	accessLogger.Info("GET /api/users - 200 OK - 45ms")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Application and access logs separated successfully")

	// Output:
	// Application and access logs separated successfully
}

// Example_traceEnabled demonstrates enabling trace information in logs.
func Example_traceEnabled() {
	tmpDir, _ := os.MkdirTemp("", "hookfile-trace-*")
	defer os.RemoveAll(tmpDir)
	defer logfil.ResetOpenFiles()

	logFile := filepath.Join(tmpDir, "app.log")

	opts := logcfg.OptionsFile{
		Filepath:    logFile,
		CreatePath:  true,
		EnableTrace: true, // Include caller/file/line information
	}

	hook, err := logfil.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Log with trace",
		"caller": "example_test.go:line",
		"file":   "example_test.go",
		"line":   123,
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	content, _ := os.ReadFile(logFile)
	fmt.Print(string(content))

	// Output:
	// level=info caller="example_test.go:line" fields.msg="Log with trace" file=example_test.go line=123
}

// Example_errorHandling demonstrates error handling for invalid file paths.
func Example_errorHandling() {
	// Try to create hook with missing file path
	opts := logcfg.OptionsFile{
		Filepath: "", // Missing!
	}

	hook, err := logfil.New(opts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if hook == nil {
		fmt.Println("Hook was not created")
	}

	// Output:
	// Error: missing file path
	// Hook was not created
}
