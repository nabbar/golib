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

	libaws "github.com/nabbar/golib/aws"
	libmon "github.com/nabbar/golib/monitor"
	moninf "github.com/nabbar/golib/monitor/info"
	montps "github.com/nabbar/golib/monitor/types"
	libreq "github.com/nabbar/golib/request"
)

const (
	defaultNameMonitor = "AWS Client"
)

func (o *mod) RegisterMonitorPool(fct montps.FuncPool) {
	if fct == nil {
		fct = func() montps.Pool {
			return nil
		}
	}

	o.p.Store(fct)
}

func (o *mod) GetMonitorNames() []string {
	if o.getPool() == nil {
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

func (o *mod) HealthCheck(ctx context.Context) error {
	if cli := o.GetAws(); cli == nil {
		return fmt.Errorf("component not started")
	} else if req := o.getRequest(); req == nil {
		return cli.Config().Check(ctx)
	} else {
		return req.HealthCheck(ctx)
	}
}

func (o *mod) getPool() montps.Pool {
	if i := o.p.Load(); i == nil {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if p := v(); p == nil {
		return nil
	} else {
		return p
	}
}

func (o *mod) getRequest() libreq.Request {
	if i := o.r.Load(); i == nil {
		return nil
	} else if v, k := i.(libreq.Request); !k {
		return nil
	} else {
		return v
	}
}

func (o *mod) setRequest(req libreq.Request) {
	if req != nil {
		o.r.Store(req)
	}
}

func (o *mod) getEndpoint(opt *libreq.OptionsHealth, aws libaws.Config) string {
	if req := o.getRequest(); req != nil && len(opt.Endpoint) > 0 {
		return opt.Endpoint
	} else {
		return aws.GetEndpoint().Host
	}
}

func (o *mod) regMonitor(opt *libreq.OptionsHealth, aws libaws.Config) error {
	var (
		e   error
		key = o.getKey()
		inf moninf.Info
		mon montps.Monitor
		res = make(map[string]interface{}, 0)
		vrs = o.getVersion()
	)

	if o.getPool() == nil {
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
		res["health"] = o.getEndpoint(opt, aws)
	}

	if inf, e = moninf.New(defaultNameMonitor); e != nil {
		return e
	} else if vrs != nil {
		inf.RegisterName(func() (string, error) {
			return fmt.Sprintf("%s %s", defaultNameMonitor, o.getKey()), nil
		})
		inf.RegisterData(func() (map[string]interface{}, error) {
			return res, nil
		})
	}

	if mon = o.getMonitor(key, inf); mon == nil {
		if mon, e = o.newMonitor(inf); e != nil {
			return e
		} else if mon == nil {
			return nil
		}
	}

	if opt.Monitor.Name != key {
		opt.Monitor.Name = key
	}

	if e = mon.SetConfig(o.x, opt.Monitor); e != nil {
		return e
	}

	mon.SetHealthCheck(o.HealthCheck)

	if e = mon.Restart(o.x.GetContext()); e != nil {
		return e
	} else if e = o.setMonitor(mon); e != nil {
		return e
	}

	return nil
}

func (o *mod) newMonitor(inf montps.Info) (montps.Monitor, error) {
	if c, e := libmon.New(o.x, inf); e != nil {
		return nil, e
	} else if c != nil {
		c.RegisterLoggerDefault(o.getLogger)
		return c, nil
	} else {
		return c, nil
	}
}

func (o *mod) getMonitor(key string, inf montps.Info) montps.Monitor {
	var (
		mon libmon.Monitor
		pol = o.getPool()
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

func (o *mod) setMonitor(mon montps.Monitor) error {
	var pol = o.getPool()

	if pol == nil {
		return nil
	}

	o.x.Store(keyMonNames, []string{mon.Name()})

	return pol.MonitorSet(mon)
}
