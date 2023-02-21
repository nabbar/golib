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

func (o *sts) MonitorAdd(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorAdd(mon)
	}
}

func (o *sts) MonitorGet(name string) montps.Monitor {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorGet(name)
	}
}

func (o *sts) MonitorSet(mon montps.Monitor) error {
	if p := o._getPool(); p == nil {
		return fmt.Errorf("monitor pool not defined")
	} else {
		return p.MonitorSet(mon)
	}
}

func (o *sts) MonitorDel(name string) {
	if p := o._getPool(); p == nil {
		return
	} else {
		p.MonitorDel(name)
	}
}

func (o *sts) MonitorList() []string {
	if p := o._getPool(); p == nil {
		return nil
	} else {
		return p.MonitorList()
	}
}

func (o *sts) MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string) {
	if p := o._getPool(); p != nil {
		p.MonitorWalk(fct, validName...)
	}
}

func (o *sts) RegisterPool(fct montps.FuncPool) {
	o.m.Lock()
	defer o.m.Unlock()
	o.p = fct
}
