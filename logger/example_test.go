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

package logger_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Example_basicLogging demonstrates basic logging at different levels.
func Example_basicLogging() {
	// Create logger
	logger := liblog.New(context.Background())
	defer logger.Close()

	// Configure for silent operation (no actual output in examples)
	logger.SetLevel(loglvl.DebugLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Log at different levels
	logger.Debug("Debug message", nil)
	logger.Info("Info message", nil)
	logger.Warning("Warning message", nil)
	logger.Error("Error message", nil)

	fmt.Println("Logged messages at multiple levels")
	// Output: Logged messages at multiple levels
}

// Example_structuredLogging demonstrates logging with structured data.
func Example_structuredLogging() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Log with structured data
	userData := map[string]interface{}{
		"user_id": 12345,
		"email":   "user@example.com",
		"role":    "admin",
	}

	logger.Info("User logged in", userData)

	fmt.Println("Logged structured data")
	// Output: Logged structured data
}

// Example_withFields demonstrates using default fields that apply to all log entries.
func Example_withFields() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Set default fields
	fields := logfld.New(context.Background())
	fields.Add("service", "api")
	fields.Add("environment", "production")
	fields.Add("version", "1.2.3")
	logger.SetFields(fields)

	// These fields will be included in all log entries
	logger.Info("Request processed", nil)
	logger.Info("Response sent", nil)

	fmt.Println("Logged with default fields")
	// Output: Logged with default fields
}

// Example_formattedMessages demonstrates using format strings in log messages.
func Example_formattedMessages() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Format strings work like fmt.Sprintf
	userID := 42
	username := "john.doe"
	duration := 150

	logger.Info("User %s (ID: %d) processed in %dms", nil, username, userID, duration)

	fmt.Println("Logged formatted message")
	// Output: Logged formatted message
}

// Example_errorChecking demonstrates the CheckError convenience method.
func Example_errorChecking() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Simulate an operation that might fail
	performOperation := func() error {
		return nil // Success
	}

	err := performOperation()
	if logger.CheckError(loglvl.ErrorLevel, loglvl.InfoLevel, "Operation completed", err) {
		fmt.Println("Error occurred")
		return
	}

	fmt.Println("Operation successful")
	// Output: Operation successful
}

// Example_entryBuilder demonstrates using the Entry builder pattern for complex logs.
func Example_entryBuilder() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Build complex log entry
	entry := logger.Entry(loglvl.InfoLevel, "Database query executed")
	entry.FieldAdd("query", "SELECT * FROM users")
	entry.FieldAdd("duration", "45ms")
	entry.FieldAdd("rows", 127)
	entry.Log()

	fmt.Println("Logged with entry builder")
	// Output: Logged with entry builder
}

// Example_accessLog demonstrates HTTP access logging.
func Example_accessLog() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Simulate HTTP request logging
	logger.Access(
		"192.168.1.100",      // Remote address
		"john.doe",           // Remote user
		time.Now(),           // Request time
		150*time.Millisecond, // Latency
		"GET",                // Method
		"/api/users",         // Path
		"HTTP/1.1",           // Protocol
		200,                  // Status
		1024,                 // Size
	).Log()

	fmt.Println("Logged access entry")
	// Output: Logged access entry
}

// Example_cloningLogger demonstrates creating independent logger clones.
func Example_cloningLogger() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Set some fields on the original
	logger.SetFields(logfld.New(context.Background()).Add("logger", "main"))

	// Clone creates an independent copy
	clonedLogger, err := logger.Clone()
	if err != nil {
		panic(err)
	}
	defer clonedLogger.Close()

	// Modify the clone without affecting original
	clonedLogger.SetLevel(loglvl.DebugLevel)
	clonedLogger.SetFields(logfld.New(context.Background()).Add("logger", "worker"))

	logger.Info("Main logger message", nil)
	clonedLogger.Debug("Cloned logger message", nil)

	fmt.Println("Used cloned logger")
	// Output: Used cloned logger
}

// Example_standardLoggerIntegration demonstrates using logger with standard log package.
func Example_standardLoggerIntegration() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Get a standard logger for third-party code
	stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)

	// Use it like standard log.Logger
	stdLogger.Println("Message from standard logger")
	stdLogger.Printf("Formatted: %s", "message")

	fmt.Println("Used standard logger integration")
	// Output: Used standard logger integration
}

// Example_ioWriterIntegration demonstrates using logger as an io.Writer.
func Example_ioWriterIntegration() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Set the level for io.Writer interface
	logger.SetIOWriterLevel(loglvl.WarnLevel)

	// Use as io.Writer
	_, _ = io.WriteString(logger, "Message written through io.Writer")

	fmt.Println("Used as io.Writer")
	// Output: Used as io.Writer
}

// Example_logDetails demonstrates the most flexible logging method.
func Example_logDetails() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// LogDetails gives full control
	err1 := errors.New("first error")
	err2 := errors.New("second error")
	errs := []error{err1, err2}

	fields := logfld.New(context.Background())
	fields.Add("operation", "database_migration")
	fields.Add("step", 3)

	data := map[string]interface{}{
		"tables_affected": 5,
		"records_updated": 1523,
	}

	logger.LogDetails(
		loglvl.ErrorLevel,
		"Migration encountered errors",
		data,
		errs,
		fields,
	)

	fmt.Println("Logged detailed entry")
	// Output: Logged detailed entry
}

// Example_multipleErrors demonstrates logging multiple errors.
func Example_multipleErrors() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.ErrorLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Multiple errors can be logged together
	err1 := errors.New("connection timeout")
	err2 := errors.New("retry limit exceeded")

	if logger.CheckError(loglvl.ErrorLevel, loglvl.NilLevel, "Operation failed", err1, err2) {
		fmt.Println("Multiple errors logged")
	}
	// Output: Multiple errors logged
}

// Example_levelManagement demonstrates dynamic level changes.
func Example_levelManagement() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Start with Info level
	logger.SetLevel(loglvl.InfoLevel)
	logger.Debug("This won't be logged", nil)
	logger.Info("This will be logged", nil)

	// Change to Debug level
	logger.SetLevel(loglvl.DebugLevel)
	logger.Debug("Now this will be logged", nil)

	currentLevel := logger.GetLevel()
	fmt.Printf("Current level: %s\n", currentLevel.String())
	// Output: Current level: Debug
}

// Example_filteringMessages demonstrates filtering log messages.
func Example_filteringMessages() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Set up filters to drop messages containing certain patterns
	logger.SetIOWriterFilter("password", "secret", "token")

	// These messages will be filtered out
	_, _ = logger.Write([]byte("User password is 12345"))
	_, _ = logger.Write([]byte("API token: abc123"))

	// This message will pass through
	_, _ = logger.Write([]byte("User logged in successfully"))

	fmt.Println("Filtering configured")
	// Output: Filtering configured
}

// Example_concurrentLogging demonstrates thread-safe concurrent logging.
func Example_concurrentLogging() {
	logger := liblog.New(context.Background())
	defer logger.Close()

	logger.SetLevel(loglvl.InfoLevel)
	_ = logger.SetOptions(&logcfg.Options{
		Stdout: &logcfg.OptionsStd{
			DisableStandard: true,
		},
	})

	// Launch multiple goroutines
	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func(id int) {
			logger.Info("Message from goroutine", map[string]interface{}{
				"goroutine_id": id,
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	fmt.Println("Concurrent logging completed")
	// Output: Concurrent logging completed
}
