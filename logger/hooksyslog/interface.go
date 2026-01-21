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
	"os"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
)

// HookSyslog is a logrus hook that writes log entries to a syslog endpoint.
// It extends the standard logrus.Hook interface with additional methods
// for lifecycle management.
//
// The hook operates asynchronously by delegating the actual writing to a shared,
// buffered aggregator. This prevents blocking the main logging goroutine.
//
// Platform support:
//   - Unix/Linux: Supports local syslog (via Unix domain sockets) and remote syslog (TCP/UDP).
//   - Windows: Supports remote syslog (TCP/UDP). Local syslog is not supported.
//
// Thread safety:
//   - Fire() is safe for concurrent calls.
//   - Close() should be called once during shutdown to release resources.
//
// Example:
//
//	opts := logcfg.OptionsSyslog{
//		Network:  "udp",
//		Host:     "syslog.example.com:514",
//		Tag:      "myapp",
//		LogLevel: []string{"info", "error"},
//	}
//	hook, _ := New(opts, &logrus.JSONFormatter{})
//	logger.AddHook(hook)
//	defer hook.Close()
type HookSyslog interface {
	logtps.Hook
}

// New creates a new HookSyslog instance with the specified configuration.
//
// This function initializes the hook and establishes a connection to the syslog
// endpoint via a shared aggregator. If an aggregator for the specified endpoint
// already exists, it is reused, and its reference count is incremented.
//
// The background writer goroutine is managed automatically by the aggregator,
// so there is no need to manually start a run loop.
//
// Parameters:
//   - opt: Configuration options including network, host, tag, facility, and filters.
//   - format: Logrus formatter for log entries (nil for default text format).
//
// Configuration:
//   - opt.Network: Protocol ("tcp", "udp", "unixgram", "unix"). Empty string implies local auto-discovery (Unix only).
//   - opt.Host: Syslog server address ("host:port" for TCP/UDP, "/dev/log" for Unix).
//   - opt.Tag: Syslog tag/application name (appears in syslog output). Defaults to process name.
//   - opt.Facility: Syslog facility ("LOCAL0"-"LOCAL7", "USER", "DAEMON", etc.).
//   - opt.LogLevel: Filter log levels (empty = all levels).
//   - opt.DisableStack: Remove "stack" field from output.
//   - opt.DisableTimestamp: Remove "time" field from output.
//   - opt.EnableTrace: Include "caller", "file", "line" fields.
//   - opt.EnableAccessLog: Write entry.Message instead of formatted fields.
//
// Returns:
//   - HookSyslog: Configured hook ready to use.
//   - error: Non-nil if unable to initialize the connection aggregator.
//
// Example:
//
//	opts := logcfg.OptionsSyslog{
//		Network:  "tcp",
//		Host:     "192.168.1.50:514",
//		Tag:      "myapp",
//		Facility: "USER",
//		LogLevel: []string{"info", "warning", "error"},
//	}
//	hook, err := New(opts, &logrus.JSONFormatter{})
//	if err != nil {
//		return nil, fmt.Errorf("failed to create syslog hook: %w", err)
//	}
//	logger.AddHook(hook)
func New(opt logcfg.OptionsSyslog, format logrus.Formatter) (HookSyslog, error) {
	var (
		LVLs = make([]logrus.Level, 0)
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, loglvl.Parse(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	if opt.Tag == "" {
		opt.Tag = os.Args[0]
	}

	n := &hks{
		m: sync.Mutex{},
		o: ohks{
			format:           format,
			levels:           LVLs,
			disableStack:     opt.DisableStack,
			disableTimestamp: opt.DisableTimestamp,
			enableTrace:      opt.EnableTrace,
			enableAccessLog:  opt.EnableAccessLog,
			network:          libptc.Parse(opt.Network),
			endpoint:         opt.Host,
			tag:              opt.Tag,
			fac:              MakeFacility(opt.Facility),
		},
		w: nil,
		r: new(atomic.Bool),
		l: new(atomic.Bool),
	}

	a, l, e := setAgg(n.o.network, n.o.endpoint)
	if e != nil {
		return nil, e
	}

	n.w = a
	n.l.Store(l)
	n.r.Store(true)

	if !l {
		n.h, _ = os.Hostname()
	}

	return n, nil
}
