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

package entry

import (
	"os"
	"strings"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// entry is the internal implementation of the Entry interface. It holds all state required
// for constructing and logging a structured log entry with logrus.
//
// The struct contains three main categories of information:
//  1. Configuration: logger function, gin context, message-only flag
//  2. Context: time, stack, caller, file, line, message
//  3. Data: custom fields, errors, arbitrary data
//
// This struct is not exported to maintain encapsulation and prevent direct field manipulation.
type entry struct {
	// log is a function that returns the logrus.Logger instance to use for logging.
	// This function-based approach allows for dynamic logger selection and lazy initialization.
	log func() *logrus.Logger

	// gin is an optional pointer to a Gin context. When set, errors are automatically
	// registered into the Gin context's error slice during logging.
	gin *ginsdk.Context

	// clean indicates whether to use message-only mode (true) or structured mode (false).
	// In message-only mode, only the message is logged without any fields or context.
	clean bool

	// Time is the timestamp of the log event. Can be zero if timestamps are disabled.
	Time time.Time `json:"time"`

	// Level defines the log level of the entry (Debug, Info, Warn, Error, Fatal, Panic, Nil).
	// This field is required and determines whether the entry is logged.
	Level loglvl.Level `json:"level"`

	// Stack is the goroutine ID or stack number. Can be 0 if not provided.
	Stack uint64 `json:"stack"`

	// Caller is the name of the calling function. Can be empty if trace is disabled,
	// not found, or the function is anonymous.
	Caller string `json:"caller"`

	// File is the source file path of the caller. Can be empty if trace is disabled,
	// not found, or anonymous.
	File string `json:"file"`

	// Line is the line number in the source file. Can be 0 if trace is disabled,
	// not found, or anonymous.
	Line uint64 `json:"line"`

	// Message is the main log message. Can be empty.
	Message string `json:"message"`

	// Error is a slice of error values. Can be nil, empty, or contain nil values.
	// Multiple errors can be logged together, and they are joined with commas.
	Error []error `json:"error"`

	// Data is arbitrary data to attach to the log entry. Can be any type that
	// can be serialized to JSON. This field is optional.
	Data interface{} `json:"data"`

	// Fields are custom structured fields to add to the log entry. This must be
	// initialized with FieldSet() before using field methods. Can be nil initially.
	Fields logfld.Fields `json:"fields"`
}

// SetEntryContext sets all context information for the log entry in a single method call.
// This includes timestamp, stack trace, caller information, file/line numbers, and the message.
//
// This method is the primary way to set contextual information and should be called before Log().
//
// Parameters:
//   - etime: Timestamp of the log event (use time.Now() or time.Time{} to disable)
//   - stack: Goroutine ID or stack number (use 0 to disable)
//   - caller: Function name of the caller (use "" if unknown or disabled)
//   - file: Source file path (use "" if unknown or disabled)
//   - line: Line number in source file (use 0 if unknown or disabled)
//   - msg: The main log message
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	e := New(loglvl.InfoLevel).SetEntryContext(
//	    time.Now(), 12345, "HandleRequest", "handler.go", 42, "Request processed")
func (e *entry) SetEntryContext(etime time.Time, stack uint64, caller, file string, line uint64, msg string) Entry {
	if e == nil {
		return nil
	}

	e.Time = etime
	e.Stack = stack
	e.Caller = caller
	e.File = file
	e.Line = line
	e.Message = msg

	return e
}

// SetMessageOnly controls whether to use message-only logging mode. When enabled (true),
// only the message text is logged using logrus Info level, ignoring all structured fields,
// context information, and errors. When disabled (false), normal structured logging is used.
//
// This mode is useful for simple console output or when you don't need structured logging.
//
// Parameters:
//   - flag: true to enable message-only mode, false for normal structured logging
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	e := New(loglvl.InfoLevel).SetMessageOnly(true)
//	e.SetEntryContext(time.Now(), 0, "", "", 0, "Simple message").Log()
func (e *entry) SetMessageOnly(flag bool) Entry {
	if e == nil {
		return nil
	}

	e.clean = flag
	return e
}

// SetLevel changes the log level of the entry. The level determines whether and how the entry
// is logged. If the logger's level is lower than the entry's level, the entry will not be logged.
//
// Special levels:
//   - NilLevel: Entry is never logged
//   - FatalLevel: Triggers os.Exit(1) after logging
//   - PanicLevel: Triggers panic after logging (logrus behavior)
//
// Parameters:
//   - lvl: The new log level (Debug, Info, Warn, Error, Fatal, Panic, Nil)
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	e := New(loglvl.InfoLevel).SetLevel(loglvl.ErrorLevel)
func (e *entry) SetLevel(lvl loglvl.Level) Entry {
	if e == nil {
		return nil
	}

	e.Level = lvl

	return e
}

// SetLogger sets the logger function that provides the logrus.Logger instance for logging.
// The function is called when Log() or Check() is invoked, allowing for lazy initialization
// or dynamic logger selection.
//
// If the logger function is nil or returns nil, the entry will not be logged.
//
// Parameters:
//   - fct: Function returning a pointer to logrus.Logger, or nil to disable logging
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	logger := logrus.New()
//	e := New(loglvl.InfoLevel).SetLogger(func() *logrus.Logger { return logger })
func (e *entry) SetLogger(fct func() *logrus.Logger) Entry {
	if e == nil {
		return nil
	}

	e.log = fct
	return e
}

// SetGinContext sets a Gin context pointer to enable automatic error registration. When a Gin
// context is set and the entry contains errors, those errors are automatically registered into
// the Gin context's error slice during logging.
//
// This integration is useful for HTTP error handling and middleware where you want errors to
// be both logged and accessible in the Gin context for response formatting.
//
// Parameters:
//   - ctx: Pointer to gin.Context, or nil to disable Gin integration
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	func handler(c *gin.Context) {
//	    e := New(loglvl.ErrorLevel).SetGinContext(c)
//	    e.ErrorAdd(true, err).Log() // Error is logged and added to c.Errors
//	}
func (e *entry) SetGinContext(ctx *ginsdk.Context) Entry {
	if e == nil {
		return nil
	}

	e.gin = ctx
	return e
}

// DataSet attaches arbitrary data to the log entry. The data can be any type that can be
// serialized to JSON by logrus. This is useful for including structured data like request
// payloads, response data, or any complex objects.
//
// The data is logged as a separate field and does not interfere with custom fields or context.
//
// Parameters:
//   - data: Any JSON-serializable data, or nil to clear data
//
// Returns:
//   - The entry itself for method chaining, or nil if entry is nil
//
// Example:
//
//	data := map[string]interface{}{"user_id": 123, "action": "login"}
//	e := New(loglvl.InfoLevel).DataSet(data)
func (e *entry) DataSet(data interface{}) Entry {
	if e == nil {
		return nil
	}

	e.Data = data
	return e
}

// Check determines whether the entry contains any non-nil errors and logs the entry accordingly.
// If errors are found, the entry is logged at its current level. If no errors are found,
// the level is changed to lvlNoErr before logging.
//
// This method is useful for conditional logging where you want to log at different levels
// based on whether errors occurred.
//
// Parameters:
//   - lvlNoErr: The log level to use if no errors are present
//
// Returns:
//   - true if the entry contains at least one non-nil error
//   - false if the entry has no errors or only nil errors
//
// Example:
//
//	e := New(loglvl.ErrorLevel).ErrorAdd(true, err)
//	if e.Check(loglvl.InfoLevel) {
//	    // Logged at ErrorLevel with errors
//	} else {
//	    // Logged at InfoLevel without errors
//	}
func (e *entry) Check(lvlNoErr loglvl.Level) bool {
	if e == nil {
		return false
	}

	var found = false
	if len(e.Error) > 0 {
		for _, er := range e.Error {
			if er == nil {
				continue
			}

			found = true
			break
		}
	}

	if !found {
		e.Level = lvlNoErr
	}

	e.Log()
	return found
}

// Log performs the actual logging operation by constructing a logrus entry with all configured
// context, fields, errors, and data, then logging it at the specified level.
//
// Behavior:
//   - Returns early if entry, logger, or fields are nil
//   - Returns early if fields have an error
//   - In message-only mode (clean=true), logs only the message using Info level
//   - Registers errors in Gin context if configured
//   - Does not log if level is NilLevel
//   - Calls os.Exit(1) after logging if level is FatalLevel
//
// Guard Conditions:
//   - Entry must not be nil
//   - Logger function must be set and return non-nil logger
//   - Fields must be set and have no error
//   - Level must not be NilLevel
//
// Example:
//
//	logger := logrus.New()
//	fields := logfld.New(nil)
//	New(loglvl.InfoLevel).
//	    SetLogger(func() *logrus.Logger { return logger }).
//	    FieldSet(fields).
//	    SetEntryContext(time.Now(), 0, "", "", 0, "Hello").
//	    Log()
func (e *entry) Log() {
	if e == nil {
		return
	} else if e.log == nil {
		return
	} else if e.Fields == nil {
		return
	} else if e.Fields.Err() != nil {
		return
	} else if e.clean {
		e._logClean()
		return
	} else if e.gin != nil && len(e.Error) > 0 {
		for _, err := range e.Error {
			_ = e.gin.Error(err)
		}
	}

	if e.Level == loglvl.NilLevel {
		return
	}

	var (
		ent *logrus.Entry
		tag = logfld.New(e.Fields).Add(logtps.FieldLevel, e.Level.String())
		log *logrus.Logger
	)

	if !e.Time.IsZero() {
		tag = tag.Add(logtps.FieldTime, e.Time.Format(time.RFC3339Nano))
	}

	if e.Stack > 0 {
		tag = tag.Add(logtps.FieldStack, e.Stack)
	}

	if e.Caller != "" {
		tag = tag.Add(logtps.FieldCaller, e.Caller)
	} else if e.File != "" {
		tag = tag.Add(logtps.FieldFile, e.File)
	}

	if e.Line > 0 {
		tag = tag.Add(logtps.FieldLine, e.Line)
	}

	if e.Message != "" {
		tag = tag.Add(logtps.FieldMessage, e.Message)
	}

	if len(e.Error) > 0 {
		var msg = make([]string, 0)

		for _, er := range e.Error {
			if er == nil {
				continue
			}
			msg = append(msg, er.Error())
		}

		tag = tag.Add(logtps.FieldError, strings.Join(msg, ", "))
	}

	if e.Data != nil {
		tag = tag.Add(logtps.FieldData, e.Data)
	}

	tag.Merge(e.Fields)

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		ent = log.WithFields(tag.Logrus())
	}

	ent.Log(e.Level.Logrus())

	if e.Level <= loglvl.FatalLevel {
		os.Exit(1)
	}
}

// _logClean is an internal method that performs message-only logging when clean mode is enabled.
// It logs only the message text using logrus Info level, ignoring all other fields and context.
//
// This method is called by Log() when SetMessageOnly(true) has been invoked.
//
// Guard Conditions:
//   - Logger function must be set and return non-nil logger
//   - Message must be set via SetEntryContext
func (e *entry) _logClean() {
	var (
		log *logrus.Logger
	)

	if e.log == nil {
		return
	} else if log = e.log(); log == nil {
		return
	} else {
		//log.SetLevel(logrus.InfoLevel)
		log.Info(e.Message)
	}
}
