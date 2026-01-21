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

package logger

import (
	"fmt"
	"time"

	logent "github.com/nabbar/golib/logger/entry"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Debug logs a message at Debug level with optional data and formatted arguments.
// Debug messages are for detailed diagnostic information useful during development.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Debug("Processing request for user %s", userData, "john.doe")
func (o *logger) Debug(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.DebugLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// Info logs a message at Info level with optional data and formatted arguments.
// Info messages are for general informational messages about normal operations.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Info("Application started on port %d", nil, 8080)
func (o *logger) Info(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.InfoLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// Warning logs a message at Warning level with optional data and formatted arguments.
// Warning messages indicate potentially harmful situations that should be reviewed.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Warning("Slow query detected: %dms", nil, queryTime)
func (o *logger) Warning(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.WarnLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// Error logs a message at Error level with optional data and formatted arguments.
// Error messages indicate failures that prevent normal operation but don't crash the application.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Error("Failed to process request", err, requestID)
func (o *logger) Error(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.ErrorLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// Fatal logs a message at Fatal level and terminates the application with os.Exit(1).
// Use Fatal only for critical errors that prevent the application from continuing.
//
// WARNING: This method calls os.Exit(1) after logging. Deferred functions will NOT run.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Fatal("Cannot connect to database", nil)
func (o *logger) Fatal(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.FatalLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// Panic logs a message at Panic level and calls panic() with the message.
// Use Panic for errors that should trigger panic recovery mechanisms.
//
// WARNING: This method triggers a panic after logging. Use recover() to catch it.
//
// Parameters:
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data (map, struct, etc.) to include in the log
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	logger.Panic("Invalid state detected", stateData)
func (o *logger) Panic(message string, data interface{}, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(loglvl.PanicLevel, fmt.Sprintf(message, args...), nil, nil, data).Log()
}

// LogDetails logs a message with complete control over all entry parameters.
// This is the most flexible logging method, allowing you to specify level, message,
// data, errors, and custom fields all at once.
//
// Parameters:
//   - lvl: Log level for this entry
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - data: Optional structured data to include in the log
//   - err: Optional slice of errors to include in the log
//   - fields: Optional additional fields to merge with default fields
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Example:
//
//	errs := []error{err1, err2}
//	fields := logfld.New(ctx).Add("request_id", "123")
//	logger.LogDetails(loglvl.ErrorLevel, "Request failed", nil, errs, fields)
func (o *logger) LogDetails(lvl loglvl.Level, message string, data interface{}, err []error, fields logfld.Fields, args ...interface{}) {
	if o == nil {
		return
	}

	o.newEntry(lvl, fmt.Sprintf(message, args...), err, fields, data).Log()
}

// CheckError logs an error if any error is provided, otherwise optionally logs success.
// This is a convenience method for error checking patterns.
//
// Parameters:
//   - lvlKO: Log level to use if errors are present
//   - lvlOK: Log level to use if no errors (use NilLevel to skip logging success)
//   - message: Message to log
//   - err: Optional errors to check
//
// Returns:
//   - true if any error was provided (and logged), false otherwise
//
// Example:
//
//	if logger.CheckError(loglvl.ErrorLevel, loglvl.InfoLevel, "Operation completed", err) {
//	    return // Error was logged
//	}
//	// Success was logged at InfoLevel
func (o *logger) CheckError(lvlKO, lvlOK loglvl.Level, message string, err ...error) bool {
	if o == nil {
		return false
	}

	ent := o.newEntry(lvlKO, message, err, nil, nil)
	return ent.Check(lvlOK)
}

// Entry creates a log entry that can be customized before logging.
// This method returns an Entry interface allowing you to add fields, set context,
// and control when the entry is actually logged.
//
// Parameters:
//   - lvl: Log level for this entry
//   - message: Format string for the log message (supports fmt.Sprintf formatting)
//   - args: Optional arguments for message formatting (used with fmt.Sprintf)
//
// Returns:
//   - logent.Entry: An entry that can be customized with fields before calling Log()
//
// Example:
//
//	entry := logger.Entry(loglvl.InfoLevel, "User action: %s", action)
//	entry.FieldAdd("user_id", userID)
//	entry.FieldAdd("timestamp", time.Now())
//	entry.Log()
func (o *logger) Entry(lvl loglvl.Level, message string, args ...interface{}) logent.Entry {
	return o.newEntry(lvl, fmt.Sprintf(message, args...), nil, nil, nil)
}

// Access creates an HTTP access log entry in standard format.
// This method generates a log entry following the Common Log Format (CLF) extended with latency.
// The entry is created at Info level with message-only mode (no stack traces or extra fields).
//
// Parameters:
//   - remoteAddr: Client IP address
//   - remoteUser: Authenticated user name (use "" for anonymous)
//   - localtime: Request timestamp
//   - latency: Request processing duration
//   - method: HTTP method (GET, POST, etc.)
//   - request: Request path
//   - proto: HTTP protocol version (HTTP/1.1, HTTP/2, etc.)
//   - status: HTTP status code
//   - size: Response size in bytes
//
// Returns:
//   - logent.Entry: A clean entry formatted for access logging
//
// Example:
//
//	logger.Access(
//	    "192.168.1.1", "john.doe",
//	    time.Now(), 150*time.Millisecond,
//	    "GET", "/api/users", "HTTP/1.1",
//	    200, 1024,
//	).Log()
func (o *logger) Access(remoteAddr, remoteUser string, localtime time.Time, latency time.Duration, method, request, proto string, status int, size int64) logent.Entry {
	var msg = fmt.Sprintf("%s - %s [%s] [%s] \"%s %s %s\" %d %d", remoteAddr, remoteUser, localtime.Format(time.RFC1123Z), latency.String(), method, request, proto, status, size)
	return o.newEntryClean(msg)
}

func (o *logger) newEntry(lvl loglvl.Level, message string, err []error, fields logfld.Fields, data interface{}) logent.Entry {
	if o == nil {
		return logent.New(loglvl.NilLevel)
	}

	var (
		fct = o.getLogrus
		ent = logent.New(lvl)
		frm = o.getCaller()
		stk = o.getStack()
		fld = o.GetFields()
	)

	// prevent overflow
	var uif uint64
	if frm.Line <= 0 {
		uif = 0
	} else {
		uif = uint64(frm.Line)
	}

	ent.ErrorSet(err)
	ent.DataSet(data)
	ent.SetEntryContext(time.Now(), stk, frm.Function, frm.File, uif, message)

	if o.x.Err() != nil {
		return ent
	}

	if fld != nil {
		ent.FieldSet(fld.Clone())
	}

	ent.SetLogger(fct)
	ent.FieldMerge(fields)

	return ent
}

func (o *logger) newEntryClean(message string) logent.Entry {
	if o == nil {
		return logent.New(loglvl.NilLevel)
	}

	return o.newEntry(loglvl.InfoLevel, message, nil, nil, nil).SetMessageOnly(true)
}
