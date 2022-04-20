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

package tls

import (
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
)

type componentTls struct {
	ctx libcfg.FuncContext
	get libcfg.FuncComponentGet
	vpr libcfg.FuncComponentViper
	key string

	fsa func(cpt libcfg.Component) liberr.Error
	fsb func(cpt libcfg.Component) liberr.Error
	fra func(cpt libcfg.Component) liberr.Error
	frb func(cpt libcfg.Component) liberr.Error

	m sync.Mutex
	t libtls.TLSConfig
}

func (c *componentTls) _getFct() (func(cpt libcfg.Component) liberr.Error, func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.t != nil {
		return c.frb, c.fra
	} else {
		return c.fsb, c.fsa
	}
}

func (c *componentTls) _runFct(fct func(cpt libcfg.Component) liberr.Error) liberr.Error {
	if fct != nil {
		return fct(c)
	}

	return nil
}

func (c *componentTls) _runCli(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	c.m.Lock()
	defer c.m.Unlock()

	var (
		err liberr.Error
		cfg *libtls.Config
		tls libtls.TLSConfig
	)

	if cfg, err = c._getConfig(getCfg); err != nil {
		return err
	} else if tls, err = cfg.New(); err != nil {
		if c.t != nil {
			return ErrorComponentReload.Error(err)
		}
		return ErrorComponentStart.Error(err)
	} else {
		c.t = tls
	}

	return nil
}

func (c *componentTls) _run(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	fb, fa := c._getFct()

	if err := c._runFct(fb); err != nil {
		return err
	} else if err = c._runCli(getCfg); err != nil {
		return err
	} else if err = c._runFct(fa); err != nil {
		return err
	}

	return nil
}

func (c *componentTls) Type() string {
	return ComponentType
}

func (c *componentTls) Init(key string, ctx libcfg.FuncContext, get libcfg.FuncComponentGet, vpr libcfg.FuncComponentViper) {
	c.m.Lock()
	defer c.m.Unlock()

	c.key = key
	c.ctx = ctx
	c.get = get
	c.vpr = vpr
}

func (c *componentTls) RegisterFuncStart(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.fsb = before
	c.fsa = after
}

func (c *componentTls) RegisterFuncReload(before, after func(cpt libcfg.Component) liberr.Error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.frb = before
	c.fra = after
}

func (c *componentTls) IsStarted() bool {
	c.m.Lock()
	defer c.m.Unlock()

	return c.t != nil
}

func (c *componentTls) IsRunning(atLeast bool) bool {
	return c.IsStarted()
}

func (c *componentTls) Start(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentTls) Reload(getCfg libcfg.FuncComponentConfigGet) liberr.Error {
	return c._run(getCfg)
}

func (c *componentTls) Stop() {
	return
}

func (c *componentTls) Dependencies() []string {
	return make([]string, 0)
}

func (c *componentTls) GetTLS() libtls.TLSConfig {
	c.m.Lock()
	defer c.m.Unlock()

	return c.t
}

func (c *componentTls) SetTLS(tls libtls.TLSConfig) {
	c.m.Lock()
	defer c.m.Unlock()

	c.t = tls
}
