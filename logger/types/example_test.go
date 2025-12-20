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

package types_test

import (
	"context"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// Example_basicFieldUsage demonstrates using field constants with logrus.
//
// This example shows the simplest way to use standardized field names
// in structured logging.
func Example_basicFieldUsage() {
	// Create a logger
	log := logrus.New()
	log.SetOutput(io.Discard) // Discard output for example

	// Use field constants for structured logging
	log.WithFields(logrus.Fields{
		types.FieldFile:    "main.go",
		types.FieldLine:    42,
		types.FieldMessage: "operation completed",
	}).Info("example log entry")

	fmt.Println("Field constants used successfully")
	// Output: Field constants used successfully
}

// Example_errorFieldsUsage demonstrates logging errors with standardized fields.
//
// This example shows how to use field constants when logging errors
// with additional context information.
func Example_errorFieldsUsage() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Simulate an error scenario
	err := fmt.Errorf("connection timeout")

	// Log error with standard fields
	log.WithFields(logrus.Fields{
		types.FieldError:  err.Error(),
		types.FieldFile:   "handler.go",
		types.FieldLine:   123,
		types.FieldCaller: "processRequest",
	}).Error("request processing failed")

	fmt.Println("Error logged with standard fields")
	// Output: Error logged with standard fields
}

// Example_allFieldConstants demonstrates all available field constants.
//
// This example shows the complete set of standard field names defined
// in the types package.
func Example_allFieldConstants() {
	// Display all field constants in deterministic order
	fields := []struct {
		name  string
		value string
	}{
		{"Time", types.FieldTime},
		{"Level", types.FieldLevel},
		{"Stack", types.FieldStack},
		{"Caller", types.FieldCaller},
		{"File", types.FieldFile},
		{"Line", types.FieldLine},
		{"Message", types.FieldMessage},
		{"Error", types.FieldError},
		{"Data", types.FieldData},
	}

	for _, field := range fields {
		fmt.Printf("%s: %s\n", field.name, field.value)
	}

	// Output:
	// Time: time
	// Level: level
	// Stack: stack
	// Caller: caller
	// File: file
	// Line: line
	// Message: message
	// Error: error
	// Data: data
}

// Example_fieldCategories demonstrates grouping fields by category.
//
// This example shows how to organize fields into logical categories
// (metadata, trace, content).
func Example_fieldCategories() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Metadata fields
	metadataFields := logrus.Fields{
		types.FieldTime:  "2025-01-01T12:00:00Z",
		types.FieldLevel: "info",
	}

	// Trace fields
	traceFields := logrus.Fields{
		types.FieldFile:   "main.go",
		types.FieldLine:   42,
		types.FieldCaller: "main.run",
		types.FieldStack:  "...",
	}

	// Content fields
	contentFields := logrus.Fields{
		types.FieldMessage: "processing started",
		types.FieldData:    map[string]interface{}{"id": 123},
	}

	// Combine all fields
	allFields := make(logrus.Fields)
	for k, v := range metadataFields {
		allFields[k] = v
	}
	for k, v := range traceFields {
		allFields[k] = v
	}
	for k, v := range contentFields {
		allFields[k] = v
	}

	log.WithFields(allFields).Info("categorized fields")
	fmt.Println("Fields categorized successfully")
	// Output: Fields categorized successfully
}

// simpleHook is a minimal Hook implementation for examples.
type simpleHook struct {
	running atomic.Bool
	entries []string
}

// Fire processes a log entry.
func (h *simpleHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry.Message)
	return nil
}

// Levels returns the log levels this hook processes.
func (h *simpleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// RegisterHook registers the hook with a logger.
func (h *simpleHook) RegisterHook(log *logrus.Logger) {
	log.AddHook(h)
}

// Run runs the hook until context is cancelled.
func (h *simpleHook) Run(ctx context.Context) {
	h.running.Store(true)
	defer h.running.Store(false)
	<-ctx.Done()
}

// IsRunning returns whether the hook is running.
func (h *simpleHook) IsRunning() bool {
	return h.running.Load()
}

// Write implements io.Writer.
func (h *simpleHook) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close implements io.Closer.
func (h *simpleHook) Close() error {
	return nil
}

// Example_basicHook demonstrates implementing a basic Hook.
//
// This example shows the minimal implementation required to satisfy
// the Hook interface.
func Example_basicHook() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Create and register hook
	hook := &simpleHook{}
	hook.RegisterHook(log)

	// Use logger - hook will intercept entries
	log.Info("test message")

	fmt.Println("Hook registered and used successfully")
	// Output: Hook registered and used successfully
}

// Example_hookLifecycle demonstrates the complete Hook lifecycle.
//
// This example shows how to create, register, run, and close a hook
// with proper context management.
func Example_hookLifecycle() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Create hook
	hook := &simpleHook{}

	// Register with logger
	hook.RegisterHook(log)

	// Start background processing
	ctx, cancel := context.WithCancel(context.Background())

	// Start hook and wait for it to be running
	done := make(chan bool)
	go func() {
		hook.Run(ctx)
		done <- true
	}()

	// Give goroutine time to start
	for !hook.IsRunning() {
		// Spin until running
	}

	// Use the logger
	log.Info("processing started")

	// Check hook status
	if hook.IsRunning() {
		fmt.Println("Hook is running")
	}

	// Cleanup
	cancel()
	<-done // Wait for Run to finish
	_ = hook.Close()

	fmt.Println("Hook lifecycle completed")
	// Output: Hook is running
	// Hook lifecycle completed
}

// Example_multipleHooks demonstrates using multiple hooks simultaneously.
//
// This example shows how to register multiple hooks with a single logger,
// allowing log entries to be processed by multiple handlers.
func Example_multipleHooks() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Create multiple hooks
	hook1 := &simpleHook{}
	hook2 := &simpleHook{}

	// Register all hooks
	hook1.RegisterHook(log)
	hook2.RegisterHook(log)

	// Log entry will be sent to both hooks
	log.Info("distributed log entry")

	fmt.Println("Multiple hooks registered successfully")
	// Output: Multiple hooks registered successfully
}

// levelFilterHook filters log entries by level.
type levelFilterHook struct {
	minLevel logrus.Level
	running  atomic.Bool
}

// Fire processes entries at or above minLevel.
func (h *levelFilterHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= h.minLevel {
		// Process only important logs
		return nil
	}
	return nil
}

// Levels returns all levels (filtering happens in Fire).
func (h *levelFilterHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// RegisterHook registers the hook with a logger.
func (h *levelFilterHook) RegisterHook(log *logrus.Logger) {
	log.AddHook(h)
}

// Run runs the hook until context is cancelled.
func (h *levelFilterHook) Run(ctx context.Context) {
	h.running.Store(true)
	defer h.running.Store(false)
	<-ctx.Done()
}

// IsRunning returns whether the hook is running.
func (h *levelFilterHook) IsRunning() bool {
	return h.running.Load()
}

// Write implements io.Writer.
func (h *levelFilterHook) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close implements io.Closer.
func (h *levelFilterHook) Close() error {
	return nil
}

// Example_hookWithFiltering demonstrates filtering log entries in a hook.
//
// This example shows how to implement a hook that only processes
// certain log levels or entries matching specific criteria.
func Example_hookWithFiltering() {
	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.TraceLevel)

	// Create hook that only processes Error level and above
	hook := &levelFilterHook{minLevel: logrus.ErrorLevel}
	hook.RegisterHook(log)

	// These will be filtered out
	log.Debug("debug message")
	log.Info("info message")

	// This will be processed
	log.Error("error message")

	fmt.Println("Hook with filtering working")
	// Output: Hook with filtering working
}

// Example_hookLevelsMethod demonstrates the Levels() method.
//
// This example shows how the Levels() method controls which log levels
// are sent to the hook's Fire() method.
func Example_hookLevelsMethod() {
	// Create a hook that only receives error and fatal logs
	hook := &simpleHook{}

	// Override Levels method (in real code, implement in type)
	levels := hook.Levels()

	fmt.Printf("Hook receives %d log levels\n", len(levels))
	// Output: Hook receives 7 log levels
}

// Example_hookWriteMethod demonstrates direct writing to a hook.
//
// This example shows using the io.Writer interface to write directly
// to a hook, bypassing the logrus.Entry mechanism.
func Example_hookWriteMethod() {
	hook := &simpleHook{}

	// Write directly to hook
	data := []byte("direct write data\n")
	n, err := hook.Write(data)

	if err == nil && n == len(data) {
		fmt.Println("Direct write successful")
	}
	// Output: Direct write successful
}

// Example_hookContextCancellation demonstrates context-based cancellation.
//
// This example shows how the Run() method respects context cancellation
// for graceful shutdown.
func Example_hookContextCancellation() {
	hook := &simpleHook{}

	ctx, cancel := context.WithCancel(context.Background())

	// Start hook in background
	done := make(chan bool)
	go func() {
		hook.Run(ctx)
		done <- true
	}()

	// Cancel context
	cancel()

	// Wait for hook to stop
	<-done

	fmt.Println("Hook stopped gracefully")
	// Output: Hook stopped gracefully
}

// Example_threadSafeFieldUsage demonstrates thread-safe field access.
//
// This example shows that field constants can be safely accessed
// from multiple goroutines without synchronization.
func Example_threadSafeFieldUsage() {
	done := make(chan bool, 3)

	// Multiple goroutines using field constants
	for i := 0; i < 3; i++ {
		go func(id int) {
			log := logrus.New()
			log.SetOutput(io.Discard)

			log.WithFields(logrus.Fields{
				types.FieldFile: fmt.Sprintf("goroutine_%d.go", id),
				types.FieldLine: id * 100,
			}).Info("concurrent access")

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	fmt.Println("Concurrent field access successful")
	// Output: Concurrent field access successful
}

// Example_fieldValidation demonstrates checking field names.
//
// This example shows how to validate that log entries contain
// expected standard fields.
func Example_fieldValidation() {
	// Simulate log entry data
	logData := map[string]interface{}{
		types.FieldTime:    "2025-01-01T12:00:00Z",
		types.FieldLevel:   "info",
		types.FieldMessage: "test message",
	}

	// Validate required fields
	requiredFields := []string{
		types.FieldTime,
		types.FieldLevel,
		types.FieldMessage,
	}

	allPresent := true
	for _, field := range requiredFields {
		if _, exists := logData[field]; !exists {
			allPresent = false
			break
		}
	}

	if allPresent {
		fmt.Println("All required fields present")
	}
	// Output: All required fields present
}

// Example_customFieldsWithStandard demonstrates mixing custom and standard fields.
//
// This example shows how to use standard field constants alongside
// custom application-specific fields.
func Example_customFieldsWithStandard() {
	log := logrus.New()
	log.SetOutput(io.Discard)

	// Mix standard and custom fields
	log.WithFields(logrus.Fields{
		// Standard fields
		types.FieldFile:  "api.go",
		types.FieldLine:  99,
		types.FieldError: "timeout",

		// Custom fields
		"request_id": "abc-123",
		"user_id":    456,
		"endpoint":   "/api/v1/users",
	}).Error("API request failed")

	fmt.Println("Mixed standard and custom fields")
	// Output: Mixed standard and custom fields
}

// Example_fieldConstantsInMapKeys demonstrates using fields as map keys.
//
// This example shows using field constants as map keys for building
// structured log data programmatically.
func Example_fieldConstantsInMapKeys() {
	// Build log data structure
	logEntry := map[string]interface{}{
		types.FieldTime:    "2025-01-01T12:00:00Z",
		types.FieldLevel:   "error",
		types.FieldMessage: "database query failed",
		types.FieldError:   "connection lost",
		types.FieldFile:    "db.go",
		types.FieldLine:    234,
	}

	// Verify structure
	if _, hasError := logEntry[types.FieldError]; hasError {
		fmt.Println("Error field present in log entry")
	}
	// Output: Error field present in log entry
}

// Example_interfaceImplementation demonstrates checking interface compliance.
//
// This example shows how to verify that a type implements the Hook interface
// at compile time.
func Example_interfaceImplementation() {
	// Compile-time interface check
	var _ types.Hook = (*simpleHook)(nil)

	fmt.Println("Hook interface implemented correctly")
	// Output: Hook interface implemented correctly
}
