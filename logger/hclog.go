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

func (o *_hclog) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel, hclog.Off:
		return
	case hclog.Trace:
		o.l.Debug(msg, nil, args...)
	case hclog.Debug:
		o.l.Debug(msg, nil, args...)
	case hclog.Info:
		o.l.Info(msg, nil, args...)
	case hclog.Warn:
		o.l.Warning(msg, nil, args...)
	case hclog.Error:
		o.l.Error(msg, nil, args...)
	}
}

func (o *_hclog) Trace(msg string, args ...interface{}) {
	o.l.Debug(msg, nil, args...)
}

func (o *_hclog) Debug(msg string, args ...interface{}) {
	o.l.Debug(msg, nil, args...)
}

func (o *_hclog) Info(msg string, args ...interface{}) {
	o.l.Info(msg, nil, args...)
}

func (o *_hclog) Warn(msg string, args ...interface{}) {
	o.l.Warning(msg, nil, args...)
}

func (o *_hclog) Error(msg string, args ...interface{}) {
	o.l.Error(msg, nil, args...)
}

func (o *_hclog) IsTrace() bool {
	return o.l.GetOptions().EnableTrace
}

func (o *_hclog) IsDebug() bool {
	return o.l.GetLevel() >= DebugLevel
}

func (o *_hclog) IsInfo() bool {
	return o.l.GetLevel() >= InfoLevel
}

func (o *_hclog) IsWarn() bool {
	return o.l.GetLevel() >= WarnLevel
}

func (o *_hclog) IsError() bool {
	return o.l.GetLevel() >= ErrorLevel
}

func (o *_hclog) ImpliedArgs() []interface{} {
	fields := o.l.GetFields()

	if i, l := fields.Load(HCLogArgs); !l {
		return make([]interface{}, 0)
	} else if v, k := i.([]interface{}); k {
		return v
	}

	return make([]interface{}, 0)
}

func (o *_hclog) With(args ...interface{}) hclog.Logger {
	o.l.SetFields(o.l.GetFields().Add(HCLogArgs, args))
	return o
}

func (o *_hclog) Name() string {
	fields := o.l.GetFields()

	if i, l := fields.Load(HCLogName); !l {
		return ""
	} else if v, k := i.(string); k {
		return v
	}

	return ""
}

func (o *_hclog) Named(name string) hclog.Logger {
	o.l.SetFields(o.l.GetFields().Add(HCLogName, name))
	return o
}

func (o *_hclog) ResetNamed(name string) hclog.Logger {
	o.l.SetFields(o.l.GetFields().Add(HCLogName, name))
	return o
}

func (o *_hclog) SetLevel(level hclog.Level) {
	switch level {
	case hclog.NoLevel, hclog.Off:
		o.l.SetLevel(NilLevel)
	case hclog.Trace:
		opt := o.l.GetOptions()
		opt.EnableTrace = true
		_ = o.l.SetOptions(opt)
		o.l.SetLevel(DebugLevel)
	case hclog.Debug:
		o.l.SetLevel(DebugLevel)
	case hclog.Info:
		o.l.SetLevel(InfoLevel)
	case hclog.Warn:
		o.l.SetLevel(WarnLevel)
	case hclog.Error:
		o.l.SetLevel(ErrorLevel)
	}
}

func (o *_hclog) GetLevel() hclog.Level {
	switch o.l.GetLevel() {
	case NilLevel:
		return hclog.NoLevel
	case DebugLevel:
		opt := o.l.GetOptions()
		if opt.EnableTrace {
			return hclog.Trace
		} else {
			return hclog.Debug
		}
	case InfoLevel:
		return hclog.Info
	case WarnLevel:
		return hclog.Warn
	case ErrorLevel:
		return hclog.Error
	default:
		return hclog.Off
	}
}

func (o *_hclog) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	var lvl Level
	switch opts.ForceLevel {
	case hclog.NoLevel, hclog.Off:
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

	return o.l.GetStdLogger(lvl, 0)
}

func (o *_hclog) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return o.l
}
