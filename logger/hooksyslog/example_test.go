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

package hooksyslog_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	logsys "github.com/nabbar/golib/logger/hooksyslog"
	libptc "github.com/nabbar/golib/network/protocol"
)

// Example_basic demonstrates the simplest use case: creating a hook that writes to local syslog.
// Note: This example uses UDP which doesn't require an actual syslog daemon to be running.
func Example_basic() {
	// Configure the hook with minimal settings
	// In this example, we use UDP protocol which doesn't fail if no server is running
	opts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514", // UDP doesn't fail without server
		Tag:      "myapp",
		LogLevel: []string{"info", "warning", "error"},
	}

	// Create the hook
	hook, err := logsys.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true, // Disable timestamp for predictable output
	})
	if err != nil {
		fmt.Printf("Error creating hook: %v\n", err)
		return
	}
	defer hook.Close()

	// Start async writer goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	// Wait for hook to be ready
	time.Sleep(100 * time.Millisecond)

	// Create and configure logger
	logger := logrus.New()
	logger.SetOutput(os.Stderr) // Use Stderr to separate from syslog output
	logger.AddHook(hook)

	// IMPORTANT: The message parameter "ignored" is NOT used by the hook in standard mode.
	// Only the fields (here "msg") are written to syslog.
	// Exception: In AccessLog mode, only the message is used and fields are ignored.
	logger.WithField("msg", "Application started successfully").Info("ignored")

	// Wait for async write
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Log sent to syslog")

	// Output:
	// Log sent to syslog
}

// Example_remoteUdp demonstrates sending logs to a remote syslog server via UDP.
func Example_remoteUdp() {
	opts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514", // Remote syslog server
		Tag:      "remote-app",
		Facility: "LOCAL0", // Use LOCAL0 facility
		LogLevel: []string{"info", "error"},
	}

	hook, err := logsys.New(opts, &logrus.JSONFormatter{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: Use fields, not message parameter
	logger.WithFields(logrus.Fields{
		"msg":      "Remote logging test",
		"service":  "api",
		"instance": "prod-1",
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Log sent to remote syslog via UDP")

	// Output:
	// Log sent to remote syslog via UDP
}

// Example_accessLog demonstrates using access log mode for HTTP request logging.
// In this mode, behavior is reversed: the message IS written, fields are IGNORED.
func Example_accessLog() {
	opts := logcfg.OptionsSyslog{
		Network:         libptc.NetworkUDP.Code(),
		Host:            "localhost:514",
		Tag:             "http-access",
		EnableAccessLog: true, // Message-only mode
		LogLevel:        []string{"info"},
	}

	hook, err := logsys.New(opts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: In AccessLog mode, behavior is REVERSED!
	// The message "GET /api/users - 200 OK - 45ms" IS output.
	// The fields (method, path, status) are IGNORED.
	logger.WithFields(logrus.Fields{
		"method": "GET",
		"path":   "/api/users",
		"status": 200,
	}).Info("GET /api/users - 200 OK - 45ms")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Access log sent to syslog")

	// Output:
	// Access log sent to syslog
}

// Example_levelFiltering demonstrates filtering logs by level.
func Example_levelFiltering() {
	opts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514",
		Tag:      "filtered-app",
		LogLevel: []string{"error", "fatal"}, // Only errors and above
	}

	hook, err := logsys.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// This will be written (error level)
	// Note: message "ignored" is NOT used, only the field "msg"
	logger.WithField("msg", "Database connection failed").Error("ignored")

	// This won't be written to syslog (wrong level)
	logger.WithField("msg", "Request completed").Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Filtered logs sent to syslog")

	// Output:
	// Filtered logs sent to syslog
}

// Example_fieldFiltering demonstrates filtering specific fields from output.
func Example_fieldFiltering() {
	// Configure to filter out stack and timestamp
	opts := logcfg.OptionsSyslog{
		Network:          libptc.NetworkUDP.Code(),
		Host:             "localhost:514",
		Tag:              "clean-app",
		DisableStack:     true,  // Remove stack fields
		DisableTimestamp: true,  // Remove time fields
		EnableTrace:      false, // Remove caller/file/line fields
		LogLevel:         []string{"info"},
	}

	hook, err := logsys.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Filtered log entry",
		"stack":  "will be filtered out",
		"caller": "will be filtered out",
		"user":   "john",
		"action": "login",
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Filtered log sent to syslog")

	// Output:
	// Filtered log sent to syslog
}

// Example_gracefulShutdown demonstrates proper shutdown with the Done channel.
func Example_gracefulShutdown() {
	opts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514",
		Tag:      "shutdown-test",
		LogLevel: []string{"info"},
	}

	hook, err := logsys.New(opts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// Send some logs
	logger.WithField("msg", "Starting shutdown sequence").Info("ignored")
	time.Sleep(100 * time.Millisecond)

	// Signal shutdown
	cancel()
	hook.Close()

	// Wait for completion
	select {
	case <-hook.Done():
		fmt.Println("Hook shutdown complete")
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout waiting for shutdown")
	}

	// Output:
	// Hook shutdown complete
}

// Example_structuredLogging demonstrates structured logging with JSON formatter.
func Example_structuredLogging() {
	opts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514",
		Tag:      "structured-app",
		LogLevel: []string{"info"},
	}

	hook, err := logsys.New(opts, &logrus.JSONFormatter{})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message parameter is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"user_id":    12345,
		"action":     "purchase",
		"amount":     99.99,
		"currency":   "USD",
		"msg":        "Purchase completed",
		"request_id": "abc-123-def",
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Structured log sent to syslog")

	// Output:
	// Structured log sent to syslog
}

// Example_multipleHooks demonstrates using multiple hooks for different destinations.
func Example_multipleHooks() {
	// Hook for errors only
	errorOpts := logcfg.OptionsSyslog{
		Network:  libptc.NetworkUDP.Code(),
		Host:     "localhost:514",
		Tag:      "errors",
		LogLevel: []string{"error", "fatal"},
	}

	errorHook, err := logsys.New(errorOpts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer errorHook.Close()

	// Hook for all levels
	allOpts := logcfg.OptionsSyslog{
		Network: libptc.NetworkUDP.Code(),
		Host:    "localhost:514",
		Tag:     "all-logs",
	}

	allHook, err := logsys.New(allOpts, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer allHook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go errorHook.Run(ctx)
	go allHook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(errorHook)
	logger.AddHook(allHook)

	// This goes to both hooks
	logger.WithField("msg", "Critical error occurred").Error("ignored")

	// This goes only to allHook
	logger.WithField("msg", "Normal operation").Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Logs sent to multiple syslog destinations")

	// Output:
	// Logs sent to multiple syslog destinations
}

// Example_traceEnabled demonstrates enabling trace information in logs.
func Example_traceEnabled() {
	opts := logcfg.OptionsSyslog{
		Network:     libptc.NetworkUDP.Code(),
		Host:        "localhost:514",
		Tag:         "trace-app",
		EnableTrace: true, // Include caller/file/line information
		LogLevel:    []string{"info"},
	}

	hook, err := logsys.New(opts, &logrus.TextFormatter{
		DisableTimestamp: true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer hook.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hook.Run(ctx)

	time.Sleep(100 * time.Millisecond)

	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.AddHook(hook)

	// IMPORTANT: message "ignored" is NOT used, only fields
	logger.WithFields(logrus.Fields{
		"msg":    "Log with trace info",
		"caller": "main.processRequest",
		"file":   "main.go",
		"line":   42,
	}).Info("ignored")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("Log with trace sent to syslog")

	// Output:
	// Log with trace sent to syslog
}
