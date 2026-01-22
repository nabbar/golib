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

import "strings"

// Severity represents the severity level of a syslog message
// according to RFC 5424. Lower numerical values indicate higher severity.
//
// The severity levels map to logrus levels as follows:
//   - Emergency (0): System is unusable
//   - Alert (1): Action must be taken immediately → logrus.PanicLevel
//   - Critical (2): Critical conditions → logrus.FatalLevel
//   - Error (3): Error conditions → logrus.ErrorLevel
//   - Warning (4): Warning conditions → logrus.WarnLevel
//   - Notice (5): Normal but significant condition
//   - Informational (6): Informational messages → logrus.InfoLevel
//   - Debug (7): Debug-level messages → logrus.DebugLevel
type Severity uint8

const (
	SeverityEmerg   Severity = iota // System is unusable
	SeverityAlert                   // Action must be taken immediately
	SeverityCrit                    // Critical conditions
	SeverityErr                     // Error conditions
	SeverityWarning                 // Warning conditions
	SeverityNotice                  // Normal but significant condition
	SeverityInfo                    // Informational messages
	SeverityDebug                   // Debug-level messages
)

// String returns the RFC 5424 name of the severity level in uppercase.
// Returns an empty string for invalid/unknown severity values.
//
// Example:
//
//	sev := SeverityInfo
//	fmt.Println(sev.String()) // Outputs: "INFO"
func (s Severity) String() string {
	switch s {
	case SeverityEmerg:
		return "EMERG"
	case SeverityAlert:
		return "ALERT"
	case SeverityCrit:
		return "CRIT"
	case SeverityErr:
		return "ERR"
	case SeverityWarning:
		return "WARNING"
	case SeverityNotice:
		return "NOTICE"
	case SeverityInfo:
		return "INFO"
	case SeverityDebug:
		return "DEBUG"
	}

	return ""
}

func (s Severity) Uint8() uint8 {
	return uint8(s)
}

// MakeSeverity converts a severity string to a Severity value.
// The conversion is case-insensitive. Returns 0 if the string doesn't match any known severity.
func MakeSeverity(severity string) Severity {
	switch strings.ToUpper(severity) {
	case SeverityEmerg.String():
		return SeverityEmerg
	case SeverityAlert.String():
		return SeverityAlert
	case SeverityCrit.String():
		return SeverityCrit
	case SeverityErr.String():
		return SeverityErr
	case SeverityWarning.String():
		return SeverityWarning
	case SeverityNotice.String():
		return SeverityNotice
	case SeverityInfo.String():
		return SeverityInfo
	case SeverityDebug.String():
		return SeverityDebug
	}

	return 0
}

// ListSeverity returns a slice containing all defined Severity levels
// in order from Emergency (0) to Debug (7).
func ListSeverity() []Severity {
	return []Severity{
		SeverityEmerg,
		SeverityAlert,
		SeverityCrit,
		SeverityErr,
		SeverityWarning,
		SeverityNotice,
		SeverityInfo,
		SeverityDebug,
	}
}
