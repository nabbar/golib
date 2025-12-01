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

package hookwriter_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loghkw "github.com/nabbar/golib/logger/hookwriter"
)

// Example_basic demonstrates the simplest use case: creating a hook that writes to a buffer.
func Example_basic() {
	var buf bytes.Buffer

	// Configure the hook with minimal settings
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true, // Disable color for predictable output
	}

	// Create the hook writing to buffer
	hook, err := loghkw.New(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true, // Disable timestamp for predictable output
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Create and configure logger (output to Discard to avoid double write)
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Use Stderr to separate from example output
	logger.AddHook(hook)

	// IMPORTANT: The message parameter "ignored" is NOT used by the hook.
	// Only the fields (here "msg") are written to the file.
	// Exception: In AccessLog mode, only the message is used and fields are ignored.
	logger.WithField("msg", "Application started").Info("ignored")

	// Print what was written by the hook
	fmt.Print(buf.String())

	// Output:
	// level=info fields.msg="Application started"
}

// Example_fileWriter demonstrates writing logs to a file with JSON formatting.
func Example_fileWriter() {
	// Create a buffer to simulate a file (for example purposes)
	var buf bytes.Buffer

	// Configure options
	opt := &logcfg.OptionsStd{
		DisableStandard:  false,
		DisableColor:     true,
		DisableStack:     true,
		DisableTimestamp: true,
	}

	// Create hook with JSON formatter
	hook, err := loghkw.New(&buf, opt, nil, &logrus.JSONFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid double write
	logger.AddHook(hook)

	// Log with fields
	logger.WithFields(logrus.Fields{
		"user_id": 123,
		"action":  "login",
		"msg":     "User logged in",
	}).Info("ignored")

	fmt.Println("Log written to file")
	// Output:
	// Log written to file
}

// Example_accessLog demonstrates using access log mode for HTTP request logging.
func Example_accessLog() {
	var buf bytes.Buffer

	// Enable access log mode
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		EnableAccessLog: true, // Message-only mode
	}

	hook, err := loghkw.New(&buf, opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Setup logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid double write
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

	// Only handle error and fatal levels
	levels := []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
	}

	hook, err := loghkw.New(buf, opt, levels, &logrus.TextFormatter{
		DisableTimestamp: true,
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid writing to buf
	logger.AddHook(hook)

	// These won't be written by the hook (wrong level)
	logger.WithField("type", "info").Info("ignored")
	logger.WithField("type", "warn").Warn("ignored")

	// This will be written by the hook (error level)
	logger.WithField("type", "error").WithField("msg", "This is an error").Error("ignored")

	fmt.Printf("Hook captured: %s", buf.String())
	// Output:
	// Hook captured: level=error fields.msg="This is an error" type=error
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

	hook, err := loghkw.New(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid double write
	logger.AddHook(hook)

	// Log with fields that will be filtered
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

// Example_multipleHooks demonstrates using multiple hooks for different outputs.
func Example_multipleHooks() {
	var infoBuf, errorBuf bytes.Buffer

	// Hook for info/debug logs
	infoOpt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}
	infoHook, _ := loghkw.New(&infoBuf, infoOpt, []logrus.Level{
		logrus.InfoLevel,
		logrus.DebugLevel,
	}, &logrus.TextFormatter{DisableTimestamp: true})

	// Hook for error logs
	errorOpt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
	}
	errorHook, _ := loghkw.New(&errorBuf, errorOpt, []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
	}, &logrus.JSONFormatter{DisableTimestamp: true})

	// Setup logger with both hooks
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid writing to buffers
	logger.AddHook(infoHook)
	logger.AddHook(errorHook)

	logger.WithField("msg", "This goes to info buffer").WithField("target", "info").Info("ignored")
	logger.WithField("msg", "This goes to error buffer").WithField("target", "error").Error("ignored")

	fmt.Printf("Info has: %s", infoBuf.String())
	fmt.Printf("Error has: %s", errorBuf.String())
	// Output:
	// Info has: level=info fields.msg="This goes to info buffer" target=info
	// Error has: {"fields.msg":"This goes to error buffer","level":"error","msg":"","target":"error"}
}

// Example_disabledHook demonstrates how to conditionally disable the hook.
func Example_disabledHook() {
	opt := &logcfg.OptionsStd{
		DisableStandard: true, // This disables the hook
	}

	hook, err := loghkw.New(os.Stdout, opt, nil, nil)
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

// Example_nilWriter demonstrates error handling for nil writer.
func Example_nilWriter() {
	opt := &logcfg.OptionsStd{
		DisableStandard: false,
	}

	hook, err := loghkw.New(nil, opt, nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if hook == nil {
		fmt.Println("Hook was not created")
	}

	// Output:
	// Error: hook writer is nil
	// Hook was not created
}

// Example_traceEnabled demonstrates enabling trace information in logs.
func Example_traceEnabled() {
	var buf bytes.Buffer

	opt := &logcfg.OptionsStd{
		DisableStandard: false,
		DisableColor:    true,
		EnableTrace:     true, // Include caller/file/line information
	}

	hook, err := loghkw.New(&buf, opt, nil, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Avoid double write
	logger.AddHook(hook)

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
