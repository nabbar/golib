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

package hooksyslog

import (
	"strings"
	"sync/atomic"

	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

// ohks holds the immutable configuration options for the hook.
// These values are set at creation time and never modified.
type ohks struct {
	format           logrus.Formatter   // Log entry formatter (nil = default)
	levels           []logrus.Level     // Enabled log levels
	disableStack     bool               // Filter out "stack" field
	disableTimestamp bool               // Filter out "time" field
	enableTrace      bool               // Keep "caller", "file", "line" fields
	enableAccessLog  bool               // Write entry.Message instead of fields

	network  libptc.NetworkProtocol // Syslog protocol (tcp, udp, unix, unixgram)
	endpoint string                 // Syslog server address

	tag string         // Syslog tag/application name
	fac SyslogFacility // Syslog facility code
}

// hks is the concrete implementation of HookSyslog.
// It uses atomic values for thread-safe channel management.
type hks struct {
	s *atomic.Value // Stores chan struct{} for Done() signal
	d *atomic.Value // Stores chan []data for buffered writes (capacity: 250)
	o ohks          // Immutable configuration
	r *atomic.Bool  // Running status flag
}

// Levels returns the log levels that this hook is configured to handle.
// This is part of the logrus.Hook interface.
//
// Returns only the levels specified in OptionsSyslog.LogLevel, or all
// levels if LogLevel was empty.
func (o *hks) Levels() []logrus.Level {
	return o.getLevel()
}

// RegisterHook adds this hook to the specified logrus logger.
// This is a convenience method equivalent to logger.AddHook(hook).
//
// Example:
//
//	hook, _ := New(opts, formatter)
//	hook.RegisterHook(logger)
func (o *hks) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

// Fire processes a log entry and queues it for writing to syslog.
// This is part of the logrus.Hook interface and is called automatically
// by logrus for each log statement.
//
// Behavior:
//   - Duplicates the entry to avoid modifying the original
//   - Applies field filtering (stack, timestamp, trace)
//   - Formats the entry using the configured formatter
//   - Maps logrus level to syslog severity
//   - Queues the formatted data to the buffered channel
//
// Access Log Mode:
//   - If EnableAccessLog is true: writes entry.Message, ignores fields
//   - If EnableAccessLog is false: formats fields, ignores Message
//
// Returns:
//   - nil on success (data queued to channel)
//   - error if formatting fails or channel is closed
//
// Note: This method returns quickly (non-blocking) as long as the
// buffer has space. Actual syslog writing happens in Run().
func (o *hks) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if o.getDisableStack() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldStack)
	}

	if o.getDisableTimestamp() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldTime)
	}

	if !o.getEnableTrace() {
		ent.Data = o.filterKey(ent.Data, logtps.FieldCaller)
		ent.Data = o.filterKey(ent.Data, logtps.FieldFile)
		ent.Data = o.filterKey(ent.Data, logtps.FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.getEnableAccessLog() {
		if len(entry.Message) > 0 {
			if !strings.HasSuffix(entry.Message, "\n") {
				entry.Message += "\n"
			}
			p = []byte(entry.Message)
		} else {
			return nil
		}
	} else {
		if len(ent.Data) < 1 {
			return nil
		}

		if f := o.getFormatter(); f != nil {
			p, e = f.Format(ent)
		} else {
			p, e = ent.Bytes()
		}

		if e != nil {
			return e
		}
	}

	switch ent.Level {
	case logrus.PanicLevel:
		_, e = o.WriteSev(SyslogSeverityAlert, p)
	case logrus.FatalLevel:
		_, e = o.WriteSev(SyslogSeverityCrit, p)
	case logrus.ErrorLevel:
		_, e = o.WriteSev(SyslogSeverityErr, p)
	case logrus.WarnLevel:
		_, e = o.WriteSev(SyslogSeverityWarning, p)
	case logrus.InfoLevel:
		_, e = o.WriteSev(SyslogSeverityInfo, p)
	case logrus.DebugLevel:
		_, e = o.WriteSev(SyslogSeverityDebug, p)
	default:
		_, e = o.Write(p)
	}
	if e != nil {
		return e
	}

	return nil
}

// filterKey removes a specific field from logrus.Fields if present.
// Used to implement DisableStack, DisableTimestamp, and EnableTrace options.
//
// Returns:
//   - Modified fields with key removed (if present)
//   - Unchanged fields if key not present or fields empty
func (o *hks) filterKey(f logrus.Fields, key string) logrus.Fields {
	if len(f) < 1 {
		return f
	}

	if _, ok := f[key]; !ok {
		return f
	} else {
		delete(f, key)
		return f
	}
}
