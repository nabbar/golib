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

package request

import (
	"context"

	montps "github.com/nabbar/golib/monitor/types"
	libreq "github.com/nabbar/golib/request"
	libver "github.com/nabbar/golib/version"
)

func (o *componentRequest) _registerMonitor(cfg *libreq.Options) error {
	var (
		e   error
		key = o._getKey()
		mon montps.Monitor
		vrs = o._getVersion()
		ctx = o._getContext()
	)

	if o.getPool() == nil {
		return nil
	} else if len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	} else if cfg == nil {
		return ErrorConfigInvalid.Error(nil)
	} else if !o.IsStarted() {
		return ErrorComponentStart.Error(nil)
	} else if ctx == nil {
		ctx = context.Background()
	}

	if mon, e = o._newMonitor(ctx, vrs); e != nil {
		return e
	}

	if m := o._getMonitor(mon.Name()); m != nil {
		mon = m
	}

	if mon == nil {
		return nil
	}

	mon.RegisterLoggerDefault(o.getLogger)

	if cfg.Health.Monitor.Name != key {
		cfg.Health.Monitor.Name = key
	}

	if e = mon.SetConfig(o.x.GetContext, cfg.Health.Monitor); e != nil {
		return e
	}

	if e = mon.Restart(o.x.GetContext()); e != nil {
		return e
	} else if e = o._setMonitor(mon); e != nil {
		return e
	}

	return nil
}

func (o *componentRequest) _newMonitor(ctx context.Context, vrs libver.Version) (montps.Monitor, error) {
	return o.getRequest().Monitor(ctx, vrs)
}

func (o *componentRequest) _getMonitor(key string) montps.Monitor {
	var (
		mon montps.Monitor
		pol = o.getPool()
	)

	if pol == nil {
		return nil
	}

	mon = pol.MonitorGet(key)
	return mon
}

func (o *componentRequest) _setMonitor(mon montps.Monitor) error {
	var pol = o.getPool()

	if pol == nil {
		return nil
	}

	return pol.MonitorSet(mon)
}
