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

package aws

import (
	"context"
	"fmt"
	"runtime"

	libreq "github.com/nabbar/golib/request"

	libaws "github.com/nabbar/golib/aws"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
)

const (
	defaultNameMonitor = "AWS Client"
)

func (o *componentAws) RegisterMonitorPool(fct montps.FuncPool) {
	o.m.Lock()
	defer o.m.Unlock()

	o.p = fct
}

func (o *componentAws) _getMonitorPool() montps.Pool {
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

func (o *componentAws) _registerMonitor(opt *libreq.OptionsHealth, aws libaws.Config) error {
	var (
		e   error
		key = o._getKey()
		inf moninf.Info
		mon montps.Monitor
		res = make(map[string]interface{}, 0)
		vrs = o._getVersion()
	)

	if o._getMonitorPool() == nil {
		return nil
	} else if len(key) < 1 {
		return ErrorComponentNotInitialized.Error(nil)
	} else if opt == nil {
		return nil
	} else if aws == nil {
		return ErrorConfigInvalid.Error(nil)
	}

	res["runtime"] = runtime.Version()[2:]

	if vrs != nil {
		res["release"] = vrs.GetRelease()
		res["build"] = vrs.GetBuild()
		res["date"] = vrs.GetDate()
		res["endpoint"] = aws.GetEndpoint().Host
		res["region"] = aws.GetRegion()
		res["health"] = o._getEndpoint(opt, aws)
	}

	if inf, e = moninf.New(defaultNameMonitor); e != nil {
		return e
	} else if vrs != nil {
		inf.RegisterName(func() (string, error) {
			return fmt.Sprintf("%s %s", defaultNameMonitor, o._getKey()), nil
		})
		inf.RegisterInfo(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon = o._getMonitor(key, inf); mon == nil {
		if mon, e = o._newMonitor(inf); e != nil {
			return e
		} else if mon == nil {
			return nil
		}
	}

	if opt.Monitor.Name != key {
		opt.Monitor.Name = key
	}

	if e = mon.SetConfig(o.x.GetContext, opt.Monitor); e != nil {
		return e
	}

	mon.SetHealthCheck(o.HealthCheck)

	if e = mon.Restart(o.x.GetContext()); e != nil {
		return e
	} else if e = o._setMonitor(mon); e != nil {
		return e
	}

	return nil
}

func (o *componentAws) _getEndpoint(opt *libreq.OptionsHealth, aws libaws.Config) string {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.r != nil && len(opt.Endpoint) > 0 {
		return opt.Endpoint
	} else {
		return aws.GetEndpoint().Host
	}
}

func (o *componentAws) _newMonitor(inf montps.Info) (montps.Monitor, error) {
	if c, e := libmon.New(o.x.GetContext, inf); e != nil {
		return nil, e
	} else if c != nil {
		c.RegisterLoggerDefault(o.getLogger)
		return c, nil
	} else {
		return c, nil
	}
}

func (o *componentAws) _getMonitor(key string, inf montps.Info) montps.Monitor {
	var (
		mon libmon.Monitor
		pol = o._getMonitorPool()
	)

	if pol == nil {
		return nil
	}

	mon = pol.MonitorGet(key)

	if mon != nil {
		mon.InfoUpd(inf)
		mon.RegisterLoggerDefault(o.getLogger)
	}

	return mon
}

func (o *componentAws) _setMonitor(mon montps.Monitor) error {
	var pol = o._getMonitorPool()

	if pol == nil {
		return nil
	}

	return pol.MonitorSet(mon)
}

func (o *componentAws) HealthCheck(ctx context.Context) error {
	o.m.RLock()
	defer o.m.RUnlock()

	if !o.IsStarted() {
		return fmt.Errorf("component not started")
	} else if o.r == nil {
		return o.a.Config().Check(ctx)
	} else {
		return o.r.HealthCheck(ctx)
	}
}
