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
	"errors"
	"io"

	"github.com/mattn/go-colorable"
	logcfg "github.com/nabbar/golib/logger/config"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// HookWriter is a logrus hook that writes log entries to an io.Writer with configurable
// filtering and formatting options.
//
// This interface extends logtps.Hook and provides integration with logrus logger for
// customized log output handling. It supports field filtering (stack, timestamp, trace),
// custom formatters, and access log mode.
type HookWriter interface {
	logtps.Hook
}

// New creates a new HookWriter instance for writing logrus entries to a custom writer.
//
// Parameters:
//   - w: The target io.Writer where log entries will be written. Must not be nil.
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
//   - HookWriter: The configured hook instance, or nil if disabled.
//   - error: "hook writer is nil" if w is nil, otherwise nil.
//
// Example:
//
//	opt := &logcfg.OptionsStd{
//	    DisableStandard: false,
//	    DisableColor:    true,
//	}
//	hook, err := hookwriter.New(os.Stdout, opt, nil, &logrus.JSONFormatter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger.AddHook(hook)
func New(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookWriter, error) {
	if w == nil {
		return nil, errors.New("hook writer is nil")
	}

	if opt == nil || opt.DisableStandard {
		return nil, nil
	} else if opt.DisableColor {
		w = colorable.NewNonColorable(w)
	}

	if len(lvls) < 1 {
		lvls = logrus.AllLevels
	}

	n := &hkstd{
		w: w,
		l: lvls,
		f: f,
		s: opt.DisableStack,
		d: opt.DisableTimestamp,
		t: opt.EnableTrace,
		c: opt.DisableColor,
		a: opt.EnableAccessLog,
	}

	return n, nil
}
