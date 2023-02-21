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

package pool

import (
	"fmt"

	montps "github.com/nabbar/golib/monitor/types"
)

func (o *pool) MonitorAdd(mon montps.Monitor) error {
	if mon == nil {
		return nil
	} else if len(mon.Name()) < 1 {
		return fmt.Errorf("monitor name cannot be empty")
	}

	if e := o.createMetrics(mon); e != nil {
		return e
	}

	if o.IsRunning() && !mon.IsRunning() {
		if e := mon.Start(o.p.GetContext()); e != nil {
			return e
		}
	}

	o.p.Store(mon.Name(), mon)
	return nil
}

func (o *pool) MonitorGet(name string) montps.Monitor {
	if len(name) < 1 {
		return nil
	} else if i, l := o.p.Load(name); !l {
		return nil
	} else if v, k := i.(montps.Monitor); !k {
		return nil
	} else {
		return v
	}
}

func (o *pool) MonitorSet(mon montps.Monitor) error {
	if mon == nil {
		return fmt.Errorf("nil monitor")
	} else if len(mon.Name()) < 1 {
		return fmt.Errorf("missing monitor name")
	} else if _, l := o.p.Load(mon.Name()); !l {
		return o.MonitorAdd(mon)
	}

	o.p.Store(mon.Name(), mon)
	return nil
}

func (o *pool) MonitorDel(name string) {
	if len(name) < 1 {
		return
	} else if i, l := o.p.LoadAndDelete(name); !l {
		return
	} else if v, k := i.(montps.Monitor); !k {
		return
	} else {
		o.deleteMetrics(v)
	}
}

func (o *pool) MonitorList() []string {
	var res = make([]string, 0)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		res = append(res, name)
		return true
	})

	return res
}

func (o *pool) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	f := func(key string, val interface{}) bool {
		var (
			ok  bool
			mon montps.Monitor
		)

		if mon, ok = val.(montps.Monitor); !ok {
			return true
		} else {
			return fct(key, mon)
		}
	}

	if len(validName) > 0 {
		o.p.WalkLimit(f, validName...)
	} else {
		o.p.Walk(f)
	}
}
