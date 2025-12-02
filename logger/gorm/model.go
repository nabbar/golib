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

package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	liblog "github.com/nabbar/golib/logger"
	logent "github.com/nabbar/golib/logger/entry"
	loglvl "github.com/nabbar/golib/logger/level"
	gorlog "gorm.io/gorm/logger"
)

// logGorm implements gorlog.Interface to bridge GORM's logging with golib logger.
// It provides thread-safe logging with configurable slow query detection and error filtering.
type logGorm struct {
	i bool                 // ignoreRecordNotFoundError: when true, ErrRecordNotFound logs as Info instead of Error
	s time.Duration        // slowThreshold: queries exceeding this duration log as Warn; 0 disables detection
	l func() liblog.Logger // logger factory: called per-log to get current logger instance
}

// LogMode sets the log level for the GORM logger and returns itself for method chaining.
// It maps GORM log levels to golib equivalents:
//   - Silent -> NilLevel (no logging)
//   - Info -> InfoLevel (detailed query logs)
//   - Warn -> WarnLevel (slow queries and warnings)
//   - Error -> ErrorLevel (errors only)
//
// This method is thread-safe and can be called concurrently.
// The returned interface is the same instance to support GORM's fluent API.
func (l *logGorm) LogMode(level gorlog.LogLevel) gorlog.Interface {
	switch level {
	case gorlog.Silent:
		l.l().SetLevel(loglvl.NilLevel)
	case gorlog.Info:
		l.l().SetLevel(loglvl.InfoLevel)
	case gorlog.Warn:
		l.l().SetLevel(loglvl.WarnLevel)
	case gorlog.Error:
		l.l().SetLevel(loglvl.ErrorLevel)
	}

	return l
}

// Info logs an informational message from GORM.
// The context is currently unused but required by GORM's interface.
// Messages are formatted with fmt.Sprintf semantics if args are provided.
//
// This method is typically called by GORM for general information like
// connection establishment or configuration changes.
func (l *logGorm) Info(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(loglvl.InfoLevel, s, i...).Log()
}

// Warn logs a warning message from GORM.
// The context is currently unused but required by GORM's interface.
// Messages are formatted with fmt.Sprintf semantics if args are provided.
//
// This method is called by GORM for non-critical issues that should be
// brought to attention, such as deprecated features or suboptimal usage.
func (l *logGorm) Warn(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(loglvl.WarnLevel, s, i...).Log()
}

// Error logs an error message from GORM.
// The context is currently unused but required by GORM's interface.
// Messages are formatted with fmt.Sprintf semantics if args are provided.
//
// This method is called by GORM for errors that occurred outside of query
// execution, such as connection failures or initialization errors.
func (l *logGorm) Error(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(loglvl.ErrorLevel, s, i...).Log()
}

// Trace logs detailed query execution information with timing, SQL, and results.
// This is the primary method called by GORM after each database operation.
//
// The log level is determined by the query outcome:
//   - ErrorLevel: When err is not nil (unless it's ErrRecordNotFound and ignored)
//   - WarnLevel: When query duration exceeds slowThreshold (if threshold > 0)
//   - InfoLevel: For successful queries within threshold
//
// All logs include structured fields:
//   - "elapsed ms": Query execution time in milliseconds (float64)
//   - "rows": Number of affected rows (int64), or "-" if rowsAffected is -1
//   - "query": The SQL query string
//
// The fc function is called to retrieve SQL and row count, allowing lazy evaluation.
//
// Parameters:
//   - ctx: Request context (currently unused but required by GORM's interface)
//   - begin: Query start time for duration calculation
//   - fc: Function returning SQL query string and affected row count
//   - err: Query execution error, or nil on success
//
// Thread safety: This method is safe for concurrent calls.
func (l *logGorm) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	var e logent.Entry
	elapsed := time.Since(begin)

	switch {
	case err != nil && (!errors.Is(err, gorlog.ErrRecordNotFound) || !l.i):
		e = l.l().Entry(loglvl.ErrorLevel, "database error with trace")
		e.ErrorAdd(true, err)
	case elapsed > l.s && l.s != 0:
		e = l.l().Entry(loglvl.WarnLevel, "database warning with trace")
		e.ErrorAdd(true, fmt.Errorf("SLOW Query >= %v", l.s))
	default:
		e = l.l().Entry(loglvl.InfoLevel, "database trace")
	}

	e.FieldAdd("elapsed ms", float64(elapsed.Nanoseconds())/1e6)

	sql, rows := fc()
	if rows == -1 {
		e.FieldAdd("rows", "-")
		e.FieldAdd("query", sql)
	} else {
		e.FieldAdd("rows", rows)
		e.FieldAdd("query", sql)
	}

	e.Log()
}
