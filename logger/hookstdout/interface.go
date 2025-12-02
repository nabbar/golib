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

package hookstdout

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
	logcfg "github.com/nabbar/golib/logger/config"
	loghkw "github.com/nabbar/golib/logger/hookwriter"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// HookStdOut is a logrus hook that writes log entries to stdout with configurable
// filtering and formatting options.
//
// This interface extends logtps.Hook and provides integration with logrus logger for
// customized stdout log output handling. It supports field filtering (stack, timestamp, trace),
// custom formatters, color output, and access log mode.
//
// The hook uses os.Stdout as the default output destination and wraps it with
// colorable support for cross-platform color output compatibility.
type HookStdOut interface {
	logtps.Hook
}

// New creates a new HookStdOut instance for writing logrus entries to stdout.
//
// This is a convenience function that calls NewWithWriter with os.Stdout as the writer.
// The hook supports color output via mattn/go-colorable for cross-platform compatibility.
//
// Parameters:
//   - opt: Configuration options controlling behavior. If nil or DisableStandard is true,
//     returns (nil, nil) to indicate the hook should be disabled.
//   - lvls: Log levels to handle. If empty or nil, defaults to logrus.AllLevels.
//   - f: Optional logrus.Formatter for entry formatting. If nil, uses entry.Bytes().
//
// Configuration options (via opt):
//   - DisableStandard: If true, returns nil hook (disabled).
//   - DisableColor: If true, wraps stdout with colorable.NewNonColorable() to disable color output.
//   - DisableStack: If true, filters out stack trace fields from log data.
//   - DisableTimestamp: If true, filters out time fields from log data.
//   - EnableTrace: If false, filters out caller/file/line fields from log data.
//   - EnableAccessLog: If true, uses message-only mode (ignores fields and formatter).
//
// Returns:
//   - HookStdOut: The configured hook instance, or nil if disabled.
//   - error: Always returns nil for this function (error handling is consistent with NewWithWriter).
//
// Example:
//
//	opt := &logcfg.OptionsStd{
//	    DisableStandard: false,
//	    DisableColor:    false,
//	}
//	hook, err := hookstdout.New(opt, nil, &logrus.TextFormatter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.AddHook(hook)
func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error) {
	return NewWithWriter(nil, opt, lvls, f)
}

// NewWithWriter creates a new HookStdOut instance with a custom io.Writer.
//
// This function allows using a custom writer instead of os.Stdout, useful for testing
// or redirecting output to buffers, files, or other destinations while maintaining
// the HookStdOut interface semantics.
//
// Parameters:
//   - w: The target io.Writer where log entries will be written. If nil, defaults to os.Stdout.
//   - opt: Configuration options controlling behavior. If nil or DisableStandard is true,
//     returns (nil, nil) to indicate the hook should be disabled.
//   - lvls: Log levels to handle. If empty or nil, defaults to logrus.AllLevels.
//   - f: Optional logrus.Formatter for entry formatting. If nil, uses entry.Bytes().
//
// Configuration options (via opt):
//   - DisableStandard: If true, returns nil hook (disabled).
//   - DisableColor: If true, wraps writer with colorable.NewNonColorable() to disable color output.
//   - DisableStack: If true, filters out stack trace fields from log data.
//   - DisableTimestamp: If true, filters out time fields from log data.
//   - EnableTrace: If false, filters out caller/file/line fields from log data.
//   - EnableAccessLog: If true, uses message-only mode (ignores fields and formatter).
//
// Returns:
//   - HookStdOut: The configured hook instance, or nil if disabled.
//   - error: Returns error from hookwriter.New if the writer validation fails.
//
// Example:
//
//	var buf bytes.Buffer
//	opt := &logcfg.OptionsStd{
//	    DisableStandard: false,
//	    DisableColor:    true,
//	}
//	hook, err := hookstdout.NewWithWriter(&buf, opt, nil, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.AddHook(hook)
func NewWithWriter(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdOut, error) {
	if w == nil {
		w = os.Stdout
	}

	if opt == nil || opt.DisableStandard {
		return nil, nil
	} else if opt.DisableColor {
		w = colorable.NewNonColorable(w)
	}

	return loghkw.New(w, opt, lvls, f)
}
