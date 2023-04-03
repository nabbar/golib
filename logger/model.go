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
	"bytes"
	"context"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	iotclo "github.com/nabbar/golib/ioutils/mapCloser"
	"github.com/sirupsen/logrus"
)

const (
	keyContext = iota
	KeyCancel
	keyLevel
	keyOptions
	keyLogrus
	keyWriter
	keyFilter

	_TraceFilterMod    = "/pkg/mod/"
	_TraceFilterVendor = "/vendor/"
)

var _selfPackage = path.Base(reflect.TypeOf(logger{}).PkgPath())

type logger struct {
	m sync.RWMutex
	x libctx.Config[uint8] // cf const key...
	f Fields               // fields map
	c iotclo.Closer
}

func defaultFormatter() logrus.TextFormatter {
	return logrus.TextFormatter{
		ForceColors:               false,
		DisableColors:             false,
		ForceQuote:                true,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          true,
		FullTimestamp:             false,
		TimestampFormat:           time.RFC3339,
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    true,
		PadLevelText:              true,
		QuoteEmptyFields:          true,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
	}
}

func (o *logger) defaultFormatter(opt *Options) logrus.Formatter {
	f := defaultFormatter()

	if opt != nil && opt.DisableColor {
		f.ForceColors = false
		f.EnvironmentOverrideColors = false
		f.DisableColors = true
	} else {
		f.ForceColors = true
		f.DisableColors = false
	}

	return &f
}

func (o *logger) defaultFormatterNoColor() logrus.Formatter {
	f := defaultFormatter()
	f.ForceColors = false
	f.EnvironmentOverrideColors = false
	f.DisableColors = true
	return &f
}

func (o *logger) contextNew() context.Context {
	ctx, cnl := context.WithCancel(o.x.GetContext())
	o.x.Store(KeyCancel, cnl)

	return ctx
}

func (o *logger) cancelCall() {
	if i, l := o.x.LoadAndDelete(KeyCancel); !l {
		return
	} else if i == nil {
		return
	} else if c, k := i.(context.CancelFunc); !k {
		return
	} else {
		c()
	}
}

func (o *logger) optionsMerge(opt *Options) {
	if !opt.InheritDefault {
		return
	}

	var no Options

	if opt.opts != nil {
		no = *opt.opts()
	} else if i, l := o.x.Load(keyOptions); !l {
		return
	} else if v, k := i.(*Options); !k {
		return
	} else if v.opts != nil {
		no = *v.opts()
	} else {
		no = *v
	}

	if opt.DisableStandard {
		no.DisableStandard = true
	}

	if opt.DisableStack {
		no.DisableStack = true
	}

	if opt.DisableTimestamp {
		no.DisableTimestamp = true
	}

	if opt.EnableTrace {
		no.EnableTrace = true
	}

	if len(opt.TraceFilter) > 0 {
		no.TraceFilter = opt.TraceFilter
	}

	if opt.DisableColor {
		no.DisableColor = true
	}

	if opt.EnableAccessLog {
		no.EnableAccessLog = true
	}

	if opt.LogFileExtend {
		no.LogFile = append(no.LogFile, opt.LogFile...)
	} else {
		no.LogFile = opt.LogFile
	}

	if opt.LogSyslogExtend {
		no.LogSyslog = append(no.LogSyslog, opt.LogSyslog...)
	} else {
		no.LogSyslog = opt.LogSyslog
	}

	*opt = no
}

func (o *logger) setLogrusLevel(lvl Level) {
	if i, l := o.x.Load(keyLogrus); !l {
		return
	} else if v, k := i.(*logrus.Logger); !k {
		return
	} else {
		v.SetLevel(lvl.Logrus())
		o.x.Store(keyLogrus, v)
	}
}

func (o *logger) getLogrus() *logrus.Logger {
	if i, l := o.x.Load(keyLogrus); !l {
		return nil
	} else if v, k := i.(*logrus.Logger); !k {
		return nil
	} else {
		return v
	}
}

func (o *logger) getStack() uint64 {
	b := make([]byte, 64)

	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	//nolint #nosec
	/* #nosec */
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

func (o *logger) getCaller() runtime.Frame {
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need.
	programCounters := make([]uintptr, 10, 255)
	n := runtime.Callers(1, programCounters)

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		more := true

		for more {
			var frame runtime.Frame
			frame, more = frames.Next()

			if strings.Contains(frame.Function, _selfPackage) {
				continue
			}

			return frame
		}
	}

	return runtime.Frame{Function: "unknown", File: "unknown", Line: 0}
}

func (o *logger) filterPath(pathname string) string {
	pathname = liberr.ConvPathFromLocal(pathname)

	if i := strings.LastIndex(pathname, _TraceFilterMod); i != -1 {
		i = i + len(_TraceFilterMod)
		pathname = pathname[i:]
	}

	if i := strings.LastIndex(pathname, _TraceFilterVendor); i != -1 {
		i = i + len(_TraceFilterVendor)
		pathname = pathname[i:]
	}

	opt := o.GetOptions()

	if opt.TraceFilter != "" {
		if i := strings.LastIndex(pathname, opt.TraceFilter); i != -1 {
			i = i + len(opt.TraceFilter)
			pathname = pathname[i:]
		}
	}

	return strings.Trim(path.Clean(pathname), liberr.PathSeparator)
}
