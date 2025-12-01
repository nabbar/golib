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

package hashicorp

import (
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

const (
	// HCLogArgs is the field key used to store hclog With() arguments in golib fields.
	// The value is a []interface{} containing key-value pairs passed to With().
	HCLogArgs = "hclog.args"

	// HCLogName is the field key used to store the logger name from Named() in golib fields.
	// The value is a string representing the logger's hierarchical name.
	HCLogName = "hclog.name"
)

// _hclog implements hclog.Logger interface, bridging to golib logger.
// It stores a function that returns a golib logger, allowing dynamic logger switching.
type _hclog struct {
	l liblog.FuncLog // logger factory function
}

// Log logs a message at the specified hclog level with optional key-value arguments.
// NoLevel and Off levels are ignored (no log output).
// All other levels are mapped to golib equivalents and logged.
func (o *_hclog) Log(level hclog.Level, msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	switch level {
	case hclog.NoLevel, hclog.Off:
		return
	case hclog.Trace:
		lg.Entry(loglvl.DebugLevel, msg, args...).Log()
	case hclog.Debug:
		lg.Entry(loglvl.DebugLevel, msg, args...).Log()
	case hclog.Info:
		lg.Entry(loglvl.InfoLevel, msg, args...).Log()
	case hclog.Warn:
		lg.Entry(loglvl.WarnLevel, msg, args...).Log()
	case hclog.Error:
		lg.Entry(loglvl.ErrorLevel, msg, args...).Log()
	}
}

// logger retrieves the current logger instance from the factory function.
// Returns nil if the factory function is nil or returns nil.
// This method is called on every log operation to support dynamic logger replacement.
func (o *_hclog) logger() liblog.Logger {
	if o.l == nil {
		return nil
	} else if lg := o.l(); lg == nil {
		return nil
	} else {
		return lg
	}
}

// Trace logs a trace-level message with optional key-value arguments.
// Mapped to golib DebugLevel since golib doesn't have a native Trace level.
func (o *_hclog) Trace(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.DebugLevel, msg, args...).Log()
}

// Debug logs a debug-level message with optional key-value arguments.
func (o *_hclog) Debug(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.DebugLevel, msg, args...).Log()
}

// Info logs an info-level message with optional key-value arguments.
func (o *_hclog) Info(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.InfoLevel, msg, args...).Log()
}

// Warn logs a warning-level message with optional key-value arguments.
func (o *_hclog) Warn(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.WarnLevel, msg, args...).Log()
}

// Error logs an error-level message with optional key-value arguments.
func (o *_hclog) Error(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.ErrorLevel, msg, args...).Log()
}

// IsTrace returns true if trace-level logging is enabled in any output sink.
// Checks EnableTrace flag in Stdout, LogFile, and LogSyslog configurations.
func (o *_hclog) IsTrace() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	if opt := lg.GetOptions(); opt == nil {
		return false
	} else if opt.Stdout != nil && opt.Stdout.EnableTrace {
		return true
	} else {
		for _, f := range opt.LogFile {
			if f.EnableTrace {
				return true
			}
		}
		for _, f := range opt.LogSyslog {
			if f.EnableTrace {
				return true
			}
		}
	}

	return false
}

// IsDebug returns true if the current log level is DebugLevel or more verbose.
func (o *_hclog) IsDebug() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.DebugLevel
}

// IsInfo returns true if the current log level is InfoLevel or more verbose.
func (o *_hclog) IsInfo() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.InfoLevel
}

// IsWarn returns true if the current log level is WarnLevel or more verbose.
func (o *_hclog) IsWarn() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.WarnLevel
}

// IsError returns true if the current log level is ErrorLevel or more verbose.
func (o *_hclog) IsError() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.ErrorLevel
}

// ImpliedArgs returns the context arguments added via With() calls.
// Arguments are stored in golib fields under the HCLogArgs key.
// Returns an empty slice if no arguments have been set or if the stored value is not []interface{}.
func (o *_hclog) ImpliedArgs() []interface{} {
	var lg = o.logger()

	if lg == nil {
		return make([]interface{}, 0)
	}

	fields := lg.GetFields()

	if i, l := fields.Get(HCLogArgs); !l {
		return make([]interface{}, 0)
	} else if v, k := i.([]interface{}); k {
		return v
	}

	return make([]interface{}, 0)
}

// With creates a logger with additional context arguments.
// The arguments are stored in golib fields and retrieved via ImpliedArgs().
// Returns self to support method chaining.
func (o *_hclog) With(args ...interface{}) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogArgs, args))
	return o
}

// Name returns the logger's name set via Named() or ResetNamed().
// The name is stored in golib fields under the HCLogName key.
// Returns empty string if no name is set or if the stored value is not a string.
func (o *_hclog) Name() string {
	var lg = o.logger()

	if lg == nil {
		return ""
	}

	fields := lg.GetFields()

	if i, l := fields.Get(HCLogName); !l {
		return ""
	} else if v, k := i.(string); k {
		return v
	}

	return ""
}

// Named creates a named sub-logger.
// The name is stored in golib fields and retrieved via Name().
// Returns self to support method chaining.
func (o *_hclog) Named(name string) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogName, name))
	return o
}

// ResetNamed resets the logger's name to the specified value.
// Functions identically to Named() in this implementation.
// Returns self to support method chaining.
func (o *_hclog) ResetNamed(name string) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogName, name))
	return o
}

// SetLevel sets the logger's log level.
// NoLevel and Off map to NilLevel (logging disabled).
// Trace enables EnableTrace flags and sets level to DebugLevel.
// Other levels map directly to golib equivalents.
func (o *_hclog) SetLevel(level hclog.Level) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	switch level {
	case hclog.NoLevel, hclog.Off:
		lg.SetLevel(loglvl.NilLevel)
	case hclog.Trace:
		opt := lg.GetOptions()

		if opt.Stdout != nil {
			opt.Stdout.EnableTrace = true
		}

		if len(opt.LogFile) > 0 {
			for i := range opt.LogFile {
				opt.LogFile[i].EnableTrace = true
			}
		}

		if len(opt.LogSyslog) > 0 {
			for i := range opt.LogSyslog {
				opt.LogSyslog[i].EnableTrace = true
			}
		}

		_ = lg.SetOptions(opt)
		lg.SetLevel(loglvl.DebugLevel)
	case hclog.Debug:
		lg.SetLevel(loglvl.DebugLevel)
	case hclog.Info:
		lg.SetLevel(loglvl.InfoLevel)
	case hclog.Warn:
		lg.SetLevel(loglvl.WarnLevel)
	case hclog.Error:
		lg.SetLevel(loglvl.ErrorLevel)
	}
}

// GetLevel returns the current log level as an hclog.Level.
// Maps golib levels back to hclog equivalents.
// DebugLevel returns Trace if EnableTrace is set, otherwise Debug.
// Unknown levels return Off.
func (o *_hclog) GetLevel() hclog.Level {
	var lg = o.logger()

	if lg == nil {
		return hclog.NoLevel
	}

	switch lg.GetLevel() {
	case loglvl.NilLevel:
		return hclog.NoLevel
	case loglvl.DebugLevel:
		if o.IsTrace() {
			return hclog.Trace
		} else {
			return hclog.Debug
		}
	case loglvl.InfoLevel:
		return hclog.Info
	case loglvl.WarnLevel:
		return hclog.Warn
	case loglvl.ErrorLevel:
		return hclog.Error
	default:
		return hclog.Off
	}
}

// StandardLogger returns a standard library *log.Logger backed by this hclog adapter.
// The ForceLevel option determines which golib level is used for all logs.
// Returns log.Default() if the logger is nil.
func (o *_hclog) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	var lg = o.logger()

	if lg == nil {
		return log.Default()
	}

	var lvl loglvl.Level
	switch opts.ForceLevel {
	case hclog.NoLevel, hclog.Off:
		lvl = loglvl.NilLevel
	case hclog.Trace:
		lvl = loglvl.DebugLevel
	case hclog.Debug:
		lvl = loglvl.DebugLevel
	case hclog.Info:
		lvl = loglvl.InfoLevel
	case hclog.Warn:
		lvl = loglvl.WarnLevel
	case hclog.Error:
		lvl = loglvl.ErrorLevel
	}

	return lg.GetStdLogger(lvl, 0)
}

// StandardWriter returns an io.Writer backed by the golib logger.
// All writes are logged through the golib logger's Write method.
// Returns os.Stdout if the logger is nil.
func (o *_hclog) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	var lg = o.logger()

	if lg == nil {
		return os.Stdout
	}

	return lg
}
