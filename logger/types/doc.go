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

// Package types provides core types, interfaces, and constants for the logger subsystem.
//
// # Overview
//
// This package defines the foundational types used across the github.com/nabbar/golib/logger ecosystem.
// It establishes standardized field names for structured logging and defines the Hook interface for
// extending logger functionality with custom output destinations and processors.
//
// # Design Philosophy
//
// The package follows these key principles:
//
//  1. Standardization: Provides consistent field names across all logger implementations
//  2. Extensibility: Hook interface allows seamless integration of custom log processors
//  3. Minimal Dependencies: Only depends on standard library and logrus
//  4. Type Safety: Strong typing for field constants prevents typos and ensures consistency
//  5. Compatibility: Full integration with logrus.Hook and io.WriteCloser interfaces
//
// # Package Architecture
//
// The package is organized into three main components:
//
//	┌─────────────────────────────────────────────────────────┐
//	│                    logger/types                         │
//	├──────────────────────┬──────────────────────────────────┤
//	│                      │                                  │
//	│   Field Constants    │         Hook Interface           │
//	│   (fields.go)        │         (hook.go)                │
//	│                      │                                  │
//	│  - FieldTime         │  Extends:                        │
//	│  - FieldLevel        │    • logrus.Hook                 │
//	│  - FieldStack        │    • io.WriteCloser              │
//	│  - FieldCaller       │                                  │
//	│  - FieldFile         │  Methods:                        │
//	│  - FieldLine         │    • RegisterHook(log)           │
//	│  - FieldMessage      │    • Run(ctx)                    │
//	│  - FieldError        │    • IsRunning()                 │
//	│  - FieldData         │    • Fire(entry)                 │
//	│                      │    • Levels()                    │
//	│                      │    • Write(p)                    │
//	│                      │    • Close()                     │
//	└──────────────────────┴──────────────────────────────────┘
//
// # Field Constants
//
// The package defines standard field names for structured logging to ensure consistency
// across different logger implementations and output formats (JSON, text, etc.).
//
// Field categories:
//
//   - Metadata fields: FieldTime, FieldLevel
//   - Trace fields: FieldStack, FieldCaller, FieldFile, FieldLine
//   - Content fields: FieldMessage, FieldError, FieldData
//
// Example usage in log entry:
//
//	{
//	  "time": "2025-01-01T12:00:00Z",      // FieldTime
//	  "level": "error",                     // FieldLevel
//	  "message": "operation failed",        // FieldMessage
//	  "error": "connection timeout",        // FieldError
//	  "file": "main.go",                    // FieldFile
//	  "line": 42,                           // FieldLine
//	  "caller": "main.processRequest",      // FieldCaller
//	  "stack": "goroutine 1 [running]...",  // FieldStack
//	  "data": {...}                         // FieldData
//	}
//
// # Hook Interface
//
// The Hook interface extends logrus.Hook with additional lifecycle management and I/O capabilities.
// This allows for sophisticated log processors that can:
//
//   - Intercept and process log entries before output
//   - Write to multiple destinations simultaneously
//   - Run in background goroutines with context-based cancellation
//   - Implement complex filtering and transformation logic
//   - Support graceful shutdown and resource cleanup
//
// Interface composition:
//
//	Hook embeds:
//	  • logrus.Hook      - Standard logrus hook integration
//	  • io.WriteCloser   - Direct write and close capabilities
//
//	Additional methods:
//	  • RegisterHook     - Self-registration with logger instance
//	  • Run             - Background execution with context control
//	  • IsRunning       - State checking for lifecycle management
//
// # Key Features
//
// Standardized Fields:
//   - Type-safe constants prevent typos in field names
//   - Consistent naming across all logger implementations
//   - Easy integration with structured logging frameworks
//   - Facilitates log parsing and analysis tools
//
// Hook Extensibility:
//   - Implement custom log processors without modifying core logger
//   - Support for multiple concurrent hooks
//   - Background processing with goroutines
//   - Context-based lifecycle management
//   - Clean shutdown via io.Closer interface
//
// Logrus Integration:
//   - Full compatibility with logrus.Hook interface
//   - Seamless integration into existing logrus-based applications
//   - Access to all logrus features (levels, formatting, fields)
//
// # Use Cases
//
// Structured Logging:
//
// Use field constants to ensure consistent field names across your application:
//
//	import "github.com/nabbar/golib/logger/types"
//
//	log.WithFields(logrus.Fields{
//	    types.FieldFile:  "handler.go",
//	    types.FieldLine:  123,
//	    types.FieldError: err.Error(),
//	}).Error("request processing failed")
//
// Custom Log Processors:
//
// Implement the Hook interface to create custom log handlers:
//
//	type EmailHook struct {
//	    smtpServer string
//	    running    atomic.Bool
//	}
//
//	func (h *EmailHook) Fire(entry *logrus.Entry) error {
//	    if entry.Level <= logrus.ErrorLevel {
//	        return h.sendEmail(entry)
//	    }
//	    return nil
//	}
//
//	func (h *EmailHook) Levels() []logrus.Level {
//	    return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel}
//	}
//
//	func (h *EmailHook) RegisterHook(log *logrus.Logger) {
//	    log.AddHook(h)
//	}
//
//	func (h *EmailHook) Run(ctx context.Context) {
//	    h.running.Store(true)
//	    defer h.running.Store(false)
//	    <-ctx.Done()
//	}
//
//	func (h *EmailHook) IsRunning() bool {
//	    return h.running.Load()
//	}
//
//	func (h *EmailHook) Write(p []byte) (n int, err error) {
//	    // Custom write logic
//	    return len(p), nil
//	}
//
//	func (h *EmailHook) Close() error {
//	    // Cleanup resources
//	    return nil
//	}
//
// Multi-Destination Logging:
//
// Use hooks to simultaneously log to multiple destinations:
//
//	logger := logrus.New()
//	fileHook := &FileHook{path: "/var/log/app.log"}
//	syslogHook := &SyslogHook{facility: "daemon"}
//	metricsHook := &MetricsHook{registry: prometheus.DefaultRegisterer}
//
//	fileHook.RegisterHook(logger)
//	syslogHook.RegisterHook(logger)
//	metricsHook.RegisterHook(logger)
//
// Log Filtering and Transformation:
//
// Implement hooks to filter or modify log entries:
//
//	type SensitiveDataFilter struct{}
//
//	func (f *SensitiveDataFilter) Fire(entry *logrus.Entry) error {
//	    if pwd, ok := entry.Data["password"]; ok {
//	        entry.Data["password"] = "***REDACTED***"
//	    }
//	    return nil
//	}
//
// # Performance Considerations
//
// Field Constants:
//   - Zero runtime overhead - constants are inlined at compile time
//   - No memory allocation for field name strings
//   - Compiler-optimized string comparisons
//
// Hook Interface:
//   - Fire() is called synchronously for every log entry
//   - Keep Fire() implementations fast to avoid blocking logging
//   - Use Run() goroutine for heavy processing (buffering, batching)
//   - Consider using channels to offload work from Fire() to Run()
//
// Recommended patterns:
//
//	// FAST: Simple filtering
//	func (h *FastHook) Fire(entry *logrus.Entry) error {
//	    if entry.Level > logrus.ErrorLevel {
//	        return nil // Skip non-error logs
//	    }
//	    h.queue <- entry // Send to background processor
//	    return nil
//	}
//
//	// SLOW: Avoid this pattern
//	func (h *SlowHook) Fire(entry *logrus.Entry) error {
//	    time.Sleep(100 * time.Millisecond) // Blocks logging!
//	    return h.sendToRemoteAPI(entry)    // Synchronous network call!
//	}
//
// # Limitations
//
// Field Constants:
//   - Field names are predefined and cannot be customized per-logger
//   - Adding new standard fields requires package modification
//   - No namespacing mechanism for field names
//   - Workaround: Use custom fields in logrus.Fields alongside standard fields
//
// Hook Interface:
//   - Fire() must return quickly to avoid blocking all logging
//   - No built-in buffering or batching mechanisms
//   - Context in Run() is not available in Fire() method
//   - Multiple hooks execute in registration order (no priority system)
//   - Error from Fire() is logged but doesn't stop other hooks
//
// Logrus Dependency:
//   - Tightly coupled to logrus.Logger and logrus.Hook interfaces
//   - Migration to other logging frameworks requires reimplementation
//   - Inherits logrus performance characteristics and limitations
//
// # Best Practices
//
// Using Field Constants:
//
//	// DO: Use constants for standard fields
//	log.WithField(types.FieldError, err.Error())
//
//	// DON'T: Hardcode field names
//	log.WithField("error", err.Error())
//
// Implementing Hooks:
//
//	// DO: Fast Fire() with background processing
//	func (h *Hook) Fire(entry *logrus.Entry) error {
//	    select {
//	    case h.queue <- entry:
//	        return nil
//	    default:
//	        return errors.New("queue full")
//	    }
//	}
//
//	// DON'T: Slow synchronous operations in Fire()
//	func (h *Hook) Fire(entry *logrus.Entry) error {
//	    return h.writeToDatabase(entry) // Blocks all logging!
//	}
//
// Context Management:
//
//	// DO: Use context for graceful shutdown
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	go hook.Run(ctx)
//
//	// DON'T: Forget to cancel context
//	go hook.Run(context.Background()) // Goroutine leak!
//
// Resource Cleanup:
//
//	// DO: Always close hooks
//	defer hook.Close()
//
//	// DO: Check IsRunning() before operations
//	if hook.IsRunning() {
//	    hook.Write(data)
//	}
//
// # Thread Safety
//
// Field Constants:
//   - Thread-safe: constants are immutable
//   - Safe for concurrent reads from multiple goroutines
//
// Hook Interface:
//   - Implementation-specific: hook implementations must handle their own synchronization
//   - Fire() may be called concurrently from multiple goroutines
//   - Run() typically executes in a single goroutine
//   - Use sync.Mutex or atomic operations to protect shared state
//
// # Integration with golib
//
// This package is used by:
//   - github.com/nabbar/golib/logger/config - Logger configuration
//   - github.com/nabbar/golib/logger/entry - Log entry management
//   - github.com/nabbar/golib/logger/fields - Field manipulation
//   - github.com/nabbar/golib/logger/gorm - GORM logger integration
//
// # External Dependencies
//
// Required:
//   - github.com/sirupsen/logrus - Structured logging library
//   - Standard library (context, io)
//
// No transitive dependencies beyond logrus.
//
// # Compatibility
//
// Minimum Go version: 1.18
//
// The package maintains semantic versioning and ensures backward compatibility
// within major versions. Field constant values are considered part of the public
// API and will not change within a major version.
//
// # Examples
//
// For comprehensive examples, see the example_test.go file.
package types
