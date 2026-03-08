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
 * furnished to do so, to subject to the following conditions:
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

// _getPool retrieves the current monitor pool by calling the registered provider function.
// This internal helper method centralizes access to the dynamically provided pool.
// It returns nil if no pool provider is registered or if the provider returns nil.
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

// MonitorAdd adds a new monitor to the registered pool.
// The monitor's name must be unique within the pool.
//
// Returns an error if the pool is not defined or if adding the monitor fails.
func (o *sts) MonitorAdd(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorAdd(mon)
	}
}

// MonitorGet retrieves a monitor by its name from the registered pool.
//
// Returns the monitor if found, otherwise nil. It also returns nil if the pool
// is not defined.
func (o *sts) MonitorGet(name string) montps.Monitor {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorGet(name)
	}
}

// MonitorSet updates an existing monitor or adds a new one to the pool.
// If a monitor with the same name already exists, it is replaced.
//
// Returns an error if the pool is not defined or if setting the monitor fails.
func (o *sts) MonitorSet(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorSet(mon)
	}
}

// MonitorDel removes a monitor from the pool by its name.
// This operation is silent and does nothing if the pool is not defined or if
// the monitor does not exist.
func (o *sts) MonitorDel(name string) {
	if p := o._getPool(); p == nil {
		return
	} else {
		p.MonitorDel(name)
	}
}

// MonitorList returns the names of all monitors currently in the pool.
//
// Returns a slice of monitor names, or nil if the pool is not defined.
func (o *sts) MonitorList() []string {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorList()
	}
}

// MonitorWalk iterates over monitors in the pool, calling the provided function for each one.
// The iteration can be filtered by providing a list of valid names.
// The iteration stops if the callback function returns false.
//
// Parameters:
//   - fct: The function to call for each monitor. Return false to stop the iteration.
//   - validName: An optional list of monitor names to include in the iteration.
//     If empty, all monitors are visited.
func (o *sts) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	if p := o._getPool(); p != nil {
		p.MonitorWalk(fct, validName...)
	}
}

// RegisterPool registers a function that provides the monitor pool.
// This dependency injection pattern allows the status package to access a monitor
// pool that can be managed and updated externally. The provider function is called
// each time monitors need to be accessed.
//
// This method must be called before using any monitor-related methods.
//
// This method is thread-safe.
func (o *sts) RegisterPool(fct montps.FuncPool) {
	o.m.Lock()
	defer o.m.Unlock()
	o.p = fct
}
