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

// MonitorAdd adds a monitor to the pool.
// If the pool is running and the monitor is not, it will be started automatically.
// Returns an error if the monitor is nil or has an empty name.
func (o *pool) MonitorAdd(mon montps.Monitor) error {
	if mon == nil {
		return nil
	} else if len(mon.Name()) < 1 {
		return fmt.Errorf("monitor name cannot be empty")
	}

	if o.IsRunning() && !mon.IsRunning() {
		if e := mon.Start(o.p.GetContext()); e != nil {
			return e
		}
	}

	o.p.Store(mon.Name(), mon)
	return nil
}

// MonitorGet retrieves a monitor from the pool by name.
// Returns nil if the name is empty or the monitor is not found.
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

// MonitorSet updates or adds a monitor in the pool.
// If the monitor doesn't exist, it will be added.
// Returns an error if the monitor is nil or has an empty name.
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

// MonitorDel removes a monitor from the pool by name.
// Does nothing if the name is empty.
func (o *pool) MonitorDel(name string) {
	if len(name) < 1 {
		return
	} else {
		o.p.Delete(name)
	}
}

// MonitorList returns a list of all monitor names in the pool.
func (o *pool) MonitorList() []string {
	var res = make([]string, 0)

	o.MonitorWalk(func(name string, val montps.Monitor) bool {
		res = append(res, name)
		return true
	})

	return res
}

// MonitorWalk iterates over all monitors in the pool, calling the provided function for each.
// The function should return true to continue iteration or false to stop.
// If validName is provided, only monitors with those names will be iterated.
func (o *pool) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	o.p.WalkLimit(func(key string, val interface{}) bool {
		if mon, ok := val.(montps.Monitor); !ok {
			o.p.Delete(key)
			return true
		} else {
			return fct(key, mon)
		}
	}, validName...)
}
