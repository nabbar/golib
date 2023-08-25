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
	liblog "github.com/nabbar/golib/logger"
	montps "github.com/nabbar/golib/monitor/types"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
	spfvpr "github.com/spf13/viper"
)

func (c *configModel) RegisterVersion(vrs libver.Version) {
	c.fct.Store(fctVersion, vrs)
}

func (c *configModel) getVersion() libver.Version {
	if i, l := c.fct.Load(fctVersion); !l {
		return nil
	} else if v, k := i.(libver.Version); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v
	}
}

func (c *configModel) RegisterFuncViper(fct libvpr.FuncViper) {
	c.fct.Store(fctViper, fct)
}

func (c *configModel) getViper() libvpr.Viper {
	if i, l := c.fct.Load(fctViper); !l {
		return nil
	} else if v, k := i.(libvpr.FuncViper); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) getSPFViper() *spfvpr.Viper {
	if v := c.getViper(); v == nil {
		return nil
	} else {
		return v.Viper()
	}
}

func (c *configModel) RegisterFuncStartBefore(fct FuncEvent) {
	c.fct.Store(fctStartBefore, fct)
}

func (c *configModel) runFuncStartBefore() error {
	if i, l := c.fct.Load(fctStartBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterFuncStartAfter(fct FuncEvent) {
	c.fct.Store(fctStartAfter, fct)
}

func (c *configModel) runFuncStartAfter() error {
	if i, l := c.fct.Load(fctStartAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterFuncReloadBefore(fct FuncEvent) {
	c.fct.Store(fctReloadBefore, fct)
}

func (c *configModel) runFuncReloadBefore() error {
	if i, l := c.fct.Load(fctReloadBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterFuncReloadAfter(fct FuncEvent) {
	c.fct.Store(fctReloadAfter, fct)
}

func (c *configModel) runFuncReloadAfter() error {
	if i, l := c.fct.Load(fctReloadAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterFuncStopBefore(fct FuncEvent) {
	c.fct.Store(fctStopBefore, fct)
}

func (c *configModel) runFuncStopBefore() error {
	if i, l := c.fct.Load(fctStopBefore); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterFuncStopAfter(fct FuncEvent) {
	c.fct.Store(fctStopAfter, fct)
}

func (c *configModel) runFuncStopAfter() error {
	if i, l := c.fct.Load(fctStopAfter); !l {
		return nil
	} else if v, k := i.(FuncEvent); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterDefaultLogger(fct liblog.FuncLog) {
	c.fct.Store(fctLoggerDef, fct)
}

func (c *configModel) getDefaultLogger() liblog.Logger {
	if i, l := c.fct.Load(fctLoggerDef); !l {
		return nil
	} else if v, k := i.(liblog.FuncLog); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}

func (c *configModel) RegisterMonitorPool(p montps.FuncPool) {
	c.fct.Store(fctMonitorPool, p)
}

func (c *configModel) getFctMonitorPool() montps.FuncPool {
	if i, l := c.fct.Load(fctMonitorPool); !l {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v
	}
}

func (c *configModel) getMonitorPool() montps.Pool {
	if i, l := c.fct.Load(fctMonitorPool); !l {
		return nil
	} else if v, k := i.(montps.FuncPool); !k {
		return nil
	} else if v == nil {
		return nil
	} else {
		return v()
	}
}
