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

package logger

import (
	"log"

	loglvl "github.com/nabbar/golib/logger/level"
)

// GetStdLogger creates and returns a new standard Go log.Logger instance that uses this logger
// as its output destination. This allows integration with third-party code expecting *log.Logger.
//
// The returned logger writes through the Logger's io.Writer interface at the specified level.
// This is useful for integrating with standard library code or third-party packages that accept
// a standard logger.
//
// Parameters:
//   - lvl: Minimum log level for messages written through the standard logger
//   - logFlags: Standard log package flags (log.LstdFlags, log.Ldate, log.Ltime, etc.)
//
// Returns:
//   - *log.Logger: A new standard logger instance writing to this logger
//
// Example:
//
//	stdLogger := logger.GetStdLogger(loglvl.InfoLevel, log.LstdFlags)
//	stdLogger.Println("Message from standard logger")
func (o *logger) GetStdLogger(lvl loglvl.Level, logFlags int) *log.Logger {
	o.SetIOWriterLevel(lvl)
	return log.New(o, "", logFlags)
}

// SetStdLogger replaces the global standard Go logger (log package) with this logger.
// All subsequent calls to log.Print*, log.Fatal*, and log.Panic* will use this logger.
//
// This is useful for capturing logs from third-party libraries that use the standard log package.
// Be cautious: this affects the global logger state.
//
// Parameters:
//   - lvl: Minimum log level for messages from the global standard logger
//   - logFlags: Standard log package flags (log.LstdFlags, log.Ldate, log.Ltime, etc.)
//
// Example:
//
//	logger.SetStdLogger(loglvl.WarnLevel, log.LstdFlags)
//	log.Println("Now routed through custom logger")
func (o *logger) SetStdLogger(lvl loglvl.Level, logFlags int) {
	o.SetIOWriterLevel(lvl)
	log.SetOutput(o)
	log.SetPrefix("")
	log.SetFlags(logFlags)
}
