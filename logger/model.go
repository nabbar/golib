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
	"io/ioutil"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	liberr "github.com/nabbar/golib/errors"

	"github.com/sirupsen/logrus"
)

const (
	_TraceFilterMod    = "/pkg/mod/"
	_TraceFilterVendor = "/vendor/"
)

var _selfPackage = path.Base(reflect.TypeOf(logger{}).PkgPath())

type logger struct {
	x context.Context
	n context.CancelFunc
	m sync.Locker
	l *atomic.Value //current level set for this logger
	o *atomic.Value //options
	s *atomic.Value //logrus logger
	f *atomic.Value //defaults fields
	w *atomic.Value //io writer level
	c _Closer       // closer
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

func (l *logger) defaultFormatter(opt *Options) logrus.Formatter {
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

func (l *logger) defaultFormatterNoColor() logrus.Formatter {
	f := defaultFormatter()
	f.ForceColors = false
	f.EnvironmentOverrideColors = false
	f.DisableColors = true
	return &f
}

func (l *logger) Clone() (Logger, error) {
	c := &logger{
		x: l.contextGet(),
		n: nil,
		m: &sync.Mutex{},
		l: new(atomic.Value),
		o: new(atomic.Value),
		s: new(atomic.Value),
		f: new(atomic.Value),
		w: new(atomic.Value),
		c: _NewCloser(),
	}

	c.setLoggerMutex(l.GetLevel())
	c.SetIOWriterLevel(l.GetIOWriterLevel())
	c.SetFields(l.GetFields())

	if err := c.SetOptions(l.GetOptions()); err != nil {
		return nil, err
	}

	return c, nil
}

func (l *logger) contextGet() context.Context {
	l.m.Lock()
	defer l.m.Unlock()

	if l.x == nil {
		l.x = context.Background()
	}

	return l.x
}

func (l *logger) contextNew() context.Context {
	ctx, cnl := context.WithCancel(l.contextGet())

	l.m.Lock()
	l.n = cnl
	l.m.Unlock()

	return ctx
}

func (l *logger) cancelCall() {
	l.m.Lock()
	defer l.m.Unlock()

	if l.n == nil {
		return
	}

	l.n()
	l.n = nil
}

func (l *logger) cancelClear() {
	l.m.Lock()

	if l.n == nil {
		l.m.Unlock()
		return
	}

	l.n()
	l.m.Unlock()
}

func (l *logger) setLoggerMutex(lvl Level) {
	if l == nil {
		return
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.l == nil {
		l.l = new(atomic.Value)
	}

	l.l.Store(lvl)
}

func (l *logger) SetLevel(lvl Level) {
	l.setLoggerMutex(lvl)
	l.setLogrusLevel(l.GetLevel())

	if opt := l.GetOptions(); opt.change != nil {
		opt.change(l)
	}
}

func (l *logger) GetLevel() Level {
	if l == nil {
		return NilLevel
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.l == nil {
		l.l = new(atomic.Value)
	}

	if i := l.l.Load(); i == nil {
		return NilLevel
	} else if o, ok := i.(Level); ok {
		return o
	}

	return NilLevel
}

func (l *logger) SetFields(field Fields) {
	if l == nil {
		return
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.f == nil {
		l.f = new(atomic.Value)
	}

	l.f.Store(field)
}

func (l *logger) GetFields() Fields {
	if l == nil {
		return NewFields()
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.f == nil {
		l.f = new(atomic.Value)
	}

	if i := l.f.Load(); i == nil {
		return NewFields()
	} else if o, ok := i.(Fields); ok {
		return o
	}

	return NewFields()
}

func (l *logger) setOptionsMutex(opt *Options) error {
	lvl := l.GetLevel()

	go func() {
		var ctx = l.contextNew()

		defer func() {
			l.cancelClear()
		}()

		select {
		case <-ctx.Done():
			_ = l.Close()
			return
		}
	}()

	obj := logrus.New()
	obj.SetLevel(lvl.Logrus())
	obj.SetFormatter(l.defaultFormatter(opt))
	obj.SetOutput(ioutil.Discard) // Send all logs to nowhere by default
	clo := _NewCloser()

	if !opt.DisableStandard {
		obj.AddHook(NewHookStandard(*opt, StdOut, []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
			logrus.TraceLevel,
		}))

		obj.AddHook(NewHookStandard(*opt, StdErr, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}))
	}

	if len(opt.LogFile) > 0 {
		for _, fopt := range opt.LogFile {
			if hook, err := NewHookFile(fopt, l.defaultFormatterNoColor()); err != nil {
				return err
			} else {
				clo.Add(hook)
				hook.RegisterHook(obj)
			}
		}
	}

	if len(opt.LogSyslog) > 0 {
		for _, lopt := range opt.LogSyslog {
			if hook, err := NewHookSyslog(lopt, l.defaultFormatterNoColor()); err != nil {
				return err
			} else {
				clo.Add(hook)
				hook.RegisterHook(obj)
			}
		}
	}

	l.setOptions(opt)

	l.m.Lock()
	defer l.m.Unlock()

	_ = l.c.Close()
	l.c = clo

	l.s = new(atomic.Value)
	l.s.Store(obj)

	return nil
}

func (l *logger) SetOptions(opt *Options) error {
	if err := l.setOptionsMutex(opt); err != nil {
		return err
	}

	if opt.init != nil {
		opt.init(l)
	}

	return nil
}

func (l *logger) GetOptions() *Options {
	if l == nil {
		return &Options{}
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.o == nil {
		l.o = new(atomic.Value)
	}

	if i := l.o.Load(); i == nil {
		return &Options{}
	} else if o, ok := i.(*Options); ok {
		return o
	}

	return &Options{}
}

func (l *logger) setOptions(opt *Options) {
	if l == nil {
		return
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.o == nil {
		l.o = new(atomic.Value)
	}

	l.o.Store(opt)
}

func (l *logger) setLogrusLevel(lvl Level) {
	if _log := l.getLog(); _log != nil {
		_log.SetLevel(lvl.Logrus())

		l.m.Lock()
		defer l.m.Unlock()

		if l.s == nil {
			return
		}

		l.s.Store(_log)
	}
}

func (l *logger) getLog() *logrus.Logger {
	if l == nil {
		return nil
	}

	l.m.Lock()
	defer l.m.Unlock()

	if l.s == nil {
		l.s = new(atomic.Value)
	}

	if i := l.s.Load(); i == nil {
		return nil
	} else if o, ok := i.(*logrus.Logger); !ok {
		return nil
	} else {
		return o
	}
}

func (l *logger) getStack() uint64 {
	b := make([]byte, 64)

	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	//nolint #nosec
	/* #nosec */
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

func (l *logger) getCaller() runtime.Frame {
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

func (l *logger) filterPath(pathname string) string {
	pathname = liberr.ConvPathFromLocal(pathname)

	if i := strings.LastIndex(pathname, _TraceFilterMod); i != -1 {
		i = i + len(_TraceFilterMod)
		pathname = pathname[i:]
	}

	if i := strings.LastIndex(pathname, _TraceFilterVendor); i != -1 {
		i = i + len(_TraceFilterVendor)
		pathname = pathname[i:]
	}

	opt := l.GetOptions()

	if opt.TraceFilter != "" {
		if i := strings.LastIndex(pathname, opt.TraceFilter); i != -1 {
			i = i + len(opt.TraceFilter)
			pathname = pathname[i:]
		}
	}

	return strings.Trim(path.Clean(pathname), liberr.PathSeparator)
}
