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

// Package hookfile provides a logrus hook implementation for file-based logging
// with configurable formatting and log levels. It's part of the golib logger package.
package hookfile

import (
	"io"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// ohkf contains the configuration options for the file hook.
// It's an unexported type to ensure immutability after creation.
// All fields are set during hook creation and never modified afterwards.
type ohkf struct {
	format     logrus.Formatter // formatter for log entries
	levels     []logrus.Level   // log levels to process
	filepath   string           // path to log file
	filemode   os.FileMode      // file permissions
	filecreate bool             // create file if missing (for rotation)
	msgMaxSize int              // defined the max size for each log message
}

// hkf is the main implementation of the HookFile interface.
// It handles writing log entries to the configured file with the specified formatting.
// The hook uses a file aggregator for efficient writes and automatic rotation detection.
type hkf struct {
	m sync.Mutex   // protects Write during error recovery
	o ohkf         // immutable config data
	w io.Writer    // aggregator writer (buffered, shared)
	r *atomic.Bool // is running flag (thread-safe)
	f []func(d logrus.Fields) logrus.Fields
	p func(e *logrus.Entry, msg string) ([]byte, error)
}

// Levels returns the log levels that this hook is configured to handle.
// Implements the logrus.Hook interface.
func (o *hkf) Levels() []logrus.Level {
	return o.getLevel()
}

// RegisterHook registers this hook with the provided logrus Logger.
// Implements the logtps.Hook interface.
func (o *hkf) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

// Fire processes a log entry and writes it to the file.
// Implements the logrus.Hook interface.
//
// The method handles various formatting options like disabling stack traces,
// timestamps, and enabling access log format based on the hook's configuration.
//
// Returns an error if the log entry could not be written to the file.
func (o *hkf) Fire(entry *logrus.Entry) error {
	// Check if this log level should be processed by this hook
	if !slices.Contains(o.getLevel(), entry.Level) {
		return nil
	}

	ent := entry.Dup()
	ent.Level = entry.Level

	for i := range o.f {
		ent.Data = o.f[i](ent.Data)
	}

	return o.writeMsg(o.p(ent, entry.Message))
}

// filterKey removes a specific key from the logrus.Fields map if it exists.
// This is used to filter out specific fields like stack traces or timestamps
// based on the hook's configuration.
//
// Parameters:
//   - f: The logrus.Fields map to filter
//   - key: The key to remove from the fields
//
// Returns the filtered fields map (same instance if key not found)
func (o *hkf) filterKey(f logrus.Fields, key string) logrus.Fields {
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

func (o *hkf) getMsgAccess(_ *logrus.Entry, msg string) ([]byte, error) {
	if len(msg) < 1 {
		return nil, nil
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	return []byte(msg), nil
}

func (o *hkf) getMsgNormal(e *logrus.Entry, _ string) ([]byte, error) {
	// Normal mode: IMPORTANT - The Message field is ignored!
	// All log data must be passed via the Data field (logrus.Fields).
	// This is because the formatter (e.g., TextFormatter) only processes
	// the Data field and ignores the Message parameter.
	// To log a message, use: entry.Data["msg"] = "your message"
	if len(e.Data) < 1 {
		return nil, nil
	}

	if f := o.getFormatter(); f != nil {
		return f.Format(e)
	}

	return e.Bytes()
}

func (o *hkf) writeMsg(p []byte, e error) error {
	if len(p) < 1 {
		return e
	}

	if e != nil {
		return e
	}

	_, e = o.w.Write(p)
	return e
}
