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

package http

import (
	liberr "github.com/nabbar/golib/errors"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

func (o *componentHttp) RegisterMonitorPool(fct montps.FuncPool) {
	o.m.Lock()
	defer o.m.Unlock()

	o.p = fct
}

func (o *componentHttp) _getMonitorPool() montps.Pool {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.p == nil {
		return nil
	} else if p := o.p(); p == nil {
		return nil
	} else {
		return p
	}
}

func (o *componentHttp) _registerMonitor(err liberr.CodeError) error {
	var (
		e   error
		key = o._getKey()
		mon []montps.Monitor
		vrs = o._getVersion()
		ctx = o.x.GetContext
	)

	if o._getMonitorPool() == nil {
		return nil
	} else if len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	} else if !o.IsStarted() {
		return ErrorComponentStart.Error(nil)
	}

	if mon, e = o._newMonitor(vrs); e != nil {
		return err.Error(e)
	} else if mon == nil {
		return nil
	}

	for _, m := range mon {
		if old := o._getMonitor(m.Name()); old != nil {
			old.InfoUpd(m.InfoGet())
			if e = old.SetConfig(ctx, m.GetConfig()); e == nil {
				m = old
			}
		}

		if e = m.Restart(ctx()); e != nil {
			return err.Error(e)
		} else if e = o._setMonitor(m); e != nil {
			return err.Error(e)
		}
	}

	return nil
}

func (o *componentHttp) _newMonitor(vrs libver.Version) ([]montps.Monitor, error) {
	o.m.RLock()
	defer o.m.RUnlock()

	if c, e := o.s.Monitor(vrs); e != nil {
		return nil, e
	} else {
		for k := range c {
			if c[k] != nil {
				c[k].RegisterLoggerDefault(o.getLogger)
			}
		}
		return c, nil
	}
}

func (o *componentHttp) _getMonitor(key string) montps.Monitor {
	var (
		mon montps.Monitor
		pol = o._getMonitorPool()
	)

	if pol == nil {
		return nil
	}

	mon = pol.MonitorGet(key)

	if mon != nil {
		mon.RegisterLoggerDefault(o.getLogger)
	}

	return mon
}

func (o *componentHttp) _setMonitor(mon montps.Monitor) error {
	var pol = o._getMonitorPool()

	if pol == nil {
		return nil
	}

	return pol.MonitorSet(mon)
}
