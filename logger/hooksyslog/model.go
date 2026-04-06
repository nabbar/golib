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
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	libptc "github.com/nabbar/golib/network/protocol"
)

// ohks holds the immutable configuration options for a hook instance.
// These values are set at creation time and are not modified during the hook's lifecycle.
type ohks struct {
	format  logrus.Formatter // format specifies the logrus formatter to use for the log entry body.
	levels  []logrus.Level   // levels defines which log levels this hook will trigger for.
	msgSize int              // defined the max size for each log message

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
	r *atomic.Bool // r is an atomic flag indicating if the hook is running (i.e., not closed).
	f []func(d logrus.Fields) logrus.Fields
	p func(e *logrus.Entry, msg string) ([]byte, error)
	l func(l logrus.Level, p []byte, e error) ([]byte, error)
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

func (o *hks) getSeverity(l logrus.Level) Severity {
	switch l {
	case logrus.PanicLevel:
		return SeverityAlert
	case logrus.FatalLevel:
		return SeverityCrit
	case logrus.ErrorLevel:
		return SeverityErr
	case logrus.WarnLevel:
		return SeverityWarning
	case logrus.InfoLevel:
		return SeverityInfo
	case logrus.DebugLevel:
		return SeverityDebug
	default:
		return SeverityInfo
	}

}

func (o *hks) getMsgAccess(e *logrus.Entry, msg string) ([]byte, error) {
	if len(msg) < 1 {
		return nil, nil
	}

	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	return o.l(e.Level, []byte(msg), nil)
}

func (o *hks) getMsgNormal(e *logrus.Entry, _ string) ([]byte, error) {
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

	p, err := e.Bytes()
	return o.l(e.Level, p, err)
}

func (o *hks) writeMsg(p []byte, e error) error {
	if len(p) < 1 {
		return e
	}

	if e != nil {
		return e
	}

	if !bytes.HasSuffix(p, []byte("\n")) {
		p = append(p, byte('\n'))
	}

	_, e = o.w.Write(p)
	return e
}

func (o *hks) locWriteMsg(l logrus.Level, p []byte, e error) ([]byte, error) {
	if len(p) < 1 {
		return nil, e
	}

	if e != nil {
		return p, e
	}

	return []byte(fmt.Sprintf(
		"<%d>%s %s[%d]: %s",
		PriorityCalc(o.o.fac, o.getSeverity(l)),
		time.Now().Format(time.Stamp),
		o.o.tag,
		os.Getpid(),
		string(p),
	)), nil
}

func (o *hks) rmtWriteMsg(l logrus.Level, p []byte, e error) ([]byte, error) {
	if len(p) < 1 {
		return nil, e
	}

	if e != nil {
		return p, e
	}

	return []byte(fmt.Sprintf(
		"<%d>%s %s %s[%d]: %s",
		PriorityCalc(o.o.fac, o.getSeverity(l)),
		time.Now().Format(time.RFC3339),
		hst,
		o.o.tag,
		os.Getpid(),
		string(p),
	)), nil
}
