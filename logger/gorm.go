/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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
	"context"
	"errors"
	"fmt"
	"time"

	gorlog "gorm.io/gorm/logger"
)

type logGorm struct {
	i bool
	s time.Duration
	l func() Logger
}

func NewGormLogger(fct func() Logger, ignoreRecordNotFoundError bool, slowThreshold time.Duration) gorlog.Interface {
	return &logGorm{
		i: ignoreRecordNotFoundError,
		s: slowThreshold,
		l: fct,
	}
}

func (l *logGorm) LogMode(level gorlog.LogLevel) gorlog.Interface {
	switch level {
	case gorlog.Silent:
		l.l().SetLevel(NilLevel)
	case gorlog.Info:
		l.l().SetLevel(InfoLevel)
	case gorlog.Warn:
		l.l().SetLevel(WarnLevel)
	case gorlog.Error:
		l.l().SetLevel(ErrorLevel)
	}

	return l
}

func (l *logGorm) Info(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(InfoLevel, s, i...).Log()
}

func (l *logGorm) Warn(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(WarnLevel, s, i...).Log()
}

func (l *logGorm) Error(ctx context.Context, s string, i ...interface{}) {
	l.l().Entry(ErrorLevel, s, i...).Log()
}

func (l *logGorm) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	var e *Entry
	elapsed := time.Since(begin)

	switch {
	case err != nil && (!errors.Is(err, gorlog.ErrRecordNotFound) || !l.i):
		e = l.l().Entry(ErrorLevel, "database error with trace")
		e.ErrorAdd(true, err)
	case elapsed > l.s && l.s != 0:
		e = l.l().Entry(WarnLevel, "database warning with trace")
		e.ErrorAdd(true, fmt.Errorf("SLOW Query >= %v", l.s))
	default:
		e = l.l().Entry(InfoLevel, "database trace")
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
