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

// data represents a single log entry in the buffered channel.
// It pairs a severity level with the formatted log data.
//
// This structure is used internally to queue log entries from Fire()
// to the background writer in Run().
type data struct {
	s SyslogSeverity // Syslog severity level
	p []byte         // Formatted log data to write
}

// newData creates a new data instance with the specified severity and payload.
//
// Parameters:
//   - s: Syslog severity level (Emergency, Alert, Critical, Error, Warning, Notice, Info, Debug)
//   - p: Formatted log data (already processed by logrus formatter)
//
// Returns:
//   - data: Initialized data structure ready to be queued
func newData(s SyslogSeverity, p []byte) data {
	return data{
		s: s,
		p: p,
	}
}
