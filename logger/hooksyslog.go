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
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type HookSyslog interface {
	logrus.Hook
	io.WriteCloser
	RegisterHook(log *logrus.Logger)
}

type syslogWrapper interface {
	io.WriteCloser

	Panic(p []byte) (n int, err error)
	Fatal(p []byte) (n int, err error)
	Error(p []byte) (n int, err error)
	Warning(p []byte) (n int, err error)
	Info(p []byte) (n int, err error)
	Debug(p []byte) (n int, err error)
}

type FuncFormatter func() logrus.Formatter

type _HookSyslog struct {
	m sync.Mutex
	w syslogWrapper
	f logrus.Formatter
	l []logrus.Level
	o _HookSyslogOptions
}

type _HookSyslogOptions struct {
	Net NetworkType
	Hst string
	Tag string
	//	Sev SyslogSeverity
	Fac SyslogFacility

	Pid bool
	Tms bool
	Trc bool
	Acc bool
}

func NewHookSyslog(opt OptionsSyslog, format logrus.Formatter) (HookSyslog, error) {
	var (
		LVLs = make([]logrus.Level, 0)
		sys  syslogWrapper
		err  error
	)

	if len(opt.LogLevel) > 0 {
		for _, ls := range opt.LogLevel {
			LVLs = append(LVLs, GetLevelString(ls).Logrus())
		}
	} else {
		LVLs = logrus.AllLevels
	}

	obj := &_HookSyslog{
		m: sync.Mutex{},
		w: sys,
		f: format,
		l: LVLs,
		o: _HookSyslogOptions{
			Net: MakeNetwork(opt.Network),
			Hst: opt.Host,
			Tag: opt.Tag,
			//			Sev: MakeSeverity(opt.Severity),
			Fac: MakeFacility(opt.Facility),
			Pid: opt.DisableStack,
			Tms: opt.DisableTimestamp,
			Trc: opt.EnableTrace,
			Acc: opt.EnableAccessLog,
		},
	}

	if h, e := obj.openCreate(); e != nil {
		return nil, e
	} else {
		_ = h.Close()
	}

	return obj, err
}

func (o *_HookSyslog) openCreate() (syslogWrapper, error) {
	return newSyslog(o.o.Net, o.o.Hst, o.o.Tag, o.o.Fac)
	//return newSyslog(o.o.Net, o.o.Hst, o.o.Tag, o.o.Sev, o.o.Fac)
}

func (o *_HookSyslog) isStack() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.o.Pid
}

func (o *_HookSyslog) isTimeStamp() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.o.Tms
}

func (o *_HookSyslog) isTrace() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.o.Trc
}

func (o *_HookSyslog) isAccessLog() bool {
	o.m.Lock()
	defer o.m.Unlock()

	return o.o.Acc
}

func (o *_HookSyslog) RegisterHook(log *logrus.Logger) {
	log.AddHook(o)
}

func (o *_HookSyslog) Levels() []logrus.Level {
	return o.l
}

func (o *_HookSyslog) Fire(entry *logrus.Entry) error {
	ent := entry.Dup()
	ent.Level = entry.Level

	if !o.isStack() {
		ent.Data = o.filterKey(ent.Data, FieldStack)
	}

	if !o.isTimeStamp() {
		ent.Data = o.filterKey(ent.Data, FieldTime)
	}

	if !o.isTrace() {
		ent.Data = o.filterKey(ent.Data, FieldCaller)
		ent.Data = o.filterKey(ent.Data, FieldFile)
		ent.Data = o.filterKey(ent.Data, FieldLine)
	}

	var (
		p []byte
		e error
	)

	if o.isAccessLog() {
		if len(entry.Message) > 0 {
			if !strings.HasSuffix(entry.Message, "\n") {
				entry.Message += "\n"
			}
			p = []byte(entry.Message)
		} else {
			return nil
		}
	} else {
		if len(ent.Data) < 1 {
			return nil
		} else if p, e = ent.Bytes(); e != nil {
			return e
		}
	}

	if _, e = o.writeLevel(ent.Level, p); e != nil {
		_ = o.Close()
		_, e = o.writeLevel(ent.Level, p)
	}

	return e
}

func (o *_HookSyslog) Write(p []byte) (n int, err error) {
	return o.writeLevel(logrus.InfoLevel, p)
}

func (o *_HookSyslog) writeLevel(lvl logrus.Level, p []byte) (n int, err error) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.w == nil {
		var e error
		if o.w, e = o.openCreate(); e != nil {
			return 0, fmt.Errorf("logrus.hooksyslog: %v", e)
		}
	}

	switch lvl {
	case logrus.PanicLevel:
		return o.w.Panic(p)
	case logrus.FatalLevel:
		return o.w.Fatal(p)
	case logrus.ErrorLevel:
		return o.w.Error(p)
	case logrus.WarnLevel:
		return o.w.Warning(p)
	case logrus.InfoLevel:
		return o.w.Info(p)
	case logrus.DebugLevel:
		return o.w.Debug(p)
	default:
		return o.w.Write(p)
	}
}

func (o *_HookSyslog) Close() error {
	o.m.Lock()
	defer o.m.Unlock()

	var e error

	if o.w != nil {
		e = o.w.Close()
	}

	o.w = nil
	return e
}

func (o *_HookSyslog) filterKey(f logrus.Fields, key string) logrus.Fields {
	if len(f) < 1 {
		return f
	}

	if _, ok := f[key]; !ok {
		return f
	} else {
		delete(f, key)
		return f
	}
}
