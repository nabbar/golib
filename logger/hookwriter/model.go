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

package hookwriter

import (
	"context"
	"io"
	"strings"

	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// hkstd implements HookWriter for writing logrus entries to an io.Writer.
//
// This internal type manages log entry processing with configurable field filtering,
// custom formatting, and access log mode. It is not thread-safe for concurrent
// Fire() calls on the same entry, but is safe when used as a logrus hook (logrus
// serializes hook calls per entry).
type hkstd struct {
	w io.Writer        // Target writer for formatted log output
	l []logrus.Level   // Log levels handled by this hook
	f logrus.Formatter // Optional formatter for entry serialization
	s bool             // DisableStack: filter stack trace fields
	d bool             // DisableTimestamp: filter time fields
	t bool             // EnableTrace: include caller/file/line fields
	c bool             // DisableColor: color output disabled
	a bool             // EnableAccessLog: message-only mode (no fields/formatting)
}

// getFormatter returns the configured logrus.Formatter.
//
// Returns nil if no formatter was configured during hook creation.
func (o *hkstd) getFormatter() logrus.Formatter {
	return o.f
}

// Run is a no-op implementation for the logtps.Hook interface.
//
// This hook does not require background processing and is always ready to handle entries.
func (o *hkstd) Run(ctx context.Context) {}

// IsRunning always returns true as this hook requires no lifecycle management.
//
// The hook is operational immediately after creation and remains active until discarded.
func (o *hkstd) IsRunning() bool {
	return true
}

// Levels returns the log levels that this hook will handle.
//
// This implements the logrus.Hook interface, allowing logrus to determine which
// entries should be passed to Fire().
func (o *hkstd) Levels() []logrus.Level {
	return o.l
}

// RegisterHook adds this hook to a logrus.Logger instance.
//
// This is a convenience method that calls logger.AddHook() with the current hook.
// After registration, the logger will route matching log entries to this hook's Fire() method.
func (o *hkstd) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

// Fire processes a logrus entry according to the hook's configuration and writes it to the output.
//
// This method implements the core logrus.Hook interface. It:
//  1. Duplicates the entry to avoid modifying the original
//  2. Filters fields based on DisableStack, DisableTimestamp, EnableTrace options
//  3. Formats the entry using the configured formatter or access log mode
//  4. Writes the formatted output to the underlying io.Writer
//
// Field filtering:
//   - If DisableStack is true, removes "stack" field
//   - If DisableTimestamp is true, removes "time" field
//   - If EnableTrace is false, removes "caller", "file", and "line" fields
//
// Formatting modes:
//   - Access log mode (EnableAccessLog=true): Uses entry.Message only, appends newline if missing
//   - Standard mode: Uses configured Formatter or entry.Bytes() for formatting
//
// Returns:
//   - nil if write succeeds or entry has no data/message to write
//   - error from formatter or writer on failure
func (o *hkstd) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if o.s {
		ent.Data = o.filterKey(ent.Data, logtps.FieldStack)
	}

	if o.d {
		ent.Data = o.filterKey(ent.Data, logtps.FieldTime)
	}

	if !o.t {
		ent.Data = o.filterKey(ent.Data, logtps.FieldCaller)
		ent.Data = o.filterKey(ent.Data, logtps.FieldFile)
		ent.Data = o.filterKey(ent.Data, logtps.FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.a {
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

	if _, e = o.Write(p); e != nil {
		return e
	}

	return nil
}

// filterKey removes a specific key from logrus.Fields if present.
//
// This is used internally to filter out fields like "stack", "time", "caller", etc.
// based on hook configuration. Returns the modified fields map.
//
// Parameters:
//   - f: The fields map to filter (may be empty)
//   - key: The field key to remove
//
// Returns:
//   - The same fields map with the key removed if it existed, or unchanged if key was absent or fields empty
func (o *hkstd) filterKey(f logrus.Fields, key string) logrus.Fields {
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
