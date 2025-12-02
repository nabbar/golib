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

// Package entry provides a flexible, chainable logger entry wrapper for structured logging with logrus.
//
// # Overview
//
// The entry package wraps logrus entries to provide a fluent, chainable API for constructing structured
// log entries with context information, custom fields, errors, and arbitrary data. It supports integration
// with Gin web framework for automatic error registration.
//
// # Design Philosophy
//
// The package follows these core principles:
//
//  1. Immutability Pattern: All setter methods return the entry itself, enabling method chaining while
//     maintaining a fluent API design.
//
//  2. Lazy Evaluation: The actual logging to logrus is deferred until the Log() or Check() method is called,
//     allowing entries to be built incrementally.
//
//  3. Flexible Context: Entries can include timestamps, stack traces, caller information, file/line numbers,
//     custom fields, arbitrary data, and multiple errors.
//
//  4. Safety First: All methods handle nil entries gracefully, returning nil or appropriate defaults rather
//     than panicking.
//
// # Key Features
//
//   - Fluent API: Chain methods for concise entry construction
//   - Multiple log levels: Debug, Info, Warn, Error, Fatal, Panic, and Nil
//   - Rich context: Time, stack, caller, file, line, and message
//   - Custom fields: Add structured key-value pairs
//   - Error handling: Multiple errors with nil filtering
//   - Data attachment: Attach arbitrary data structures
//   - Gin integration: Automatic error registration in Gin context
//   - Message-only mode: Simple logging without structured fields
//
// # Architecture
//
// The package consists of the following components:
//
//	┌──────────────────────────────────────┐
//	│           Entry Interface            │
//	│  - Configuration methods             │
//	│  - Field management                  │
//	│  - Error management                  │
//	│  - Logging methods                   │
//	└──────────────┬───────────────────────┘
//	               │
//	               ▼
//	┌──────────────────────────────────────┐
//	│         entry Implementation         │
//	│                                      │
//	│  ┌────────────────────────────────┐  │
//	│  │  Configuration State           │  │
//	│  │  - Logger function             │  │
//	│  │  - Gin context pointer         │  │
//	│  │  - Message-only flag           │  │
//	│  └────────────────────────────────┘  │
//	│               │                      │
//	│               ▼                      │
//	│  ┌────────────────────────────────┐  │
//	│  │  Context Information           │  │
//	│  │  - Time, Stack, Caller         │  │
//	│  │  - File, Line, Message         │  │
//	│  └────────────────────────────────┘  │
//	│               │                      │
//	│               ▼                      │
//	│  ┌────────────────────────────────┐  │
//	│  │  Data & Fields                 │  │
//	│  │  - Custom fields               │  │
//	│  │  - Errors slice                │  │
//	│  │  - Arbitrary data              │  │
//	│  └────────────────────────────────┘  │
//	│               │                      │
//	│               ▼                      │
//	│       Log to logrus                  │
//	└──────────────────────────────────────┘
//
// # Entry Lifecycle
//
// A typical entry follows this lifecycle:
//
//  1. Creation: New(level) creates an entry with initial state
//  2. Configuration: Set logger, level, gin context, message mode
//  3. Context: Set time, stack, caller, file, line, message
//  4. Fields: Add, merge, or set custom structured fields
//  5. Errors: Add or set error information
//  6. Data: Attach arbitrary data structures
//  7. Logging: Call Log() or Check() to output to logrus
//
// # Logging Behavior
//
// The Log() method behavior depends on the entry configuration:
//
// Normal Mode (clean=false):
//   - Includes all context fields (time, stack, caller, file, line)
//   - Includes custom fields and data
//   - Formats errors as comma-separated string
//   - Respects log level filtering
//   - Registers errors in Gin context if configured
//
// Message-Only Mode (clean=true):
//   - Outputs only the message text
//   - Ignores all context fields and custom fields
//   - Uses simple Info level logging
//   - Suitable for console output or simple logs
//
// Special Cases:
//   - NilLevel entries are never logged
//   - Nil logger or nil fields prevent logging
//   - FatalLevel triggers os.Exit(1) after logging
//
// # Level Hierarchy
//
// Log levels from most to least verbose:
//   - PanicLevel: System panic situations
//   - FatalLevel: Fatal errors (triggers exit)
//   - ErrorLevel: Error conditions
//   - WarnLevel: Warning conditions
//   - InfoLevel: Informational messages
//   - DebugLevel: Debug information
//   - TraceLevel: Detailed trace information
//   - NilLevel: Disabled logging
//
// # Field Management
//
// Fields are managed through the logger/fields package and provide:
//   - Type-safe key-value storage
//   - Merge operations for combining field sets
//   - Deletion of specific keys
//   - Integration with logrus.Fields
//
// Fields must be initialized with FieldSet() before using FieldAdd(), FieldMerge(), or FieldClean().
//
// # Error Handling
//
// The package supports multiple error patterns:
//
// Standard Errors:
//   - Add individual errors with ErrorAdd()
//   - Set entire error slice with ErrorSet()
//   - Clear errors with ErrorClean()
//
// Error Filtering:
//   - ErrorAdd(cleanNil=true, ...) filters out nil errors
//   - ErrorAdd(cleanNil=false, ...) includes nil errors
//
// Wrapped Errors:
//   - Automatically unwraps github.com/nabbar/golib/errors
//   - Extracts error slices from wrapped errors
//   - Supports fmt.Errorf %w wrapping
//
// # Integration Points
//
// Logrus Integration:
//   - Requires logger function returning *logrus.Logger
//   - Uses logrus.Entry for actual logging
//   - Supports all logrus formatters and hooks
//
// Gin Integration:
//   - SetGinContext() enables automatic error registration
//   - Errors are added to Gin context error slice
//   - Useful for HTTP error handling and middleware
//
// # Performance Considerations
//
// Memory:
//   - Entry struct is lightweight (~300 bytes base)
//   - Fields and errors use slice allocation
//   - No memory pooling (entries are short-lived)
//
// CPU:
//   - Method chaining has minimal overhead
//   - Field operations delegate to logger/fields package
//   - Log() method performs string building and formatting
//
// Concurrency:
//   - Entries are NOT thread-safe
//   - Create separate entry per goroutine
//   - Logger function should return thread-safe logger
//
// # Use Cases
//
// Simple Logging:
//
//	entry.New(loglvl.InfoLevel).
//	    SetLogger(loggerFunc).
//	    FieldSet(fields).
//	    SetEntryContext(time.Now(), 0, "", "", 0, "Hello").
//	    Log()
//
// Error Logging:
//
//	entry.New(loglvl.ErrorLevel).
//	    SetLogger(loggerFunc).
//	    FieldSet(fields).
//	    ErrorAdd(true, err1, err2).
//	    SetEntryContext(time.Now(), 0, "func", "file.go", 42, "Failed").
//	    Log()
//
// Structured Logging with Data:
//
//	entry.New(loglvl.InfoLevel).
//	    SetLogger(loggerFunc).
//	    FieldSet(fields).
//	    FieldAdd("user_id", userID).
//	    DataSet(requestData).
//	    SetEntryContext(time.Now(), 0, "", "", 0, "Request processed").
//	    Log()
//
// Gin Error Registration:
//
//	e := entry.New(loglvl.ErrorLevel).
//	    SetLogger(loggerFunc).
//	    SetGinContext(c).
//	    FieldSet(fields).
//	    ErrorAdd(true, dbError).
//	    SetEntryContext(time.Now(), 0, "", "", 0, "Database failed")
//	e.Log() // Registers error in c.Errors
//
// Conditional Logging:
//
//	e := entry.New(loglvl.ErrorLevel).
//	    SetLogger(loggerFunc).
//	    FieldSet(fields).
//	    ErrorAdd(true, err)
//	if e.Check(loglvl.InfoLevel) {
//	    // Has errors, logged at ErrorLevel
//	} else {
//	    // No errors, logged at InfoLevel
//	}
//
// # Best Practices
//
// DO:
//   - Always call FieldSet() before FieldAdd/FieldMerge/FieldClean
//   - Use method chaining for concise entry construction
//   - Filter nil errors with cleanNil=true in production
//   - Create new entry per log statement
//   - Set logger function that returns valid logger
//
// DON'T:
//   - Share entries across goroutines
//   - Call FieldAdd/FieldMerge/FieldClean without FieldSet
//   - Ignore FatalLevel effects (os.Exit)
//   - Mutate fields/errors after logging
//   - Use PanicLevel in production code
//
// # Limitations
//
//   - No automatic caller detection (must provide manually)
//   - No log buffering or batching
//   - No built-in sampling or rate limiting
//   - Fields must be initialized before use
//   - Not thread-safe (by design)
//
// # Dependencies
//
// Standard Library:
//   - os: For exit on fatal level
//   - strings: For error message joining
//   - time: For timestamp handling
//
// External:
//   - github.com/sirupsen/logrus: Core logging engine
//   - github.com/gin-gonic/gin: Gin framework integration
//   - github.com/nabbar/golib/logger/fields: Field management
//   - github.com/nabbar/golib/logger/level: Log level definitions
//   - github.com/nabbar/golib/logger/types: Type constants
//   - github.com/nabbar/golib/errors: Error utilities
//
// # Package Status
//
// This package is production-ready and stable. It is widely used in nabbar/golib ecosystem
// for structured logging across multiple packages.
package entry
