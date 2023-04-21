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
	"io"
	"sync"

	iotclo "github.com/nabbar/golib/ioutils/mapCloser"
	"github.com/sirupsen/logrus"
)

func (o *logger) Clone() Logger {
	if o == nil {
		return nil
	}

	o.m.RLock()
	defer o.m.RUnlock()

	return &logger{
		x: o.x.Clone(nil),
		m: sync.RWMutex{},
		f: o.f.FieldsClone(nil),
		c: o.c.Clone(),
	}
}

func (o *logger) SetLevel(lvl Level) {
	o.x.Store(keyLevel, lvl)
	o.setLogrusLevel(o.GetLevel())

	if opt := o.GetOptions(); opt.change != nil {
		opt.change(o)
	}
}

func (o *logger) GetLevel() Level {
	if o == nil {
		return NilLevel
	} else if o.x == nil {
		return NilLevel
	} else if i, l := o.x.Load(keyLevel); !l {
		return NilLevel
	} else if v, k := i.(Level); !k {
		return NilLevel
	} else {
		return v
	}
}

func (o *logger) SetFields(field Fields) {
	if o == nil {
		return
	}

	if field != nil {
		o.m.Lock()
		defer o.m.Unlock()
		o.f = field
	}
}

func (o *logger) GetFields() Fields {
	if o == nil {
		return NewFields(context.Background)
	}

	o.m.RLock()
	defer o.m.RUnlock()
	return o.f.FieldsClone(nil)
}

func (o *logger) SetOptions(opt *Options) error {
	var (
		ctx, cnl = context.WithCancel(o.x.GetContext())
		lvl      = o.GetLevel()
		obj      = logrus.New()
	)

	defer cnl()

	o.optionsMerge(opt)

	obj.SetLevel(lvl.Logrus())
	obj.SetFormatter(o.defaultFormatter(opt))
	obj.SetOutput(io.Discard) // Send all logs to nowhere by default

	clo := iotclo.New(func() context.Context {
		return ctx
	})

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
			if hook, err := NewHookFile(fopt, o.defaultFormatterNoColor()); err != nil {
				return err
			} else {
				clo.Add(hook)
				hook.RegisterHook(obj)
			}
		}
	}

	if len(opt.LogSyslog) > 0 {
		for _, lopt := range opt.LogSyslog {
			if hook, err := NewHookSyslog(lopt, o.defaultFormatterNoColor()); err != nil {
				return err
			} else {
				clo.Add(hook)
				hook.RegisterHook(obj)
			}
		}
	}

	o.x.Store(keyOptions, opt)
	o.x.Store(keyLogrus, obj)

	_ = o.Close()
	o.c.Add(clo.Get()...)
	clo.Clean()

	if opt.init != nil {
		opt.init(o)
	}

	return nil
}

func (o *logger) GetOptions() *Options {
	if o == nil {
		return &Options{}
	} else if o.x == nil {
		return &Options{}
	} else if i, l := o.x.Load(keyOptions); !l {
		return &Options{}
	} else if v, k := i.(*Options); !k {
		return &Options{}
	} else {
		return v
	}
}
