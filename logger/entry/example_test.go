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

package entry_test

import (
	"errors"
	"fmt"
	"time"

	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	"github.com/sirupsen/logrus"
)

// Example_basicLogging demonstrates the simplest usage of the entry package
// for basic structured logging with logrus.
func Example_basicLogging() {
	// Create a logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create fields for structured data
	fields := logfld.New(nil)

	// Create and log an entry
	logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		SetEntryContext(time.Now(), 0, "", "", 0, "Application started").
		Log()
}

// Example_errorLogging demonstrates logging with error information.
func Example_errorLogging() {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	fields := logfld.New(nil)

	// Create an error entry
	err := errors.New("database connection failed")

	logent.New(loglvl.ErrorLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		ErrorAdd(true, err). // cleanNil=true filters out nil errors
		SetEntryContext(time.Now(), 0, "ConnectDB", "db.go", 42, "Failed to connect to database").
		Log()
}

// Example_multipleErrors demonstrates handling multiple errors in a single entry.
func Example_multipleErrors() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Collect multiple errors
	err1 := errors.New("connection timeout")
	err2 := errors.New("retry failed")

	logent.New(loglvl.ErrorLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		ErrorAdd(true, err1, err2).
		SetEntryContext(time.Now(), 0, "", "", 0, "Multiple failures occurred").
		Log()
}

// Example_structuredData demonstrates adding custom structured data to log entries.
func Example_structuredData() {
	logger := logrus.New()

	// Create fields with custom data
	fields := logfld.New(nil)

	// Create entry with custom fields
	logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		FieldAdd("user_id", 12345).
		FieldAdd("action", "login").
		FieldAdd("ip_address", "192.168.1.1").
		SetEntryContext(time.Now(), 0, "", "", 0, "User logged in").
		Log()
}

// Example_dataAttachment demonstrates attaching arbitrary data structures.
func Example_dataAttachment() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Create data structure
	data := map[string]interface{}{
		"request_id":  "req-123",
		"duration_ms": 450,
		"status_code": 200,
	}

	logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		DataSet(data).
		SetEntryContext(time.Now(), 0, "", "", 0, "Request completed").
		Log()
}

// Example_methodChaining demonstrates the fluent API for building complex entries.
func Example_methodChaining() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Build complex entry with method chaining
	logent.New(loglvl.WarnLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		FieldAdd("component", "auth").
		FieldAdd("attempt", 3).
		ErrorAdd(true, errors.New("invalid credentials")).
		DataSet(map[string]string{"username": "user@example.com"}).
		SetEntryContext(time.Now(), 0, "Authenticate", "auth.go", 100, "Authentication failed").
		Log()
}

// Example_conditionalLogging demonstrates using Check() for conditional logging
// with different levels based on error presence.
func Example_conditionalLogging() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Simulate operation that may fail
	var err error
	// err = performOperation()

	entry := logent.New(loglvl.ErrorLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		ErrorAdd(true, err)

	// Check will log at ErrorLevel if errors exist, InfoLevel otherwise
	hasErrors := entry.Check(loglvl.InfoLevel)

	if hasErrors {
		fmt.Println("Operation failed with errors")
	} else {
		fmt.Println("Operation succeeded")
	}
}

// Example_messageOnly demonstrates simple message-only logging without structured fields.
func Example_messageOnly() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Log only the message, ignoring all fields
	logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		SetMessageOnly(true).
		SetEntryContext(time.Now(), 0, "", "", 0, "Simple console message").
		Log()
}

// Example_fieldManagement demonstrates managing custom fields throughout entry lifecycle.
func Example_fieldManagement() {
	logger := logrus.New()

	// Create base fields
	baseFields := logfld.New(nil)
	baseFields.Add("app", "myapp")
	baseFields.Add("version", "1.0.0")

	// Create additional fields
	reqFields := logfld.New(nil)
	reqFields.Add("request_id", "req-456")

	// Build entry with merged fields
	entry := logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(baseFields).
		FieldAdd("endpoint", "/api/users").
		FieldMerge(reqFields).
		SetEntryContext(time.Now(), 0, "", "", 0, "API request")

	// Log the entry
	entry.Log()

	// Clean specific fields for reuse
	entry.FieldClean("request_id").
		FieldAdd("request_id", "req-457").
		SetEntryContext(time.Now(), 0, "", "", 0, "Next request").
		Log()
}

// Example_contextInformation demonstrates logging with detailed context information
// including stack traces, caller information, and file/line numbers.
func Example_contextInformation() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Log with full context information
	logent.New(loglvl.DebugLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields).
		SetEntryContext(
			time.Now(),                // timestamp
			12345,                     // goroutine stack number
			"ProcessRequest",          // caller function name
			"handler.go",              // file name
			78,                        // line number
			"Processing user request", // message
		).
		Log()
}

// Example_errorManagement demonstrates various error management operations.
func Example_errorManagement() {
	logger := logrus.New()
	fields := logfld.New(nil)

	// Create entry
	entry := logent.New(loglvl.ErrorLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields)

	// Add errors incrementally
	entry.ErrorAdd(true, errors.New("first error"))
	entry.ErrorAdd(true, errors.New("second error"))

	// Log current errors
	entry.SetEntryContext(time.Now(), 0, "", "", 0, "Multiple errors").Log()

	// Clean errors and add new ones
	entry.ErrorClean().
		ErrorAdd(true, errors.New("new error after cleanup")).
		SetEntryContext(time.Now(), 0, "", "", 0, "After cleanup").
		Log()

	// Set errors directly with a slice
	errs := []error{
		errors.New("error from slice 1"),
		errors.New("error from slice 2"),
	}
	entry.ErrorSet(errs).
		SetEntryContext(time.Now(), 0, "", "", 0, "Set errors").
		Log()
}

// Example_levelControl demonstrates controlling log levels dynamically.
func Example_levelControl() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel) // Allow all levels

	fields := logfld.New(nil)

	// Create entry and change level
	entry := logent.New(loglvl.DebugLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(fields)

	// Log at debug level
	entry.SetEntryContext(time.Now(), 0, "", "", 0, "Debug message").Log()

	// Change to info level and log again
	entry.SetLevel(loglvl.InfoLevel).
		SetEntryContext(time.Now(), 0, "", "", 0, "Info message").
		Log()

	// Change to error level
	entry.SetLevel(loglvl.ErrorLevel).
		ErrorAdd(true, errors.New("error occurred")).
		SetEntryContext(time.Now(), 0, "", "", 0, "Error message").
		Log()
}

// Example_complexWorkflow demonstrates a complete, complex logging workflow
// combining all major features.
func Example_complexWorkflow() {
	// Setup
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	baseFields := logfld.New(nil)
	baseFields.Add("service", "payment-api")
	baseFields.Add("environment", "production")

	// Simulate processing
	userID := 12345
	amount := 99.99
	var processingErr error

	// Build comprehensive log entry
	entry := logent.New(loglvl.InfoLevel).
		SetLogger(func() *logrus.Logger { return logger }).
		FieldSet(baseFields).
		FieldAdd("user_id", userID).
		FieldAdd("amount", amount).
		FieldAdd("currency", "USD")

	// Add transaction data
	txData := map[string]interface{}{
		"transaction_id": "tx-789",
		"payment_method": "credit_card",
		"timestamp":      time.Now().Unix(),
	}
	entry.DataSet(txData)

	// Check for errors
	if processingErr != nil {
		entry.SetLevel(loglvl.ErrorLevel).
			ErrorAdd(true, processingErr).
			SetEntryContext(time.Now(), 0, "ProcessPayment", "payment.go", 156, "Payment processing failed")
	} else {
		entry.SetEntryContext(time.Now(), 0, "ProcessPayment", "payment.go", 180, "Payment processed successfully")
	}

	// Log the entry
	entry.Log()
}
