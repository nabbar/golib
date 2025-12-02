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
	"sync/atomic"

	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
	"github.com/sirupsen/logrus"
)

// HookSyslog is a logrus hook that writes log entries to syslog.
// It extends the standard logrus.Hook interface with additional methods
// for lifecycle management and direct syslog writing.
//
// The hook operates asynchronously using a buffered channel (capacity: 250)
// to prevent blocking the logging goroutine. A background goroutine (started
// via Run) processes the buffered entries and writes them to syslog.
//
// Platform support:
//   - Unix/Linux: Uses log/syslog with TCP, UDP, or Unix domain sockets
//   - Windows: Uses Windows Event Log via golang.org/x/sys/windows/svc/eventlog
//
// Thread safety:
//   - Fire() is safe for concurrent calls (buffered channel)
//   - WriteSev() is safe for concurrent calls (buffered channel)
//   - Done() and Close() should only be called once during shutdown
//
// Example:
//
//	opts := logcfg.OptionsSyslog{
//		Network:  "unixgram",
//		Host:     "/dev/log",
//		Tag:      "myapp",
//		LogLevel: []string{"info", "error"},
//	}
//	hook, _ := New(opts, &logrus.JSONFormatter{})
//	go hook.Run(ctx)
//	logger.AddHook(hook)
type HookSyslog interface {
	logtps.Hook

	// Done returns a receive-only channel that is closed when the hook's
	// Run goroutine terminates. This allows graceful shutdown coordination.
	//
	// The channel is closed when:
	//   - The context passed to Run() is cancelled
	//   - Close() is called on the hook
	//
	// Use this to wait for all buffered log entries to be written before
	// terminating the application:
	//
	//	cancel() // Stop the Run goroutine
	//	hook.Close() // Close the channels
	//	<-hook.Done() // Wait for completion
	//
	// Note: This channel is safe to read from multiple goroutines, but
	// each reader will only receive the close signal once.
	Done() <-chan struct{}
	// WriteSev writes a log entry to the syslog buffer with the specified
	// severity level and data.
	//
	// This method bypasses the logrus.Entry mechanism and directly writes
	// to syslog. It's useful for custom logging scenarios or when you need
	// explicit control over the severity level.
	//
	// Parameters:
	//   - s: Syslog severity (Emergency, Alert, Critical, Error, Warning, Notice, Info, Debug)
	//   - p: Log message data (will be sent as-is to syslog)
	//
	// Returns:
	//   - n: Number of bytes accepted (len(p)) if successful
	//   - err: Error if the channel is closed or buffer is full
	//
	// Behavior:
	//   - Non-blocking if buffer has space (typical case)
	//   - Blocks if buffer is full (250 entries) until space is available
	//   - Returns error if Close() was called (channel closed)
	//
	// The data is queued to a buffered channel and written asynchronously
	// by the Run() goroutine. There's no guarantee of immediate delivery.
	//
	// Example:
	//
	//	hook, _ := New(opts, nil)
	//	go hook.Run(ctx)
	//	_, err := hook.WriteSev(SyslogSeverityInfo, []byte("Custom log entry"))
	//	if err != nil {
	//		log.Printf("Failed to write: %v", err)
	//	}
	WriteSev(s SyslogSeverity, p []byte) (n int, err error)
}

// New creates a new HookSyslog instance with the specified configuration.
//
// This function initializes the hook but does NOT start the background writer
// goroutine. You must call Run(ctx) in a separate goroutine after creating
// the hook.
//
// Parameters:
//   - opt: Configuration options including network, host, tag, facility, and filters
//   - format: Logrus formatter for log entries (nil for default text format)
//
// Configuration:
//   - opt.Network: Protocol ("tcp", "udp", "unixgram", "unix", "" for local)
//   - opt.Host: Syslog server address ("host:port" for TCP/UDP, "/dev/log" for Unix)
//   - opt.Tag: Syslog tag/application name (appears in syslog output)
//   - opt.Facility: Syslog facility ("LOCAL0"-"LOCAL7", "USER", "DAEMON", etc.)
//   - opt.LogLevel: Filter log levels (empty = all levels)
//   - opt.DisableStack: Remove "stack" field from output
//   - opt.DisableTimestamp: Remove "time" field from output
//   - opt.EnableTrace: Include "caller", "file", "line" fields
//   - opt.EnableAccessLog: Write entry.Message instead of formatted fields
//
// Returns:
//   - HookSyslog: Configured hook ready to use (call Run to start)
//   - error: Non-nil if unable to connect to syslog (validates connection)
//
// The function validates the syslog connection by opening and immediately
// closing it. This ensures early detection of configuration errors.
//
// Example:
//
//	opts := logcfg.OptionsSyslog{
//		Network:  "unixgram",
//		Host:     "/dev/log",
//		Tag:      "myapp",
//		Facility: "USER",
//		LogLevel: []string{"info", "warning", "error"},
//	}
//	hook, err := New(opts, &logrus.JSONFormatter{})
//	if err != nil {
//		return fmt.Errorf("failed to create syslog hook: %w", err)
//	}
//	go hook.Run(context.Background())
//	defer hook.Close()
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

	n := &hks{
		s: new(atomic.Value),
		d: new(atomic.Value),
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
			//sev : MakeSeverity(opt.Severity),
		},
		r: new(atomic.Bool),
	}

	n.s.Store(make(chan struct{}))
	n.d.Store(make(chan []data, 250))

	if h, e := n.getSyslog(); e != nil {
		return nil, e
	} else {
		_ = h.Close()
	}

	return n, nil
}
