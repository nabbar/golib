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
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	logtps "github.com/nabbar/golib/logger/types"
	libptc "github.com/nabbar/golib/network/protocol"
)

// ohks holds the immutable configuration options for a hook instance.
// These values are set at creation time and are not modified during the hook's lifecycle.
type ohks struct {
	format           logrus.Formatter // format specifies the logrus formatter to use for the log entry body.
	levels           []logrus.Level   // levels defines which log levels this hook will trigger for.
	disableStack     bool             // disableStack controls the removal of the "stack" field from log entries.
	disableTimestamp bool             // disableTimestamp controls the removal of the "time" field from log entries.
	enableTrace      bool             // enableTrace controls the inclusion of "caller", "file", and "line" fields.
	enableAccessLog  bool             // enableAccessLog switches the hook to a mode where only the entry's message is logged.

	network  libptc.NetworkProtocol // network stores the protocol for the syslog connection (e.g., tcp, udp).
	endpoint string                 // endpoint is the network address of the syslog server.

	tag string   // tag is the application name or tag to be included in the syslog message.
	fac Facility // fac is the syslog facility code (e.g., USER, LOCAL0).
}

// hks is the concrete implementation of the HookSyslog interface.
// It holds the hook's configuration and its runtime state.
type hks struct {
	m sync.Mutex   // m provides exclusive access to the writer, primarily for recovery scenarios.
	o ohks         // o contains the immutable configuration options for the hook.
	w io.Writer    // w is the writer pointing to the shared, buffered connection aggregator.
	h string       // h stores the local hostname, cached for use in syslog messages for remote endpoints.
	r *atomic.Bool // r is an atomic flag indicating if the hook is running (i.e., not closed).
	l *atomic.Bool // l is an atomic flag indicating if the connection is to a local syslog endpoint.
}

// Levels returns the slice of logrus levels that this hook is configured to handle.
// This method is part of the logrus.Hook interface.
func (o *hks) Levels() []logrus.Level {
	return o.getLevel()
}

// RegisterHook adds this hook to the provided logrus logger instance.
// This is a convenience method equivalent to `logger.AddHook(hook)`.
func (o *hks) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

// Fire is the entry point for processing a log entry. It is called by logrus
// for each log message that matches the configured levels.
//
// The method performs the following steps:
//  1. Filters fields based on the hook's configuration (stack, timestamp, trace).
//  2. Formats the log entry into a byte slice. In "access log" mode, it uses the raw message;
//     otherwise, it uses the configured logrus formatter.
//  3. Maps the logrus level to the corresponding RFC 5424 syslog severity.
//  4. Constructs the full syslog message string, prepending the priority, timestamp,
//     hostname (for remote logs), and tag.
//  5. Writes the final message to the underlying shared aggregator.
//
// This method is non-blocking under normal conditions, as the actual network I/O
// is handled asynchronously by the aggregator.
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

	// In access log mode, use the raw message. Otherwise, use the formatter.
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

	// Map logrus level to syslog severity.
	var sev Severity
	switch ent.Level {
	case logrus.PanicLevel:
		sev = SeverityAlert
	case logrus.FatalLevel:
		sev = SeverityCrit
	case logrus.ErrorLevel:
		sev = SeverityErr
	case logrus.WarnLevel:
		sev = SeverityWarning
	case logrus.InfoLevel:
		sev = SeverityInfo
	case logrus.DebugLevel:
		sev = SeverityDebug
	default:
		sev = SeverityInfo
	}

	// Construct the full RFC 5424 syslog message.
	if o.l.Load() {
		// Format for local syslog (e.g., using time.Stamp and no hostname).
		p = []byte(fmt.Sprintf(
			"<%d>%s %s[%d]: %s",
			PriorityCalc(o.o.fac, sev),
			time.Now().Format(time.Stamp),
			o.o.tag,
			os.Getpid(),
			string(p),
		))
	} else {
		// Format for remote syslog (e.g., using RFC3339 and including hostname).
		p = []byte(fmt.Sprintf(
			"<%d>%s %s %s[%d]: %s",
			PriorityCalc(o.o.fac, sev),
			time.Now().Format(time.RFC3339),
			o.h,
			o.o.tag,
			os.Getpid(),
			string(p),
		))
	}

	if !bytes.HasSuffix(p, []byte("\n")) {
		p = append(p, byte('\n'))
	}

	// Write the formatted message to the aggregator.
	_, e = o.Write(p)

	if e != nil {
		return e
	}

	return nil
}

// filterKey removes a specific key from a logrus.Fields map.
// This is a utility function used to implement the field filtering options.
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
