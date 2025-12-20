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
	"bytes"
	"strings"

	loglvl "github.com/nabbar/golib/logger/level"
)

// Close stops all logging hooks and releases associated resources.
// This method should always be called when done with the logger, typically via defer.
//
// The method:
//   - Stops all background hook goroutines
//   - Closes file handles
//   - Disconnects from syslog
//   - Waits up to 200ms for pending log entries to be written
//
// Returns:
//   - error: Always returns nil (satisfies io.Closer interface)
//
// Example:
//
//	logger := logger.New(ctx)
//	defer logger.Close() // Ensures cleanup
func (o *logger) Close() error {
	if o != nil && o.hasCloser() {
		o.switchCloser(nil)
	}
	return nil
}

// Write implements io.Writer by creating a log entry from the provided bytes.
// This allows the logger to be used anywhere an io.Writer is expected.
//
// The method:
//   - Trims whitespace from the message
//   - Applies configured filters (drops messages containing filter patterns)
//   - Logs at the level set by SetIOWriterLevel (defaults to NilLevel)
//   - Includes default fields from GetFields()
//
// Parameters:
//   - p: Bytes to write (typically a log message)
//
// Returns:
//   - n: Number of bytes written (always len(p))
//   - err: Always nil (errors are logged internally)
//
// Example:
//
//	logger.SetIOWriterLevel(loglvl.InfoLevel)
//	io.WriteString(logger, "Message from io.Writer")
func (o *logger) Write(p []byte) (n int, err error) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	val := strings.TrimSpace(string(o.IOWriterFilter(p)))

	if len(val) < 1 {
		return len(p), nil
	}

	o.newEntry(o.GetIOWriterLevel(), val, nil, o.GetFields(), nil).Log()
	return len(p), nil
}

// SetIOWriterLevel sets the log level used for Write() method calls.
// This is separate from the main logger level and applies only to io.Writer interface usage.
//
// Parameters:
//   - lvl: Log level for io.Writer operations
//
// Example:
//
//	logger.SetIOWriterLevel(loglvl.WarnLevel)
//	io.WriteString(logger, "This will be logged at Warn level")
func (o *logger) SetIOWriterLevel(lvl loglvl.Level) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	o.x.Store(keyWriter, lvl)
}

// GetIOWriterLevel returns the current log level for Write() method calls.
//
// Returns:
//   - loglvl.Level: The io.Writer log level, or NilLevel if not set
//
// Example:
//
//	level := logger.GetIOWriterLevel()
//	if level == loglvl.NilLevel {
//	    logger.SetIOWriterLevel(loglvl.InfoLevel)
//	}
func (o *logger) GetIOWriterLevel() loglvl.Level {
	if o == nil {
		return loglvl.NilLevel
	} else if o.x == nil {
		return loglvl.NilLevel
	} else if i, l := o.x.Load(keyWriter); !l {
		return loglvl.NilLevel
	} else if v, k := i.(loglvl.Level); !k {
		return loglvl.NilLevel
	} else {
		return v
	}
}

// SetIOWriterFilter replaces all filter patterns with the provided patterns.
// Messages containing any of these patterns will be dropped (not logged).
// This applies only to Write() method calls, not direct logging methods.
//
// Pass an empty slice to clear all filters.
//
// Parameters:
//   - pattern: Substrings to filter out (case-sensitive)
//
// Example:
//
//	// Drop any messages containing sensitive data
//	logger.SetIOWriterFilter("password", "token", "secret")
func (o *logger) SetIOWriterFilter(pattern ...string) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	var p = make([][]byte, 0, len(pattern))
	for _, s := range pattern {
		p = append(p, []byte(s))
	}

	o.x.Store(keyFilter, p)
}

// AddIOWriterFilter adds filter patterns to the existing list.
// Unlike SetIOWriterFilter, this method appends to the current filters
// instead of replacing them.
//
// Parameters:
//   - pattern: Additional substrings to filter out (case-sensitive)
//
// Example:
//
//	logger.SetIOWriterFilter("password")
//	logger.AddIOWriterFilter("token", "secret") // Now filters all three
func (o *logger) AddIOWriterFilter(pattern ...string) {
	if o == nil {
		return
	} else if o.x == nil {
		return
	}

	var p = make([][]byte, 0, len(pattern))

	if i, l := o.x.Load(keyFilter); !l {
		// nothing
	} else if v, k := i.([][]byte); !k {
		// nothing
	} else {
		p = append(make([][]byte, 0, len(pattern)+len(v)), v...)
	}

	for _, s := range pattern {
		p = append(p, []byte(s))
	}

	o.x.Store(keyFilter, p)
}

func (o *logger) IOWriterFilter(p []byte) []byte {
	if o == nil {
		return p
	} else if o.x == nil {
		return p
	} else if i, l := o.x.Load(keyFilter); !l {
		return p
	} else if v, k := i.([][]byte); !k {
		return p
	} else {
		for _, b := range v {
			if bytes.Contains(p, b) {
				return make([]byte, 0)
			}
		}

		return p
	}
}
