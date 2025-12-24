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

// Package hookfile provides a logrus hook implementation for writing logs to files with various formatting options.
// It supports log rotation detection, custom formatters, and different log levels. The hook can be configured
// to enable/disable stack traces, timestamps, and access log formatting.
//
// # Important Usage Notes
//
// When using this hook in normal mode (not access log mode), all log data MUST be passed via the
// logrus.Entry.Data field. The Message parameter is ignored by the formatter. For example:
//
//	logger.WithField("msg", "User logged in").WithField("user", "john").Info("")
//
// NOT:
//
//	logger.Info("User logged in") // This message will be ignored!
//
// # Log Rotation
//
// The hook automatically detects external log rotation (e.g., by logrotate) when CreatePath and Create are enabled.
// It uses inode comparison to detect when the log file has been moved/renamed and automatically
// reopens the file at the configured path. The sync timer runs every second to check for rotation.
//
// # Thread Safety
//
// The hook is thread-safe and can be used concurrently from multiple goroutines. It uses an
// aggregator pattern to manage writes to the same file from multiple hooks efficiently.
//
// # Related Packages
//
// This package integrates with:
//   - github.com/sirupsen/logrus - The logging framework
//   - github.com/nabbar/golib/logger/config - Configuration structures
//   - github.com/nabbar/golib/ioutils/aggregator - Buffered file writing with rotation support
//   - github.com/nabbar/golib/logger/types - Common logger types and interfaces
package hookfile

import (
	"sync/atomic"

	libiot "github.com/nabbar/golib/ioutils"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

// HookFile defines the interface for a logrus hook that writes logs to files.
// It embeds the base Hook interface from golib/logger/types.
type HookFile interface {
	logtps.Hook
}

// New creates and initializes a new file hook with the specified options and formatter.
//
// Parameters:
//   - opt: Configuration options for the file hook including file path, permissions, and log levels
//   - format: The logrus.Formatter to use for formatting log entries
//
// Returns:
//   - HookFile: The initialized file hook instance
//   - error: An error if the hook could not be created (e.g., invalid file path)
//
// The function will create necessary directories if CreatePath is enabled in options.
// For automatic log rotation support, both CreatePath and Create must be enabled.
// If no log levels are specified, it will log all levels by default.
//
// Example usage:
//
//	opts := logcfg.OptionsFile{
//	    Filepath:   "/var/log/myapp.log",
//	    CreatePath: true,
//	    Create:     true,
//	    FileMode:   0644,
//	    PathMode:   0755,
//	    LogLevel:   []string{"info", "warning", "error"},
//	}
//	hook, err := New(opts, &logrus.TextFormatter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	logger := logrus.New()
//	logger.AddHook(hook)
//	// Remember to use Data field for messages:
//	logger.WithField("msg", "Application started").Info("")
func New(opt logcfg.OptionsFile, format logrus.Formatter) (HookFile, error) {
	if opt.Filepath == "" {
		return nil, errMissingFilePath
	}

	var LVLs = make([]logrus.Level, 0)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, loglvl.Parse(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	if opt.FileMode == 0 {
		opt.FileMode = 0644
	}

	if opt.PathMode == 0 {
		opt.PathMode = 0755
	}

	if opt.CreatePath {
		if e := libiot.PathCheckCreate(true, opt.Filepath, opt.FileMode.FileMode(), opt.PathMode.FileMode()); e != nil {
			return nil, e
		}
	}

	a, e := setAgg(opt.Filepath, opt.FileMode.FileMode(), opt.Create)
	if e != nil {
		return nil, e
	}

	n := &hkf{
		o: ohkf{
			format:           format,
			levels:           LVLs,
			disableStack:     opt.DisableStack,
			disableTimestamp: opt.DisableTimestamp,
			enableTrace:      opt.EnableTrace,
			enableAccessLog:  opt.EnableAccessLog,
			filepath:         opt.Filepath,
			filemode:         opt.FileMode.FileMode(),
			filecreate:       opt.Create,
		},
		w: a,
		r: new(atomic.Bool),
	}
	n.r.Store(true)

	return n, nil
}
