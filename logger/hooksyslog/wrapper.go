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

import "io"

// Wrapper is a platform-specific interface for writing to syslog.
// It abstracts the differences between Unix syslog and Windows Event Log.
//
// Implementations:
//   - _Syslog: Unix/Linux implementation using log/syslog
//   - _WinLog: Windows implementation using golang.org/x/sys/windows/svc/eventlog
//
// The interface provides severity-specific methods that map to syslog/event log
// severity levels. Each method writes the provided byte slice to the underlying
// logging system.
type Wrapper interface {
	io.WriteCloser

	// Panic writes a message with ALERT severity (syslog) or ERROR (Windows).
	// This is the highest severity level after EMERGENCY.
	Panic(p []byte) (n int, err error)

	// Fatal writes a message with CRITICAL severity (syslog) or ERROR (Windows).
	// Used for critical conditions that require immediate attention.
	Fatal(p []byte) (n int, err error)

	// Error writes a message with ERROR severity.
	// Used for error conditions.
	Error(p []byte) (n int, err error)

	// Warning writes a message with WARNING severity.
	// Used for warning conditions.
	Warning(p []byte) (n int, err error)

	// Info writes a message with INFORMATIONAL severity.
	// Used for informational messages.
	Info(p []byte) (n int, err error)

	// Debug writes a message with DEBUG severity (syslog) or INFO (Windows).
	// Used for debug-level messages.
	Debug(p []byte) (n int, err error)
}
