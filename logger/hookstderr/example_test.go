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

package hookstderr_test

import (
	"bytes"
	"fmt"
	"io"

	logcfg "github.com/nabbar/golib/logger/config"
	loghks "github.com/nabbar/golib/logger/hookstderr"
	"github.com/sirupsen/logrus"
)

// Example_basic demonstrates the simplest use case: creating a hook that writes errors to a buffer.
func Example_basic() {
	var buf bytes.Buffer

	// Configure the hook with minimal settings
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true, // Disable color for predictable output
	}

	// Create the hook writing to buffer (simulating stderr)
	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true, // Disable timestamp for predictable output
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Create and configure logger (discard default output to avoid double write)
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// Log an error message with a field (required for hook to write)
	logger.WithField("msg", "Application error occurred").Error("ignored")

	// Print what was written by the hook to stderr
	fmt.Print(buf.String())

	// Output:
	// level=error fields.msg="Application error occurred"
}

// Example_errorLogging demonstrates writing error logs to stderr with JSON formatting.
func Example_errorLogging() {
	// Create a buffer to simulate stderr (for example purposes)
	var buf bytes.Buffer

	// Configure options for error logging
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     true,
		DisableTimestamp: true,
	}

	// Create hook with JSON formatter for structured error logs
	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.JSONFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// Log error with fields
	logger.WithFields(logrus.Fields{
		"error_code": "E500",
		"request_id": "abc123",
		"msg":        "Database connection failed",
	}).Error("ignored")

	fmt.Println("Error logged to stderr")
	// Output:
	// Error logged to stderr
}

// Example_levelFiltering demonstrates filtering errors by level.
func Example_levelFiltering() {
	var buf = bytes.NewBuffer(make([]byte, 0))

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}

	// Only handle error, fatal, and panic levels (typical for stderr)
	levels := []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	hook, err := loghks.NewWithWriter(buf, opt, levels, &logrus.TextFormatter{
		DisableTimestamp: true,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// These won't be written by the hook (wrong level)
	logger.WithField("type", "info").Info("ignored")
	logger.WithField("type", "warn").Warn("ignored")

	// This will be written by the hook (error level)
	logger.WithField("type", "error").WithField("msg", "Critical error").Error("ignored")

	fmt.Printf("Stderr captured: %s", buf.String())
	// Output:
	// Stderr captured: level=error fields.msg="Critical error" type=error
}

// Example_fieldFiltering demonstrates filtering specific fields from error output.
func Example_fieldFiltering() {
	var buf bytes.Buffer

	// Configure to filter out stack and timestamp from errors
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     true,  // Remove stack fields
		DisableTimestamp: true,  // Remove time fields
		EnableTrace:      false, // Remove caller/file/line fields
	}

	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// Log error with fields that will be filtered
	logger.WithFields(logrus.Fields{
		"msg":    "Clean error message",
		"stack":  "trace info",
		"caller": "main.go:123",
		"module": "database",
	}).Error("ignored")

	// Only "module" field remains after filtering
	fmt.Print(buf.String())
	// Output:
	// level=error fields.msg="Clean error message" module=database
}

// Example_separateStreams demonstrates using both stdout and stderr hooks.
func Example_separateStreams() {
	var stdoutBuf, stderrBuf bytes.Buffer

	// Hook for info logs (stdout)
	stdoutOpt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}
	// Note: We would use hookstdout.NewWithWriter for stdout in real code
	// For this example, we simulate it with a buffer
	infoHook, _ := loghks.NewWithWriter(&stdoutBuf, stdoutOpt, []logrus.Level{
		logrus.InfoLevel,
		logrus.DebugLevel,
	}, &logrus.TextFormatter{DisableTimestamp: true})

	// Hook for error logs (stderr)
	stderrOpt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}
	errorHook, _ := loghks.NewWithWriter(&stderrBuf, stderrOpt, []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
	}, &logrus.JSONFormatter{DisableTimestamp: true})

	// Setup logger with both hooks
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(infoHook)
	logger.AddHook(errorHook)

	logger.WithField("msg", "Normal operation").WithField("stream", "stdout").Info("ignored")
	logger.WithField("msg", "Error occurred").WithField("stream", "stderr").Error("ignored")

	fmt.Printf("Stdout: %s", stdoutBuf.String())
	fmt.Printf("Stderr: %s", stderrBuf.String())
	// Output:
	// Stdout: level=info fields.msg="Normal operation" stream=stdout
	// Stderr: {"fields.msg":"Error occurred","level":"error","msg":"","stream":"stderr"}
}

// Example_accessLog demonstrates using access log mode for clean error messages.
func Example_accessLog() {
	var buf bytes.Buffer

	// Enable access log mode for message-only output
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		EnableAccessLog: true, // Message-only mode
	}

	hook, err := loghks.NewWithWriter(&buf, opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// Log error (fields are ignored in access log mode)
	logger.WithFields(logrus.Fields{
		"status_code": 500,
		"path":        "/api/users",
	}).Error("500 Internal Server Error - /api/users - 123ms")

	fmt.Print(buf.String())
	// Output:
	// 500 Internal Server Error - /api/users - 123ms
}

// Example_disabledHook demonstrates how to conditionally disable the stderr hook.
func Example_disabledHook() {
	opt := &logcfg.OptionsStd{
		DisableStandard: true, // This disables the hook
	}

	hook, err := loghks.New(opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if hook == nil {
		fmt.Println("Stderr hook is disabled")
	} else {
		fmt.Println("Stderr hook is enabled")
	}

	// Output:
	// Stderr hook is disabled
}

// Example_customWriter demonstrates using a custom writer for testing.
func Example_customWriter() {
	var buf bytes.Buffer

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}

	// Use custom buffer instead of os.Stderr
	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	logger.WithField("msg", "Error captured in buffer").Error("ignored")

	// Buffer now contains the error log
	fmt.Printf("Captured: %s", buf.String())
	// Output:
	// Captured: level=error fields.msg="Error captured in buffer"
}

// Example_traceEnabled demonstrates enabling trace information in error logs.
func Example_traceEnabled() {
	var buf bytes.Buffer

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
		EnableTrace:     true, // Include caller/file/line information
	}

	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	logger.WithFields(logrus.Fields{
		"msg":    "Error with trace info",
		"caller": "example_test.go:line",
		"file":   "example_test.go",
		"line":   456,
		"module": "auth",
	}).Error("ignored")

	// Trace fields are included because EnableTrace is true
	fmt.Print(buf.String())
	// Output:
	// level=error caller="example_test.go:line" fields.msg="Error with trace info" file=example_test.go line=456 module=auth
}

// Example_errorWithStack demonstrates logging errors with stack traces.
func Example_errorWithStack() {
	var buf bytes.Buffer

	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     false, // Include stack traces
		DisableTimestamp: true,
	}

	hook, err := loghks.NewWithWriter(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard default output
	logger.AddHook(hook)

	// Log error with stack trace
	logger.WithFields(logrus.Fields{
		"msg":   "Error with stack",
		"stack": "goroutine 1 [running]:\nmain.main()",
		"code":  "E001",
	}).Error("ignored")

	fmt.Print(buf.String())
	// Output:
	// level=error code=E001 fields.msg="Error with stack" stack="goroutine 1 [running]:\nmain.main()"
}
