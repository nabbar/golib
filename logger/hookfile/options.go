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

// Package hookfile provides file-based logging hooks for logrus.
// This file contains getter methods for accessing the hook's configuration options.
package hookfile

import (
	"github.com/sirupsen/logrus"
)

// getFormatter returns the logrus.Formatter used by this hook.
// The formatter is responsible for converting log entries into byte slices.
func (o *hkf) getFormatter() logrus.Formatter {
	return o.o.format
}

// getLevel returns the log levels that this hook is configured to process.
// If no levels are explicitly set, it defaults to all log levels.
func (o *hkf) getLevel() []logrus.Level {
	return o.o.levels
}

// getDisableStack indicates whether stack traces are disabled for this hook.
// When true, stack traces will not be included in the log output.
func (o *hkf) getDisableStack() bool {
	return o.o.disableStack
}

// getDisableTimestamp indicates whether timestamps are disabled for this hook.
// When true, timestamps will not be included in the log output.
func (o *hkf) getDisableTimestamp() bool {
	return o.o.disableTimestamp
}

// getEnableTrace indicates whether trace information is enabled for this hook.
// When true, additional tracing information will be included in the log output.
func (o *hkf) getEnableTrace() bool {
	return o.o.enableTrace
}

// getEnableAccessLog indicates whether access log format is enabled.
// When true, logs will be written in a simplified access log format.
func (o *hkf) getEnableAccessLog() bool {
	return o.o.enableAccessLog
}
