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

package types

// Standard field name constants for structured logging.
//
// These constants define the canonical field names used across the logger subsystem
// to ensure consistency in structured log output. They are used as keys in logrus.Fields
// maps and appear in formatted log output (JSON, text, etc.).
//
// The constants are organized into three logical categories:
//
// Metadata fields: Contain information about the log entry itself
//   - FieldTime: Timestamp when the log entry was created
//   - FieldLevel: Severity level (debug, info, warn, error, fatal)
//
// Trace fields: Provide execution context and debugging information
//   - FieldStack: Full stack trace (usually for errors)
//   - FieldCaller: Function or method name that generated the log
//   - FieldFile: Source code file name
//   - FieldLine: Line number in source code file
//
// Content fields: Carry the actual log message and associated data
//   - FieldMessage: Primary log message text
//   - FieldError: Error message or description
//   - FieldData: Additional structured data (maps, objects, etc.)
//
// Example usage:
//
//	log.WithFields(logrus.Fields{
//	    types.FieldFile:  "handler.go",
//	    types.FieldLine:  123,
//	    types.FieldError: err.Error(),
//	}).Error("request failed")
const (
	// FieldTime is the field name for log entry timestamp.
	// Typically formatted as RFC3339: "2025-01-01T12:00:00Z"
	FieldTime = "time"

	// FieldLevel is the field name for log severity level.
	// Common values: "debug", "info", "warn", "error", "fatal", "panic"
	FieldLevel = "level"

	// FieldStack is the field name for full stack trace.
	// Contains multi-line stack trace starting from the point of logging.
	// Usually included only for error and fatal level logs.
	FieldStack = "stack"

	// FieldCaller is the field name for calling function identifier.
	// Typically formatted as "package.function" or "package.Type.method".
	// Example: "main.processRequest" or "server.Handler.ServeHTTP"
	FieldCaller = "caller"

	// FieldFile is the field name for source code file.
	// Contains the file name (not full path) where the log was generated.
	// Example: "handler.go", "server.go"
	FieldFile = "file"

	// FieldLine is the field name for source code line number.
	// Contains the line number within FieldFile where the log was generated.
	// Type: integer. Example: 42, 123
	FieldLine = "line"

	// FieldMessage is the field name for primary log message.
	// Contains the main descriptive text of the log entry.
	// This is the human-readable description of what happened.
	FieldMessage = "message"

	// FieldError is the field name for error description.
	// Contains the error message or error string when an error occurred.
	// Typically populated from err.Error() or error descriptions.
	FieldError = "error"

	// FieldData is the field name for additional structured data.
	// Contains any extra contextual information as structured data.
	// Can hold maps, slices, or any JSON-serializable data structure.
	FieldData = "data"
)
