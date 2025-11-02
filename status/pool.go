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

package status

import (
	"fmt"

	montps "github.com/nabbar/golib/monitor/types"
)

// _getPool retrieves the current monitor pool by calling the registered function.
// Returns nil if no pool function is registered or if the function returns nil.
// This is an internal helper method.
//
// Returns the monitor pool or nil.
func (o *sts) _getPool() montps.Pool {
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

// MonitorAdd adds a new monitor to the pool.
// The monitor's name must be unique within the pool.
//
// Parameters:
//   - mon: the monitor to add
//
// Returns an error if the pool is not defined or if the monitor cannot be added.
// See github.com/nabbar/golib/monitor/types for Monitor interface details.
func (o *sts) MonitorAdd(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorAdd(mon)
	}
}

// MonitorGet retrieves a monitor by name from the pool.
//
// Parameters:
//   - name: the name of the monitor to retrieve
//
// Returns the monitor if found, nil otherwise or if pool is not defined.
// See github.com/nabbar/golib/monitor/types for Monitor interface details.
func (o *sts) MonitorGet(name string) montps.Monitor {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorGet(name)
	}
}

// MonitorSet updates or adds a monitor in the pool.
// If a monitor with the same name exists, it's replaced; otherwise, it's added.
//
// Parameters:
//   - mon: the monitor to set
//
// Returns an error if the pool is not defined or if the monitor cannot be set.
// See github.com/nabbar/golib/monitor/types for Monitor interface details.
func (o *sts) MonitorSet(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorSet(mon)
	}
}

// MonitorDel removes a monitor from the pool by name.
// Does nothing if the pool is not defined or if the monitor doesn't exist.
//
// Parameters:
//   - name: the name of the monitor to remove
func (o *sts) MonitorDel(name string) {
	if p := o._getPool(); p == nil {
		return
	} else {
		p.MonitorDel(name)
	}
}

// MonitorList returns the names of all monitors in the pool.
//
// Returns a slice of monitor names, or nil if the pool is not defined.
func (o *sts) MonitorList() []string {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorList()
	}
}

// MonitorWalk iterates over monitors in the pool, calling the provided function for each.
// The iteration stops if the function returns false.
//
// Parameters:
//   - fct: function called for each monitor; return false to stop iteration
//   - validName: optional list of monitor names to filter; if empty, all monitors are visited
//
// Does nothing if the pool is not defined.
// See github.com/nabbar/golib/monitor/types for Monitor interface details.
func (o *sts) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	if p := o._getPool(); p != nil {
		p.MonitorWalk(fct, validName...)
	}
}

// RegisterPool registers a function that returns the monitor pool.
// The function is called each time monitors need to be accessed,
// allowing the pool to be dynamically updated.
//
// This method must be called before using any monitor-related methods.
//
// Parameters:
//   - fct: function that returns the current monitor pool
//
// This method is thread-safe.
// See github.com/nabbar/golib/monitor/types for FuncPool and Pool details.
func (o *sts) RegisterPool(fct montps.FuncPool) {
	o.m.Lock()
	defer o.m.Unlock()
	o.p = fct
}
