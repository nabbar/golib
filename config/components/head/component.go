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

package head

import (
	cfgtps "github.com/nabbar/golib/config/types"
	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

const (
	ComponentType = "head"

	keyCptKey = iota + 1
	keyCptDependencies
	keyFctViper
	keyFctGetCpt
	keyCptVersion
	keyCptLogger
	keyFctStaBef
	keyFctStaAft
	keyFctRelBef
	keyFctRelAft
	keyFctMonitorPool
)

func (o *componentHead) Type() string {
	return ComponentType
}

func (o *componentHead) Init(key string, ctx libctx.FuncContext, get cfgtps.FuncCptGet, vpr libvpr.FuncViper, vrs libver.Version, log liblog.FuncLog) {
	o.m.Lock()
	defer o.m.Unlock()

	if o.x == nil {
		o.x = libctx.NewConfig[uint8](ctx)
	} else {
		x := libctx.NewConfig[uint8](ctx)
		x.Merge(o.x)
		o.x = x
	}

	o.x.Store(keyCptKey, key)
	o.x.Store(keyFctGetCpt, get)
	o.x.Store(keyFctViper, vpr)
	o.x.Store(keyCptVersion, vrs)
	o.x.Store(keyCptLogger, log)
}

func (o *componentHead) RegisterFuncStart(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctStaBef, before)
	o.x.Store(keyFctStaAft, after)
}

func (o *componentHead) RegisterFuncReload(before, after cfgtps.FuncCptEvent) {
	o.x.Store(keyFctRelBef, before)
	o.x.Store(keyFctRelAft, after)
}

func (o *componentHead) IsStarted() bool {
	o.m.RLock()
	defer o.m.RUnlock()

	return o.h != nil
}

func (o *componentHead) IsRunning() bool {
	return o.IsStarted()
}

func (o *componentHead) Start() error {
	return o._run()
}

func (o *componentHead) Reload() error {
	return o._run()
}

func (o *componentHead) Stop() {
	o.m.Lock()
	defer o.m.Unlock()

	o.h = nil
	return
}

func (o *componentHead) Dependencies() []string {
	o.m.RLock()
	defer o.m.RUnlock()

	var def = make([]string, 0)

	if o == nil {
		return def
	} else if o.x == nil {
		return def
	} else if i, l := o.x.Load(keyCptDependencies); !l {
		return def
	} else if v, k := i.([]string); !k {
		return def
	} else if len(v) > 0 {
		return v
	} else {
		return def
	}
}

func (o *componentHead) SetDependencies(d []string) error {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.x == nil {
		return ErrorComponentNotInitialized.Error(nil)
	} else {
		o.x.Store(keyCptDependencies, d)
		return nil
	}
}

func (o *componentHead) getLogger() liblog.Logger {
	if i, l := o.x.Load(keyCptLogger); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else {
		return v()
	}
}
