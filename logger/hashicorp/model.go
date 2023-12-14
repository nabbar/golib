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

package hashicorp

import (
	"io"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

const (
	HCLogArgs = "hclog.args"
	HCLogName = "hclog.name"
)

type _hclog struct {
	l liblog.FuncLog
}

func (o *_hclog) Log(level hclog.Level, msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	switch level {
	case hclog.NoLevel, hclog.Off:
		return
	case hclog.Trace:
		lg.Entry(loglvl.DebugLevel, msg, args...).Log()
	case hclog.Debug:
		lg.Entry(loglvl.DebugLevel, msg, args...).Log()
	case hclog.Info:
		lg.Entry(loglvl.InfoLevel, msg, args...).Log()
	case hclog.Warn:
		lg.Entry(loglvl.WarnLevel, msg, args...).Log()
	case hclog.Error:
		lg.Entry(loglvl.ErrorLevel, msg, args...).Log()
	}
}

func (o *_hclog) logger() liblog.Logger {
	if o.l == nil {
		return nil
	} else if lg := o.l(); lg == nil {
		return nil
	} else {
		return lg
	}
}

func (o *_hclog) Trace(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.DebugLevel, msg, args...).Log()
}

func (o *_hclog) Debug(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.DebugLevel, msg, args...).Log()
}

func (o *_hclog) Info(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.InfoLevel, msg, args...).Log()
}

func (o *_hclog) Warn(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.WarnLevel, msg, args...).Log()
}

func (o *_hclog) Error(msg string, args ...interface{}) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	lg.Entry(loglvl.ErrorLevel, msg, args...).Log()
}

func (o *_hclog) IsTrace() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	if opt := lg.GetOptions(); opt == nil {
		return false
	} else if opt.Stdout != nil && opt.Stdout.EnableTrace {
		return true
	} else {
		for _, f := range opt.LogFile {
			if f.EnableTrace {
				return true
			}
		}
		for _, f := range opt.LogSyslog {
			if f.EnableTrace {
				return true
			}
		}
	}

	return false
}

func (o *_hclog) IsDebug() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.DebugLevel
}

func (o *_hclog) IsInfo() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.InfoLevel
}

func (o *_hclog) IsWarn() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.WarnLevel
}

func (o *_hclog) IsError() bool {
	var lg = o.logger()

	if lg == nil {
		return false
	}

	return lg.GetLevel() >= loglvl.ErrorLevel
}

func (o *_hclog) ImpliedArgs() []interface{} {
	var lg = o.logger()

	if lg == nil {
		return make([]interface{}, 0)
	}

	fields := lg.GetFields()

	if i, l := fields.Load(HCLogArgs); !l {
		return make([]interface{}, 0)
	} else if v, k := i.([]interface{}); k {
		return v
	}

	return make([]interface{}, 0)
}

func (o *_hclog) With(args ...interface{}) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogArgs, args))
	return o
}

func (o *_hclog) Name() string {
	var lg = o.logger()

	if lg == nil {
		return ""
	}

	fields := lg.GetFields()

	if i, l := fields.Load(HCLogName); !l {
		return ""
	} else if v, k := i.(string); k {
		return v
	}

	return ""
}

func (o *_hclog) Named(name string) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogName, name))
	return o
}

func (o *_hclog) ResetNamed(name string) hclog.Logger {
	var lg = o.logger()

	if lg == nil {
		return o
	}

	lg.SetFields(lg.GetFields().Add(HCLogName, name))
	return o
}

func (o *_hclog) SetLevel(level hclog.Level) {
	var lg = o.logger()

	if lg == nil {
		return
	}

	switch level {
	case hclog.NoLevel, hclog.Off:
		lg.SetLevel(loglvl.NilLevel)
	case hclog.Trace:
		opt := lg.GetOptions()

		if opt.Stdout != nil {
			opt.Stdout.EnableTrace = true
		}

		if len(opt.LogFile) > 0 {
			for i := range opt.LogFile {
				opt.LogFile[i].EnableTrace = true
			}
		}

		if len(opt.LogSyslog) > 0 {
			for i := range opt.LogSyslog {
				opt.LogSyslog[i].EnableTrace = true
			}
		}

		_ = lg.SetOptions(opt)
		lg.SetLevel(loglvl.DebugLevel)
	case hclog.Debug:
		lg.SetLevel(loglvl.DebugLevel)
	case hclog.Info:
		lg.SetLevel(loglvl.InfoLevel)
	case hclog.Warn:
		lg.SetLevel(loglvl.WarnLevel)
	case hclog.Error:
		lg.SetLevel(loglvl.ErrorLevel)
	}
}

func (o *_hclog) GetLevel() hclog.Level {
	var lg = o.logger()

	if lg == nil {
		return hclog.NoLevel
	}

	switch lg.GetLevel() {
	case loglvl.NilLevel:
		return hclog.NoLevel
	case loglvl.DebugLevel:
		if o.IsTrace() {
			return hclog.Trace
		} else {
			return hclog.Debug
		}
	case loglvl.InfoLevel:
		return hclog.Info
	case loglvl.WarnLevel:
		return hclog.Warn
	case loglvl.ErrorLevel:
		return hclog.Error
	default:
		return hclog.Off
	}
}

func (o *_hclog) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	var lg = o.logger()

	if lg == nil {
		return log.Default()
	}

	var lvl loglvl.Level
	switch opts.ForceLevel {
	case hclog.NoLevel, hclog.Off:
		lvl = loglvl.NilLevel
	case hclog.Trace:
		lvl = loglvl.DebugLevel
	case hclog.Debug:
		lvl = loglvl.DebugLevel
	case hclog.Info:
		lvl = loglvl.InfoLevel
	case hclog.Warn:
		lvl = loglvl.WarnLevel
	case hclog.Error:
		lvl = loglvl.ErrorLevel
	}

	return lg.GetStdLogger(lvl, 0)
}

func (o *_hclog) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	var lg = o.logger()

	if lg == nil {
		return os.Stdout
	}

	return lg
}
