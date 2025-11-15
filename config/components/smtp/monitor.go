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

package smtp

import (
	"context"

	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
)

func (o *mod) RegisterMonitorPool(fct montps.FuncPool) {
	if fct == nil {
		fct = func() montps.Pool {
			return nil
		}
	}

	o.x.Store(keyFctMonitorPool, fct)
}

func (o *mod) _getMonitorPool() montps.Pool {
	if i, l := o.x.Load(keyFctMonitorPool); !l || i == nil {
		return nil
	} else if f, k := i.(montps.FuncPool); !k || f == nil {
		return nil
	} else {
		return f()
	}
}

func (o *mod) _registerMonitor(cfg *montps.Config) error {
	var (
		e   error
		key = o.getKey()
		mon montps.Monitor
		vrs = o.getVersion()
	)

	if o._getMonitorPool() == nil {
		return nil
	} else if len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	} else if cfg == nil {
		return ErrorConfigInvalid.Error(nil)
	} else if !o.IsStarted() {
		return ErrorComponentStart.Error(nil)
	}

	if mon = o._getMonitor(key); mon == nil {
		if mon, e = o._newMonitor(o.x, vrs); e != nil {
			return e
		} else if mon == nil {
			return nil
		}
	}

	mon.RegisterLoggerDefault(o.getLogger)

	if cfg.Name != key {
		cfg.Name = key
	}

	if e = mon.SetConfig(o.x, *cfg); e != nil {
		return e
	}

	if e = mon.Restart(o.x.GetContext()); e != nil {
		return e
	} else if e = o._setMonitor(mon); e != nil {
		return e
	}

	return nil
}

func (o *mod) _newMonitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	if s := o.s.Load(); s != nil {
		return s.Monitor(ctx, vrs)
	}

	return nil, ErrorComponentNotInitialized.Error(nil)
}

func (o *mod) _getMonitor(key string) montps.Monitor {
	var (
		mon montps.Monitor
		pol = o._getMonitorPool()
	)

	if pol == nil {
		return nil
	}

	mon = pol.MonitorGet(key)
	return mon
}

func (o *mod) _setMonitor(mon montps.Monitor) error {
	var pol = o._getMonitorPool()

	if pol == nil {
		return nil
	}

	return pol.MonitorSet(mon)
}
