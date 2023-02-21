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
	"sync"

	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
)

type componentLog struct {
	m sync.RWMutex
	x libctx.Config[uint8]

	l liblog.Logger
	v liblog.Level
}

func (o *componentLog) Log() liblog.Logger {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.l != nil {
		return o.l.Clone()
	}

	return nil
}

func (o *componentLog) SetLevel(lvl liblog.Level) {
	o.m.Lock()
	defer o.m.Unlock()

	o.v = lvl

	if o.l == nil {
		return
	}

	o.l.SetLevel(lvl)
}

func (o *componentLog) GetLevel() liblog.Level {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.v
}

func (o *componentLog) SetField(fields liblog.Fields) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.l == nil {
		return
	}

	o.l.SetFields(fields)
}

func (o *componentLog) GetField() liblog.Fields {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.l == nil {
		return nil
	}

	return o.l.GetFields()
}

func (o *componentLog) GetOptions() *liblog.Options {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.l == nil {
		return nil
	}

	return o.l.GetOptions()
}

func (o *componentLog) SetOptions(opt *liblog.Options) liberr.Error {
	o.m.Lock()
	defer o.m.Unlock()

	if o.l == nil {
		return ErrorComponentNotInitialized.Error(nil)
	}

	if e := o.l.SetOptions(opt); e != nil {
		return ErrorConfigInvalid.ErrorParent(e)
	}

	return nil
}
