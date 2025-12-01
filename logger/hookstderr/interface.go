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

package hookstderr

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"
	logcfg "github.com/nabbar/golib/logger/config"
	loghkw "github.com/nabbar/golib/logger/hookwriter"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// HookStdErr is a logrus hook that writes log entries to standard error (stderr) with configurable
// filtering and formatting options.
//
// This interface extends logtps.Hook and provides integration with logrus logger for
// customized error log output handling. It is specifically designed for writing to stderr,
// which is conventionally used for error messages and diagnostic output.
//
// The hook supports field filtering (stack, timestamp, trace), custom formatters, and access log mode.
// It delegates its implementation to the hookwriter package, providing a stderr-specific wrapper.
type HookStdErr interface {
	logtps.Hook
}

// New creates a new HookStdErr instance for writing logrus entries to standard error (stderr).
//
// This function is a convenience wrapper that creates a hook writing to os.Stderr.
// For custom writer destinations, use NewWithWriter instead.
//
// Parameters:
//   - opt: Configuration options controlling behavior. If nil or DisableStandard is true,
//     returns (nil, nil) to indicate the hook should be disabled.
//   - lvls: Log levels to handle. If empty or nil, defaults to logrus.AllLevels.
//   - f: Optional logrus.Formatter for entry formatting. If nil, uses entry.Bytes().
//
// Configuration options (via opt):
//   - DisableStandard: If true, returns nil hook (disabled).
//   - DisableColor: If true, wraps stderr with colorable.NewNonColorable() to disable color output.
//   - DisableStack: If true, filters out stack trace fields from log data.
//   - DisableTimestamp: If true, filters out time fields from log data.
//   - EnableTrace: If false, filters out caller/file/line fields from log data.
//   - EnableAccessLog: If true, uses message-only mode (ignores fields and formatter).
//
// Returns:
//   - HookStdErr: The configured hook instance writing to stderr, or nil if disabled.
//   - error: An error if there is an issue creating the hook (e.g., from underlying hookwriter.New).
//
// Example:
//
//	opt := &logcfg.OptionsStd{
//	    DisableStandard: false,
//	    DisableColor:    true,
//	}
//	hook, err := hookstderr.New(opt, nil, &logrus.JSONFormatter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.AddHook(hook)
func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error) {
	return NewWithWriter(nil, opt, lvls, f)
}

// NewWithWriter creates a new HookStdErr instance for writing logrus entries to a custom writer.
//
// This function allows specifying a custom io.Writer instead of os.Stderr. This is useful for:
//   - Testing with buffers or mock writers
//   - Redirecting stderr to files or network destinations
//   - Wrapping stderr with additional processing layers
//
// Parameters:
//   - w: The target io.Writer where log entries will be written. If nil, defaults to os.Stderr.
//   - opt: Configuration options controlling behavior. If nil or DisableStandard is true,
//     returns (nil, nil) to indicate the hook should be disabled.
//   - lvls: Log levels to handle. If empty or nil, defaults to logrus.AllLevels.
//   - f: Optional logrus.Formatter for entry formatting. If nil, uses entry.Bytes().
//
// Configuration behavior:
//   - When DisableColor is true, the writer is wrapped with colorable.NewNonColorable(w)
//     to strip ANSI color escape sequences from output.
//   - When DisableColor is false, the writer is used as-is, allowing color output if
//     the writer supports it (e.g., terminal stderr).
//
// Returns:
//   - HookStdErr: The configured hook instance, or nil if disabled.
//   - error: An error if there is an issue creating the hook (e.g., from underlying hookwriter.New).
//
// Example with buffer for testing:
//
//	var buf bytes.Buffer
//	opt := &logcfg.OptionsStd{DisableStandard: false}
//	hook, err := hookstderr.NewWithWriter(&buf, opt, nil, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.AddHook(hook)
func NewWithWriter(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error) {
	if w == nil {
		w = os.Stderr
	}

	if opt == nil || opt.DisableStandard {
		return nil, nil
	} else if opt.DisableColor {
		w = colorable.NewNonColorable(w)
	}

	return loghkw.New(w, opt, lvls, f)
}
