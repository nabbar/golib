/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package logger

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
)

const (
	HCLogArgs = "hclog.args"
	HCLogName = "hclog.name"
)

type _hclog struct {
	l Logger
}

func (l *_hclog) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Off, hclog.NoLevel:
		return
	case hclog.Trace:
		l.l.Debug(msg, nil, args...)
	case hclog.Debug:
		l.l.Debug(msg, nil, args...)
	case hclog.Info:
		l.l.Info(msg, nil, args...)
	case hclog.Warn:
		l.l.Warning(msg, nil, args...)
	case hclog.Error:
		l.l.Error(msg, nil, args...)
	}
}

func (l *_hclog) Trace(msg string, args ...interface{}) {
	l.l.Debug(msg, nil, args...)
}

func (l *_hclog) Debug(msg string, args ...interface{}) {
	l.l.Debug(msg, nil, args...)
}

func (l *_hclog) Info(msg string, args ...interface{}) {
	l.l.Info(msg, nil, args...)
}

func (l *_hclog) Warn(msg string, args ...interface{}) {
	l.l.Warning(msg, nil, args...)
}

func (l *_hclog) Error(msg string, args ...interface{}) {
	l.l.Error(msg, nil, args...)
}

func (l *_hclog) IsTrace() bool {
	return l.l.GetOptions().EnableTrace
}

func (l *_hclog) IsDebug() bool {
	return l.l.GetLevel() >= DebugLevel
}

func (l *_hclog) IsInfo() bool {
	return l.l.GetLevel() >= InfoLevel
}

func (l *_hclog) IsWarn() bool {
	return l.l.GetLevel() >= WarnLevel
}

func (l *_hclog) IsError() bool {
	return l.l.GetLevel() >= ErrorLevel
}

func (l *_hclog) ImpliedArgs() []interface{} {
	fields := l.l.GetFields()

	if a, ok := fields[HCLogArgs]; !ok {
		return make([]interface{}, 0)
	} else if s, ok := a.([]interface{}); ok {
		return s
	}

	return make([]interface{}, 0)
}

func (l *_hclog) With(args ...interface{}) hclog.Logger {
	l.l.SetFields(l.l.GetFields().Add(HCLogArgs, args))
	return l
}

func (l *_hclog) Name() string {
	fields := l.l.GetFields()

	if a, ok := fields[HCLogName]; !ok {
		return ""
	} else if s, ok := a.(string); ok {
		return s
	}

	return ""
}

func (l *_hclog) Named(name string) hclog.Logger {
	l.l.SetFields(l.l.GetFields().Add(HCLogName, name))
	return l
}

func (l *_hclog) ResetNamed(name string) hclog.Logger {
	l.l.SetFields(l.l.GetFields().Add(HCLogName, name))
	return l
}

func (l *_hclog) SetLevel(level hclog.Level) {
	switch level {
	case hclog.Off, hclog.NoLevel:
		l.l.SetLevel(NilLevel)
	case hclog.Trace:
		l.l.SetLevel(DebugLevel)
	case hclog.Debug:
		l.l.SetLevel(DebugLevel)
	case hclog.Info:
		l.l.SetLevel(InfoLevel)
	case hclog.Warn:
		l.l.SetLevel(WarnLevel)
	case hclog.Error:
		l.l.SetLevel(ErrorLevel)
	}
}

func (l *_hclog) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	var lvl Level
	switch opts.ForceLevel {
	case hclog.Off, hclog.NoLevel:
		lvl = NilLevel
	case hclog.Trace:
		lvl = DebugLevel
	case hclog.Debug:
		lvl = DebugLevel
	case hclog.Info:
		lvl = InfoLevel
	case hclog.Warn:
		lvl = WarnLevel
	case hclog.Error:
		lvl = ErrorLevel
	}

	return l.l.GetStdLogger(lvl, 0)
}

func (l *_hclog) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return l.l
}
