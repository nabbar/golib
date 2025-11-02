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

package config

import (
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

// RegisterVersion registers the application version information.
// This version is made available to all components during initialization.
func (o *model) RegisterVersion(vrs libver.Version) {
	o.fct.Store(fctVersion, vrs)
}

// RegisterFuncViper registers the Viper configuration provider function.
// Components will use this function to access configuration values.
func (o *model) RegisterFuncViper(fct libvpr.FuncViper) {
	o.fct.Store(fctViper, fct)
}

// RegisterFuncStartBefore registers a hook to execute before starting components.
// This hook runs before any component Start() method is called.
func (o *model) RegisterFuncStartBefore(fct FuncEvent) {
	o.fct.Store(fctStartBefore, fct)
}

// RegisterFuncStartAfter registers a hook to execute after starting all components.
// This hook runs after all component Start() methods have completed successfully.
func (o *model) RegisterFuncStartAfter(fct FuncEvent) {
	o.fct.Store(fctStartAfter, fct)
}

// RegisterFuncReloadBefore registers a hook to execute before reloading components.
// This hook runs before any component Reload() method is called.
func (o *model) RegisterFuncReloadBefore(fct FuncEvent) {
	o.fct.Store(fctReloadBefore, fct)
}

// RegisterFuncReloadAfter registers a hook to execute after reloading all components.
// This hook runs after all component Reload() methods have completed successfully.
func (o *model) RegisterFuncReloadAfter(fct FuncEvent) {
	o.fct.Store(fctReloadAfter, fct)
}

// RegisterFuncStopBefore registers a hook to execute before stopping components.
// This hook runs before any component Stop() method is called.
func (o *model) RegisterFuncStopBefore(fct FuncEvent) {
	o.fct.Store(fctStopBefore, fct)
}

// RegisterFuncStopAfter registers a hook to execute after stopping all components.
// This hook runs after all component Stop() methods have completed.
func (o *model) RegisterFuncStopAfter(fct FuncEvent) {
	o.fct.Store(fctStopAfter, fct)
}

// RegisterMonitorPool registers the monitor pool provider function.
// Components can use this to register health checks and metrics.
func (o *model) RegisterMonitorPool(p montps.FuncPool) {
	o.fct.Store(fctMonitorPool, p)
}

func (o *model) getVersion() libver.Version {
	if i, l := o.fct.Load(fctVersion); !l {
		return nil
	} else if v, k := i.(libver.Version); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v
	}
}

func (o *model) getViper() libvpr.Viper {
	if i, l := o.fct.Load(fctViper); !l {
		return nil
	} else if v, k := i.(libvpr.FuncViper); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncStartBefore() error {
	if i, l := o.fct.Load(fctStartBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncStartAfter() error {
	if i, l := o.fct.Load(fctStartAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncReloadBefore() error {
	if i, l := o.fct.Load(fctReloadBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncReloadAfter() error {
	if i, l := o.fct.Load(fctReloadAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncStopBefore() error {
	if i, l := o.fct.Load(fctStopBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) runFuncStopAfter() error {
	if i, l := o.fct.Load(fctStopAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (o *model) getFctMonitorPool() montps.FuncPool {
	if i, l := o.fct.Load(fctMonitorPool); !l {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v
	}
}

func (o *model) getMonitorPool() montps.Pool {
	if i, l := o.fct.Load(fctMonitorPool); !l {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}
