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

package hookstdout_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghko "github.com/nabbar/golib/logger/hookstdout"
)

// Example_basic demonstrates the simplest use case: creating a hook that writes to stdout.
func Example_basic() {
	// Note: Using a buffer for predictable test output
	var buf bytes.Buffer

	// Configure the hook with minimal settings
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true, // Disable color for predictable output
	}

	// Create the hook writing to buffer (simulating stdout)
	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true, // Disable timestamp for predictable output
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Create and configure logger (output to Stderr to avoid double write)
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: The message parameter "ignored" in Info() is NOT used by the hook.
	// Only the fields (here "msg") are written to the output.
	// Exception: In AccessLog mode, only the message is used and fields are ignored.
	logger.WithField("msg", "Application started").Info("ignored")

	// Print what was written by the hook
	fmt.Print(buf.String())

	// Output:
	// level=info fields.msg="Application started"
}

// Example_coloredOutput demonstrates using colored output for console applications.
func Example_coloredOutput() {
	var buf bytes.Buffer

	// Enable color output
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    false, // Enable color output
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: The message "ignored" is NOT output - only fields are used
	logger.WithField("msg", "Colored log output").Info("ignored")

	fmt.Println("Log written with colors (ANSI codes present)")
	// Output:
	// Log written with colors (ANSI codes present)
}

// Example_jsonFormatter demonstrates writing structured JSON logs to stdout.
func Example_jsonFormatter() {
	var buf bytes.Buffer

	// Configure options
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     true,
		DisableTimestamp: true,
	}

	// Create hook with JSON formatter
	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.JSONFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: The message "ignored" is NOT output - only fields are used
	logger.WithFields(logrus.Fields{
		"user_id": 123,
		"action":  "login",
		"msg":     "User logged in",
	}).Info("ignored")

	fmt.Println("Log written to stdout as JSON")
	// Output:
	// Log written to stdout as JSON
}

// Example_accessLog demonstrates using access log mode for HTTP request logging.
func Example_accessLog() {
	var buf bytes.Buffer

	// Enable access log mode
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		EnableAccessLog: true, // Message-only mode
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
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

	fmt.Print(buf.String())
	// Output:
	// GET /api/users - 200 OK - 45ms
}

// Example_levelFiltering demonstrates filtering logs by level.
func Example_levelFiltering() {
	var buf = bytes.NewBuffer(make([]byte, 0))

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}

	// Only handle info and debug levels (not errors)
	levels := []logrus.Level{
		logrus.InfoLevel,
		logrus.DebugLevel,
	}

	hook, err := loghko.NewWithWriter(buf, opt, levels, &logrus.TextFormatter{
		DisableTimestamp: true,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// This will be written by the hook (info level)
	// Note: message "ignored" is NOT used, only the field "msg"
	logger.WithField("msg", "Info message").Info("ignored")

	// This won't be written by the hook (wrong level)
	logger.WithField("msg", "Error message").Error("ignored")

	fmt.Printf("Hook captured: %s", buf.String())
	// Output:
	// Hook captured: level=info fields.msg="Info message"
}

// Example_fieldFiltering demonstrates filtering specific fields from output.
func Example_fieldFiltering() {
	var buf bytes.Buffer

	// Configure to filter out stack and timestamp
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     true,  // Remove stack fields
		DisableTimestamp: true,  // Remove time fields
		EnableTrace:      false, // Remove caller/file/line fields
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// Log with fields that will be filtered
	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Filtered log",
		"stack":  "trace info",
		"caller": "main.go:123",
		"user":   "john",
	}).Info("ignored")

	// Only "user" field remains after filtering
	fmt.Print(buf.String())
	// Output:
	// level=info fields.msg="Filtered log" user=john
}

// Example_disabledHook demonstrates how to conditionally disable the hook.
func Example_disabledHook() {
	opt := &logcfg.OptionsStd{
		DisableStandard: true, // This disables the hook
	}

	hook, err := loghko.New(opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if hook == nil {
		fmt.Println("Hook is disabled")
	} else {
		fmt.Println("Hook is enabled")
	}

	// Output:
	// Hook is disabled
}

// Example_traceEnabled demonstrates enabling trace information in logs.
func Example_traceEnabled() {
	var buf bytes.Buffer

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
		EnableTrace:     true, // Include caller/file/line information
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Log with trace info",
		"caller": "example_test.go:line",
		"file":   "example_test.go",
		"line":   123,
		"user":   "john",
	}).Info("ignored")

	// Trace fields are included because EnableTrace is true
	fmt.Print(buf.String())
	// Output:
	// level=info caller="example_test.go:line" fields.msg="Log with trace info" file=example_test.go line=123 user=john
}

// Example_cliApplication demonstrates a typical CLI application setup.
func Example_cliApplication() {
	var buf bytes.Buffer

	// Minimal, clean output for CLI
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     false, // Colors enabled for interactive terminals
		DisableStack:     true,
		DisableTimestamp: true,
		EnableTrace:      false,
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only field "msg"
	logger.WithField("msg", "Processing files...").Info("ignored")

	fmt.Println("CLI log written with colors")
	// Output:
	// CLI log written with colors
}

// Example_dockerContainer demonstrates JSON logging for containerized applications.
func Example_dockerContainer() {
	var buf bytes.Buffer

	// Structured JSON logs without colors for container log drivers
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true, // No colors for log aggregation
	}

	hook, err := loghko.NewWithWriter(&buf, opt, nil, &logrus.JSONFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":       "Container started",
		"container": "app-1",
		"image":     "myapp:latest",
	}).Info("ignored")

	fmt.Println("JSON log written for container stdout")
	// Output:
	// JSON log written for container stdout
}
