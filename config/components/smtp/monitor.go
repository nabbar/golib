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

func (o *mod) GetMonitorNames() []string {
	if o.getMonPool() == nil {
		return nil
	}

	if i, l := o.x.Load(keyMonNames); l && i != nil {
		if v, k := i.([]string); k && len(v) > 0 {
			return v
		}
	}

	if key := o.getKey(); len(key) > 0 {
		return []string{key}
	}

	return nil
}

func (o *mod) getMonPool() montps.Pool {
	if i, l := o.x.Load(keyFctMonitorPool); !l || i == nil {
		return nil
	} else if f, k := i.(montps.FuncPool); !k || f == nil {
		return nil
	} else {
		return f()
	}
}

func (o *mod) regMonitor(cfg *montps.Config) error {
	var (
		e   error
		key = o.getKey()
		mon montps.Monitor
		vrs = o.getVersion()
	)

	if o.getMonPool() == nil {
		return nil
	} else if len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	} else if cfg == nil {
		return ErrorConfigInvalid.Error(nil)
	} else if !o.IsStarted() {
		return ErrorComponentStart.Error(nil)
	}

	if mon = o.getMonitor(key); mon == nil {
		if mon, e = o.newMonitor(o.x, vrs); e != nil {
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
	} else if e = o.setMonitor(mon); e != nil {
		return e
	}

	return nil
}

func (o *mod) newMonitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	if s := o.s.Load(); s != nil {
		return s.Monitor(ctx, vrs)
	}

	return nil, ErrorComponentNotInitialized.Error(nil)
}

func (o *mod) getMonitor(key string) montps.Monitor {
	var (
		mon montps.Monitor
		pol = o.getMonPool()
	)

	if pol == nil {
		return nil
	}

	mon = pol.MonitorGet(key)
	return mon
}

func (o *mod) setMonitor(mon montps.Monitor) error {
	var pol = o.getMonPool()

	if pol == nil {
		return nil
	}

	o.x.Store(keyMonNames, []string{mon.Name()})

	return pol.MonitorSet(mon)
}
