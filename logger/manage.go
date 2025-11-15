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
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	libctx "github.com/nabbar/golib/context"
	iotclo "github.com/nabbar/golib/ioutils/mapCloser"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	logfil "github.com/nabbar/golib/logger/hookfile"
	logerr "github.com/nabbar/golib/logger/hookstderr"
	logout "github.com/nabbar/golib/logger/hookstdout"
	logsys "github.com/nabbar/golib/logger/hooksyslog"
	loglvl "github.com/nabbar/golib/logger/level"
	logtps "github.com/nabbar/golib/logger/types"
	"github.com/sirupsen/logrus"
)

func (o *logger) switchCloser(c iotclo.Closer) {
	if o == nil {
		return
	} else if c == nil {
		c = o.newCloser()
	}

	i := o.c.Swap(c)

	if i == nil {
		return
	} else if v, k := i.(iotclo.Closer); k && v != nil {
		go func() {
			// temp waiting all still calling log finish
			// Increased delay to allow pending logs to complete
			time.Sleep(200 * time.Millisecond)
			_ = v.Close()
		}()
	}
}

func (o *logger) newCloser() iotclo.Closer {
	if o == nil || o.x == nil {
		return nil
	}

	return iotclo.New(o.x.GetContext())
}

func (o *logger) hasCloser() bool {
	if o == nil || o.x == nil {
		return false
	}

	if i := o.c.Load(); i != nil {
		if _, k := i.(iotclo.Closer); k {
			return true
		}
	}

	return false
}

func (o *logger) Clone() (Logger, error) {
	if o == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	l := &logger{
		m: sync.RWMutex{},
		x: libctx.New[uint8](o.x),
		f: logfld.New(o.x),
		c: new(atomic.Value),
	}

	l.SetLevel(o.GetLevel())
	l.SetFields(o.GetFields())
	if e := l.SetOptions(o.GetOptions()); e != nil {
		return nil, e
	}

	return l, nil
}

func (o *logger) RegisterFuncUpdateLogger(fct func(log Logger)) {
	o.x.Store(keyFctUpdLog, fct)
}

func (o *logger) runFuncUpdateLogger() {
	if i, l := o.x.Load(keyFctUpdLog); !l {
		return
	} else if f, k := i.(func(log Logger)); !k {
		return
	} else if f == nil {
		return
	} else {
		f(o)
	}
}

func (o *logger) RegisterFuncUpdateLevel(fct func(log Logger)) {
	o.x.Store(keyFctUpdLvl, fct)
}

func (o *logger) runFuncUpdateLevel() {
	if i, l := o.x.Load(keyFctUpdLvl); !l {
		return
	} else if f, k := i.(func(log Logger)); !k {
		return
	} else if f == nil {
		return
	} else {
		f(o)
	}
}

func (o *logger) SetLevel(lvl loglvl.Level) {
	o.x.Store(keyLevel, lvl)
	o.setLogrusLevel(o.GetLevel())
	o.runFuncUpdateLevel()
}

func (o *logger) GetLevel() loglvl.Level {
	if o == nil {
		return loglvl.NilLevel
	} else if o.x == nil {
		return loglvl.NilLevel
	} else if i, l := o.x.Load(keyLevel); !l {
		return loglvl.NilLevel
	} else if v, k := i.(loglvl.Level); !k {
		return loglvl.NilLevel
	} else {
		return v
	}
}

func (o *logger) SetFields(field logfld.Fields) {
	if o == nil {
		return
	}
	o.f.Clean()
	o.f.Merge(field)
}

func (o *logger) GetFields() logfld.Fields {
	if o == nil {
		return logfld.New(context.Background())
	}

	return o.f.Clone()
}

func (o *logger) SetOptions(opt *logcfg.Options) error {
	var (
		lvl = o.GetLevel()
		obj = logrus.New()
		hkl = make([]logtps.Hook, 0)
	)

	o.optionsMerge(opt)

	obj.SetLevel(lvl.Logrus())
	obj.SetFormatter(o.defaultFormatter(nil))
	obj.SetOutput(io.Discard) // Send all logs to nowhere by default

	if opt.Stdout != nil && !opt.Stdout.DisableStandard {
		f := o.defaultFormatter(opt.Stdout)
		l := []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
			logrus.TraceLevel,
		}

		if h, e := logout.New(opt.Stdout, l, f); e != nil {
			return e
		} else {
			hkl = append(hkl, h)
		}

		l = []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}

		if h, e := logerr.New(opt.Stdout, l, f); e != nil {
			return e
		} else {
			hkl = append(hkl, h)
		}
	}

	if len(opt.LogFile) > 0 {
		for _, f := range opt.LogFile {
			if h, e := logfil.New(f, o.defaultFormatterNoColor()); e != nil {
				return e
			} else {
				hkl = append(hkl, h)
			}
		}
	}

	if len(opt.LogSyslog) > 0 {
		for _, s := range opt.LogSyslog {
			if h, e := logsys.New(s, o.defaultFormatterNoColor()); e != nil {
				return e
			} else {
				hkl = append(hkl, h)
			}
		}
	}

	if len(hkl) > 0 {
		var clo = o.newCloser()

		for _, h := range hkl {
			clo.Add(h)
			h.RegisterHook(obj)
			go h.Run(o.x.GetContext())
		}

		// waiting hook has start !!
		for _, h := range hkl {
			for i := 0; i < 10; i++ {
				if h.IsRunning() {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
		}

		o.switchCloser(clo)
	} else if o.hasCloser() {
		o.switchCloser(nil)
	}

	o.x.Store(keyOptions, opt)
	o.x.Store(keyLogrus, obj)
	o.runFuncUpdateLogger()

	return nil
}

func (o *logger) GetOptions() *logcfg.Options {
	if o == nil {
		return &logcfg.Options{}
	} else if o.x == nil {
		return &logcfg.Options{}
	} else if i, l := o.x.Load(keyOptions); !l {
		return &logcfg.Options{}
	} else if v, k := i.(*logcfg.Options); !k {
		return &logcfg.Options{}
	} else {
		return v
	}
}
