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
	"fmt"

	"github.com/sirupsen/logrus"
)

// getFormatter returns the configured logrus formatter.
// Returns nil if no formatter was specified (uses logrus default).
func (o *hks) getFormatter() logrus.Formatter {
	return o.o.format
}

// getLevel returns the configured log levels for this hook.
func (o *hks) getLevel() []logrus.Level {
	return o.o.levels
}

// getDisableStack returns true if stack field filtering is enabled.
func (o *hks) getDisableStack() bool {
	return o.o.disableStack
}

// getDisableTimestamp returns true if timestamp field filtering is enabled.
func (o *hks) getDisableTimestamp() bool {
	return o.o.disableTimestamp
}

// getEnableTrace returns true if trace fields (caller, file, line) should be kept.
func (o *hks) getEnableTrace() bool {
	return o.o.enableTrace
}

// getEnableAccessLog returns true if access log mode is enabled
// (write entry.Message instead of formatted fields).
func (o *hks) getEnableAccessLog() bool {
	return o.o.enableAccessLog
}

// getSyslog creates a new platform-specific syslog writer.
// Returns a Wrapper implementation based on the current OS.
func (o *hks) getSyslog() (Wrapper, error) {
	return newSyslog(o.o.network, o.o.endpoint, o.o.tag, o.o.fac)
}

// getSyslogInfo returns a human-readable description of the syslog configuration.
// Used in error messages to identify the syslog destination.
func (o *hks) getSyslogInfo() string {
	return fmt.Sprintf("syslog to '%s %s' with tag '%s'", o.o.network.Code(), o.o.endpoint, o.o.tag)
}
