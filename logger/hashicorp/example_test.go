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

package hashicorp_test

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	liblog "github.com/nabbar/golib/logger"
	loghc "github.com/nabbar/golib/logger/hashicorp"
)

// Example_basic demonstrates the simplest usage of the HashiCorp logger adapter.
// This is suitable for basic integration where default behavior is acceptable.
func Example_basic() {
	// Create a mock logger (in real usage, use a configured golib logger)
	mockLogger := NewMockLogger()

	// Create hclog adapter
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Use like any hclog.Logger
	hcLogger.Info("application started")

	fmt.Println("HashiCorp logger adapter created")
	// Output:
	// HashiCorp logger adapter created
}

// Example_withLevels demonstrates logging at different severity levels.
func Example_withLevels() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Log at different levels
	hcLogger.Trace("trace level message")
	hcLogger.Debug("debug level message")
	hcLogger.Info("info level message")
	hcLogger.Warn("warning level message")
	hcLogger.Error("error level message")

	fmt.Printf("Logged %d messages\n", len(mockLogger.entries))
	// Output:
	// Logged 5 messages
}

// Example_setLevel demonstrates dynamic log level adjustment.
func Example_setLevel() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Set to Info level
	hcLogger.SetLevel(hclog.Info)
	fmt.Println("Level set to Info")

	// Set to Debug level
	hcLogger.SetLevel(hclog.Debug)
	fmt.Println("Level set to Debug")

	// Set to Warn level
	hcLogger.SetLevel(hclog.Warn)
	fmt.Println("Level set to Warn")

	// Output:
	// Level set to Info
	// Level set to Debug
	// Level set to Warn
}

// Example_levelChecks demonstrates checking log levels before expensive operations.
func Example_levelChecks() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	hcLogger.SetLevel(hclog.Debug)

	// Check level before expensive operations
	if hcLogger.IsDebug() {
		fmt.Println("Debug logging is enabled")
	}

	if hcLogger.IsInfo() {
		fmt.Println("Info logging is enabled")
	}

	// Output:
	// Debug logging is enabled
	// Info logging is enabled
}

// Example_withContext demonstrates adding context fields to logs.
func Example_withContext() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Add context with key-value pairs
	contextLogger := hcLogger.With("request_id", "req-123", "user", "alice")

	// All subsequent logs include the context
	contextLogger.Info("processing request")
	contextLogger.Warn("slow operation detected")

	fmt.Println("Logs include context fields")
	// Output:
	// Logs include context fields
}

// Example_namedLogger demonstrates creating named sub-loggers for components.
func Example_namedLogger() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Create named loggers for different components
	databaseLogger := hcLogger.Named("database")
	cacheLogger := hcLogger.Named("cache")
	apiLogger := hcLogger.Named("api")

	// Each logger includes its name
	databaseLogger.Info("connection established")
	cacheLogger.Info("cache initialized")
	apiLogger.Info("server started")

	fmt.Println("Created 3 named loggers")
	// Output:
	// Created 3 named loggers
}

// Example_resetNamed demonstrates resetting logger names.
func Example_resetNamed() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Set initial name
	hcLogger.Named("component-a")

	// Reset to new name
	hcLogger.ResetNamed("component-b")

	fmt.Println("Logger name reset")
	// Output:
	// Logger name reset
}

// Example_setDefault demonstrates setting the global default hclog logger.
func Example_setDefault() {
	mockLogger := NewMockLogger()

	// Set as default hclog logger
	loghc.SetDefault(func() liblog.Logger { return mockLogger })

	// Now hclog.Default() returns our adapter
	defaultLogger := hclog.Default()
	defaultLogger.Info("using global default logger")

	fmt.Println("Default logger set")
	// Output:
	// Default logger set
}

// Example_standardLogger demonstrates getting a standard library logger.
func Example_standardLogger() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Get standard library logger
	opts := &hclog.StandardLoggerOptions{
		ForceLevel: hclog.Info,
	}
	stdLogger := hcLogger.StandardLogger(opts)

	// Use like *log.Logger
	stdLogger.Println("standard library log message")

	fmt.Println("Standard logger created")
	// Output:
	// Standard logger created
}

// Example_standardWriter demonstrates getting an io.Writer for logging.
func Example_standardWriter() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Get io.Writer
	opts := &hclog.StandardLoggerOptions{}
	writer := hcLogger.StandardWriter(opts)

	// Use like any io.Writer
	writer.Write([]byte("message to writer\n"))

	fmt.Println("Standard writer obtained")
	// Output:
	// Standard writer obtained
}

// Example_nilSafe demonstrates graceful handling of nil logger.
func Example_nilSafe() {
	// Create adapter with nil logger function
	hcLogger := loghc.New(nil)

	// All operations are safe and won't panic
	hcLogger.Info("this is safe")
	hcLogger.Debug("no panic")
	hcLogger.Error("graceful handling")

	fmt.Println("Nil logger handled gracefully")
	// Output:
	// Nil logger handled gracefully
}

// Example_logMethod demonstrates using the generic Log method.
func Example_logMethod() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Log with explicit level
	hcLogger.Log(hclog.Info, "info message")
	hcLogger.Log(hclog.Warn, "warn message")
	hcLogger.Log(hclog.Error, "error message")

	// NoLevel and Off are ignored
	hcLogger.Log(hclog.NoLevel, "ignored")
	hcLogger.Log(hclog.Off, "also ignored")

	fmt.Printf("Logged %d messages\n", len(mockLogger.entries))
	// Output:
	// Logged 3 messages
}

// Example_impliedArgs demonstrates retrieving stored context arguments.
func Example_impliedArgs() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Add context
	hcLogger.With("key1", "value1", "key2", "value2")

	// Store args in fields
	mockLogger.fields = mockLogger.fields.Add(loghc.HCLogArgs, []interface{}{"key", "val"})

	// Retrieve implied args
	args := hcLogger.ImpliedArgs()
	fmt.Printf("Retrieved %d arguments\n", len(args))

	// Output:
	// Retrieved 2 arguments
}

// Example_name demonstrates retrieving logger name.
func Example_name() {
	mockLogger := NewMockLogger()
	hcLogger := loghc.New(func() liblog.Logger { return mockLogger })

	// Set name
	mockLogger.fields = mockLogger.fields.Add(loghc.HCLogName, "my-component")

	// Retrieve name
	name := hcLogger.Name()
	fmt.Printf("Logger name: %s\n", name)

	// Output:
	// Logger name: my-component
}
