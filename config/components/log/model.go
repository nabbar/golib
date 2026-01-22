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

package log

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	logfld "github.com/nabbar/golib/logger/fields"
	loglvl "github.com/nabbar/golib/logger/level"
)

type mod struct {
	x libctx.Config[uint8]
	l libatm.Value[liblog.Logger]
	r *atomic.Bool
	v *atomic.Uint32
}

func (o *mod) Log() liblog.Logger {
	if i := o.getLog(); i == nil {
		return nil
	} else {
		return i
	}
}

func (o *mod) LogClone() liblog.Logger {
	if i := o.getLog(); i == nil {
		return nil
	} else if l, e := i.Clone(); e != nil {
		return nil
	} else {
		return l
	}
}

func (o *mod) SetLevel(lvl loglvl.Level) {
	o.v.Store(lvl.Uint32())

	if i := o.getLog(); i == nil {
		return
	} else {
		i.SetLevel(lvl)
		o.setLog(i)
	}
}

func (o *mod) GetLevel() loglvl.Level {
	if i := o.getLog(); i == nil {
		return loglvl.ParseFromUint32(o.v.Load())
	} else if l := i.GetLevel(); l.Uint32() != o.v.Load() {
		o.SetLevel(l)
	}
	return loglvl.ParseFromUint32(o.v.Load())
}

func (o *mod) SetField(fields logfld.Fields) {
	if i := o.getLog(); i == nil {
		return
	} else {
		i.SetFields(fields)
		o.setLog(i)
	}
}

func (o *mod) GetField() logfld.Fields {
	if i := o.getLog(); i == nil {
		return logfld.New(o.x)
	} else {
		return i.GetFields()
	}
}

func (o *mod) GetOptions() *logcfg.Options {
	if i := o.getLog(); i == nil {
		return nil
	} else {
		return i.GetOptions()
	}
}

func (o *mod) SetOptions(opt *logcfg.Options) error {
	if i := o.getLog(); i == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else {
		e := i.SetOptions(opt)
		o.setLog(i)
		return ErrorConfigInvalid.IfError(e)
	}
}

func (o *mod) getLog() liblog.Logger {
	if !o.r.Load() {
		return nil
	} else {
		return o.l.Load()
	}
}

func (o *mod) setLog(l liblog.Logger) {
	if l == nil {
		o.r.Store(false)
	} else {
		o.r.Store(true)
		o.l.Store(l)
	}
}
